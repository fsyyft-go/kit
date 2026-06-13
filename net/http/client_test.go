// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	stdhttp "net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requestSnapshot 记录本地回显服务观察到的请求关键信息。
//
// 该结构只保存测试断言需要的字段，不表示完整 HTTP 请求。
type requestSnapshot struct {
	Method      string `json:"method"`       // Method 记录服务端观察到的 HTTP 方法。
	Path        string `json:"path"`         // Path 记录服务端观察到的 URL path，不包含 query。
	ContentType string `json:"content_type"` // ContentType 记录 Content-Type 请求头。
	Body        string `json:"body"`         // Body 记录服务端读取到的完整请求体。
}

// newRequestEchoServer 构造用于验证客户端请求语义的本地 HTTP 服务。
//
// 该辅助函数返回一个仅监听本地随机端口的 httptest.Server。默认、/notfound 与 /error 分支
// 会把收到的请求方法、路径、Content-Type 与请求体编码为 JSON；/head 和 /timeout 分支用于覆盖
// 响应头与超时等特殊行为。
//
// 参数：
//   - t: 测试上下文，用于注册服务关闭清理逻辑并标记辅助函数调用栈。
//
// 返回：
//   - *httptest.Server: 已启动的本地 HTTP 测试服务。
func newRequestEchoServer(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if r.URL.Path == "/timeout" {
			select {
			case <-r.Context().Done():
				return
			case <-time.After(500 * time.Millisecond):
				w.WriteHeader(stdhttp.StatusGatewayTimeout)
				return
			}
		}

		body, _ := io.ReadAll(r.Body)
		snapshot := requestSnapshot{
			Method:      r.Method,
			Path:        r.URL.Path,
			ContentType: r.Header.Get("Content-Type"),
			Body:        string(body),
		}

		switch r.URL.Path {
		case "/head":
			w.Header().Set("X-Test", "head-ok")
			w.WriteHeader(stdhttp.StatusNoContent)
		case "/notfound":
			w.WriteHeader(stdhttp.StatusNotFound)
			_ = json.NewEncoder(w).Encode(snapshot)
		case "/error":
			w.WriteHeader(stdhttp.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(snapshot)
		default:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(snapshot)
		}
	}))
	t.Cleanup(server.Close)

	return server
}

// readResponseBody 读取并关闭 HTTP 响应体。
//
// 该辅助函数集中处理响应体读取和关闭断言，确保测试不会泄漏响应资源。
//
// 参数：
//   - t: 测试上下文，用于报告读取或关闭响应体失败。
//   - resp: 需要读取并关闭的 HTTP 响应。
//
// 返回：
//   - []byte: 响应体的完整字节内容。
func readResponseBody(t *testing.T, resp *stdhttp.Response) []byte {
	t.Helper()

	require.NotNil(t, resp)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	return body
}

// closeResponseBody 关闭非空 HTTP 响应体。
//
// 该辅助函数用于无需读取响应体内容的场景，保证每个成功返回的响应都被显式关闭。
//
// 参数：
//   - t: 测试上下文，用于报告关闭响应体失败。
//   - resp: 可能包含响应体的 HTTP 响应。
func closeResponseBody(t *testing.T, resp *stdhttp.Response) {
	t.Helper()

	if resp == nil || resp.Body == nil {
		return
	}
	require.NoError(t, resp.Body.Close())
}

// decodeRequestSnapshot 从响应体中解析请求快照。
//
// 该辅助函数通过 readResponseBody 读取并关闭响应体，用于验证本地 echo server 观察到的
// 请求方法、路径、Content-Type 头和请求体是否符合客户端契约。
//
// 参数：
//   - t: 测试上下文，用于报告响应读取或 JSON 解析失败。
//   - resp: 本地 echo server 返回的 HTTP 响应。
//
// 返回：
//   - requestSnapshot: 由服务端观察并回传的请求快照。
func decodeRequestSnapshot(t *testing.T, resp *stdhttp.Response) requestSnapshot {
	t.Helper()

	body := readResponseBody(t, resp)
	var snapshot requestSnapshot
	require.NoError(t, json.Unmarshal(body, &snapshot))

	return snapshot
}

