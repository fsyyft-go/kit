// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package middleware 汇总用于 Kratos 服务端请求处理的中间件子包。
//
// 当前子包包括 basicauth 和 validate：basicauth 提供基于 HTTP Basic
// Authentication 的服务端认证中间件；validate 提供调用请求对象
// Validate() error 方法的校验中间件。调用方应直接导入所需子包，并按
// Kratos middleware.Middleware 契约接入服务端链路。
//
// 本包本身仅作为分类入口，不直接导出中间件构造函数。各子包的错误返回、
// 默认配置和自定义回调语义在对应 package comment 与函数文档中说明。
package middleware
