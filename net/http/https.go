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
	// 定义默认的获取证书方法为 HEAD。
	defaultGetCertificatesMethod = "HEAD"
)

// GetCertificatesExpirestime 获取证书剩余有效期。
//
// 参数：
// - ctx context.Context 上下文对象，用于控制超时与取消。
// - requestURL string 需要进行证书请求的 URL 地址。
// - method string （可选）请求时使用的方式，如果为空值，则默认使用 HEAD。
// - address string （可选）指定的服务器 Endpoint，包含 IP 和端口，例如：127.0.0.1:443。
// - timeout time.Duration 超时时间。
//
// 返回值：
// - int 证书剩余有效天数，-99 表示获取证书失败，-98 表示未获取到证书。
// - error 错误信息。
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

// GetCertificates 获取证书。
//
// 参数：
// - ctx context.Context 上下文对象，用于控制超时与取消。
// - requestURL string 需要进行证书请求的 URL 地址。
// - method string （可选）请求时使用的方式，如果为空值，则默认使用 HEAD。
// - address string （可选）指定的服务器 Endpoint，包含 IP 和端口，例如：127.0.0.1:443。
// - timeout time.Duration 超时时间。
//
// 返回值：
// - []*x509.Certificate 获取到的证书链。
// - error 错误信息。
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
		request, _ := http.NewRequestWithContext(ctx, method, requestURL, reader)
		// 发送 HTTP 请求，获取响应。
		if resp, err = client.Do(request); nil == err {
			// 确保响应体关闭，防止资源泄漏。
			defer func() { _ = resp.Body.Close() }()
			// 获取 TLS 握手中的对端证书链。
			certs = resp.TLS.PeerCertificates
			// 如果请求失败，尝试从错误中提取证书。
		} else if cert, errCheckError := checkError(err); nil == errCheckError {
			// 只包含根证书，不包含证书链时，可能会出现 x509.UnknownAuthorityError。
			certs = []*x509.Certificate{cert}
			// 错误已处理，置为 nil。
			err = nil
		}
		// URL 解析失败，返回错误。
	} else {
		// 地址不正确。
		err = e
	}

	// 返回证书链和错误信息。
	return certs, err
}

// generaterClient 构建自定义的 http.Client。
//
// 参数：
// - url *url.URL 需要进行证书请求的 URL 地址的 url.URL 表示形式。
// - address string （可选）指定的服务器 Endpoint，包含 IP 和端口，例如：127.0.0.1:443。
// - timeout time.Duration 超时时间。
//
// 返回值：
// - *http.Client 构建好的自定义 HTTP 客户端。
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

// generateTLSConfig 构建 tls.Config。
// 应用于 TLS 的 SNI 扩展协议。
//
// 参数：
// - url *url.URL 需要进行证书请求的 URL 地址的 url.URL 表示形式。
//
// 返回值：
// - *tls.Config 构建好的 TLS 配置。
// - string 服务器主机名。
func generateTLSConfig(url *url.URL) (*tls.Config, string) {
	// serverName 用于存储主机名。
	serverName := url.Host

	// 如果地址中包含有端口，需要把端口移除。
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

// generaterDialContext 构建 http.Transport 的 DialContext。
//
// 参数：
// - address string （可选）指定的服务器 Endpoint，包含 IP 和端口，例如：127.0.0.1:443。
//
// 返回值：
// - func(ctx context.Context, network, addr string) (net.Conn, error) 自定义的 DialContext 函数。
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

// checkError 异常检查。
// 尝试从错误中提取 x509 证书。
//
// 参数：
// - clientError error 需要检查的错误。
//
// 返回值：
// - *x509.Certificate 提取到的证书，若无则为 nil。
// - error 错误信息。
func checkError(clientError error) (*x509.Certificate, error) {
	// certificate 用于存储提取到的证书。
	var certificate *x509.Certificate
	// err 用于存储错误信息。
	var err error

	// urlError 用于存储 URL 相关的错误。
	var urlError *url.Error
	// 判断错误类型是否为 URL 错误，且内部错误不为 nil。
	if errors.As(clientError, &urlError) && nil != urlError.Err {
		// unknownAuthorityError 用于存储未知证书颁发机构的错误。
		var unknownAuthorityError x509.UnknownAuthorityError
		// 判断内部错误类型是否为 UnknownAuthorityError，且证书不为 nil。
		if errors.As(urlError.Err, &unknownAuthorityError) && nil != unknownAuthorityError.Cert {
			// 提取证书。
			certificate = unknownAuthorityError.Cert
		}
	}

	// 返回提取到的证书和错误信息。
	return certificate, err
}