// TestNewClient_Configuration 验证 NewClient 按默认值与 Option 组合构造客户端。
//
// 该测试通过表驱动用例覆盖默认传输层、连接池参数、代理函数、自定义 Hook、自定义 Transport
// 以及默认 HookManager 的装配语义，确保客户端初始化契约稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewClient_Configuration(t *testing.T) {
	proxyURL, err := url.Parse("http://proxy.local:8080")
	require.NoError(t, err)
	proxyFn := func(*stdhttp.Request) (*url.URL, error) {
		return proxyURL, nil
	}
	customTransport := &stdhttp.Transport{MaxConnsPerHost: 7}
	customHook := &recordingHook{}

	tests := []struct {
		name        string
		description string
		giveOptions []Option
		assert      func(t *testing.T, c *client)
	}{
		{
			name:        "success/defaults",
			description: "验证 NewClient 在无 Option 时使用包级默认配置并创建默认 HookManager。",
			assert: func(t *testing.T, c *client) {
				assert.Equal(t, nameDefault, c.name)
				assert.Equal(t, timeoutDefault, c.timeout)
				assert.Equal(t, timeoutDefault, c.client.Timeout)
				require.NotNil(t, c.transport)
				assert.Equal(t, maxConnsPerHostDefault, c.transport.MaxConnsPerHost)
				assert.Equal(t, maxIdleConnsPerHostDefault, c.transport.MaxIdleConnsPerHost)
				assert.Equal(t, maxIdleConnsDefault, c.transport.MaxIdleConns)
				require.NotNil(t, c.transport.TLSClientConfig)
				assert.Equal(t, tlsInsecureSkipVerifyDefault, c.transport.TLSClientConfig.InsecureSkipVerify)
				require.IsType(t, &HookManager{}, c.hook)
				manager := c.hook.(*HookManager)
				assert.Len(t, manager.hooks, 2)
			},
		},
		{
			name:        "success/options-override-fields",
			description: "验证 Option 能覆盖名称、超时、代理、连接池与日志相关配置。",
			giveOptions: []Option{
				WithName("custom-client"),
				WithTimeout(123 * time.Millisecond),
				WithProxy(proxyFn),
				WithMaxConnsPerHost(11),
				WithMaxIdleConnsPerHost(12),
				WithMaxIdleConns(13),
				WithLogSlow(0),
				WithLogError(false),
				WithTraceEnable(false),
				WithLogger(nil),
			},
			assert: func(t *testing.T, c *client) {
				assert.Equal(t, "custom-client", c.name)
				assert.Equal(t, 123*time.Millisecond, c.timeout)
				assert.Equal(t, 123*time.Millisecond, c.client.Timeout)
				assert.Equal(t, 11, c.transport.MaxConnsPerHost)
				assert.Equal(t, 12, c.transport.MaxIdleConnsPerHost)
				assert.Equal(t, 13, c.transport.MaxIdleConns)
				gotProxyURL, err := c.transport.Proxy(httptest.NewRequest(stdhttp.MethodGet, "http://example.test", nil))
				require.NoError(t, err)
				assert.Equal(t, proxyURL, gotProxyURL)
				assert.Nil(t, c.logger)
				require.IsType(t, &HookManager{}, c.hook)
				manager := c.hook.(*HookManager)
				assert.Empty(t, manager.hooks)
			},
		},
		{
			name:        "success/trace-hook-enabled",
			description: "验证开启 trace 且关闭慢日志与错误日志时仅装配 trace hook。",
			giveOptions: []Option{
				WithTraceEnable(true),
				WithLogSlow(0),
				WithLogError(false),
			},
			assert: func(t *testing.T, c *client) {
				require.IsType(t, &HookManager{}, c.hook)
				manager := c.hook.(*HookManager)
				require.Len(t, manager.hooks, 1)
				assert.IsType(t, &traceHook{}, manager.hooks[0])
			},
		},
		{
			name:        "success/custom-transport-and-hook",
			description: "验证显式提供 Transport 与 Hook 时 NewClient 保留调用方传入的实例。",
			giveOptions: []Option{
				WithTransport(customTransport),
				WithHook(customHook),
				WithTraceEnable(true),
			},
			assert: func(t *testing.T, c *client) {
				assert.Same(t, customTransport, c.transport)
				transport, ok := c.client.Transport.(*stdhttp.Transport)
				require.True(t, ok)
				assert.Same(t, customTransport, transport)
				assert.Same(t, customHook, c.hook)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := NewClient(tt.giveOptions...)

			require.IsType(t, &client{}, got)
			actual := got.(*client)
			require.NotNil(t, actual.client)
			tt.assert(t, actual)
		})
	}
}

