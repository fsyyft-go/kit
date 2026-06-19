// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"net/http"
	"net/url"
	"time"

	kitlog "github.com/fsyyft-go/kit/log"
)

type (
	// Option 定义修改 HTTP client 配置的函数。
	//
	// Option 由 [NewClient] 按传入顺序执行，后传入的配置可覆盖先前写入的同一字段。
	//
	// 参数：
	//   - c: 待修改的 client 配置实例，由 NewClient 创建并传入。
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
// 该名称当前用于配置保存，默认 Hook 日志不会自动输出该字段。
//
// 参数：
//   - name: 自定义客户端名称，空字符串会按原值写入。
//
// 返回：
//   - Option: 应用于 [NewClient] 的客户端名称配置项。
func WithName(name string) Option {
	return func(c *client) {
		c.name = name
	}
}

// WithTimeout 设置 HTTP 客户端总超时时间。
//
// timeout 会写入底层 http.Client.Timeout；非正值表示不设置整体超时，正值会限制包含连接、重定向和读取响应体在内的完整请求周期。
//
// 参数：
//   - timeout: 自定义客户端总超时时间。
//
// 返回：
//   - Option: 应用于 [NewClient] 的超时时间配置项。
func WithTimeout(timeout time.Duration) Option {
	return func(c *client) {
		c.timeout = timeout
	}
}

// WithTraceEnable 控制是否为默认 HookManager 自动注入 traceHook。
//
// 仅在未通过 [WithHook] 提供自定义 Hook 时生效。
//
// 参数：
//   - enable: true 表示启用默认 traceHook 注入，false 表示不注入。
//
// 返回：
//   - Option: 应用于 [NewClient] 的 traceHook 开关配置项。
func WithTraceEnable(enable bool) Option {
	return func(c *client) {
		c.traceEnable = enable
	}
}

// WithProxy 设置 HTTP 客户端代理函数。
//
// 该选项只在使用 NewClient 内置 Transport 时生效；通过 [WithTransport] 提供自定义 Transport 后，代理行为由自定义 Transport 决定。
//
// 参数：
//   - proxy: 自定义代理函数，用于根据请求返回代理 URL；可为 nil，含义与 http.Transport.Proxy 一致。
//
// 返回：
//   - Option: 应用于 [NewClient] 的代理配置项。
func WithProxy(proxy func(*http.Request) (*url.URL, error)) Option {
	return func(c *client) {
		c.proxy = proxy
	}
}

// WithTransport 设置自定义的 http.Transport。
//
// 当该选项最终写入非 nil transport 时，NewClient 不再根据 proxy、TLS 和连接池默认参数构造内置 Transport；传入 nil 时继续使用默认构造逻辑。Transport 的生命周期由调用方负责。
//
// 参数：
//   - transport: 自定义的 HTTP 传输层配置；为 nil 时 NewClient 会继续构造内置 Transport。
//
// 返回：
//   - Option: 应用于 [NewClient] 的 Transport 配置项。
func WithTransport(transport *http.Transport) Option {
	return func(c *client) {
		c.transport = transport
	}
}

// WithMaxConnsPerHost 设置每个主机的最大连接数。
//
// 该选项只在使用 NewClient 内置 Transport 时生效，取值语义与 http.Transport.MaxConnsPerHost 保持一致。
//
// 参数：
//   - maxConnsPerHost: 自定义的每主机最大连接数。
//
// 返回：
//   - Option: 应用于 [NewClient] 的连接池配置项。
func WithMaxConnsPerHost(maxConnsPerHost int) Option {
	return func(c *client) {
		c.maxConnsPerHost = maxConnsPerHost
	}
}

// WithMaxIdleConnsPerHost 设置每个主机的最大空闲连接数。
//
// 该选项只在使用 NewClient 内置 Transport 时生效，取值语义与 http.Transport.MaxIdleConnsPerHost 保持一致。
//
// 参数：
//   - maxIdleConnsPerHost: 自定义的每主机最大空闲连接数。
//
// 返回：
//   - Option: 应用于 [NewClient] 的连接池配置项。
func WithMaxIdleConnsPerHost(maxIdleConnsPerHost int) Option {
	return func(c *client) {
		c.maxIdleConnsPerHost = maxIdleConnsPerHost
	}
}

// WithMaxIdleConns 设置所有主机的最大空闲连接数。
//
// 该选项只在使用 NewClient 内置 Transport 时生效，取值语义与 http.Transport.MaxIdleConns 保持一致。
//
// 参数：
//   - maxIdleConns: 自定义的全局最大空闲连接数。
//
// 返回：
//   - Option: 应用于 [NewClient] 的连接池配置项。
func WithMaxIdleConns(maxIdleConns int) Option {
	return func(c *client) {
		c.maxIdleConns = maxIdleConns
	}
}

// WithLogSlow 设置默认慢请求日志 Hook 的阈值。
//
// 仅在未通过 [WithHook] 提供自定义 Hook 时生效；当阈值小于等于 0 时，不会自动安装慢请求 Hook。
//
// 参数：
//   - logSlow: 默认慢请求日志 Hook 的阈值。
//
// 返回：
//   - Option: 应用于 [NewClient] 的慢请求日志配置项。
func WithLogSlow(logSlow time.Duration) Option {
	return func(c *client) {
		c.logSlow = logSlow
	}
}

// WithLogError 控制是否为默认 HookManager 自动注入错误日志 Hook。
//
// 仅在未通过 [WithHook] 提供自定义 Hook 时生效。
//
// 参数：
//   - logError: true 表示启用默认错误日志 Hook，false 表示不注入。
//
// 返回：
//   - Option: 应用于 [NewClient] 的错误日志 Hook 开关配置项。
func WithLogError(logError bool) Option {
	return func(c *client) {
		c.logError = logError
	}
}

// WithHook 设置自定义 Hook。
//
// 当该选项最终写入非 nil hook 时，NewClient 不再自动组装 logSlow、traceEnable 和 logError 对应的默认 HookManager；传入 nil 时继续按这些选项组装默认 HookManager。
//
// 参数：
//   - hook: 自定义的 Hook 实现；为 nil 时 NewClient 会继续组装默认 HookManager。
//
// 返回：
//   - Option: 应用于 [NewClient] 的 Hook 配置项。
func WithHook(hook Hook) Option {
	return func(c *client) {
		c.hook = hook
	}
}

// WithLogger 设置 HTTP 客户端的日志记录器。
//
// 该日志记录器只会传递给自动组装的默认慢请求、trace 和错误日志 Hook；如果通过 WithHook 提供自定义 Hook，NewClient 不会把 logger 自动注入到自定义 Hook。调用方应提供可用的 Logger。
//
// 参数：
//   - logger: 自定义的日志记录器实现。
//
// 返回：
//   - Option: 应用于 [NewClient] 的日志配置项。
func WithLogger(logger kitlog.Logger) Option {
	return func(c *client) {
		c.logger = logger
	}
}
