// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package basicauth 提供用于 Kratos 服务端的 Basic Authentication 中间件。
//
// Server 从 transport.ServerContext 的 Authorization 请求头中解析 Basic
// 凭据，并使用 WithValidator 提供的回调校验用户名和密码。校验失败时，
// 中间件会设置 WWW-Authenticate 响应头并返回 ErrInvalidBasicAuth。
//
// 默认 validator 始终拒绝认证，realm 默认为 Restricted，因此公开服务通常
// 需要显式提供凭据校验逻辑。若上下文中没有服务端 transport 信息，
// 中间件会跳过认证并继续调用后续处理器。
package basicauth