// TestClient_RequestMethods 验证客户端便捷请求方法的 HTTP 传输语义。
//
// 该测试通过本地 httptest.Server 覆盖 GET、HEAD、POST、表单 POST、JSON POST、HTTP 错误状态码
// 与客户端超时，确保方法、请求体、Content-Type、状态码和错误处理行为稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestClient_RequestMethods(t *testing.T) {
	server := newRequestEchoServer(t)
	client := NewClient(
		WithTimeout(50*time.Millisecond),
		WithLogSlow(0),
		WithLogError(false),
		WithTraceEnable(false),
	)

	tests := []struct {
		name        string
		description string
		givePath    string
		giveCall    func(ctx context.Context, client Client, targetURL string) (*stdhttp.Response, error)
		assert      func(t *testing.T, resp *stdhttp.Response, err error)
	}{
		{
			name:        "success/get",
			description: "验证 Get 使用 GET 方法发送请求并返回服务端回显的请求快照。",
			givePath:    "/get",
			giveCall: func(ctx context.Context, client Client, targetURL string) (*stdhttp.Response, error) {
				return client.Get(ctx, targetURL)
			},
			assert: func(t *testing.T, resp *stdhttp.Response, err error) {
				require.NoError(t, err)
				assert.Equal(t, stdhttp.StatusOK, resp.StatusCode)
				snapshot := decodeRequestSnapshot(t, resp)
				assert.Equal(t, stdhttp.MethodGet, snapshot.Method)
				assert.Equal(t, "/get", snapshot.Path)
				assert.Empty(t, snapshot.Body)
			},
		},
		{
			name:        "success/head",
			description: "验证 Head 使用 HEAD 方法并保留响应状态码与响应头。",
			givePath:    "/head",
			giveCall: func(ctx context.Context, client Client, targetURL string) (*stdhttp.Response, error) {
				return client.Head(ctx, targetURL)
			},
			assert: func(t *testing.T, resp *stdhttp.Response, err error) {
				require.NoError(t, err)
				defer closeResponseBody(t, resp)
				assert.Equal(t, stdhttp.StatusNoContent, resp.StatusCode)
				assert.Equal(t, "head-ok", resp.Header.Get("X-Test"))
			},
		},
		{
			name:        "success/post-body",
			description: "验证 Post 使用 POST 方法并完整传输调用方提供的请求体。",
			givePath:    "/post",
			giveCall: func(ctx context.Context, client Client, targetURL string) (*stdhttp.Response, error) {
				return client.Post(ctx, targetURL, strings.NewReader("plain-body"))
			},
			assert: func(t *testing.T, resp *stdhttp.Response, err error) {
				require.NoError(t, err)
				assert.Equal(t, stdhttp.StatusOK, resp.StatusCode)
				snapshot := decodeRequestSnapshot(t, resp)
				assert.Equal(t, stdhttp.MethodPost, snapshot.Method)
				assert.Equal(t, "plain-body", snapshot.Body)
				assert.Empty(t, snapshot.ContentType)
			},
		},
		{
			name:        "success/post-form",
			description: "验证 PostForm 使用标准表单编码并设置 form-urlencoded Content-Type。",
			givePath:    "/form",
			giveCall: func(ctx context.Context, client Client, targetURL string) (*stdhttp.Response, error) {
				return client.PostForm(ctx, targetURL, url.Values{"a": {"1"}, "b": {"2"}})
			},
			assert: func(t *testing.T, resp *stdhttp.Response, err error) {
				require.NoError(t, err)
				assert.Equal(t, stdhttp.StatusOK, resp.StatusCode)
				snapshot := decodeRequestSnapshot(t, resp)
				assert.Equal(t, stdhttp.MethodPost, snapshot.Method)
				assert.Equal(t, "a=1&b=2", snapshot.Body)
				assert.Equal(t, "application/x-www-form-urlencoded; charset=utf-8", snapshot.ContentType)
			},
		},
		{
			name:        "success/post-json",
			description: "验证 PostJSON 序列化 JSON 请求体并设置 application/json Content-Type。",
			givePath:    "/json",
			giveCall: func(ctx context.Context, client Client, targetURL string) (*stdhttp.Response, error) {
				return client.PostJSON(ctx, targetURL, map[string]any{"x": 1, "y": "z"})
			},
			assert: func(t *testing.T, resp *stdhttp.Response, err error) {
				require.NoError(t, err)
				assert.Equal(t, stdhttp.StatusOK, resp.StatusCode)
				snapshot := decodeRequestSnapshot(t, resp)
				assert.Equal(t, stdhttp.MethodPost, snapshot.Method)
				assert.JSONEq(t, `{"x":1,"y":"z"}`, snapshot.Body)
				assert.Equal(t, "application/json; charset=utf-8", snapshot.ContentType)
			},
		},
		{
			name:        "success/not-found-status",
			description: "验证 HTTP 404 状态码按普通响应返回而不会被包装为请求错误。",
			givePath:    "/notfound",
			giveCall: func(ctx context.Context, client Client, targetURL string) (*stdhttp.Response, error) {
				return client.Get(ctx, targetURL)
			},
			assert: func(t *testing.T, resp *stdhttp.Response, err error) {
				require.NoError(t, err)
				assert.Equal(t, stdhttp.StatusNotFound, resp.StatusCode)
				snapshot := decodeRequestSnapshot(t, resp)
				assert.Equal(t, "/notfound", snapshot.Path)
			},
		},
		{
			name:        "success/server-error-status",
			description: "验证 HTTP 500 状态码按普通响应返回而不会被包装为请求错误。",
			givePath:    "/error",
			giveCall: func(ctx context.Context, client Client, targetURL string) (*stdhttp.Response, error) {
				return client.Get(ctx, targetURL)
			},
			assert: func(t *testing.T, resp *stdhttp.Response, err error) {
				require.NoError(t, err)
				assert.Equal(t, stdhttp.StatusInternalServerError, resp.StatusCode)
				snapshot := decodeRequestSnapshot(t, resp)
				assert.Equal(t, "/error", snapshot.Path)
			},
		},
		{
			name:        "error/client-timeout",
			description: "验证客户端超时时会返回请求错误并且不产生可用响应。",
			givePath:    "/timeout",
			giveCall: func(ctx context.Context, client Client, targetURL string) (*stdhttp.Response, error) {
				return client.Get(ctx, targetURL)
			},
			assert: func(t *testing.T, resp *stdhttp.Response, err error) {
				require.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), "Client.Timeout")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			resp, err := tt.giveCall(t.Context(), client, server.URL+tt.givePath)

			tt.assert(t, resp, err)
		})
	}
}

