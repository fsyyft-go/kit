# HTTP Transport 包

`http` 包提供了 Kratos HTTP 服务器到 Gin 引擎的转换功能，让你能够在 Kratos 框架中使用 Gin 的路由和中间件能力。

## 特性

- 将 Kratos HTTP 路由转换为 Gin 路由
- 支持路径参数和查询参数的转换
- 提供路由信息获取功能
- 保持 Kratos 的上下文和中间件兼容性

## 快速开始

### 基本使用

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

## 路由转换说明

### 路径参数转换

Kratos 的路径参数会被自动转换为 Gin 的路径参数格式：

```go
// Kratos 路由
srv.Handle("/users/{id}", handler)

// 转换后的 Gin 路由
// /users/:id
```

### 查询参数处理

查询参数会被正确保留和处理：

```go
// Kratos 路由
srv.Handle("/search?q={query}", handler)

// 转换后的 Gin 路由
// /search
// 查询参数可以通过 c.Query("q") 访问
```

## 最佳实践

1. 路由定义
   - 使用 Kratos 的路由定义方式
   - 避免在路径中包含复杂的正则表达式
   - 保持路径参数命名的一致性

2. 性能优化
   - 避免重复调用 Parse 函数
   - 合理组织路由结构
   - 注意路由数量对性能的影响

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../../../LICENSE) 文件。 