# Middleware 包

`middleware` 包提供了两个基于 Kratos 的中间件实现：验证中间件（validate）和基本认证中间件（basicauth）。

## 特性

### 验证中间件 (validate)
- 支持对实现了 validator 接口的请求进行验证
- 提供验证失败时的自定义回调处理
- 支持通过 WithValidateCallback 选项设置自定义回调

### 基本认证中间件 (basicauth)
- 提供标准的 HTTP Basic Authentication 实现
- 支持自定义验证器
- 支持设置认证域（realm）
- 完整的错误处理机制

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
        // 自定义验证失败处理逻辑
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

## 验证中间件详细说明

### 请求验证

验证中间件会检查请求对象是否实现了 validator 接口：

```go
type validator interface {
    Validate() error
}
```

如果请求对象实现了该接口，中间件会调用其 Validate() 方法进行验证。

### 自定义验证回调

可以通过 WithValidateCallback 选项设置验证失败时的处理逻辑：

```go
type ValidateCallback func(ctx context.Context, req interface{}, err error) (interface{}, error)

// 示例：自定义错误响应
validate.Validator(
    validate.WithValidateCallback(func(ctx context.Context, req interface{}, err error) (interface{}, error) {
        return nil, errors.BadRequest("VALIDATION_ERROR", err.Error())
    }),
)
```

## 基本认证中间件详细说明

### 认证头格式

基本认证中间件期望请求头中包含标准的 Basic Authentication 头：

```
Authorization: Basic base64(username:password)
```

### 自定义验证器

可以通过 WithValidator 选项设置自定义的认证逻辑：

```go
type Validator func(ctx context.Context, username, password string) bool

// 示例：使用数据库验证用户
basicauth.Server(
    basicauth.WithValidator(func(ctx context.Context, username, password string) bool {
        return db.ValidateUser(username, password)
    }),
)
```

### 设置认证域

可以通过 WithRealm 选项设置认证域，这会影响浏览器显示的认证提示：

```go
basicauth.Server(
    basicauth.WithRealm("Restricted Area"),
)
```

## 最佳实践

1. 验证中间件
   - 为复杂的请求结构体实现 validator 接口
   - 在验证回调中提供有意义的错误信息
   - 避免在验证逻辑中执行耗时操作

2. 基本认证中间件
   - 使用 HTTPS 保护认证信息
   - 实现合适的密码验证逻辑
   - 设置有意义的认证域名称
   - 注意错误处理和安全日志

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../../LICENSE) 文件。 