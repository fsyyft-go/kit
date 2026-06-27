// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	kitlog "github.com/fsyyft-go/kit/log"
)

type (
	// Client 定义支持请求 Hook、默认配置和常用 HTTP 方法的客户端接口。
	//
	// 返回的 *http.Response 在非 nil 时由调用方负责关闭 Body；请求错误语义与标准库 http.Client 保持一致。
	Client interface {
		// Do 执行自定义的 HTTP 请求。
		//
		// ctx 用于传递给 HookContext；请求本身使用 req 已携带的上下文，Do 不会用 ctx 重写 req.Context()。
		//
		// 参数：
		//   - ctx: 传递给 HookContext 的上下文。
		//   - req: 待发送的 HTTP 请求对象。
		//
		// 返回：
		//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
		//   - error: Hook Before 失败或底层 HTTP 请求失败时返回错误；Hook After 的错误会被忽略。
		Do(ctx context.Context, req *http.Request) (*http.Response, error)
		// Head 发送 HTTP HEAD 请求。
		//
		// 参数：
		//   - ctx: 请求上下文，用于创建 HTTP 请求并控制其生命周期。
		//   - url: 请求地址。
		//
		// 返回：
		//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
		//   - error: 请求创建失败、Hook Before 失败或底层 HTTP 请求失败时返回错误。
		Head(ctx context.Context, url string) (*http.Response, error)
		// Get 发送 HTTP GET 请求。
		//
		// 参数：
		//   - ctx: 请求上下文，用于创建 HTTP 请求并控制其生命周期。
		//   - url: 请求地址。
		//
		// 返回：
		//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
		//   - error: 请求创建失败、Hook Before 失败或底层 HTTP 请求失败时返回错误。
		Get(ctx context.Context, url string) (*http.Response, error)
		// Post 发送 HTTP POST 请求。
		//
		// 参数：
		//   - ctx: 请求上下文，用于创建 HTTP 请求并控制其生命周期。
		//   - url: 请求地址。
		//   - body: 请求体；可为 nil。
		//
		// 返回：
		//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
		//   - error: 请求创建失败、Hook Before 失败或底层 HTTP 请求失败时返回错误。
		Post(ctx context.Context, url string, body io.Reader) (*http.Response, error)
		// PostForm 发送 application/x-www-form-urlencoded 表单 POST 请求。
		//
		// 参数：
		//   - ctx: 请求上下文，用于创建 HTTP 请求并控制其生命周期。
		//   - url: 请求地址。
		//   - data: 表单数据，会通过 url.Values.Encode 编码到请求体。
		//
		// 返回：
		//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
		//   - error: 请求创建失败、Hook Before 失败或底层 HTTP 请求失败时返回错误。
		PostForm(ctx context.Context, url string, data url.Values) (*http.Response, error)
		// PostJSON 发送 application/json POST 请求。
		//
		// 参数：
		//   - ctx: 请求上下文，用于创建 HTTP 请求并控制其生命周期。
		//   - url: 请求地址。
		//   - data: 待编码为 JSON 的请求体数据。
		//
		// 返回：
		//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
		//   - error: JSON 编码失败、请求创建失败、Hook Before 失败或底层 HTTP 请求失败时返回错误。
		PostJSON(ctx context.Context, url string, data any) (*http.Response, error)
	}

	// client 为 HTTP 客户端的具体实现。
	//
	// 通过组合 http.Client 及自定义配置，实现统一的 HTTP 请求封装。
	client struct {
		name                string                                // 客户端名称。
		timeout             time.Duration                         // 超时时间。
		traceEnable         bool                                  // 开启追踪。
		proxy               func(*http.Request) (*url.URL, error) // 网络代理配置。
		maxConnsPerHost     int                                   // 每主机最大连接数。
		maxIdleConnsPerHost int                                   // 每主机最大空闲连接数。
		maxIdleConns        int                                   // 全局最大空闲连接数。

		transport *http.Transport // 传输层配置。

		hook     Hook          // 钩子实现。
		logSlow  time.Duration // 慢请求阈值。
		logError bool          // 是否记录错误。

		logger kitlog.Logger // 日志记录器。

		client *http.Client // 标准库 HTTP 客户端。
	}
)