// TestClient_RequestCreationErrors 验证客户端在请求构造阶段返回错误。
//
// 该测试覆盖各便捷方法的非法 URL 分支，以及 PostJSON 在 JSON 序列化失败时不会发起网络请求的错误语义。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestClient_RequestCreationErrors(t *testing.T) {
	client := NewClient(
		WithLogSlow(0),
		WithLogError(false),
		WithTraceEnable(false),
	)

	tests := []struct {
		name        string
		description string
		giveCall    func(ctx context.Context, client Client) (*stdhttp.Response, error)
	}{
		{
			name:        "error/head-invalid-url",
			description: "验证 Head 在 URL 无法解析时直接返回构造错误。",
			giveCall: func(ctx context.Context, client Client) (*stdhttp.Response, error) {
				return client.Head(ctx, "http://[::1")
			},
		},
		{
			name:        "error/get-invalid-url",
			description: "验证 Get 在 URL 无法解析时直接返回构造错误。",
			giveCall: func(ctx context.Context, client Client) (*stdhttp.Response, error) {
				return client.Get(ctx, "http://[::1")
			},
		},
		{
			name:        "error/post-invalid-url",
			description: "验证 Post 在 URL 无法解析时直接返回构造错误。",
			giveCall: func(ctx context.Context, client Client) (*stdhttp.Response, error) {
				return client.Post(ctx, "http://[::1", strings.NewReader("body"))
			},
		},
		{
			name:        "error/post-form-invalid-url",
			description: "验证 PostForm 在 URL 无法解析时直接返回构造错误。",
			giveCall: func(ctx context.Context, client Client) (*stdhttp.Response, error) {
				return client.PostForm(ctx, "http://[::1", url.Values{"a": {"1"}})
			},
		},
		{
			name:        "error/post-json-invalid-url",
			description: "验证 PostJSON 在序列化成功但 URL 无法解析时返回请求构造错误。",
			giveCall: func(ctx context.Context, client Client) (*stdhttp.Response, error) {
				return client.PostJSON(ctx, "http://[::1", map[string]any{"a": 1})
			},
		},
		{
			name:        "error/post-json-marshal",
			description: "验证 PostJSON 遇到不可序列化数据时返回 JSON 序列化错误。",
			giveCall: func(ctx context.Context, client Client) (*stdhttp.Response, error) {
				return client.PostJSON(ctx, "http://example.test", func() {})
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			resp, err := tt.giveCall(t.Context(), client)

			require.Error(t, err)
			assert.Nil(t, resp)
		})
	}
}

