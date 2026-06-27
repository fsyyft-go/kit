// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// defaultGetCertificatesMethod 是获取证书时使用的默认 HTTP 方法。
	defaultGetCertificatesMethod = "HEAD"
)

// GetCertificatesExpirestime 返回请求目标提取到的首张证书剩余有效天数。
//
// 本函数仅基于 [GetCertificates] 返回切片中的首张证书计算有效期，
// 不校验证书链完整性，也不推断中间证书或根证书的剩余有效期。
// 剩余天数通过 time.Until(cert.NotAfter).Hours()/24 向零截断，
// 已过期证书可能返回负数。
//
// 参数：
//   - ctx: 请求上下文，用于控制证书请求的取消和超时。
//   - requestURL: 目标请求地址，用于生成请求并解析 TLS ServerName。
//   - method: 请求使用的 HTTP 方法；为空时使用 HEAD。
//   - address: 可选的实际拨号地址，格式通常为 host:port；非空时只覆盖底层 Dial 目标，不改变 requestURL，也不改变基于 requestURL 生成的 SNI。
//   - timeout: HTTP 客户端的整体超时时间；零值表示不设置 http.Client 超时。
//
// 返回：
//   - int: 首张证书剩余的整天数；返回 -99 表示 GetCertificates 返回错误，返回 -98 表示请求过程未提取到任何证书。
//   - error: GetCertificates 返回的错误；当返回值为 -98 时通常为 nil。
func GetCertificatesExpirestime(ctx context.Context, requestURL, method, address string, timeout time.Duration) (int, error) {
	// exitDays 用于存储证书剩余天数。
	var exitDays int
	// err 用于存储错误信息。
	var err error

	// certs 用于存储获取到的证书链。
	var certs []*x509.Certificate

	// 调用 GetCertificates 获取证书链。
	certs, err = GetCertificates(ctx, requestURL, method, address, timeout)

	// 如果获取证书时发生错误，返回 -99。
	if nil != err {
		exitDays = -99
		// 如果未获取到证书，返回 -98。
	} else if len(certs) == 0 {
		exitDays = -98
		// 正常获取到证书时，计算末级证书的剩余有效天数。
	} else {
		// 只检查末级证书，颁发机构的有效期不需要我们操心。
		cert := certs[0]

		// 计算证书剩余有效期（天数）。
		exitDays = int(time.Until(cert.NotAfter).Hours() / 24)
	}

	// 返回剩余天数和错误信息。
	return exitDays, err
}

// GetCertificates 请求目标地址并提取 TLS 对端证书。
//
// 本函数使用基于 requestURL 主机名的 TLS ServerName 发起请求；主机名提取沿用
// generateTLSConfig 的规则，因此 IPv6 字面量主机存在按第一个冒号截断的既有限制。
// 本函数保留标准库默认的证书校验策略。成功收到响应时优先从响应的 TLS
// 状态读取 PeerCertificates。若 TLS 握手因未知 CA 失败，会尝试从错误链中的
// x509.UnknownAuthorityError 提取其中携带的证书并将该错误视为已处理；
// 这只用于在证书不受本地信任时尽量取回证书，不表示证书链校验成功。
//
// 参数：
//   - ctx: 请求上下文，用于控制请求取消和超时。
//   - requestURL: 目标请求地址，用于创建请求并解析 TLS ServerName。
//   - method: 请求使用的 HTTP 方法；为空时使用 HEAD。
//   - address: 可选的实际拨号地址，格式通常为 host:port；非空时只覆盖底层 Dial 目标，不改变 requestURL，也不改变基于 requestURL 生成的 SNI。
//   - timeout: HTTP 客户端的整体超时时间；零值表示不设置 http.Client 超时。
//
// 返回：
//   - []*x509.Certificate: 提取到的对端证书切片；成功建立 TLS 连接时通常为 PeerCertificates，未知 CA 场景下可能只包含从错误中提取的一张证书；非 TLS 响应或未取到证书时返回空切片。
//   - error: URL 解析、请求创建、网络访问、上下文取消或 TLS 握手中的未处理错误；当仅因未知 CA 失败且已成功提取证书时返回 nil。
func GetCertificates(ctx context.Context, requestURL, method, address string, timeout time.Duration) ([]*x509.Certificate, error) {
	// certs 用于存储获取到的证书链。
	var certs []*x509.Certificate
	// err 用于存储错误信息。
	var err error

	// 解析请求的 URL。
	if u, e := url.Parse(requestURL); nil == e {
		// resp 用于存储 HTTP 响应。
		var resp *http.Response

		// 如果未指定 method，则使用默认方法 HEAD。
		if len(method) == 0 {
			method = defaultGetCertificatesMethod
		}

		// 生成自定义的 HTTP 客户端。
		client := generaterClient(u, address, timeout)

		// 创建空的请求体。
		reader := strings.NewReader(string([]byte{}))
		// 创建带上下文的 HTTP 请求。
		request, requestErr := http.NewRequestWithContext(ctx, method, requestURL, reader)
		if nil != requestErr {
			err = requestErr
			// 发送 HTTP 请求，获取响应。
		} else if resp, err = client.Do(request); nil == err {
			// 确保响应体关闭，防止资源泄漏。
			defer func() { _ = resp.Body.Close() }()
			// 获取 TLS 握手中的对端证书链。
			if nil != resp.TLS {
				certs = resp.TLS.PeerCertificates
			}
			// 如果请求失败，尝试从错误中提取证书。
		} else if cert, errCheckError := checkError(err); nil == errCheckError {
			// 只包含根证书，不包含证书链时，可能会出现 x509.UnknownAuthorityError。
			certs = []*x509.Certificate{cert}
			// 错误已处理，置为 nil。
			err = nil
		} else {
			err = errCheckError
		}
		// URL 解析失败，返回错误。
	} else {
		// 地址不正确。
		err = e
	}

	// 返回证书链和错误信息。
	return certs, err
}

