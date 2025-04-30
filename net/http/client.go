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

	kiglog "github.com/fsyyft-go/kit/log"
)

type (
	// Client 定义了 HTTP 客户端的接口。
	//
	// 提供常用的 HTTP 请求方法。
	Client interface {
		// Do 执行自定义的 HTTP 请求。
		//
		// 参数：
		//   - ctx context.Context：请求上下文。
		//   - req *http.Request：HTTP 请求对象。
		// 返回值：
		//   - *http.Response：HTTP 响应对象。
		//   - error：请求错误信息。
		Do(ctx context.Context, req *http.Request) (*http.Response, error)
		// Head 发送 HTTP HEAD 请求。
		//
		// 参数：
		//   - ctx context.Context：请求上下文。
		//   - url string：请求地址。
		// 返回值：
		//   - *http.Response：HTTP 响应对象。
		//   - error：请求错误信息。
		Head(ctx context.Context, url string) (*http.Response, error)
		// Get 发送 HTTP GET 请求。
		//
		// 参数：
		//   - ctx context.Context：请求上下文。
		//   - url string：请求地址。
		// 返回值：
		//   - *http.Response：HTTP 响应对象。
		//   - error：请求错误信息。
		Get(ctx context.Context, url string) (*http.Response, error)
		// Post 发送 HTTP POST 请求。
		//
		// 参数：
		//   - ctx context.Context：请求上下文。
		//   - url string：请求地址。
		//   - body io.Reader：请求体。
		// 返回值：
		//   - *http.Response：HTTP 响应对象。
		//   - error：请求错误信息。
		Post(ctx context.Context, url string, body io.Reader) (*http.Response, error)
		// PostForm 发送表单 POST 请求。
		//
		// 参数：
		//   - ctx context.Context：请求上下文。
		//   - url string：请求地址。
		//   - data url.Values：表单数据。
		// 返回值：
		//   - *http.Response：HTTP 响应对象。
		//   - error：请求错误信息。
		PostForm(ctx context.Context, url string, data url.Values) (*http.Response, error)
		// PostJSON 发送 JSON POST 请求。
		//
		// 参数：
		//   - ctx context.Context：请求上下文。
		//   - url string：请求地址。
		//   - data any：JSON 数据。
		// 返回值：
		//   - *http.Response：HTTP 响应对象。
		//   - error：请求错误信息。
		PostJSON(ctx context.Context, url string, data any) (*http.Response, error)
	}

	// client 为 HTTP 客户端的具体实现。
	//
	// 通过组合 http.Client 及自定义配置，实现统一的 HTTP 请求封装。
	client struct {
		name                string                                // 客户端名称。
		timeout             time.Duration                         // 超时时间。
		proxy               func(*http.Request) (*url.URL, error) // 网络代理配置。
		maxConnsPerHost     int                                   // 每主机最大连接数。
		maxIdleConnsPerHost int                                   // 每主机最大空闲连接数。
		maxIdleConns        int                                   // 全局最大空闲连接数。

		transport *http.Transport // 传输层配置。

		hook Hook // 钩子实现。

		logger kiglog.Logger // 日志记录器。

		client *http.Client // 标准库 HTTP 客户端。
	}
)

