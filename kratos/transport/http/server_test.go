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

// TestGetRouter 测试从 Kratos HTTP Server 获取 mux.Router 功能
func TestGetRouter(t *testing.T) {
	// 创建一个 Kratos HTTP 服务器
	srv := kratoshttp.NewServer()

	// 获取路由器
	router := getRouter(srv)

	// 验证获取的路由器不为空
	assert.NotNil(t, router)
	assert.IsType(t, &mux.Router{}, router)
}

// TestGetPaths 测试获取 HTTP 服务器中注册的所有路由信息
func TestGetPaths(t *testing.T) {
	// 创建一个 Kratos HTTP 服务器
	srv := kratoshttp.NewServer()

	// 注册一些测试路由
	srv.Handle("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.HandlePrefix("/prefix", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// 获取所有路由信息
	routes := GetPaths(srv)

	// 验证路由信息不为空
	assert.NotNil(t, routes)
}

// TestRouteInfo 测试 RouteInfo 结构体
func TestRouteInfo(t *testing.T) {
	// 创建一个 RouteInfo 实例
	routeInfo := RouteInfo{
		method: "GET",
		path:   "/test",
	}

	// 验证字段值
	assert.Equal(t, "GET", routeInfo.method)
	assert.Equal(t, "/test", routeInfo.path)
}

// TestBasicParse 测试基本的 Parse 功能
func TestBasicParse(t *testing.T) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 创建一个 Kratos HTTP 服务器
	srv := kratoshttp.NewServer()

	// 创建一个 Gin 引擎
	engine := gin.New()

	// 验证 Parse 不会导致 panic
	assert.NotPanics(t, func() {
		Parse(srv, engine)
	})
}

// TestParseWithPathProcessing 测试 Parse 函数中的路径处理
func TestParseWithPathProcessing(t *testing.T) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 创建一个 Kratos HTTP 服务器
	srv := kratoshttp.NewServer()

	// 创建一个处理函数，它将返回简单的响应
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test path processing"))
	})

	// 手动添加路由
	router := getRouter(srv)
	router.Handle("/path/with/query?param=value", handler).Methods("GET")
	router.Handle("", handler).Methods("GET")

	// 创建一个 Gin 引擎
	engine := gin.New()

	// 将 Kratos 路由解析到 Gin 引擎中
	Parse(srv, engine)

	// 测试不带查询参数的路径
	req1 := httptest.NewRequest("GET", "/path/with/query", nil)
	resp1 := httptest.NewRecorder()
	engine.ServeHTTP(resp1, req1)

	// 测试空路径变为根路径
	req2 := httptest.NewRequest("GET", "/", nil)
	resp2 := httptest.NewRecorder()
	engine.ServeHTTP(resp2, req2)
}

// TestParseWithNilParams 测试 Parse 函数处理 nil 参数
func TestParseWithNilParams(t *testing.T) {
	// 测试 nil 服务器
	assert.NotPanics(t, func() {
		Parse(nil, gin.New())
	})

	// 测试 nil 引擎
	assert.NotPanics(t, func() {
		Parse(kratoshttp.NewServer(), nil)
	})

	// 测试两个都是 nil
	assert.NotPanics(t, func() {
		Parse(nil, nil)
	})
}

// TestParseWithNoRoutes 测试处理没有路由的情况
func TestParseWithNoRoutes(t *testing.T) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 创建一个没有注册任何路由的 Kratos HTTP 服务器
	srv := kratoshttp.NewServer()

	// 创建一个 Gin 引擎
	engine := gin.New()

	// 将 Kratos 路由解析到 Gin 引擎
	Parse(srv, engine)

	// 获取所有路由信息
	routes := GetPaths(srv)

	// 验证返回空路由列表而不是 nil 或导致崩溃
	assert.NotNil(t, routes)
}

// TestGetPathsWithNilRouter 测试处理 nil 路由器的情况
func TestGetPathsWithNilRouter(t *testing.T) {
	// 创建一个特殊的空服务器
	// 注意: 实际使用中不应该出现这种情况，这里仅用于测试代码的健壮性
	var srv *kratoshttp.Server

	// 预期获取路径应该优雅地处理 nil 路由器的情况
	assert.NotPanics(t, func() {
		routes := GetPaths(srv)
		// 对于 nil 服务器，预期返回空切片而不是 nil
		assert.NotNil(t, routes)
		assert.Empty(t, routes)
	})
}

// TestGetPathsWithRouteErrors 测试 GetPaths 处理路由错误的情况
func TestGetPathsWithRouteErrors(t *testing.T) {
	// 创建一个 Kratos HTTP 服务器
	srv := kratoshttp.NewServer()

	// 注册一个复杂路径，可能在获取路径模板或方法时产生错误
	complexPath := "/test/{param:.*}"
	srv.Handle(complexPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// 验证 GetPaths 不会因为处理错误而抛出 panic
	assert.NotPanics(t, func() {
		routes := GetPaths(srv)
		assert.NotNil(t, routes)
	})
}

// BenchmarkParse 基准测试 Parse 函数的性能
func BenchmarkParse(b *testing.B) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 创建一个 Kratos HTTP 服务器
	srv := kratoshttp.NewServer()

	// 注册一些测试路由
	srv.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Handle("/api/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	// 重置基准计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		// 为每次迭代创建新的 Gin 引擎
		engine := gin.New()

		// 将 Kratos 路由解析到 Gin 引擎
		Parse(srv, engine)
	}
}
