// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// HTTP 客户端配置选项定义及默认参数。
// 通过 Option 机制灵活配置 HTTP 客户端行为。

package http

import (
	"net/http"
	"net/url"
	"time"

	kitlog "github.com/fsyyft-go/kit/log"
)

type (
	// Option 定义用于配置 client 的函数类型。
	//
	// 通过 Option，可以灵活地设置 client 的各项参数。
	Option func(c *client)
)

// 以下为 HTTP 客户端的默认参数配置。
// 可通过 Option 机制覆盖。
var (
	// nameDefault 为 HTTP 客户端默认名称。
	nameDefault = "kit-defulat-http-client"
	// timeoutDefault 为 HTTP 客户端默认超时时间。
	timeoutDefault = 30 * time.Second
	// traceEnableDefault 为 HTTP 客户端默认开启追踪。
	traceEnableDefault = false
	// proxyDefault 为 HTTP 客户端默认网络代理配置。
	proxyDefault = http.ProxyFromEnvironment
	// maxConnsPerHostDefault 为每个主机的最大连接数默认值。
	maxConnsPerHostDefault = 128
	// maxIdleConnsPerHostDefault 为每个主机的最大空闲连接数默认值。
	maxIdleConnsPerHostDefault = 128
	// maxIdleConnsDefault 为所有主机的最大空闲连接数默认值。
	maxIdleConnsDefault = 1024

	// logSlowDefault 为慢请求阈值默认值。
	logSlowDefault = 10 * time.Second
	// logErrorDefault 为是否记录错误默认值。
	logErrorDefault = true

	// dialTimeoutDefault 为拨号超时时间默认值。
	dialTimeoutDefault = 5 * time.Second
	// dialKeepAliveDefault 为拨号保持活动时间默认值。
	dialKeepAliveDefault = 90 * time.Second
	// forceAttemptHTTP2Default 控制是否强制尝试 HTTP2，默认开启。
	forceAttemptHTTP2Default = true
	// idleConnTimeoutDefault 为空闲连接超时时间默认值。
	idleConnTimeoutDefault = 90 * time.Second
	// tlsInsecureSkipVerifyDefault 控制是否跳过 TLS 证书校验，默认跳过。
	tlsInsecureSkipVerifyDefault = true
	// tlsHandshakeTimeoutDefault 为 TLS 握手超时时间默认值。
	tlsHandshakeTimeoutDefault = 3 * time.Second
	// expectContinueTimeoutDefault 为 Expect-Continue 超时时间默认值。
	expectContinueTimeoutDefault = 1 * time.Second
)

// WithName 设置 HTTP 客户端名称。
//
// 参数：
//   - name string：自定义客户端名称。
//
// 返回值：
//   - Option：用于设置客户端名称的配置项。
func WithName(name string) Option {
	return func(c *client) {
		c.name = name
	}
}

// WithTimeout 设置 HTTP 客户端超时时间。
//
// 参数：
//   - timeout time.Duration：自定义超时时间。
//
// 返回值：
//   - Option：用于设置超时时间的配置项。
func WithTimeout(timeout time.Duration) Option {
	return func(c *client) {
		c.timeout = timeout
	}
}

// WithTraceEnable 控制是否为默认 HookManager 自动注入 traceHook。
//
// 仅在未通过 WithHook 提供自定义 Hook 时生效。
//
// 参数：
//   - enable：是否启用默认 traceHook 注入。
//
// 返回值：
//   - Option：用于设置 traceEnable 的配置项。
func WithTraceEnable(enable bool) Option {
	return func(c *client) {
		c.traceEnable = enable
	}
}

// WithProxy 设置 HTTP 客户端代理。
//
// 参数：
//   - proxy func(*http.Request) (*url.URL, error)：自定义代理函数，用于根据请求返回代理 URL。
//
// 返回值：
//   - Option：用于设置代理的配置项。
func WithProxy(proxy func(*http.Request) (*url.URL, error)) Option {
	return func(c *client) {
		c.proxy = proxy
	}
}

// WithTransport 设置自定义的 http.Transport。
//
// 提供该选项后，NewClient 不再根据 proxy、TLS 和连接池默认参数构造内置 Transport。
//
// 参数：
//   - transport：自定义的 HTTP 传输层配置。
//
// 返回值：
//   - Option：用于设置 transport 的配置项。
func WithTransport(transport *http.Transport) Option {
	return func(c *client) {
		c.transport = transport
	}
}

// WithMaxConnsPerHost 设置每个主机的最大连接数。
//
// 参数：
//   - maxConnsPerHost int：自定义的最大连接数。
//
// 返回值：
//   - Option：用于设置最大连接数的配置项。
func WithMaxConnsPerHost(maxConnsPerHost int) Option {
	return func(c *client) {
		c.maxConnsPerHost = maxConnsPerHost
	}
}

// WithMaxIdleConnsPerHost 设置每个主机的最大空闲连接数。
//
// 参数：
//   - maxIdleConnsPerHost int：自定义的最大空闲连接数。
//
// 返回值：
//   - Option：用于设置最大空闲连接数的配置项。
func WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) Option {
	return func(c *client) {
		c.maxIdleConnsPerHost = maxIdleConnsPerHost
	}
}

// WithMaxIdleConns 设置所有主机的最大空闲连接数。
//
// 参数：
//   - maxIdleConns int：自定义的最大空闲连接数。
//
// 返回值：
//   - Option：用于设置最大空闲连接数的配置项。
func WithMaxIdleConns(maxIdleConns int) Option {
	return func(c *client) {
		c.maxIdleConns = maxIdleConns
	}
}

// WithLogSlow 设置默认慢请求日志 Hook 的阈值。
//
// 仅在未通过 WithHook 提供自定义 Hook 时生效；当阈值小于等于 0 时，不会自动安装慢请求 Hook。
//
// 参数：
//   - logSlow：默认慢请求日志 Hook 的阈值。
//
// 返回值：
//   - Option：用于设置慢请求阈值的配置项。
func WithLogSlow(logSlow time.Duration) Option {
	return func(c *client) {
		c.logSlow = logSlow
	}
}

// WithLogError 控制是否为默认 HookManager 自动注入错误日志 Hook。
//
// 仅在未通过 WithHook 提供自定义 Hook 时生效。
//
// 参数：
//   - logError：是否启用默认错误日志 Hook。
//
// 返回值：
//   - Option：用于设置错误日志 Hook 开关的配置项。
func WithLogError(logError bool) Option {
	return func(c *client) {
		c.logError = logError
	}
}

// WithHook 设置自定义 Hook。
//
// 提供该选项后，NewClient 不会再自动组装 logSlow、traceEnable 和 logError 对应的默认 HookManager。
//
// 参数：
//   - hook：自定义的 Hook 实现。
//
// 返回值：
//   - Option：用于设置 Hook 的配置项。
func WithHook(hook Hook) Option {
	return func(c *client) {
		c.hook = hook
	}
}

// WithLogger 设置 HTTP 客户端的日志记录器。
//
// 参数：
//   - logger kitlog.Logger：自定义的日志记录器实现。
//
// 返回值：
//   - Option：用于设置日志记录器的配置项。
func WithLogger(logger kitlog.Logger) Option {
	return func(c *client) {
		c.logger = logger
	}
}
