# OTP 包

`otp` 包提供了基于时间的一次性密码（TOTP）算法的实现，支持各种常见哈希算法和自定义选项。

## 特性

- 符合 RFC 6238 标准的 TOTP 算法实现
- 支持多种哈希算法（SHA1、SHA256、SHA512）
- 支持自定义密码长度（位数）
- 支持自定义有效期（秒）
- 支持时间窗口设置
- 支持生成兼容的二维码 URL
- 支持发行者和标签自定义
- 线程安全

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/fsyyft-go/kit/crypto/otp"
)

func main() {
    // 使用 Base32 编码的密钥种子
    secretKeyBase32 := "JBSWY3DPEHPK3PXP"
    
    // 创建 OTP 实例
    oneTimePassword, err := otp.NewOneTimePassword(secretKeyBase32)
    if err != nil {
        panic(err)
    }
    
    // 生成当前密码
    password, err := oneTimePassword.Password()
    if err != nil {
        panic(err)
    }
    fmt.Printf("当前密码: %s\n", password)
    
    // 验证密码
    isValid := oneTimePassword.VeryfyPassword(password)
    fmt.Printf("密码验证: %v\n", isValid)
    
    // 生成 TOTP URL（用于二维码）
    url := oneTimePassword.GenerateURL()
    fmt.Printf("TOTP URL: %s\n", url)
}
```

### 使用功能选项

```go
// 创建带自定义选项的 OTP 实例
oneTimePassword, err := otp.NewOneTimePassword(secretKeyBase32,
    otp.WithSHA256(),                // 使用 SHA256 哈希算法
    otp.WithDigits(8),               // 生成 8 位密码
    otp.WithPeriodSeconds(60),       // 密码有效期为 60 秒
    otp.WithWindowSize(5),           // 验证时间窗口为 5
    otp.WithIssuer("MyCompany"),     // 设置发行者
    otp.WithLabel("user@example.com"), // 设置标签
)
```

### 快捷方法

```go
// 直接验证密码，无需创建实例
isValid := otp.VeryfyPassword(secretKeyBase32, "123456")

// 直接生成 TOTP URL，无需创建实例
url := otp.GenerateURL(secretKeyBase32, 
    otp.WithIssuer("MyCompany"),
    otp.WithLabel("user@example.com"))
```

## 配置选项

### 哈希算法

```go
// 使用 SHA256 算法（默认为 SHA1）
otp.WithSHA256()

// 使用 SHA512 算法
otp.WithSHA512()
```

### 密码长度

```go
// 设置密码长度为 8 位（默认为 6 位）
otp.WithDigits(8)
```

### 有效期

```go
// 设置密码有效期为 60 秒（默认为 30 秒）
otp.WithPeriodSeconds(60)
```

### 时间窗口

```go
// 设置验证时间窗口为 5（默认为 10）
// 这意味着验证时会检查前后 5 个时间周期的密码
otp.WithWindowSize(5)
```

### URL 配置

```go
// 设置发行者（会显示在验证器应用中）
otp.WithIssuer("MyCompany")

// 设置标签（通常是用户标识，如邮箱）
otp.WithLabel("user@example.com")
```

## 最佳实践

1. 密钥管理
   - 安全存储密钥，避免泄露
   - 建议使用加密方式存储用户密钥
   - 确保密钥的 Base32 格式正确（无填充）

2. 算法选择
   - SHA1 是最广泛支持的算法，兼容性最好
   - SHA256/SHA512 提供更高的安全性，但部分验证器可能不支持

3. 密码长度
   - 6 位是标准长度，大多数验证器支持
   - 增加位数可以提高安全性，但需验证器支持

4. 时间窗口
   - 较大的时间窗口提高用户体验但降低安全性
   - 较小的时间窗口提高安全性但可能导致验证失败

5. 兼容性考虑
   - 使用标准参数确保与主流验证器应用兼容
   - 测试生成的 URL 是否可被常见验证器识别

## 应用场景

- 双因素认证（2FA）
- 安全登录系统
- API 访问控制
- 一次性验证码
- 临时授权码生成

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../../LICENSE) 文件。 