// NewClient 创建一个新的 HTTP 客户端实例。
//
// 当未显式提供 Transport 时，NewClient 会构造默认 http.Transport，并将
// TLSClientConfig.InsecureSkipVerify 设为 true，也就是默认跳过 TLS 证书校验；
// 如需启用证书校验，调用方必须通过 WithTransport 显式提供自定义 Transport 并调整 TLS 配置。
// 当未显式提供 Hook 时，会按 logSlow、traceEnable 和 logError 选项自动组装默认 HookManager。
//
// 参数：
//   - opts: 用于覆盖默认超时、连接池、Transport、Hook 和日志配置的可选项，按传入顺序应用。
//
// 返回：
//   - Client: 按给定选项构造的 HTTP 客户端实例。
func NewClient(opts ...Option) Client {
	c := &client{
		name:                nameDefault,
		timeout:             timeoutDefault,
		traceEnable:         traceEnableDefault,
		proxy:               proxyDefault,
		maxConnsPerHost:     maxConnsPerHostDefault,
		maxIdleConnsPerHost: maxIdleConnsPerHostDefault,
		maxIdleConns:        maxIdleConnsDefault,
		logSlow:             logSlowDefault,
		logError:            logErrorDefault,
		logger:              kitlog.GetLogger(),
	}

	for _, opt := range opts {
		opt(c)
	}

	if nil == c.transport {
		c.transport = &http.Transport{
			Proxy: c.proxy,
			DialContext: (&net.Dialer{
				Timeout:   dialTimeoutDefault,
				KeepAlive: dialKeepAliveDefault,
			}).DialContext,
			ForceAttemptHTTP2: forceAttemptHTTP2Default,
			IdleConnTimeout:   idleConnTimeoutDefault,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: tlsInsecureSkipVerifyDefault,
			},
			TLSHandshakeTimeout:   tlsHandshakeTimeoutDefault,
			ExpectContinueTimeout: expectContinueTimeoutDefault,
			MaxIdleConns:          c.maxIdleConns,
			MaxConnsPerHost:       c.maxConnsPerHost,
			MaxIdleConnsPerHost:   c.maxIdleConnsPerHost,
		}
	}

	if nil == c.hook {
		hm := NewHookManager()
		if c.logSlow > 0 {
			ls := NewSlowHook(c.logger, c.logSlow)
			hm.AddHook(ls)
		}
		if c.traceEnable {
			th := NewTraceHook(c.logger)
			hm.AddHook(th)
		}
		if c.logError {
			le := NewLogErrorHook(c.logger)
			hm.AddHook(le)
		}
		c.hook = hm
	}

	c.client = &http.Client{
		Timeout:   c.timeout,
		Transport: c.transport,
	}

	return c
}

// Do 执行自定义的 HTTP 请求。
//
// ctx 用于传递给 HookContext；请求本身使用 req 已携带的上下文，Do 不会用 ctx 重写 req.Context()。
//
// 参数：
//   - ctx: 传递给 HookContext 的上下文。
//   - req: 待发送的 HTTP 请求对象。
//
// 返回：
//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
//   - error: Hook Before 失败或底层 HTTP 请求失败时返回错误；Hook After 的错误会被忽略。
func (c *client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if nil != c.hook {
		hc := NewHookContext(ctx, req.Method, req.URL.String(), req)
		if err := c.hook.Before(hc); nil != err {
			// Before 失败时请求尚未发送，直接返回该错误，避免带着不完整的 Hook 状态继续执行。
			return nil, err
		}
		resp, err := c.client.Do(hc.Request())
		hc.SetResult(resp, err)
		_ = c.hook.After(hc) // 请求已经完成，After 只做收尾观察逻辑，不覆盖原始响应和错误。
		return resp, err
	} else {
		return c.client.Do(req)
	}
}

// Head 发送 HTTP HEAD 请求。
//
// 参数：
//   - ctx: 请求上下文，用于创建 HTTP 请求并控制其生命周期。
//   - url: 请求地址。
//
// 返回：
//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
//   - error: 请求创建失败、Hook Before 失败或底层 HTTP 请求失败时返回错误。
func (c *client) Head(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(ctx, req)
}

// Get 发送 HTTP GET 请求。
//
// 参数：
//   - ctx: 请求上下文，用于创建 HTTP 请求并控制其生命周期。
//   - url: 请求地址。
//
// 返回：
//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
//   - error: 请求创建失败、Hook Before 失败或底层 HTTP 请求失败时返回错误。
func (c *client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(ctx, req)
}

// Post 发送 HTTP POST 请求。
//
// 参数：
//   - ctx: 请求上下文，用于创建 HTTP 请求并控制其生命周期。
//   - url: 请求地址。
//   - body: 请求体；可为 nil。
//
// 返回：
//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
//   - error: 请求创建失败、Hook Before 失败或底层 HTTP 请求失败时返回错误。
func (c *client) Post(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	return c.Do(ctx, req)
}