// NewClient 创建一个新的 HTTP 客户端实例。
//
// 参数：
//   - opts ...Option：可选配置项。
//
// 返回值：
//   - Client：HTTP 客户端实例。
func NewClient(opts ...Option) Client {
	c := &client{
		name:                nameDefault,
		timeout:             timeoutDefault,
		proxy:               proxyDefault,
		maxConnsPerHost:     maxConnsPerHostDefault,
		maxIdleConnsPerHost: maxIdleConnsPerHostDefault,
		maxIdleConns:        maxIdleConnsDefault,
		logger:              kiglog.GetLogger(),
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
// 参数：
//   - ctx context.Context：请求上下文。
//   - req *http.Request：HTTP 请求对象。
//
// 返回值：
//   - *http.Response：HTTP 响应对象。
//   - error：请求错误信息。
func (c *client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if nil != c.hook {
		hc := NewHookContext(ctx, req.Method, req.URL.String(), req)
		_ = c.hook.Before(hc)
		resp, err := c.client.Do(req)
		hc.SetResult(resp, err)
		_ = c.hook.After(hc)
		return resp, err
	} else {
		return c.client.Do(req)
	}
}

// Head 发送 HTTP HEAD 请求。
//
// 参数：
//   - ctx context.Context：请求上下文。
//   - url string：请求地址。
//
// 返回值：
//   - *http.Response：HTTP 响应对象。
//   - error：请求错误信息。
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
//   - ctx context.Context：请求上下文。
//   - url string：请求地址。
//
// 返回值：
//   - *http.Response：HTTP 响应对象。
//   - error：请求错误信息。
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
//   - ctx context.Context：请求上下文。
//   - url string：请求地址。
//   - body io.Reader：请求体。
//
// 返回值：
//   - *http.Response：HTTP 响应对象。
//   - error：请求错误信息。
func (c *client) Post(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	return c.Do(ctx, req)
}

// PostForm 发送表单 POST 请求。
//
// 参数：
//   - ctx context.Context：请求上下文。
//   - url string：请求地址。
//   - data url.Values：表单数据。
//
// 返回值：
//   - *http.Response：HTTP 响应对象。
//   - error：请求错误信息。
func (c *client) PostForm(ctx context.Context, url string, data url.Values) (*http.Response, error) {
	body := strings.NewReader(data.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	return c.Do(ctx, req)
}

// PostJSON 发送 JSON POST 请求。
//
// 参数：
//   - ctx context.Context：请求上下文。
//   - url string：请求地址。
//   - data any：JSON 数据。
//
// 返回值：
//   - *http.Response：HTTP 响应对象。
//   - error：请求错误信息。
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
	// clientDefault 为全局默认 HTTP 客户端实例。
	clientDefault       Client
	clientDefaultLocker sync.Locker
)

// clientDef 获取全局默认 HTTP 客户端实例。
//
// 返回值：
//   - Client：全局默认 HTTP 客户端。
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

// Do 执行自定义的 HTTP 请求（全局默认客户端）。
//
// 参数：
//   - ctx context.Context：请求上下文。
//   - req *http.Request：HTTP 请求对象。
//
// 返回值：
//   - *http.Response：HTTP 响应对象。
//   - error：请求错误信息。
func Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return clientDef().Do(ctx, req)
}

// Head 发送 HTTP HEAD 请求（全局默认客户端）。
//
// 参数：
//   - ctx context.Context：请求上下文。
//   - url string：请求地址。
//
// 返回值：
//   - *http.Response：HTTP 响应对象。
//   - error：请求错误信息。
func Head(ctx context.Context, url string) (*http.Response, error) {
	return clientDef().Head(ctx, url)
}

// Get 发送 HTTP GET 请求（全局默认客户端）。
//
// 参数：
//   - ctx context.Context：请求上下文。
//   - url string：请求地址。
//
// 返回值：
//   - *http.Response：HTTP 响应对象。
//   - error：请求错误信息。
func Get(ctx context.Context, url string) (*http.Response, error) {
	return clientDef().Get(ctx, url)
}

// Post 发送 HTTP POST 请求（全局默认客户端）。
//
// 参数：
//   - ctx context.Context：请求上下文。
//   - url string：请求地址。
//   - body io.Reader：请求体。
//
// 返回值：
//   - *http.Response：HTTP 响应对象。
//   - error：请求错误信息。
func Post(ctx context.Context, url string, body io.Reader) (*http.Response, error) {
	return clientDef().Post(ctx, url, body)
}

// PostForm 发送表单 POST 请求（全局默认客户端）。
//
// 参数：
//   - ctx context.Context：请求上下文。
//   - url string：请求地址。
//   - data url.Values：表单数据。
//
// 返回值：
//   - *http.Response：HTTP 响应对象。
//   - error：请求错误信息。
func PostForm(ctx context.Context, url string, data url.Values) (*http.Response, error) {
	return clientDef().PostForm(ctx, url, data)
}

// PostJSON 发送 JSON POST 请求（全局默认客户端）。
//
// 参数：
//   - ctx context.Context：请求上下文。
//   - url string：请求地址。
//   - data any：JSON 数据。
//
// 返回值：
//   - *http.Response：HTTP 响应对象。
//   - error：请求错误信息。
func PostJSON(ctx context.Context, url string, data any) (*http.Response, error) {
	return clientDef().PostJSON(ctx, url, data)
}
