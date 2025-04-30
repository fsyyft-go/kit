// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	kitlog "github.com/fsyyft-go/kit/log"
)

//
// 单元测试设计说明：
//
// 本文件系统性覆盖 HTTP Client 的全部核心功能，包括：
// - GET/POST/HEAD/表单/JSON 请求的正常与异常流程
// - 超时、错误、代理、Option 配置、全局方法、钩子、trace 等边界与扩展能力
// - 所有测试均基于本地 httptest.Server，避免外部依赖，保证可重复性
// - 用例采用表格驱动，断言全部使用 testify，注释前置、中文标点
// - 钩子、trace、Option、全局方法等均有独立专项测试
// - 便于 CI/CD 自动化与后续维护
//
// 使用方法：
//   go test -v ./net/http
//

// mockHandler 用于本地 HTTP Server，支持多种请求类型与响应。
func mockHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/get":
		w.Header().Set("X-Test", "ok")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("get-ok"))
	case "/head":
		w.Header().Set("X-Test", "ok")
		w.WriteHeader(http.StatusNoContent)
	case "/post":
		b, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("post-" + string(b)))
	case "/form":
		r.ParseForm()
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		_, _ = w.Write([]byte(r.Form.Encode()))
	case "/json":
		var m map[string]any
		_ = json.NewDecoder(r.Body).Decode(&m)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(m)
	case "/timeout":
		time.Sleep(300 * time.Millisecond) // 保证超时
		w.WriteHeader(http.StatusOK)
	case "/error":
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error"))
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// TestClient_AllScenarios 覆盖所有核心请求类型与边界场景。
func TestClient_AllScenarios(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer ts.Close()

	client := NewClient(
		WithTimeout(50 * time.Millisecond), // 必然超时
	)
	type testCase struct {
		name   string
		method string
		path   string
		body   io.Reader
		form   url.Values
		json   any
		want   int
		check  func(*testing.T, *http.Response, error)
	}

	cases := []testCase{
		{
			name:   "GET 正常",
			method: http.MethodGet,
			path:   "/get",
			want:   http.StatusOK,
			check: func(t *testing.T, resp *http.Response, err error) {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				if resp != nil {
					b, _ := io.ReadAll(resp.Body)
					_ = resp.Body.Close()
					t.Logf("GET body: %s", string(b))
					assert.Equal(t, "get-ok", string(b))
				}
			},
		},
		{
			name:   "HEAD 正常",
			method: http.MethodHead,
			path:   "/head",
			want:   http.StatusNoContent,
			check: func(t *testing.T, resp *http.Response, err error) {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusNoContent, resp.StatusCode)
				if resp != nil {
					_ = resp.Body.Close()
				}
			},
		},
		{
			name:   "POST 正常",
			method: http.MethodPost,
			path:   "/post",
			body:   strings.NewReader("abc"),
			want:   http.StatusOK,
			check: func(t *testing.T, resp *http.Response, err error) {
				assert.NoError(t, err)
				if resp != nil {
					b, _ := io.ReadAll(resp.Body)
					_ = resp.Body.Close()
					t.Logf("POST body: %s", string(b))
					assert.Contains(t, string(b), "post-abc")
				}
			},
		},
		{
			name:   "POST 表单",
			method: http.MethodPost,
			path:   "/form",
			form:   url.Values{"a": {"1"}, "b": {"2"}},
			want:   http.StatusOK,
			check: func(t *testing.T, resp *http.Response, err error) {
				assert.NoError(t, err)
				if resp != nil {
					b, _ := io.ReadAll(resp.Body)
					_ = resp.Body.Close()
					t.Logf("FORM body: %s", string(b))
					assert.Contains(t, string(b), "a=1")
					assert.Contains(t, string(b), "b=2")
				}
			},
		},
		{
			name:   "POST JSON",
			method: http.MethodPost,
			path:   "/json",
			json:   map[string]any{"x": 1, "y": "z"},
			want:   http.StatusOK,
			check: func(t *testing.T, resp *http.Response, err error) {
				assert.NoError(t, err)
				if resp != nil {
					var m map[string]any
					_ = json.NewDecoder(resp.Body).Decode(&m)
					_ = resp.Body.Close()
					t.Logf("JSON body: %+v", m)
					assert.Equal(t, float64(1), m["x"])
					assert.Equal(t, "z", m["y"])
				}
			},
		},
		{
			name:   "超时",
			method: http.MethodGet,
			path:   "/timeout",
			want:   0,
			check: func(t *testing.T, resp *http.Response, err error) {
				assert.Error(t, err)
				assert.Nil(t, resp)
			},
		},
		{
			name:   "404",
			method: http.MethodGet,
			path:   "/notfound",
			want:   http.StatusNotFound,
			check: func(t *testing.T, resp *http.Response, err error) {
				assert.NoError(t, err)
				if resp != nil {
					assert.Equal(t, http.StatusNotFound, resp.StatusCode)
					_ = resp.Body.Close()
				}
			},
		},
		{
			name:   "500",
			method: http.MethodGet,
			path:   "/error",
			want:   http.StatusInternalServerError,
			check: func(t *testing.T, resp *http.Response, err error) {
				assert.NoError(t, err)
				if resp != nil {
					assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
					_ = resp.Body.Close()
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			url := ts.URL + c.path
			var resp *http.Response
			var err error
			switch {
			case c.form != nil:
				resp, err = client.PostForm(context.Background(), url, c.form)
			case c.json != nil:
				resp, err = client.PostJSON(context.Background(), url, c.json)
			case c.method == http.MethodGet:
				resp, err = client.Get(context.Background(), url)
			case c.method == http.MethodHead:
				resp, err = client.Head(context.Background(), url)
			case c.method == http.MethodPost && c.body != nil:
				resp, err = client.Post(context.Background(), url, c.body)
			}
			c.check(t, resp, err)
		})
	}
}

