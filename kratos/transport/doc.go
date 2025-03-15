// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package transport 提供了 Kratos 框架的传输层实现，主要集成了 Gin 框架的 HTTP 服务功能。

主要组件：

1. http - HTTP 服务实现：
  - Gin 框架集成
  - 路由管理和转换
  - 中间件支持
  - 错误处理机制

基本用法：

1. 创建 HTTP 服务：

	// 创建 Gin 引擎
	engine := gin.Default()

	// 创建 Kratos HTTP 服务器
	srv := http.NewServer()

	// 注册路由
	srv.HandleFunc("/hello", HelloHandler)

	// 将 Kratos 路由解析到 Gin
	http.Parse(srv, engine)

2. 路由处理：

	// 定义处理函数
	func HelloHandler(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	    return &HelloResponse{
	        Message: "Hello " + req.Name,
	    }, nil
	}

	// 注册路由
	srv.HandleFunc("/hello", HelloHandler)

3. 中间件使用：

	// 添加全局中间件
	srv.Use(middleware.Chain(
	    recovery.Recovery(),
	    tracing.Server(),
	))

	// 添加路由中间件
	srv.HandleFunc("/secure", SecureHandler, auth.BasicAuth())

服务功能：

1. HTTP 服务：
  - 支持 RESTful API
  - 支持中间件链
  - 支持路由组
  - 支持参数绑定

2. 路由管理：
  - 路由注册
  - 路由分组
  - 路由参数
  - 路由中间件

性能优化：

1. 请求处理：
  - 高效路由匹配
  - 请求池复用
  - 响应缓存

2. 并发处理：
  - 连接池管理
  - 请求限流
  - 超时控制

使用建议：

1. 服务配置：
  - 合理设置超时
  - 配置连接池
  - 设置请求限制

2. 路由设计：
  - 合理组织路由
  - 使用路由组
  - 添加适当中间件

3. 错误处理：
  - 统一错误响应
  - 记录错误日志
  - 提供错误详情

注意事项：

1. 服务管理：
  - 优雅启动关闭
  - 处理连接泄漏
  - 监控服务状态

2. 安全考虑：
  - 验证请求来源
  - 限制请求大小
  - 防止 DoS 攻击

3. 性能调优：
  - 监控性能指标
  - 优化响应时间
  - 控制资源使用

更多示例请参考 example/transport 目录。
*/
package transport
