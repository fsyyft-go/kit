// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// TestGetRouter 测试从 Kratos HTTP Server 获取 mux.Router 功能。
func TestGetRouter(t *testing.T) {
	// 创建一个 Kratos HTTP 服务器。
	srv := kratoshttp.NewServer()

	// 获取路由器。
	router := getRouter(srv)

	// 验证获取的路由器不为空。
	assert.NotNil(t, router)
	assert.IsType(t, &mux.Router{}, router)
}

// TestRouteInfo 测试 RouteInfo 结构体。
func TestRouteInfo(t *testing.T) {
	// 创建一个 RouteInfo 实例。
	routeInfo := RouteInfo{
		method: "GET",
		path:   "/test",
	}

	// 验证字段值。
	assert.Equal(t, "GET", routeInfo.method)
	assert.Equal(t, "/test", routeInfo.path)
}

// TestGetPathsScenarios 测试 GetPaths 函数在各种场景下的行为。
func TestGetPathsScenarios(t *testing.T) {
	// 定义测试用例。
	tests := []struct {
		name          string                    // 测试用例名称。
		setupServer   func() *kratoshttp.Server // 准备服务器的函数。
		expectEmpty   bool                      // 是否期望返回空列表。
		validatePaths func([]RouteInfo)         // 验证路径列表的函数。
	}{
		{
			name: "正常路由",
			setupServer: func() *kratoshttp.Server {
				srv := kratoshttp.NewServer()
				srv.Handle("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				srv.HandlePrefix("/prefix", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				return srv
			},
			expectEmpty: false,
			validatePaths: func(routes []RouteInfo) {
				assert.NotEmpty(t, routes)
			},
		},
		{
			name: "空服务器",
			setupServer: func() *kratoshttp.Server {
				return nil
			},
			expectEmpty: true,
			validatePaths: func(routes []RouteInfo) {
				assert.Empty(t, routes)
			},
		},
		{
			name: "复杂路径",
			setupServer: func() *kratoshttp.Server {
				srv := kratoshttp.NewServer()
				complexPath := "/test/{param:.*}"
				srv.Handle(complexPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				return srv
			},
			expectEmpty: false,
			validatePaths: func(routes []RouteInfo) {
				assert.NotEmpty(t, routes)
			},
		},
		{
			name: "无路由",
			setupServer: func() *kratoshttp.Server {
				return kratoshttp.NewServer()
			},
			expectEmpty: false,
			validatePaths: func(routes []RouteInfo) {
				// 返回空路由列表而不是 nil。
				assert.NotNil(t, routes)
			},
		},
	}

	// 执行测试用例。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 准备服务器。
			srv := tt.setupServer()

			// 预期 GetPaths 不会导致 panic。
			assert.NotPanics(t, func() {
				// 获取所有路由信息。
				routes := GetPaths(srv)

				// 验证路由信息。
				assert.NotNil(t, routes)
				if tt.expectEmpty {
					assert.Empty(t, routes)
				}
				if tt.validatePaths != nil {
					tt.validatePaths(routes)
				}
			})
		})
	}
}

// TestParseBasicScenarios 测试 Parse 函数的基本场景。
func TestParseBasicScenarios(t *testing.T) {
	// 设置 Gin 为测试模式。
	gin.SetMode(gin.TestMode)

	// 定义测试用例。
	tests := []struct {
		name        string                    // 测试用例名称。
		setupServer func() *kratoshttp.Server // 准备服务器的函数。
		setupEngine func() *gin.Engine        // 准备 Gin 引擎的函数。
	}{
		{
			name: "基本解析",
			setupServer: func() *kratoshttp.Server {
				return kratoshttp.NewServer()
			},
			setupEngine: func() *gin.Engine {
				return gin.New()
			},
		},
		{
			name: "无路由服务器",
			setupServer: func() *kratoshttp.Server {
				return kratoshttp.NewServer()
			},
			setupEngine: func() *gin.Engine {
				return gin.New()
			},
		},
	}

	// 执行测试用例。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 准备服务器和引擎。
			srv := tt.setupServer()
			engine := tt.setupEngine()

			// 验证 Parse 不会导致 panic。
			assert.NotPanics(t, func() {
				Parse(srv, engine)
			})

			// 对于无路由服务器场景，验证返回空路由列表而不是 nil。
			if tt.name == "无路由服务器" {
				routes := GetPaths(srv)
				assert.NotNil(t, routes)
			}
		})
	}
}