// TestClient_Option 覆盖 Option 配置项。
func TestClient_Option(t *testing.T) {
	logger, _ := kitlog.NewStdLogger("")
	client := NewClient(
		WithName("test-opt"),
		WithTimeout(123*time.Millisecond),
		WithTraceEnable(true),
		WithProxy(nil),
		WithMaxConnsPerHost(1),
		WithMaxIdleConnsPerHost(1),
		WithMaxIdleConns(1),
		WithLogger(logger),
	)
	assert.NotNil(t, client)
}

// mockHook 用于测试钩子调用。
type mockHook struct {
	before, after atomic.Bool
}

// Before 钩子前置逻辑。
func (h *mockHook) Before(ctx *HookContext) error {
	h.before.Store(true)
	return nil
}

// After 钩子后置逻辑。
func (h *mockHook) After(ctx *HookContext) error {
	h.after.Store(true)
	return nil
}

// TestClient_Hook 测试钩子机制。
func TestClient_Hook(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer ts.Close()

	hook := &mockHook{}
	client := NewClient(WithHook(hook))
	resp, err := client.Get(context.Background(), ts.URL+"/get")
	if resp != nil {
		_ = resp.Body.Close()
	}
	assert.NoError(t, err)
	assert.True(t, hook.before.Load())
	assert.True(t, hook.after.Load())
}

// TestClient_Trace 测试 trace hook 注入。
func TestClient_Trace(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer ts.Close()

	logger, _ := kitlog.NewStdLogger("")
	client := NewClient(WithTraceEnable(true), WithLogger(logger))
	resp, err := client.Get(context.Background(), ts.URL+"/get")
	if resp != nil {
		_ = resp.Body.Close()
	}
	assert.NoError(t, err)
}

// TestClient_GlobalFuncs 测试全局方法。
func TestClient_GlobalFuncs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer ts.Close()

	resp, err := Get(context.Background(), ts.URL+"/get")
	if resp != nil {
		_ = resp.Body.Close()
	}
	assert.NoError(t, err)

	resp, err = Head(context.Background(), ts.URL+"/head")
	if resp != nil {
		_ = resp.Body.Close()
	}
	assert.NoError(t, err)

	resp, err = Post(context.Background(), ts.URL+"/post", strings.NewReader("abc"))
	if resp != nil {
		_ = resp.Body.Close()
	}
	assert.NoError(t, err)

	resp, err = PostForm(context.Background(), ts.URL+"/form", url.Values{"a": {"1"}})
	if resp != nil {
		_ = resp.Body.Close()
	}
	assert.NoError(t, err)

	resp, err = PostJSON(context.Background(), ts.URL+"/json", map[string]any{"x": 1})
	if resp != nil {
		_ = resp.Body.Close()
	}
	assert.NoError(t, err)
}