type recordingHook struct {
	beforeCalls   atomic.Int32
	afterCalls    atomic.Int32
	beforeErr     error
	afterErr      error
	beforeContext *HookContext
	afterContext  *HookContext
}

// Before 记录 Hook 前置调用并按配置返回错误。
//
// 该辅助方法用于验证 client.Do 是否在请求发送前调用 Hook，并在前置 Hook 失败时中止请求。
//
// 参数：
//   - ctx: HTTP 请求生命周期中的 Hook 上下文。
//
// 返回：
//   - error: 预设的前置 Hook 错误；未设置时为 nil。
func (h *recordingHook) Before(ctx *HookContext) error {
	h.beforeCalls.Add(1)
	h.beforeContext = ctx
	return h.beforeErr
}

// After 记录 Hook 后置调用并按配置返回错误。
//
// 该辅助方法用于验证 client.Do 是否在请求完成后传递响应和错误信息，并确认后置 Hook 错误不会覆盖请求结果。
//
// 参数：
//   - ctx: HTTP 请求生命周期中的 Hook 上下文。
//
// 返回：
//   - error: 预设的后置 Hook 错误；未设置时为 nil。
func (h *recordingHook) After(ctx *HookContext) error {
	h.afterCalls.Add(1)
	h.afterContext = ctx
	return h.afterErr
}

// TestClient_DoHookLifecycle 验证 Do 方法的 Hook 生命周期语义。
//
// 该测试覆盖前置 Hook 成功、前置 Hook 阻断请求，以及后置 Hook 错误被忽略的行为，确保 Hook
// 与底层 HTTP 请求之间的执行顺序和结果传递稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestClient_DoHookLifecycle(t *testing.T) {
	var serverHits atomic.Int32
	server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		serverHits.Add(1)
		w.WriteHeader(stdhttp.StatusAccepted)
		_, _ = w.Write([]byte("accepted"))
	}))
	t.Cleanup(server.Close)

	errBefore := errors.New("before hook failed")
	errAfter := errors.New("after hook failed")

	tests := []struct {
		name           string
		description    string
		giveHook       *recordingHook
		wantErrIs      error
		wantStatusCode int
		wantHitDelta   int32
		assert         func(t *testing.T, hook *recordingHook, resp *stdhttp.Response)
	}{
		{
			name:           "success/hook-observes-request-and-result",
			description:    "验证前置 Hook 可观察请求信息，后置 Hook 可观察响应结果。",
			giveHook:       &recordingHook{},
			wantStatusCode: stdhttp.StatusAccepted,
			wantHitDelta:   1,
			assert: func(t *testing.T, hook *recordingHook, resp *stdhttp.Response) {
				assert.Equal(t, int32(1), hook.beforeCalls.Load())
				assert.Equal(t, int32(1), hook.afterCalls.Load())
				require.NotNil(t, hook.beforeContext)
				require.NotNil(t, hook.afterContext)
				assert.Equal(t, stdhttp.MethodGet, hook.beforeContext.Method())
				assert.Equal(t, resp, hook.afterContext.OriginResult())
				assert.NoError(t, hook.afterContext.OriginError())
				assert.False(t, hook.afterContext.EndTime().IsZero())
			},
		},
		{
			name:         "error/before-hook-stops-request",
			description:  "验证前置 Hook 返回错误时 Do 不会发送 HTTP 请求且不会调用后置 Hook。",
			giveHook:     &recordingHook{beforeErr: errBefore},
			wantErrIs:    errBefore,
			wantHitDelta: 0,
			assert: func(t *testing.T, hook *recordingHook, resp *stdhttp.Response) {
				assert.Nil(t, resp)
				assert.Equal(t, int32(1), hook.beforeCalls.Load())
				assert.Equal(t, int32(0), hook.afterCalls.Load())
			},
		},
		{
			name:           "success/after-hook-error-ignored",
			description:    "验证后置 Hook 返回错误不会覆盖底层 HTTP 请求的成功结果。",
			giveHook:       &recordingHook{afterErr: errAfter},
			wantStatusCode: stdhttp.StatusAccepted,
			wantHitDelta:   1,
			assert: func(t *testing.T, hook *recordingHook, resp *stdhttp.Response) {
				assert.Equal(t, int32(1), hook.beforeCalls.Load())
				assert.Equal(t, int32(1), hook.afterCalls.Load())
				assert.Equal(t, resp, hook.afterContext.OriginResult())
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			startHits := serverHits.Load()
			client := NewClient(WithHook(tt.giveHook))
			req, err := stdhttp.NewRequestWithContext(t.Context(), stdhttp.MethodGet, server.URL, nil)
			require.NoError(t, err)

			resp, err := client.Do(t.Context(), req)
			defer closeResponseBody(t, resp)

			if tt.wantErrIs != nil {
				require.ErrorIs(t, err, tt.wantErrIs)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			}
			assert.Equal(t, tt.wantHitDelta, serverHits.Load()-startHits)
			tt.assert(t, tt.giveHook, resp)
		})
	}
}

