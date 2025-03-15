// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package kratos 提供了对 go-kratos 框架的扩展和增强功能，包括配置解析、中间件和传输层的增强实现。

主要组件：

1. config - 配置管理：
  - 提供灵活的配置解码器
  - 支持点分隔键的配置展开
  - 支持自定义配置解析
  - 集成 Kratos 配置系统

2. middleware - 中间件：
  - basicauth：HTTP 基本认证中间件
  - validate：请求验证中间件

3. transport - 传输层：
  - http：Gin 框架集成
  - 路由管理和转换
  - 中间件支持

基本功能：

1. 配置解码：

	// 创建配置解码器
	decoder := config.NewDecoder(
	    config.WithResolve(func(target map[string]interface{}) error {
	        // 自定义配置处理逻辑
	        return nil
	    }),
	)

	// 解码配置
	var cfg map[string]interface{}
	err := decoder.Decode(kv, cfg)

2. HTTP 服务集成：

	// 创建 Gin 引擎
	engine := gin.Default()

	// 创建 Kratos HTTP 服务器
	srv := http.NewServer()

	// 注册路由
	srv.HandleFunc("/hello", HelloHandler)

	// 将 Kratos 路由解析到 Gin
	http.Parse(srv, engine)

3. 中间件使用：

	// 基本认证中间件
	auth := basicauth.NewMiddleware(
	    basicauth.WithValidator(func(username, password string) bool {
	        return username == "admin" && password == "password"
	    }),
	)

	// 请求验证中间件
	validate := validate.NewMiddleware()

配置功能：

1. 解码器选项：
  - 支持自定义解析函数
  - 支持配置验证
  - 支持默认值设置

2. 配置格式：
  - 支持多种编码格式
  - 支持嵌套配置
  - 支持环境变量

HTTP 功能：

1. 路由管理：
  - 支持路由信息提取
  - 支持路由转换
  - 支持中间件链

2. 服务集成：
  - Gin 框架适配
  - 中间件支持
  - 错误处理

中间件功能：

1. 基本认证：
  - 用户验证
  - 安全头处理
  - 自定义验证器

2. 请求验证：
  - 参数验证
  - 自定义验证规则
  - 错误处理

使用建议：

1. 配置管理：
  - 使用点分隔的配置键
  - 实现自定义解析逻辑
  - 注意配置验证

2. HTTP 服务：
  - 合理组织路由
  - 正确处理中间件顺序
  - 注意错误处理

3. 中间件使用：
  - 选择合适的中间件
  - 自定义验证逻辑
  - 处理异常情况

注意事项：

1. 配置处理：
  - 验证配置完整性
  - 处理配置更新
  - 注意类型转换

2. 服务集成：
  - 注意路由冲突
  - 处理并发请求
  - 合理设置超时

3. 安全考虑：
  - 验证请求来源
  - 保护敏感信息
  - 限制访问权限

更多示例请参考 example/kratos 目录。
*/
package kratos
