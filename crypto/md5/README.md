# MD5 包

`md5` 包提供了计算字符串 MD5 哈希值的简便功能，支持错误处理和无错误版本。

## 特性

- 简单易用的字符串 MD5 哈希计算
- 支持带错误处理的哈希计算
- 支持忽略错误的简化版函数
- 轻量级设计，无外部依赖
- 使用标准库 crypto/md5 实现，安全可靠

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/fsyyft-go/kit/crypto/md5"
)

func main() {
    // 计算字符串的 MD5 哈希值，忽略可能的错误
    hash := md5.HashStringWithoutError("hello world")
    fmt.Println(hash) // 输出: 5eb63bbbe01eeed093cb22bb8f5acdc3
    
    // 计算字符串的 MD5 哈希值，处理可能的错误
    hashWithError, err := md5.HashString("hello world")
    if err != nil {
        fmt.Printf("计算哈希值时发生错误: %v\n", err)
        return
    }
    fmt.Println(hashWithError) // 输出: 5eb63bbbe01eeed093cb22bb8f5acdc3
}
```

### 错误处理

在实际应用中，字符串计算 MD5 几乎不可能失败。如果您不需要处理错误，可以使用 `HashStringWithoutError` 函数：

```go
// 忽略错误的简化版
hash := md5.HashStringWithoutError("hello world")
```

如果您需要处理可能的错误，可以使用 `HashString` 函数：

```go
// 带错误处理的完整版
hash, err := md5.HashString("hello world")
if err != nil {
    // 处理错误
    return
}
```

### 使用场景

MD5 适用于以下场景：

- 数据完整性校验
- 快速内容标识
- 简单的文件指纹
- 缓存键生成

**注意**：MD5 不适用于安全敏感场景，如密码存储、数字签名等。请使用更安全的哈希算法（如 SHA-256）或专门的密码哈希函数（如 bcrypt）。

## API 说明

### HashString

```go
func HashString(source string) (string, error)
```

计算字符串的 MD5 哈希值，并返回可能发生的错误。

参数：
- `source`：需要计算哈希值的源字符串。

返回值：
- `string`：计算得到的 MD5 哈希值的十六进制字符串表示。
- `error`：操作过程中可能发生的错误。

### HashStringWithoutError

```go
func HashStringWithoutError(source string) string
```

计算字符串的 MD5 哈希值，忽略可能发生的错误。该函数是 HashString 的简化版本，适用于确定不会发生错误的场景。

参数：
- `source`：需要计算哈希值的源字符串。

返回值：
- `string`：计算得到的 MD5 哈希值的十六进制字符串表示。

## 最佳实践

1. 选择合适的函数
   - 对于一般场景，使用 `HashStringWithoutError`
   - 当需要严格错误处理时，使用 `HashString`

2. 性能考虑
   - MD5 是相对轻量级的哈希算法
   - 对于大量小字符串的哈希计算，性能表现良好
   - 对于超大字符串（MB级别），请注意内存使用

3. 安全考虑
   - 不要使用 MD5 存储密码或其他敏感信息
   - 不要依赖 MD5 进行安全验证
   - 对于安全敏感场景，请使用更强的哈希算法

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../../LICENSE) 文件。 