// generaterClient 构建用于提取证书的 HTTP 客户端。
//
// 返回的客户端使用自定义 Transport：TLS 配置由 generateTLSConfig 生成，
// DialContext 由 generaterDialContext 生成。本函数不跳过 TLS 证书校验。
//
// 参数：
//   - url: 已解析的目标 URL，用于生成 TLS 配置；必须非 nil。
//   - address: 可选的实际拨号地址；非空时传给自定义 DialContext 作为连接目标。
//   - timeout: HTTP 客户端的整体超时时间；零值表示不设置 http.Client 超时。
//
// 返回：
//   - *http.Client: 带自定义 TLS 配置和 DialContext 的 HTTP 客户端。
func generaterClient(url *url.URL, address string, timeout time.Duration) *http.Client {
	// 生成 TLS 配置。
	tlsConfig, _ := generateTLSConfig(url)

	// 构建自定义的 HTTP 传输层，设置 TLS 配置和自定义 DialContext。
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		DialContext:     generaterDialContext(address),
	}

	// 构建自定义的 HTTP 客户端，设置传输层和超时时间。
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	// 返回自定义的 HTTP 客户端。
	return client
}

// generateTLSConfig 根据目标 URL 构建 TLS 配置。
//
// 本函数从 url.Host 提取主机名并写入 tls.Config.ServerName，
// 以便 TLS 握手按 requestURL 的主机名发送 SNI。当前实现按第一个冒号
// 截断端口，因此对 IPv6 字面量主机不会得到标准化主机名。
//
// 参数：
//   - url: 已解析的目标 URL，必须非 nil。
//
// 返回：
//   - *tls.Config: 仅设置 ServerName 的 TLS 配置，保持标准库默认的证书校验行为。
//   - string: 从 url.Host 提取出的服务器主机名。
func generateTLSConfig(url *url.URL) (*tls.Config, string) {
	// serverName 用于存储主机名。
	serverName := url.Host

	// 如果地址中包含有端口，按既有实现取第一个冒号前的主机名部分。
	if strings.Contains(serverName, ":") {
		serverName = serverName[:strings.Index(serverName, ":")]
	}

	// 构建 TLS 配置，指定 ServerName 以支持 SNI。
	tlsConfig := &tls.Config{
		ServerName: serverName, // 服务端多证书时，指定主机名可以区分出对应的证书。
	}

	// 返回 TLS 配置和主机名。
	return tlsConfig, serverName
}

// generaterDialContext 构建覆盖拨号目标的 DialContext。
//
// 参数：
//   - address: 可选的实际拨号地址，格式通常为 host:port。
//
// 返回：
//   - func(ctx context.Context, network, addr string) (net.Conn, error): 自定义 DialContext；当 address 非空时忽略 Transport 传入的 addr，直接拨号到 address，否则沿用原始 addr。
func generaterDialContext(address string) func(ctx context.Context, network, addr string) (net.Conn, error) {
	// 返回自定义的 DialContext 函数。
	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		// a 用于存储实际连接的地址。
		a := addr
		// 如果指定了地址，则使用指定的地址。
		if len(address) > 0 {
			a = address
		}
		// 使用 net.Dialer 进行连接。
		return (&net.Dialer{}).DialContext(ctx, network, a)
	}
	// 返回自定义的 DialContext。
	return dialContext
}

// checkError 检查客户端错误中是否携带可提取的证书。
//
// 本函数只处理经 errors.As 可识别的 *url.Error 包装的
// x509.UnknownAuthorityError；其它错误保持原样返回。
//
// 参数：
//   - clientError: 待检查的客户端错误。
//
// 返回：
//   - *x509.Certificate: 从 x509.UnknownAuthorityError 中提取到的证书；无法提取时为 nil。
//   - error: 未识别或未处理的原始错误；成功提取证书时返回 nil。
func checkError(clientError error) (*x509.Certificate, error) {
	// certificate 用于存储提取到的证书。
	var certificate *x509.Certificate
	// err 用于存储错误信息。
	err := clientError

	// urlError 用于存储 URL 相关的错误。
	var urlError *url.Error
	// 判断错误类型是否为 URL 错误，且内部错误不为 nil。
	if errors.As(clientError, &urlError) && nil != urlError.Err {
		// unknownAuthorityError 用于存储未知证书颁发机构的错误。
		var unknownAuthorityError x509.UnknownAuthorityError
		// 判断内部错误类型是否为 UnknownAuthorityError，且错误中携带证书。
		if errors.As(urlError.Err, &unknownAuthorityError) && nil != unknownAuthorityError.Cert {
			// 提取证书。
			certificate = unknownAuthorityError.Cert
			// 错误已被解释为可用的证书信息。
			err = nil
		}
	}

	// 返回提取到的证书和错误信息。
	return certificate, err
}
