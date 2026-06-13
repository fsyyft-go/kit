// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"net"
	stdhttp "net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newLocalTLSServer 构造使用自签名证书的本地 TLS 测试服务。
//
// 该辅助函数仅在本地随机端口启动 httptest TLS server，用于验证证书提取逻辑，避免任何外部网络依赖。
//
// 参数：
//   - t: 测试上下文，用于注册服务关闭清理逻辑并标记辅助函数调用栈。
//
// 返回：
//   - *httptest.Server: 已启动的本地 TLS 测试服务。
func newLocalTLSServer(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewTLSServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("X-Method", r.Method)
		w.WriteHeader(stdhttp.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(server.Close)

	return server
}

// newSelfSignedCertificate 构造用于错误链测试的自签名 x509 证书。
//
// 该辅助函数生成内存中的 RSA 私钥和证书模板，便于构造 x509.UnknownAuthorityError 测试数据。
//
// 参数：
//   - t: 测试上下文，用于报告密钥或证书生成失败并标记辅助函数调用栈。
//
// 返回：
//   - *x509.Certificate: 已解析的自签名证书。
func newSelfSignedCertificate(t *testing.T) *x509.Certificate {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "example.test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		DNSNames:     []string{"example.test"},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(derBytes)
	require.NoError(t, err)

	return cert
}

// TestGetCertificates_LocalTLS 验证 GetCertificates 从本地 TLS 服务提取证书链。
//
// 该测试通过 httptest.NewTLSServer 覆盖默认 HEAD 方法、自定义 GET 方法以及 address 覆盖拨号地址的行为，
// 确保证书读取逻辑无需外部网络即可稳定验证。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGetCertificates_LocalTLS(t *testing.T) {
	server := newLocalTLSServer(t)
	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	requestURLWithHost := "https://example.test"

	tests := []struct {
		name        string
		description string
		giveURL     string
		giveMethod  string
		giveAddress string
		assert      func(t *testing.T, certs []*x509.Certificate)
	}{
		{
			name:        "success/default-head-method",
			description: "验证 method 为空时默认使用 HEAD 并从本地 TLS 响应中提取服务端证书。",
			giveURL:     server.URL,
			assert: func(t *testing.T, certs []*x509.Certificate) {
				require.NotEmpty(t, certs)
				assert.True(t, certs[0].NotAfter.After(time.Now()))
			},
		},
		{
			name:        "success/custom-get-method",
			description: "验证显式指定 GET 方法时仍能从 TLS 握手状态中提取证书链。",
			giveURL:     server.URL,
			giveMethod:  stdhttp.MethodGet,
			assert: func(t *testing.T, certs []*x509.Certificate) {
				require.NotEmpty(t, certs)
				assert.True(t, certs[0].NotAfter.After(time.Now()))
			},
		},
		{
			name:        "success/address-override",
			description: "验证 requestURL 主机与拨号地址不一致时使用 address 覆盖连接端点并保留 SNI 主机名。",
			giveURL:     requestURLWithHost,
			giveAddress: serverURL.Host,
			assert: func(t *testing.T, certs []*x509.Certificate) {
				require.NotEmpty(t, certs)
				assert.True(t, certs[0].NotAfter.After(time.Now()))
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			certs, err := GetCertificates(t.Context(), tt.giveURL, tt.giveMethod, tt.giveAddress, time.Second)

			require.NoError(t, err)
			tt.assert(t, certs)
		})
	}
}

// TestGetCertificates_ErrorScenarios 验证 GetCertificates 的错误与无 TLS 响应分支。
//
// 该测试覆盖非法 URL、请求创建失败、拨号失败以及普通 HTTP 响应无 TLS 证书的边界行为，确保错误不会被误吞。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGetCertificates_ErrorScenarios(t *testing.T) {
	plainServer := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
	}))
	t.Cleanup(plainServer.Close)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	closedAddress := listener.Addr().String()
	require.NoError(t, listener.Close())

	tests := []struct {
		name        string
		description string
		giveURL     string
		giveMethod  string
		giveAddress string
		wantErr     bool
		assert      func(t *testing.T, certs []*x509.Certificate, err error)
	}{
		{
			name:        "error/invalid-url-parse",
			description: "验证 requestURL 无法解析时返回 URL 解析错误。",
			giveURL:     "http://[::1",
			wantErr:     true,
			assert: func(t *testing.T, certs []*x509.Certificate, err error) {
				assert.Empty(t, certs)
			},
		},
		{
			name:        "error/invalid-method",
			description: "验证 method 包含非法字符导致请求构造失败时返回错误。",
			giveURL:     "https://example.test",
			giveMethod:  "BAD\nMETHOD",
			wantErr:     true,
			assert: func(t *testing.T, certs []*x509.Certificate, err error) {
				assert.Empty(t, certs)
				assert.Contains(t, err.Error(), "invalid method")
			},
		},
		{
			name:        "error/dial-failure-preserved",
			description: "验证连接失败时保留原始请求错误而不是误判为成功。",
			giveURL:     "https://example.test",
			giveAddress: closedAddress,
			wantErr:     true,
			assert: func(t *testing.T, certs []*x509.Certificate, err error) {
				assert.Empty(t, certs)
			},
		},
		{
			name:        "boundary/plain-http-no-tls",
			description: "验证普通 HTTP 响应没有 TLS 状态时返回空证书链且不发生 panic。",
			giveURL:     plainServer.URL,
			assert: func(t *testing.T, certs []*x509.Certificate, err error) {
				assert.NoError(t, err)
				assert.Empty(t, certs)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			certs, err := GetCertificates(t.Context(), tt.giveURL, tt.giveMethod, tt.giveAddress, 200*time.Millisecond)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			tt.assert(t, certs, err)
		})
	}
}