// TestClient_DoWithoutHook 验证 Do 在未配置 Hook 时直接委托给底层标准库客户端。
//
// 该测试显式清空客户端 Hook，覆盖无 Hook 分支并确认请求仍能通过本地服务完成。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestClient_DoWithoutHook(t *testing.T) {
	server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusCreated)
		_, _ = w.Write([]byte(r.Method))
	}))
	t.Cleanup(server.Close)

	client := NewClient(
		WithLogSlow(0),
		WithLogError(false),
		WithTraceEnable(false),
	).(*client)
	client.hook = nil
	req, err := stdhttp.NewRequestWithContext(t.Context(), stdhttp.MethodPost, server.URL, strings.NewReader("body"))
	require.NoError(t, err)

	resp, err := client.Do(t.Context(), req)

	require.NoError(t, err)
	assert.Equal(t, stdhttp.StatusCreated, resp.StatusCode)
	assert.Equal(t, stdhttp.MethodPost, string(readResponseBody(t, resp)))
}

type fakeClient struct {
	calls []fakeClientCall
}

type fakeClientCall struct {
	Operation string
	Method    string
	URL       string
	Body      string
	Form      url.Values
	JSON      any
}

// Do 记录全局 Do 包装函数传入的原始请求。
//
// 该辅助方法实现 Client 接口，用于验证包级全局函数是否委托给 clientDefault。
//
// 参数：
//   - ctx: 请求上下文，本 fake 不读取该值。
//   - req: 调用方传入的 HTTP 请求。
//
// 返回：
//   - *http.Response: 固定的成功响应。
//   - error: 始终为 nil。
func (f *fakeClient) Do(ctx context.Context, req *stdhttp.Request) (*stdhttp.Response, error) {
	body := ""
	if req.Body != nil {
		data, _ := io.ReadAll(req.Body)
		body = string(data)
	}
	f.calls = append(f.calls, fakeClientCall{
		Operation: "Do",
		Method:    req.Method,
		URL:       req.URL.String(),
		Body:      body,
	})
	return newFakeResponse(), nil
}

// Head 记录全局 Head 包装函数传入的 URL。
//
// 该辅助方法实现 Client 接口，用于验证包级 Head 函数的委托行为。
//
// 参数：
//   - ctx: 请求上下文，本 fake 不读取该值。
//   - targetURL: 调用方传入的请求地址。
//
// 返回：
//   - *http.Response: 固定的成功响应。
//   - error: 始终为 nil。
func (f *fakeClient) Head(ctx context.Context, targetURL string) (*stdhttp.Response, error) {
	f.calls = append(f.calls, fakeClientCall{Operation: "Head", Method: stdhttp.MethodHead, URL: targetURL})
	return newFakeResponse(), nil
}

