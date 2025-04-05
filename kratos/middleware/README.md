# Kratos Middleware

## 简介

`kratos/middleware` 包提供了一组强大的中间件实现，用于扩展 Kratos 框架的功能。目前包含两个核心中间件：验证中间件（validate）和基本认证中间件（basicauth）。这些中间件旨在简化常见的 Web 服务功能实现，提供可靠的请求验证和认证机制。

### 主要特性

#### 验证中间件 (validate)
- 支持对实现了 validator 接口的请求进行自动验证
- 提供灵活的验证失败回调机制
- 支持自定义验证错误处理
- 与 Kratos 验证系统无缝集成
- 完整的测试覆盖

#### 基本认证中间件 (basicauth)
- 标准的 HTTP Basic Authentication 实现
- 支持自定义认证验证器
- 可配置的认证域（realm）设置
- 完整的错误处理机制
- 安全的认证头解析

### 设计理念

本包的设计遵循以下原则：

- **可扩展性**：支持自定义验证器和认证逻辑
- **安全性**：遵循 HTTP 认证最佳实践
- **易用性**：简单的 API 设计，开箱即用
- **可靠性**：完整的错误处理和边界检查
- **可测试性**：高测试覆盖率和完整的测试用例

## 安装

### 前置条件

- Go 版本要求：>= 1.16
- 依赖要求：
  - github.com/go-kratos/kratos/v2
  - github.com/stretchr/testify (用于测试)

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/kratos/middleware
```

## 快速开始

### 验证中间件

```go
import (
    "github.com/fsyyft-go/kit/kratos/middleware/validate"
)

// 使用默认验证处理
srv.Use(validate.Validator())

// 使用自定义验证失败回调
srv.Use(validate.Validator(
    validate.WithValidateCallback(func(ctx context.Context, req interface{}, err error) (interface{}, error) {
        return nil, errors.BadRequest("CUSTOM_ERROR", err.Error())
    }),
))
```

### 基本认证中间件

```go
import (
    "github.com/fsyyft-go/kit/kratos/middleware/basicauth"
)

// 使用默认验证器
srv.Use(basicauth.Server())

// 使用自定义验证器和认证域
srv.Use(basicauth.Server(
    basicauth.WithValidator(func(ctx context.Context, username, password string) bool {
        return username == "admin" && password == "secret"
    }),
    basicauth.WithRealm("My Service"),
))
```

## 详细指南

### 验证中间件

#### 1. 实现 validator 接口

```go
type MyRequest struct {
    Name string `validate:"required"`
    Age  int    `validate:"gte=0,lte=120"`
}

func (r *MyRequest) Validate() error {
    if r.Name == "" {
        return errors.New("name is required")
    }
    if r.Age < 0 || r.Age > 120 {
        return errors.New("age must be between 0 and 120")
    }
    return nil
}
```

#### 2. 自定义验证回调

```go
validate.Validator(
    validate.WithValidateCallback(func(ctx context.Context, req interface{}, err error) (interface{}, error) {
        // 自定义错误响应
        return nil, errors.BadRequest("VALIDATION_ERROR", err.Error())
    }),
)
```

### 基本认证中间件

#### 1. 自定义验证器

```go
basicauth.Server(
    basicauth.WithValidator(func(ctx context.Context, username, password string) bool {
        // 实现数据库验证
        return db.ValidateUser(username, password)
    }),
)
```

#### 2. 设置认证域

```go
basicauth.Server(
    basicauth.WithRealm("Restricted Area"),
)
```

### 最佳实践

#### 验证中间件
- 为复杂的请求结构体实现 validator 接口
- 在验证回调中提供详细的错误信息
- 避免在验证逻辑中执行耗时操作
- 合理组织验证规则

#### 基本认证中间件
- 使用 HTTPS 保护认证信息
- 实现安全的密码验证逻辑
- 设置有意义的认证域名称
- 注意错误处理和安全日志

## API 文档

### 验证中间件

```go
// 验证器接口
type validator interface {
    Validate() error
}

// 验证回调函数
type ValidateCallback func(ctx context.Context, req interface{}, err error) (interface{}, error)

// 创建验证器中间件
func Validator(opts ...Option) middleware.Middleware

// 设置验证回调
func WithValidateCallback(fn ValidateCallback) Option
```

### 基本认证中间件

```go
// 认证验证器
type CredentialValidator func(ctx context.Context, username, password string) bool

// 创建基本认证中间件
func Server(opts ...Option) middleware.Middleware

// 设置验证器
func WithValidator(validator CredentialValidator) Option

// 设置认证域
func WithRealm(realm string) Option
```

## 性能指标

| 操作 | 性能指标 | 说明 |
|------|----------|------|
| 验证请求 | O(1) | 单个请求的验证时间 |
| 基本认证 | O(1) | 认证头解析和验证时间 |
| 中间件链执行 | O(n) | n 为中间件数量 |

## 测试覆盖率

| 包 | 覆盖率 |
|------|--------|
| middleware/validate | >95% |
| middleware/basicauth | >95% |

## 调试指南

### 常见问题排查

#### 1. 验证失败

- 检查请求结构体是否正确实现了 validator 接口
- 验证错误信息是否足够详细
- 确认验证规则的合理性

#### 2. 认证失败

- 检查认证头格式是否正确
- 验证用户名和密码是否正确编码
- 确认验证器实现是否正确

## 相关文档

- [Kratos 中间件文档](https://go-kratos.dev/docs/component/middleware/)
- [HTTP Basic Auth 规范](https://tools.ietf.org/html/rfc7617)

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

本文的大部分信息，由 AI 使用[模板](../../ai/templates/docs/package_readme_template.md)根据[提示词](../../ai/prompts/docs/package_readme_generator.md)自动生成，如有任何问题，请随时联系我。