// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package middleware 提供了一组用于 Kratos HTTP 服务的中间件，包括基本认证和请求验证等功能。

主要组件：

1. basicauth - 基本认证中间件：
  - 实现 HTTP 基本认证
  - 支持自定义验证器
  - 支持安全头处理

2. validate - 请求验证中间件：
  - 支持请求参数验证
  - 支持自定义验证规则
  - 提供错误处理机制

基本用法：

1. 基本认证中间件：

	// 创建基本认证中间件
	auth := basicauth.NewMiddleware(
	    basicauth.WithValidator(func(username, password string) bool {
	        return username == "admin" && password == "password"
	    }),
	)

	// 在路由中使用
	router.Use(auth)

2. 请求验证中间件：

	// 创建验证中间件
	validate := validate.NewMiddleware()

	// 在路由中使用
	router.Use(validate)

	// 定义带验证规则的请求结构
	type Request struct {
	    Name  string `validate:"required"`
	    Email string `validate:"required,email"`
	    Age   int    `validate:"gte=0,lte=130"`
	}

中间件功能：

1. 基本认证：
  - 用户名密码验证
  - 自定义验证逻辑
  - 安全头处理
  - 错误响应定制

2. 请求验证：
  - 参数类型验证
  - 自定义验证规则
  - 验证错误处理
  - 支持嵌套结构

安全特性：

1. 认证安全：
  - 密码加密存储
  - 防止暴力破解
  - 安全头处理

2. 数据验证：
  - 输入数据净化
  - 类型安全检查
  - 防止注入攻击

性能优化：

1. 验证器缓存：
  - 缓存验证规则
  - 减少解析开销
  - 优化内存使用

2. 并发处理：
  - 线程安全设计
  - 最小化锁竞争
  - 高效请求处理

使用建议：

1. 中间件配置：
  - 合理设置超时
  - 配置错误处理
  - 自定义验证规则

2. 安全实践：
  - 使用强密码策略
  - 实施访问控制
  - 记录安全日志

3. 性能考虑：
  - 避免重复验证
  - 合理使用缓存
  - 控制中间件数量

注意事项：

1. 认证处理：
  - 保护敏感信息
  - 处理认证失败
  - 实现会话管理

2. 验证规则：
  - 定义合适规则
  - 处理验证错误
  - 提供错误信息

3. 错误处理：
  - 统一错误格式
  - 提供详细信息
  - 避免信息泄露

更多示例请参考 example/middleware 目录。
*/
package middleware