// Get 记录全局 Get 包装函数传入的 URL。
//
// 该辅助方法实现 Client 接口，用于验证包级 Get 函数的委托行为。
//
// 参数：
//   - ctx: 请求上下文，本 fake 不读取该值。
//   - targetURL: 调用方传入的请求地址。
//
// 返回：
//   - *http.Response: 固定的成功响应。
//   - error: 始终为 nil。
func (f *fakeClient) Get(ctx context.Context, targetURL string) (*stdhttp.Response, error) {
	f.calls = append(f.calls, fakeClientCall{Operation: "Get", Method: stdhttp.MethodGet, URL: targetURL})
	return newFakeResponse(), nil
}

// Post 记录全局 Post 包装函数传入的 URL 与请求体。
//
// 该辅助方法实现 Client 接口，用于验证包级 Post 函数的委托行为。
//
// 参数：
//   - ctx: 请求上下文，本 fake 不读取该值。
//   - targetURL: 调用方传入的请求地址。
//   - body: 调用方传入的请求体。
//
// 返回：
//   - *http.Response: 固定的成功响应。
//   - error: 始终为 nil。
func (f *fakeClient) Post(ctx context.Context, targetURL string, body io.Reader) (*stdhttp.Response, error) {
	data, _ := io.ReadAll(body)
	f.calls = append(f.calls, fakeClientCall{Operation: "Post", Method: stdhttp.MethodPost, URL: targetURL, Body: string(data)})
	return newFakeResponse(), nil
}

// PostForm 记录全局 PostForm 包装函数传入的 URL 与表单数据。
//
// 该辅助方法实现 Client 接口，用于验证包级 PostForm 函数的委托行为。
//
// 参数：
//   - ctx: 请求上下文，本 fake 不读取该值。
//   - targetURL: 调用方传入的请求地址。
//   - data: 调用方传入的表单数据。
//
// 返回：
//   - *http.Response: 固定的成功响应。
//   - error: 始终为 nil。
func (f *fakeClient) PostForm(ctx context.Context, targetURL string, data url.Values) (*stdhttp.Response, error) {
	f.calls = append(f.calls, fakeClientCall{Operation: "PostForm", Method: stdhttp.MethodPost, URL: targetURL, Form: data})
	return newFakeResponse(), nil
}

// PostJSON 记录全局 PostJSON 包装函数传入的 URL 与 JSON 数据。
//
// 该辅助方法实现 Client 接口，用于验证包级 PostJSON 函数的委托行为。
//
// 参数：
//   - ctx: 请求上下文，本 fake 不读取该值。
//   - targetURL: 调用方传入的请求地址。
//   - data: 调用方传入的 JSON 数据。
//
// 返回：
//   - *http.Response: 固定的成功响应。
//   - error: 始终为 nil。
func (f *fakeClient) PostJSON(ctx context.Context, targetURL string, data any) (*stdhttp.Response, error) {
	f.calls = append(f.calls, fakeClientCall{Operation: "PostJSON", Method: stdhttp.MethodPost, URL: targetURL, JSON: data})
	return newFakeResponse(), nil
}

// newFakeResponse 构造 fakeClient 使用的固定 HTTP 响应。
//
// 该辅助函数为全局函数委托测试提供可关闭的响应体，避免测试泄漏资源。
//
// 返回：
//   - *http.Response: 状态码为 299 的固定响应。
func newFakeResponse() *stdhttp.Response {
	return &stdhttp.Response{
		StatusCode: 299,
		Body:       io.NopCloser(strings.NewReader("fake")),
	}
}

