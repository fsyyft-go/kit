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
	"regexp"
	"strings"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/middleware"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
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
	serverAccessor struct {
		// Server 是标准库 http.Server 实例，处理 HTTP 请求和响应。
		*http.Server

		// listener 是网络监听器，用于接受连接。
		_ net.Listener

		// tlsConf 是 TLS 配置，用于 HTTPS 连接。
		_ *tls.Config

		// endpoint 是服务器 URL 信息。
		_ *url.URL

		// err 是服务器错误信息。
		_ error

		// network 是服务器网络地址。
		_ string

		// path 是服务器路径。
		_ string

		// timeout 是请求超时时间。
		_ time.Duration

		// filter 是 HTTP 过滤器函数列表。
		_ []kratoshttp.FilterFunc

		// matcher 是中间件匹配器。
		_ matcher

		// decodeRequest 是请求解码函数，用于解析请求参数。
		_ kratoshttp.DecodeRequestFunc

		// decodeHeader 是请求头解码函数。
		_ kratoshttp.DecodeRequestFunc

		// decodeBody 是请求体解码函数。
		_ kratoshttp.DecodeRequestFunc

		// encodeResponse 是响应编码函数。
		_ kratoshttp.EncodeResponseFunc

		// encodeError 是错误编码函数。
		_ kratoshttp.EncodeErrorFunc

		// enableCompression 是否启用压缩。
		_ bool

		// router 是 gorilla/mux 路由器实例，用于 HTTP 路由管理。
		router *mux.Router
	}

	// RouteInfo 结构体存储路由信息。
	RouteInfo struct {
		// method 是 HTTP 请求方法（GET、POST、PUT 等）。
		method string

		// path 是路由路径模板。
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
func getRouter(s *kratoshttp.Server) *mux.Router {
	// 检查输入参数是否为空，避免空指针异常。
	if nil == s {
		return nil
	}

	// 将 kratoshttp.Server 指针转换为 serverAccessor 指针，以访问私有字段。
	sa := (*serverAccessor)(unsafe.Pointer(s))
	return sa.router
}

// GetPaths 获取 HTTP 服务器中注册的所有路由信息。
//
// 参数：
//   - s：kratos http.Server 指针。
//
// 返回值：
//   - []RouteInfo：包含所有注册路由信息的切片。
func GetPaths(s *kratoshttp.Server) []RouteInfo {
	// 初始化空的路由信息切片。
	routeInfos := make([]RouteInfo, 0)

	// 获取路由器实例。
	router := getRouter(s)

	// 如果路由器为空，直接返回空切片。
	if router == nil {
		return routeInfos
	}

	// 遍历路由器中的所有路由。
	_ = router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		// 获取路由路径模板。
		path, err := route.GetPathTemplate()
		if err != nil {
			// 如果获取路径模板失败，跳过此路由，继续处理下一个。
			return nil
		}

		// 获取路由支持的 HTTP 方法。
		method, err := route.GetMethods()
		if err != nil {
			// 如果获取方法失败，默认使用 GET 方法。
			method = []string{"GET"}
		}

		// 为每个 HTTP 方法创建路由信息对象并添加到结果切片中。
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
//
// 参数：
//   - s：kratos http.Server 指针。
//   - e：gin.Engine 指针。
func Parse(s *kratoshttp.Server, e *gin.Engine) {
	// 检查输入参数是否为空。
	if s == nil || e == nil {
		return
	}

	// 获取所有路由信息。
	routeInfos := GetPaths(s)

	// 遍历所有路由信息并注册到 Gin 引擎。
	for _, routeInfo := range routeInfos {
		// 将 Mux 路径格式转换为 Gin 路径格式。
		path := parsePath(routeInfo.path)

		// 在 Gin 中注册路由处理函数。
		e.Handle(routeInfo.method, path, func(c *gin.Context) {
			// 将请求代理到 Kratos HTTP 服务器处理。
			s.ServeHTTP(c.Writer, c.Request)
		})
	}
}

// parsePath 将 Mux 格式路由转换为 Gin 格式路由。
//
// 参数：
//   - path：Mux 格式的路由路径。
//
// 返回值：
//   - string：转换后的 Gin 格式路由路径。
//
// 示例：
//   - /users/{id} -> /users/:id
//   - /users/{name:[a-z]+} -> /users/:name
//   - /users/{id}/posts/{postId} -> /users/:id/posts/:postId
func parsePath(path string) string {
	// 处理空路径的情况。
	if path == "" {
		return "/"
	}

	// 保存原始路径末尾是否有斜杠的状态。
	hasTrailingSlash := strings.HasSuffix(path, "/")

	// 使用正则表达式处理路径参数，包括带正则表达式的参数。
	// 匹配 {param} 或 {param:pattern} 格式。
	regexPattern := regexp.MustCompile(`{([^:}]+)(?::[^}]*)?}`)
	path = regexPattern.ReplaceAllString(path, ":$1")

	// 处理多个连续的斜杠，将其替换为单个斜杠。
	path = regexp.MustCompile(`/+`).ReplaceAllString(path, "/")

	// 确保路径以斜杠开头。
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// 处理路径末尾的斜杠。
	if path != "/" {
		// 移除末尾的斜杠。
		path = strings.TrimRight(path, "/")
		// 如果原始路径有末尾斜杠，则保留。
		if hasTrailingSlash {
			path += "/"
		}
	}

	return path
}
