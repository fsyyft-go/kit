// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// package http 提供 Kratos HTTP 服务器与 Gin 框架的集成功能。
package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/middleware"
	kratos_http "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/mux"
)

type (
	// matcher 接口定义了中间件匹配器的基本行为。
	// 用于管理和匹配 HTTP 操作的中间件。
	matcher interface {
		// Use 添加全局中间件。
		Use(ms ...middleware.Middleware)

		// Add 为特定选择器添加中间件。
		Add(selector string, ms ...middleware.Middleware)

		// Match 根据操作名匹配并返回相应的中间件列表。
		Match(operation string) []middleware.Middleware
	}

	// serverAccessor 是一个用于访问 kratos http.Server 内部字段的结构体。
	// 通过 unsafe.Pointer 转换实现对私有字段的访问。
	// 字段说明：
	// - Server: 底层的 http.Server 实例。
	// - router: gorilla/mux 路由器实例。
	// 其他字段使用空标识符 _ 占位，保持内存布局与原始结构体一致。
	serverAccessor struct {
		// 标准库 http.Server 实例，处理 HTTP 请求和响应。
		*http.Server

		// 网络监听器，用于接受连接。
		_ net.Listener

		// TLS 配置，用于 HTTPS 连接。
		_ *tls.Config

		// 服务器 URL 信息。
		_ *url.URL

		// 服务器错误信息。
		_ error

		// 服务器网络地址。
		_ string

		// 服务器路径。
		_ string

		// 请求超时时间。
		_ time.Duration

		// HTTP 过滤器函数列表。
		_ []kratos_http.FilterFunc

		// 中间件匹配器。
		_ matcher

		// 请求解码函数，用于解析请求参数。
		_ kratos_http.DecodeRequestFunc

		// 请求头解码函数。
		_ kratos_http.DecodeRequestFunc

		// 请求体解码函数。
		_ kratos_http.DecodeRequestFunc

		// 响应编码函数。
		_ kratos_http.EncodeResponseFunc

		// 错误编码函数。
		_ kratos_http.EncodeErrorFunc

		// 是否启用压缩。
		_ bool

		// gorilla/mux 路由器实例，用于 HTTP 路由管理。
		router *mux.Router
	}

	// RouteInfo 结构体存储路由信息。
	// 包含 HTTP 方法和路径。
	RouteInfo struct {
		// HTTP 请求方法（GET、POST、PUT 等）。
		method string

		// 路由路径模板。
		path string
	}
)

// getRouter 从 kratos http.Server 中获取 mux.Router 实例。
// 使用 unsafe.Pointer 实现对私有字段的访问。
//
// 参数：
//   - s：kratos http.Server 指针。
//
// 返回值：
//   - *mux.Router：gorilla/mux 路由器指针。
func getRouter(s *kratos_http.Server) *mux.Router {
	// 检查是否为 nil，避免空指针异常
	if nil == s {
		return nil
	}

	// 将 kratoshttp.Server 指针转换为 serverAccessor 指针，以访问私有字段。
	sa := (*serverAccessor)(unsafe.Pointer(s))
	return sa.router
}

// GetPaths 获取 HTTP 服务器中注册的所有路由信息。
// 遍历 mux.Router 中的所有路由，提取方法和路径信息。
//
// 参数：
//   - s：kratos http.Server 指针
//
// 返回值：
//   - []RouteInfo：包含所有注册路由信息的切片
func GetPaths(s *kratos_http.Server) []RouteInfo {
	// 初始化空的路由信息切片。
	routeInfos := make([]RouteInfo, 0)

	// 获取路由器实例。
	router := getRouter(s)

	// 如果路由器为 nil，直接返回空切片
	if nil == router {
		return routeInfos
	}

	// 遍历路由器中的所有路由。
	_ = router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		// 获取路由路径模板。
		path, err := route.GetPathTemplate()
		if nil != err {
			// 如果获取路径模板失败，跳过此路由，继续处理下一个
			return nil
		}

		// 获取路由支持的 HTTP 方法。
		method, err := route.GetMethods()
		if nil != err {
			// 如果获取方法失败，假设此路由支持 GET 方法
			method = []string{"GET"}
		}

		// 为每个 HTTP 方法创建一个路由信息对象并添加到结果切片中。
		for _, m := range method {
			routeInfos = append(routeInfos, RouteInfo{
				method: m,
				path:   path,
			})
		}

		return nil
	})
	return routeInfos
}

// Parse 将 kratos http.Server 中的路由注册到 gin.Engine 中。
// 这个函数允许将 Kratos 的路由处理逻辑与 Gin 框架集成。
//
// 参数：
//   - s：kratos http.Server 指针。
//   - e：gin.Engine 指针。
func Parse(s *kratos_http.Server, e *gin.Engine) {
	// 检查参数是否为 nil
	if nil == s || nil == e {
		return
	}

	// 获取所有路由信息。
	routeInfos := GetPaths(s)

	// 遍历所有路由信息并注册到 Gin 引擎。
	for _, routeInfo := range routeInfos {
		// 处理带有查询参数的路径。
		path := routeInfo.path

		// 查找第一个问号位置，分割路径和查询参数部分。
		if idx := strings.Index(path, "?"); idx >= 0 {
			path = path[:idx]
		}

		// 确保路径不为空，空路径设置为根路径。
		if path == "" {
			path = "/"
		}

		// 在 Gin 中注册路由处理函数。
		// 将请求代理到 Kratos HTTP 服务器处理。
		e.Handle(routeInfo.method, path, func(c *gin.Context) {
			s.ServeHTTP(c.Writer, c.Request)
		})
	}
}