// TestParseWithPathProcessing 测试 Parse 函数中的路径处理。
func TestParseWithPathProcessing(t *testing.T) {
	// 设置 Gin 为测试模式。
	gin.SetMode(gin.TestMode)

	// 定义测试用例。
	tests := []struct {
		name        string                              // 测试用例名称。
		setupRoute  func(*mux.Router, http.HandlerFunc) // 设置路由的函数。
		testRequest func(*gin.Engine)                   // 测试请求的函数。
	}{
		{
			name: "处理查询参数",
			setupRoute: func(router *mux.Router, handler http.HandlerFunc) {
				router.Handle("/path/with/query?param=value", handler).Methods("GET")
			},
			testRequest: func(engine *gin.Engine) {
				req := httptest.NewRequest("GET", "/path/with/query", nil)
				resp := httptest.NewRecorder()
				engine.ServeHTTP(resp, req)
				// 这里不验证响应内容，只确保处理不会崩溃。
			},
		},
		{
			name: "处理空路径",
			setupRoute: func(router *mux.Router, handler http.HandlerFunc) {
				router.Handle("", handler).Methods("GET")
			},
			testRequest: func(engine *gin.Engine) {
				req := httptest.NewRequest("GET", "/", nil)
				resp := httptest.NewRecorder()
				engine.ServeHTTP(resp, req)
				// 这里不验证响应内容，只确保处理不会崩溃。
			},
		},
	}

	// 执行测试用例。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建服务器。
			srv := kratoshttp.NewServer()

			// 创建处理函数。
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("test path processing"))
			})

			// 设置路由。
			router := getRouter(srv)
			tt.setupRoute(router, handler)

			// 创建 Gin 引擎。
			engine := gin.New()

			// 将 Kratos 路由解析到 Gin 引擎中。
			Parse(srv, engine)

			// 测试请求。
			tt.testRequest(engine)
		})
	}
}

// TestParseWithNilParams 测试 Parse 函数处理 nil 参数。
func TestParseWithNilParams(t *testing.T) {
	// 定义测试用例。
	tests := []struct {
		name   string             // 测试用例名称。
		server *kratoshttp.Server // 服务器。
		engine *gin.Engine        // Gin 引擎。
	}{
		{
			name:   "nil 服务器",
			server: nil,
			engine: gin.New(),
		},
		{
			name:   "nil 引擎",
			server: kratoshttp.NewServer(),
			engine: nil,
		},
		{
			name:   "全部 nil",
			server: nil,
			engine: nil,
		},
	}

	// 执行测试用例。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证 Parse 不会导致 panic。
			assert.NotPanics(t, func() {
				Parse(tt.server, tt.engine)
			})
		})
	}
}

// BenchmarkParse 基准测试 Parse 函数的性能。
func BenchmarkParse(b *testing.B) {
	// 设置 Gin 为测试模式。
	gin.SetMode(gin.TestMode)

	// 创建一个 Kratos HTTP 服务器。
	srv := kratoshttp.NewServer()

	// 注册一些测试路由。
	srv.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Handle("/api/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// 重置基准计时器。
	b.ResetTimer()

	// 运行基准测试。
	for i := 0; i < b.N; i++ {
		// 为每次迭代创建新的 Gin 引擎。
		engine := gin.New()

		// 将 Kratos 路由解析到 Gin 引擎。
		Parse(srv, engine)
	}
}
