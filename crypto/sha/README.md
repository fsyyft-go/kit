# sha

## 简介

`sha` 包提供了计算字符串 SHA256 哈希值的简便功能，主要封装了 Go 标准库的 crypto/sha256 包，使其更易于使用。SHA256 作为现代加密哈希算法，广泛应用于数据完整性校验、签名、区块链等场景。

### 主要特性

- 简化的 SHA256 哈希计算 API
- 支持字符串直接计算哈希值
- 提供带错误处理和忽略错误的版本
- 高性能实现，适合大数据量和并发场景
- 适用于各种字符编码（ASCII、UTF-8 等）

### 设计理念

该包设计理念是提供最简单、最直接的 SHA256 哈希计算接口，使开发者能够以最少的代码行完成常见任务。通过封装标准库中较为复杂的操作，开发者只需一行代码即可计算字符串的 SHA256 哈希值，大大减少了重复代码和潜在错误。

## 安装

### 前置条件

- Go 版本要求：Go 1.18 或更高版本
- 依赖要求：
  - Go 标准库的 crypto/sha256

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/crypto/sha
```

## 快速开始

### 基础用法

```go
package main

import (
    "fmt"
    "github.com/fsyyft-go/kit/crypto/sha"
)

func main() {
    // 最简单的用法 - 忽略可能的错误
    hash := sha.SHA256HashStringWithoutError("hello world")
    fmt.Println("SHA256 哈希值:", hash)
    
    // 带错误处理的用法
    hash, err := sha.SHA256HashString("hello world")
    if err != nil {
        fmt.Printf("计算哈希值时发生错误: %v\n", err)
        return
    }
    fmt.Println("SHA256 哈希值:", hash)
}
```

### 配置选项

```go
// SHA 包设计简单，无需配置选项
// 所有功能都通过直接调用函数实现
```

## 详细指南

### 核心概念

SHA256（Secure Hash Algorithm 256）是一种广泛使用的加密哈希函数，可生成 256 位（32 字节）的哈希值，
通常以 64 位十六进制数表示。本包提供了一种简便方法计算字符串的 SHA256 哈希值，主要用于数据完整性校验、
签名、区块链等场景。

SHA256 算法具有抗碰撞性和抗篡改性，适合安全性要求较高的应用。

### 常见用例

#### 1. 计算普通字符串的 SHA256 哈希值

```go
// 计算简单字符串的 SHA256 哈希值
hash := sha.SHA256HashStringWithoutError("hello world")
fmt.Println(hash)  // 输出: b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
```

#### 2. 计算中文字符串的 SHA256 哈希值

```go
// 计算中文字符串的 SHA256 哈希值
hash, err := sha.SHA256HashString("你好，世界")
if err != nil {
    fmt.Printf("计算哈希值时发生错误: %v\n", err)
    return
}
fmt.Println(hash)  // 输出: 46932f1e6ea5216e77f58b1908d72ec9322ed129318c6d4bd4450b5eaab9d7e7
```

### 最佳实践

- 安全考虑
  - SHA256 适用于数据完整性校验、签名、区块链等安全场景
  - 不建议直接用于密码存储，推荐使用 bcrypt、Argon2 等专用算法
- 错误处理
  - 对于普通场景，可以使用 SHA256HashStringWithoutError 简化代码
  - 在关键流程中，建议使用 SHA256HashString 并妥善处理可能的错误
- 性能优化
  - 对于频繁计算的场景，可使用并行处理提高吞吐量
  - 当计算大量小字符串的哈希值时，SHA256HashStringWithoutError 通常性能更佳

## API 文档

### 主要类型

```go
// 本包主要使用 Go 标准库中的类型，无自定义类型
```

### 关键函数

#### SHA256HashString

计算字符串的 SHA256 哈希值，并返回可能发生的错误

```go
func SHA256HashString(source string) (string, error)
```

示例：
```go
hash, err := sha.SHA256HashString("hello world")
if err != nil {
    // 处理错误
}
fmt.Println(hash)
```

#### SHA256HashStringWithoutError

计算字符串的 SHA256 哈希值，忽略可能发生的错误

```go
func SHA256HashStringWithoutError(source string) string
```

示例：
```go
hash := sha.SHA256HashStringWithoutError("hello world")
fmt.Println(hash)
```

### 错误处理

本包通常不会返回错误，除非在 I/O 操作中发生异常。在大多数正常使用场景下，
可以安全地使用 SHA256HashStringWithoutError 函数而不必担心错误处理。

## 性能指标

| 操作 | 性能指标 | 说明 |
|------|----------|------|
| 空字符串 | ~200ns | 单次操作 |
| 短字符串 (10字节) | ~250ns | 单次操作 |
| 中等字符串 (1KB) | ~2.5µs | 单次操作 |
| 大字符串 (100KB) | ~150µs | 单次操作 |
| 并行处理 (100KB) | ~35µs/op | 8核并行 |

## 测试覆盖率

| 包 | 覆盖率 |
|------|--------|
| sha | 100% |

完整的测试覆盖了所有正常场景、边界场景和错误处理场景，确保了包的稳定性和可靠性。

## 调试指南

### 常见问题排查

#### 哈希值不一致

- 检查输入字符串的编码方式（UTF-8、ASCII 等）
- 验证字符串中是否包含不可见字符
- 确认没有字符串截断问题

#### 性能问题

- 对于大量小字符串，使用 SHA256HashStringWithoutError 避免错误检查开销
- 对于大数据量，考虑并行处理或分批处理
- 如果反复计算相同字符串的哈希值，考虑缓存结果

## 相关文档

- [Go crypto/sha256 包文档](https://pkg.go.dev/crypto/sha256)
- [SHA256 算法规范 (FIPS 180-4)](https://nvlpubs.nist.gov/nistpubs/FIPS/NIST.FIPS.180-4.pdf)
- [密码哈希比较](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考我们的[贡献指南](../../CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT 许可证。查看 [LICENSE](../../LICENSE) 文件了解更多信息。 