// TestGetCertificatesExpirestime 验证证书剩余有效期的状态码语义。
//
// 该测试覆盖成功获取证书、获取证书失败返回 -99，以及普通 HTTP 无证书返回 -98 的分支。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGetCertificatesExpirestime(t *testing.T) {
	tlsServer := newLocalTLSServer(t)
	plainServer := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
	}))
	t.Cleanup(plainServer.Close)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	closedAddress := listener.Addr().String()
	require.NoError(t, listener.Close())

	tests := []struct {
		name        string
		description string
		giveURL     string
		giveAddress string
		wantErr     bool
		assert      func(t *testing.T, days int)
	}{
		{
			name:        "success/local-tls-days",
			description: "验证成功获取本地 TLS 证书时返回非错误哨兵值的剩余天数。",
			giveURL:     tlsServer.URL,
			assert: func(t *testing.T, days int) {
				assert.NotEqual(t, -99, days)
				assert.NotEqual(t, -98, days)
			},
		},
		{
			name:        "error/get-certificates-failed",
			description: "验证证书获取失败时返回 -99 并保留错误。",
			giveURL:     "https://example.test",
			giveAddress: closedAddress,
			wantErr:     true,
			assert: func(t *testing.T, days int) {
				assert.Equal(t, -99, days)
			},
		},
		{
			name:        "boundary/no-certificates",
			description: "验证请求成功但未获取到 TLS 证书时返回 -98。",
			giveURL:     plainServer.URL,
			assert: func(t *testing.T, days int) {
				assert.Equal(t, -98, days)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			days, err := GetCertificatesExpirestime(t.Context(), tt.giveURL, "", tt.giveAddress, 200*time.Millisecond)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			tt.assert(t, days)
		})
	}
}

// TestGenerateTLSConfig 验证 TLS 配置中的 SNI 主机名生成规则。
//
// 该测试覆盖带端口主机、不带端口主机以及 IPv6 字面量主机的 serverName 派生行为。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGenerateTLSConfig(t *testing.T) {
	tests := []struct {
		name           string
		description    string
		giveRawURL     string
		wantServerName string
	}{
		{
			name:           "success/host-with-port",
			description:    "验证带端口的主机会在 TLS ServerName 中移除端口。",
			giveRawURL:     "https://example.test:8443/path",
			wantServerName: "example.test",
		},
		{
			name:           "success/host-without-port",
			description:    "验证不带端口的主机会原样作为 TLS ServerName。",
			giveRawURL:     "https://example.test/path",
			wantServerName: "example.test",
		},
		{
			name:           "boundary/ipv6-host-current-behavior",
			description:    "验证当前实现对 IPv6 字面量主机按第一个冒号截断的既有行为。",
			giveRawURL:     "https://[::1]:443/path",
			wantServerName: "[",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			parsedURL, err := url.Parse(tt.giveRawURL)
			require.NoError(t, err)
			config, serverName := generateTLSConfig(parsedURL)

			require.NotNil(t, config)
			assert.Equal(t, tt.wantServerName, serverName)
			assert.Equal(t, tt.wantServerName, config.ServerName)
		})
	}
}

// TestGeneraterClient 验证证书查询客户端的传输层配置。
//
// 该测试确认 generaterClient 会设置超时、自定义 DialContext 与基于 URL 主机的 TLS ServerName。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestGeneraterClient(t *testing.T) {
	parsedURL, err := url.Parse("https://example.test:8443/path")
	require.NoError(t, err)

	client := generaterClient(parsedURL, "127.0.0.1:443", 321*time.Millisecond)

	require.NotNil(t, client)
	assert.Equal(t, 321*time.Millisecond, client.Timeout)
	transport, ok := client.Transport.(*stdhttp.Transport)
	require.True(t, ok)
	require.NotNil(t, transport.TLSClientConfig)
	assert.Equal(t, "example.test", transport.TLSClientConfig.ServerName)
	assert.NotNil(t, transport.DialContext)
}