// TestClient_GlobalFunctions 验证包级全局函数委托给默认客户端。
//
// 该测试使用手写 fake 替换 clientDefault，覆盖 Do、Head、Get、Post、PostForm 与 PostJSON 的委托语义，
// 并通过 Cleanup 恢复全局状态。
//
// 参数：
//   - t: 测试上下文，用于运行子测试、注册全局状态恢复逻辑并报告断言失败。
func TestClient_GlobalFunctions(t *testing.T) {
	originalClientDefault := clientDefault
	fake := &fakeClient{}
	clientDefault = fake
	t.Cleanup(func() {
		clientDefault = originalClientDefault
	})

	tests := []struct {
		name        string
		description string
		giveCall    func(t *testing.T) (*stdhttp.Response, error)
		wantCall    fakeClientCall
	}{
		{
			name:        "success/do",
			description: "验证全局 Do 将原始请求委托给 clientDefault.Do。",
			giveCall: func(t *testing.T) (*stdhttp.Response, error) {
				req, err := stdhttp.NewRequestWithContext(t.Context(), stdhttp.MethodPatch, "http://example.test/do", strings.NewReader("body"))
				require.NoError(t, err)
				return Do(t.Context(), req)
			},
			wantCall: fakeClientCall{Operation: "Do", Method: stdhttp.MethodPatch, URL: "http://example.test/do", Body: "body"},
		},
		{
			name:        "success/head",
			description: "验证全局 Head 将 URL 委托给 clientDefault.Head。",
			giveCall: func(t *testing.T) (*stdhttp.Response, error) {
				return Head(t.Context(), "http://example.test/head")
			},
			wantCall: fakeClientCall{Operation: "Head", Method: stdhttp.MethodHead, URL: "http://example.test/head"},
		},
		{
			name:        "success/get",
			description: "验证全局 Get 将 URL 委托给 clientDefault.Get。",
			giveCall: func(t *testing.T) (*stdhttp.Response, error) {
				return Get(t.Context(), "http://example.test/get")
			},
			wantCall: fakeClientCall{Operation: "Get", Method: stdhttp.MethodGet, URL: "http://example.test/get"},
		},
		{
			name:        "success/post",
			description: "验证全局 Post 将 URL 和请求体委托给 clientDefault.Post。",
			giveCall: func(t *testing.T) (*stdhttp.Response, error) {
				return Post(t.Context(), "http://example.test/post", strings.NewReader("payload"))
			},
			wantCall: fakeClientCall{Operation: "Post", Method: stdhttp.MethodPost, URL: "http://example.test/post", Body: "payload"},
		},
		{
			name:        "success/post-form",
			description: "验证全局 PostForm 将 URL 和表单数据委托给 clientDefault.PostForm。",
			giveCall: func(t *testing.T) (*stdhttp.Response, error) {
				return PostForm(t.Context(), "http://example.test/form", url.Values{"a": {"1"}})
			},
			wantCall: fakeClientCall{Operation: "PostForm", Method: stdhttp.MethodPost, URL: "http://example.test/form", Form: url.Values{"a": {"1"}}},
		},
		{
			name:        "success/post-json",
			description: "验证全局 PostJSON 将 URL 和 JSON 数据委托给 clientDefault.PostJSON。",
			giveCall: func(t *testing.T) (*stdhttp.Response, error) {
				return PostJSON(t.Context(), "http://example.test/json", map[string]any{"x": 1})
			},
			wantCall: fakeClientCall{Operation: "PostJSON", Method: stdhttp.MethodPost, URL: "http://example.test/json", JSON: map[string]any{"x": 1}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			callCount := len(fake.calls)
			resp, err := tt.giveCall(t)
			defer closeResponseBody(t, resp)

			require.NoError(t, err)
			require.Len(t, fake.calls, callCount+1)
			assert.Equal(t, tt.wantCall, fake.calls[callCount])
		})
	}
}

// TestClient_ClientDefLazyInitialization 验证默认客户端的惰性初始化和复用语义。
//
// 该测试显式清空 clientDefault，确认 clientDef 首次调用会创建客户端，后续调用复用同一实例，并在测试结束后恢复全局状态。
//
// 参数：
//   - t: 测试上下文，用于注册全局状态恢复逻辑并报告断言失败。
func TestClient_ClientDefLazyInitialization(t *testing.T) {
	originalClientDefault := clientDefault
	clientDefault = nil
	t.Cleanup(func() {
		clientDefault = originalClientDefault
	})

	first := clientDef()
	second := clientDef()

	require.IsType(t, &client{}, first)
	require.IsType(t, &client{}, second)
	firstClient := first.(*client)
	secondClient := second.(*client)
	assert.Same(t, firstClient, secondClient)
	require.IsType(t, &client{}, clientDefault)
	assert.Same(t, firstClient, clientDefault.(*client))
}
