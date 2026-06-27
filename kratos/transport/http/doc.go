// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package http 提供 Kratos HTTP Server 与 Gin Engine 之间的路由桥接工具。
//
// Parse 会遍历已注册到 kratoshttp.Server 的 mux 路由，并将其转换为等价的 Gin 路由，
// 再把请求回送给原 kratoshttp.Server 处理。
// GetPaths 提供路由提取辅助，主要用于本包桥接逻辑和调试场景；当前 RouteInfo 的字段未导出，
// 包外调用方无法直接读取其中的 method 和 path。
// 本包不创建 HTTP server，也不替换 Kratos 中间件、编解码或错误处理链语义；
// 它仅复用已有注册结果完成 Gin 侧挂载。
// 实现通过 unsafe 访问 kratoshttp.Server 内部 router 布局，升级 Kratos 版本后需要重新核对结构字段位置。
package http