// TestGeneraterDialContext 验证自定义拨号地址选择逻辑。
//
// 该测试通过本地监听器覆盖默认使用传入 addr 与显式 address 覆盖连接端点两种行为，避免固定端口依赖。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGeneraterDialContext(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = listener.Close()
	})
	accepted := make(chan net.Conn, 2)
	acceptDone := make(chan struct{})
	go func() {
		defer close(acceptDone)
		for i := 0; i < 2; i++ {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			accepted <- conn
		}
	}()
	t.Cleanup(func() {
		for {
			select {
			case conn := <-accepted:
				_ = conn.Close()
			default:
				return
			}
		}
	})

	tests := []struct {
		name        string
		description string
		giveAddress string
		giveAddr    string
	}{
		{
			name:        "success/use-request-addr",
			description: "验证 address 为空时 DialContext 使用请求传入的 addr。",
			giveAddr:    listener.Addr().String(),
		},
		{
			name:        "success/use-override-address",
			description: "验证 address 非空时 DialContext 忽略请求 addr 并连接覆盖地址。",
			giveAddress: listener.Addr().String(),
			giveAddr:    "127.0.0.1:1",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			dialContext := generaterDialContext(tt.giveAddress)
			conn, err := dialContext(t.Context(), "tcp", tt.giveAddr)

			require.NoError(t, err)
			require.NoError(t, conn.Close())
			select {
			case acceptedConn := <-accepted:
				require.NoError(t, acceptedConn.Close())
			case <-time.After(time.Second):
				require.Fail(t, "listener did not accept connection")
			}
		})
	}

	_ = listener.Close()
	select {
	case <-acceptDone:
	case <-time.After(time.Second):
		require.Fail(t, "accept loop did not stop")
	}
}

// TestCheckError 验证从请求错误链中提取未知 CA 证书的行为。
//
// 该测试覆盖可提取 UnknownAuthorityError 证书、普通 URL 错误保留原始错误，以及非 URL 错误保留原始错误。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestCheckError(t *testing.T) {
	cert := newSelfSignedCertificate(t)
	ordinaryErr := errors.New("dial failed")

	tests := []struct {
		name        string
		description string
		giveErr     error
		wantCert    *x509.Certificate
		wantErrIs   error
	}{
		{
			name:        "success/unknown-authority-certificate",
			description: "验证 URL 错误链中存在 UnknownAuthorityError 时提取证书并清除错误。",
			giveErr: &url.Error{
				Op:  "Get",
				URL: "https://example.test",
				Err: x509.UnknownAuthorityError{Cert: cert},
			},
			wantCert: cert,
		},
		{
			name:        "error/url-error-without-certificate",
			description: "验证 URL 错误链不包含可提取证书时保留原始错误。",
			giveErr: &url.Error{
				Op:  "Get",
				URL: "https://example.test",
				Err: ordinaryErr,
			},
			wantErrIs: ordinaryErr,
		},
		{
			name:        "error/non-url-error",
			description: "验证非 URL 错误无法提取证书并原样返回错误。",
			giveErr:     ordinaryErr,
			wantErrIs:   ordinaryErr,
		},
		{
			name:        "error/url-error-nil-inner",
			description: "验证 URL 错误内部错误为空时保留原始 URL 错误。",
			giveErr: &url.Error{
				Op:  "Get",
				URL: "https://example.test",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			gotCert, gotErr := checkError(tt.giveErr)

			if tt.wantErrIs != nil {
				require.ErrorIs(t, gotErr, tt.wantErrIs)
			} else if tt.wantCert != nil {
				require.NoError(t, gotErr)
			} else {
				require.Error(t, gotErr)
			}
			assert.Same(t, tt.wantCert, gotCert)
		})
	}
}

// TestGetCertificates_UntrustedServerFallsBackToCertificate 验证未知 CA 错误可转换为证书结果。
//
// 该测试使用标准库 http.Client 访问本地自签名 TLS 服务产生 UnknownAuthorityError，随后验证 checkError
// 可从错误链中提取服务端证书，覆盖真实 TLS 校验失败场景。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestGetCertificates_UntrustedServerFallsBackToCertificate(t *testing.T) {
	server := newLocalTLSServer(t)

	client := &stdhttp.Client{Timeout: time.Second}
	resp, err := client.Get(server.URL)
	defer closeResponseBody(t, resp)
	require.Error(t, err)

	cert, checkErr := checkError(err)

	require.NoError(t, checkErr)
	require.NotNil(t, cert)
	assert.True(t, cert.NotAfter.After(time.Now()))
}

// TestGetCertificates_WithContextCancellation 验证证书请求遵循调用方 context 取消。
//
// 该测试使用已取消 context 调用本地 TLS 服务，确保 GetCertificates 返回 context.Canceled 并且不返回证书链。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestGetCertificates_WithContextCancellation(t *testing.T) {
	server := newLocalTLSServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	certs, err := GetCertificates(ctx, server.URL, "", "", time.Second)

	require.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, certs)
}

// TestGenerateTLSConfig_CompatibleWithTLSConfig 验证生成的 TLS 配置可用于标准库 TLS 配置结构。
//
// 该测试补充断言 generateTLSConfig 返回的配置没有跳过证书校验，避免把证书查询客户端和通用客户端默认配置混淆。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestGenerateTLSConfig_CompatibleWithTLSConfig(t *testing.T) {
	parsedURL, err := url.Parse("https://example.test")
	require.NoError(t, err)

	config, serverName := generateTLSConfig(parsedURL)

	require.IsType(t, &tls.Config{}, config)
	assert.Equal(t, "example.test", serverName)
	assert.False(t, config.InsecureSkipVerify)
}
