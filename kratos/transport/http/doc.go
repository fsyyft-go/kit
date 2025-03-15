// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package http 提供了基于 Gin 框架的 HTTP 服务实现，用于 Kratos 框架的传输层。
//
// 主要功能：
//   - 基于 Gin 的服务实现
//   - 路由管理和转换
//   - 中间件支持
//   - 错误处理机制
//
// 基本用法：
//
//	// 创建 Gin 引擎
//	engine := gin.Default()
//
//	// 创建 HTTP 服务器
//	srv := NewServer()
//
//	// 注册路由
//	srv.HandleFunc("/hello", HelloHandler)
//
//	// 解析路由到 Gin
//	Parse(srv, engine)
//
// 服务配置：
//   - WithAddress：设置监听地址
//   - WithTimeout：设置超时时间
//   - WithMiddleware：添加中间件
//   - WithErrorHandler：设置错误处理器
//
// 注意事项：
//   - 设置适当的超时和错误处理
//   - 合理使用中间件
//   - 注意请求限制和安全防护
//   - 监控服务状态和性能指标
package http
