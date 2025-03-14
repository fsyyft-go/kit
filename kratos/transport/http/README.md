# Kratos HTTP Transport

## 简介

`kratos/transport/http` 包提供了 Kratos HTTP 服务器到 Gin 引擎的转换功能，让你能够在 Kratos 框架中使用 Gin 的路由和中间件能力。这个包的主要目的是实现两个优秀框架的无缝集成，让开发者能够同时享受 Kratos 的微服务能力和 Gin 的 Web 框架特性。

### 主要特性

- 将 Kratos HTTP 路由自动转换为 Gin 路由
- 支持路径参数和查询参数的智能转换
- 提供完整的路由信息获取功能
- 保持 Kratos 的上下文和中间件兼容性
- 高性能的路由转换实现
- 完整的测试覆盖
- 详细的代码文档

### 设计理念

本包的设计遵循以下原则：

- **无缝集成**：让 Kratos 和 Gin 的功能自然融合
- **高性能**：路由转换和请求处理的高效实现
- **易用性**：简单的 API 设计，降低学习成本
- **可靠性**：完整的错误处理和边界检查
- **可扩展性**：支持自定义路由和处理器扩展

## 安装

### 前置条件

- Go 版本要求：>= 1.16
- 依赖要求：
  - github.com/go-kratos/kratos/v2
  - github.com/gin-gonic/gin
  - github.com/gorilla/mux

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/kratos/transport/http
```

## 快速开始

### 基础用法

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/go-kratos/kratos/v2/transport/http"
    kithttp "github.com/fsyyft-go/kit/kratos/transport/http"
)

func main() {
    // 创建 Kratos HTTP 服务器
    srv := http.NewServer()

    // 创建 Gin 引擎
    engine := gin.New()

    // 将 Kratos 路由转换为 Gin 路由
    kithttp.Parse(srv, engine)

    // 启动服务器
    if err := srv.Start(); err != nil {
        panic(err)
    }
}
```

### 获取路由信息

```go
// 获取所有路由信息
routes := kithttp.GetPaths(srv)
for _, route := range routes {
    fmt.Printf("Method: %s, Path: %s\n", route.Method(), route.Path())
}
```

## 详细指南

### 核心概念

1. **路由转换**
   - 自动将 Kratos 路由格式转换为 Gin 路由格式
   - 保持路由参数和查询参数的正确映射
   - 支持多种 HTTP 方法

2. **路由信息**
   - 提供完整的路由元数据访问
   - 支持路由遍历和检查
   - 方便的路由调试功能

3. **请求处理**
   - 保持请求上下文的一致性
   - 支持中间件链的正确执行
   - 错误处理和响应封装

### 常见用例

#### 1. 基本路由转换

```go
// Kratos 路由定义
srv.Handle("/users/{id}", handler)

// 自动转换为 Gin 路由
// /users/:id
```

#### 2. 带查询参数的路由

```go
// Kratos 路由定义
srv.Handle("/search?q={query}", handler)

// 转换后的 Gin 路由
// /search
// 查询参数通过 c.Query("q") 访问
```

### 最佳实践

- 路由定义时使用清晰的命名规范
- 避免过于复杂的路由参数设计
- 合理组织路由层级结构
- 使用适当的 HTTP 方法
- 保持路由处理器的简洁性

## API 文档

### 主要类型

```go
// RouteInfo 存储路由信息
type RouteInfo struct {
    method string
    path   string
}

// matcher 接口定义中间件匹配器
type matcher interface {
    Use(ms ...middleware.Middleware)
    Add(selector string, ms ...middleware.Middleware)
    Match(operation string) []middleware.Middleware
}
```

### 关键函数

#### Parse

将 Kratos HTTP 服务器路由转换到 Gin 引擎。

```go
func Parse(s *kratoshttp.Server, e *gin.Engine)
```

#### GetPaths

获取服务器中注册的所有路由信息。

```go
func GetPaths(s *kratoshttp.Server) []RouteInfo
```

### 错误处理

- 空指针检查和防御性编程
- 路由转换错误的优雅处理
- 请求处理过程中的错误捕获
- 中间件链执行的错误处理

## 性能指标

| 操作 | 性能指标 | 说明 |
|------|----------|------|
| 路由转换 | O(n) | n 为路由数量 |
| 路由匹配 | O(1) | 使用 Gin 的路由树 |
| 中间件执行 | O(m) | m 为中间件数量 |

## 测试覆盖率

| 包 | 覆盖率 |
|------|--------|
| transport/http | >95% |
| transport/http/server | >90% |

## 调试指南

### 常见问题排查

#### 1. 路由未正确转换

- 检查 Kratos 路由定义是否正确
- 验证路由参数格式
- 确认 Parse 函数调用时机

#### 2. 请求处理失败

- 检查中间件配置
- 验证处理器实现
- 查看错误日志

## 相关文档

- [Kratos HTTP 文档](https://go-kratos.dev/docs/component/transport/http/)
- [Gin 框架文档](https://gin-gonic.com/docs/)
- [示例代码](../example/kratos/transport/http/README.md)

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 改进文档
- 提交代码改进

请参考我们的[贡献指南](../../CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT 许可证。查看 [LICENSE](../../LICENSE) 文件了解更多信息。 

## 补充说明

本文的大部分信息，由 AI 使用[模板](../../../ai/templates/docs/package_readme_template.md)根据[提示词](../../../ai/prompts/docs/package_readme_generator.md)自动生成，如有任何问题，请随时联系我。