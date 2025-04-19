# md5

## 简介

`md5` 包提供了计算字符串 MD5 哈希值的简便功能，主要封装了 Go 标准库的 crypto/md5 包，使其更易于使用。虽然 MD5 在安全敏感场景下不再推荐使用，但在数据完整性校验、缓存键生成等非安全关键场景中仍然有广泛应用。

### 主要特性

- 简化的 MD5 哈希计算 API
- 支持字符串直接计算哈希值
- 提供带错误处理和忽略错误的版本
- 高性能实现
- 适用于各种字符编码（ASCII、UTF-8 等）

### 设计理念

该包的设计理念是提供最简单、最直接的 MD5 哈希计算接口，使开发者能够以最少的代码行完成常见任务。通过封装标准库中较为复杂的操作，开发者只需一行代码即可计算字符串的 MD5 哈希值，大大减少了重复代码和潜在错误。

## 安装

### 前置条件

- Go 版本要求：Go 1.18 或更高版本
- 依赖要求：
  - Go 标准库的 crypto/md5

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/crypto/md5
```

## 快速开始

### 基础用法

```go
package main

import (
    "fmt"
    "github.com/fsyyft-go/kit/crypto/md5"
)

func main() {
    // 最简单的用法 - 忽略可能的错误
    hash := md5.HashStringWithoutError("hello world")
    fmt.Println("MD5 哈希值:", hash)
    
    // 带错误处理的用法
    hash, err := md5.HashString("hello world")
    if err != nil {
        fmt.Printf("计算哈希值时发生错误: %v\n", err)
        return
    }
    fmt.Println("MD5 哈希值:", hash)
}
```

### 配置选项

```go
// MD5 包设计简单，无需配置选项
// 所有功能都通过直接调用函数实现
```

## 详细指南

### 核心概念

MD5（Message Digest Algorithm 5）是一种广泛使用的密码散列函数，可生成 128 位（16 字节）的散列值（哈希值），
通常以 32 位十六进制数表示。本包提供了一种简便方法计算字符串的 MD5 哈希值，主要用于数据完整性校验、
缓存键生成等非安全场景。

虽然 MD5 算法已不再被认为是密码学安全的（存在碰撞攻击风险），但由于其速度快、实现简单，
在不涉及安全性的场景中仍然被广泛使用。

### 常见用例

#### 1. 计算普通字符串的 MD5 哈希值

```go
// 计算简单字符串的 MD5 哈希值
hash := md5.HashStringWithoutError("hello world")
fmt.Println(hash)  // 输出: 5eb63bbbe01eeed093cb22bb8f5acdc3
```

#### 2. 计算中文字符串的 MD5 哈希值

```go
// 计算中文字符串的 MD5 哈希值
hash, err := md5.HashString("你好，世界")
if err != nil {
    fmt.Printf("计算哈希值时发生错误: %v\n", err)
    return
}
fmt.Println(hash)  // 输出: dbefd3ada018615b35588a01e216ae6e
```

### 最佳实践

- 安全考虑
  - 不要在密码存储等安全要求较高的场景中使用 MD5
  - 对于密码哈希，推荐使用 bcrypt 或 Argon2 等专用算法
  - MD5 主要适用于数据完整性校验、缓存键生成等非安全关键场景

- 错误处理
  - 对于普通场景，可以使用 HashStringWithoutError 简化代码
  - 在关键流程中，建议使用 HashString 并妥善处理可能的错误

- 性能优化
  - 对于频繁计算的场景，可以考虑使用对象池减少内存分配
  - 对于大量数据，可以使用并行处理提高吞吐量
  - 当计算大量小字符串的哈希值时，HashStringWithoutError 通常性能更佳

## API 文档

### 主要类型

```go
// 本包主要使用 Go 标准库中的类型，无自定义类型
```

### 关键函数

#### HashString

计算字符串的 MD5 哈希值，并返回可能发生的错误

```go
func HashString(source string) (string, error)
```

示例：
```go
hash, err := md5.HashString("hello world")
if err != nil {
    // 处理错误
}
fmt.Println(hash)
```

#### HashStringWithoutError

计算字符串的 MD5 哈希值，忽略可能发生的错误

```go
func HashStringWithoutError(source string) string
```

示例：
```go
hash := md5.HashStringWithoutError("hello world")
fmt.Println(hash)
```

### 错误处理

本包通常不会返回错误，除非在 I/O 操作中发生异常。在大多数正常使用场景下，
可以安全地使用 HashStringWithoutError 函数而不必担心错误处理。

## 性能指标

| 操作 | 性能指标 | 说明 |
|------|----------|------|
| 空字符串 | ~100ns | 单次操作 |
| 短字符串 (10字节) | ~150ns | 单次操作 |
| 中等字符串 (1KB) | ~1.5µs | 单次操作 |
| 大字符串 (100KB) | ~100µs | 单次操作 |
| 并行处理 (100KB) | ~25µs/op | 8核并行 |

## 测试覆盖率

| 包 | 覆盖率 |
|------|--------|
| md5 | 96% |

## 调试指南

### 常见问题排查

#### 哈希值不一致

- 检查输入字符串的编码方式（UTF-8、ASCII 等）
- 验证字符串中是否包含不可见字符
- 确认没有字符串截断问题

#### 性能问题

- 对于大量小字符串，使用 HashStringWithoutError 避免错误检查开销
- 对于大数据量，考虑并行处理或分批处理
- 如果反复计算相同字符串的哈希值，考虑缓存结果

## 相关文档

- [Go crypto/md5 包文档](https://pkg.go.dev/crypto/md5)
- [MD5 算法规范 (RFC 1321)](https://tools.ietf.org/html/rfc1321)
- [MD5 安全考虑 (RFC 6151)](https://tools.ietf.org/html/rfc6151)
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