// PostForm 发送 application/x-www-form-urlencoded 表单 POST 请求。
//
// 参数：
//   - ctx: 请求上下文，用于创建 HTTP 请求并控制其生命周期。
//   - url: 请求地址。
//   - data: 表单数据，会通过 url.Values.Encode 编码到请求体。
//
// 返回：
//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
//   - error: 请求创建失败、Hook Before 失败或底层 HTTP 请求失败时返回错误。
func (c *client) PostForm(ctx context.Context, url string, data url.Values) (*http.Response, error) {
	body := strings.NewReader(data.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	return c.Do(ctx, req)
}

// PostJSON 发送 application/json POST 请求。
//
// 参数：
//   - ctx: 请求上下文，用于创建 HTTP 请求并控制其生命周期。
//   - url: 请求地址。
//   - data: 待编码为 JSON 的请求体数据。
//
// 返回：
//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
//   - error: JSON 编码失败、请求创建失败、Hook Before 失败或底层 HTTP 请求失败时返回错误。
func (c *client) PostJSON(ctx context.Context, url string, data any) (*http.Response, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(jsonBytes)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	return c.Do(ctx, req)
}

var (
	// clientDefault 为全局默认 HTTP 客户端实例，首次调用包级辅助函数时懒加载创建。
	clientDefault       Client
	clientDefaultLocker sync.Locker = &sync.Mutex{}
)

// clientDef 返回懒加载的包级默认 HTTP 客户端实例。
//
// 首次调用会使用 NewClient 的默认配置创建实例，后续包级 Do、Get、Post 等辅助函数都会复用该实例。
//
// 参数：无。
//
// 返回：
//   - Client: 包级默认 HTTP 客户端实例。
func clientDef() Client {
	if nil != clientDefault {
		return clientDefault
	}
	clientDefaultLocker.Lock()
	defer clientDefaultLocker.Unlock()

	if nil == clientDefault {
		clientDefault = NewClient()
	}

	return clientDefault
}

// Do 使用全局默认客户端执行自定义的 HTTP 请求。
//
// ctx 用于传递给 HookContext；请求本身使用 req 已携带的上下文，Do 不会用 ctx 重写 req.Context()。
//
// 参数：
//   - ctx: 传递给 HookContext 的上下文。
//   - req: 待发送的 HTTP 请求对象。
//
// 返回：
//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
//   - error: Hook Before 失败或底层 HTTP 请求失败时返回错误；Hook After 的错误会被忽略。
func Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return clientDef().Do(ctx, req)
}

// Head 使用全局默认客户端发送 HTTP HEAD 请求。
//
// 参数：
//   - ctx: 请求上下文，用于创建 HTTP 请求并控制其生命周期。
//   - url: 请求地址。
//
// 返回：
//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
//   - error: 请求创建失败、Hook Before 失败或底层 HTTP 请求失败时返回错误。
func Head(ctx context.Context, url string) (*http.Response, error) {
	return clientDef().Head(ctx, url)
}

// Get 使用全局默认客户端发送 HTTP GET 请求。
//
// 参数：
//   - ctx: 请求上下文，用于创建 HTTP 请求并控制其生命周期。
//   - url: 请求地址。
//
// 返回：
//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
//   - error: 请求创建失败、Hook Before 失败或底层 HTTP 请求失败时返回错误。
func Get(ctx context.Context, url string) (*http.Response, error) {
	return clientDef().Get(ctx, url)
}

// Post 使用全局默认客户端发送 HTTP POST 请求。
//
// 参数：
//   - ctx: 请求上下文，用于创建 HTTP 请求并控制其生命周期。
//   - url: 请求地址。
//   - body: 请求体；可为 nil。
//
// 返回：
//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
//   - error: 请求创建失败、Hook Before 失败或底层 HTTP 请求失败时返回错误。
func Post(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	return clientDef().Post(ctx, url, body)
}

// PostForm 使用全局默认客户端发送 application/x-www-form-urlencoded 表单 POST 请求。
//
// 参数：
//   - ctx: 请求上下文，用于创建 HTTP 请求并控制其生命周期。
//   - url: 请求地址。
//   - data: 表单数据，会通过 url.Values.Encode 编码到请求体。
//
// 返回：
//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
//   - error: 请求创建失败、Hook Before 失败或底层 HTTP 请求失败时返回错误。
func PostForm(ctx context.Context, url string, data url.Values) (*http.Response, error) {
	return clientDef().PostForm(ctx, url, data)
}

// PostJSON 使用全局默认客户端发送 application/json POST 请求。
//
// 参数：
//   - ctx: 请求上下文，用于创建 HTTP 请求并控制其生命周期。
//   - url: 请求地址。
//   - data: 待编码为 JSON 的请求体数据。
//
// 返回：
//   - *http.Response: HTTP 响应对象；非 nil 时调用方负责关闭 Body。
//   - error: JSON 编码失败、请求创建失败、Hook Before 失败或底层 HTTP 请求失败时返回错误。
func PostJSON(ctx context.Context, url string, data any) (*http.Response, error) {
	return clientDef().PostJSON(ctx, url, data)
}
