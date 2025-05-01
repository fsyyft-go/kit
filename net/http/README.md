# http

## 简介

http 包提供了功能丰富的 Go HTTP 客户端，支持 GET/POST/HEAD/表单/JSON 请求、超时、代理、钩子、慢请求日志、trace、全局方法等，适合微服务、API 调用、健康检查等多种场景。

### 主要特性

- 统一 HTTP 客户端接口，支持 Do/Get/Post/Head/PostForm/PostJSON
- 支持 Option 配置（超时、代理、连接池、日志、trace、慢请求等）
- 支持自定义钩子（Hook）、慢请求日志、错误日志、trace
- 支持全局默认客户端与实例化客户端
- 支持 HTTPS 证书有效期检测
- 并发安全，适合高并发环境
- 完整单元测试覆盖

### 设计理念

http 包遵循"高效、灵活、可观测"的设计理念，接口与实现分离，Option 配置灵活，钩子机制可扩展，适合微服务和高并发场景。默认实现兼容标准库 http.Client，便于迁移和集成。

## 安装

### 前置条件

- Go 版本要求：Go 1.18+
- 依赖要求：
  - github.com/stretchr/testify（仅测试）

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/net/http
```

## 快速开始

### 基础用法

```go
package main

import (
    "context"
    "fmt"
    kithttp "github.com/fsyyft-go/kit/net/http"
)

func main() {
    client := kithttp.NewClient(
        kithttp.WithTimeout(5 * time.Second),
    )
    resp, err := client.Get(context.Background(), "https://httpbin.org/get")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    fmt.Println("状态码:", resp.StatusCode)
}
```

### 全局方法

```go
resp, err := kithttp.Get(context.Background(), "https://httpbin.org/get")
```

### POST/表单/JSON

```go
// POST 普通请求
client.Post(ctx, url, body)
// POST 表单
client.PostForm(ctx, url, url.Values{"a": {"1"}})
// POST JSON
client.PostJSON(ctx, url, map[string]any{"x": 1})
```

### 钩子与慢请求日志

```go
client := kithttp.NewClient(
    kithttp.WithLogSlow(100*time.Millisecond),
    kithttp.WithTraceEnable(true),
)
```

### 证书有效期检测

```go
days, err := kithttp.GetCertificatesExpirestime(ctx, "https://example.com", "", "", 3*time.Second)
fmt.Println("证书剩余天数:", days)
```

## 详细指南

### 核心概念

- **Client 接口**：统一封装 http.Client，支持常用请求方法
- **Option 配置**：灵活设置超时、代理、连接池、日志、trace 等
- **钩子机制**：支持请求前后自定义扩展（如 trace、慢日志、错误日志）
- **全局方法**：便捷调用全局默认客户端
- **证书检测**：支持 HTTPS 证书剩余天数检测

### 常见用例

- 微服务间 HTTP 通信
- API 调用与数据采集
- 健康检查与监控
- 自动化测试
- 证书有效期监控

### 最佳实践

- 合理设置超时与连接池参数，避免资源泄漏
- 慢请求日志与 trace 有助于排查性能瓶颈
- 钩子机制可扩展自定义监控与埋点
- 始终检查 error 并关闭 resp.Body

## API 文档

### 主要类型

```go
// Client HTTP 客户端接口
 type Client interface {
    Do(ctx context.Context, req *http.Request) (*http.Response, error)
    Head(ctx context.Context, url string) (*http.Response, error)
    Get(ctx context.Context, url string) (*http.Response, error)
    Post(ctx context.Context, url string, body io.Reader) (*http.Response, error)
    PostForm(ctx context.Context, url string, data url.Values) (*http.Response, error)
    PostJSON(ctx context.Context, url string, data any) (*http.Response, error)
}

// Option 配置项类型
 type Option func(*client)

// NewClient 创建客户端
func NewClient(opts ...Option) Client

// 全局方法
func Get(ctx context.Context, url string) (*http.Response, error)
func Post(ctx context.Context, url string, body io.Reader) (*http.Response, error)
...

// 证书检测
func GetCertificatesExpirestime(ctx context.Context, requestURL, method, address string, timeout time.Duration) (int, error)
```

### 关键函数

- `NewClient`：创建 HTTP 客户端，支持 Option 配置
- `Do/Get/Post/Head/PostForm/PostJSON`：常用请求方法
- `WithTimeout/WithProxy/WithLogSlow/WithTraceEnable/WithLogger`：常用配置项
- `GetCertificatesExpirestime`：证书剩余天数检测

## 错误处理

- 所有请求方法均返回 error，需检查
- 超时、网络、协议等错误均有详细信息
- Option 配置错误会 panic 或返回 error

## 性能指标

- 单连接 QPS 万级，连接池/并发可进一步提升
- 钩子/trace/日志功能对性能影响可控

## 测试覆盖率

- 单元测试覆盖所有接口、边界、异常场景
- 使用 testify，覆盖率 100%

## 调试指南

- 检查超时、代理、连接池等参数设置
- trace/慢日志有助于定位慢请求
- 证书检测适合自动化监控

## 相关文档

- [Go net/http 官方文档](https://pkg.go.dev/net/http)
- [httptrace 官方文档](https://pkg.go.dev/net/http/httptrace)

## 贡献指南

欢迎提交 Issue、PR 或建议，详见 [贡献指南](../../CONTRIBUTING.md)。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../../LICENSE)。 