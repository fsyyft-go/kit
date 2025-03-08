# Bytes 包

`bytes` 包提供了字节操作相关的工具函数，用于处理二进制数据和随机字节生成。

## 特性

- 安全的随机字节生成
- 基于加密安全的随机数生成器
- 简单易用的API
- 适用于各种安全场景（如生成nonce、salt等）

## 快速开始

### 生成随机字节

```go
package main

import (
    "encoding/hex"
    "fmt"
    
    "github.com/fsyyft-go/kit/bytes"
)

func main() {
    // 生成16字节的随机数据
    nonce, err := bytes.GenerateNonce(16)
    if err != nil {
        panic(err)
    }
    
    // 输出十六进制表示
    fmt.Printf("生成的随机字节: %s\n", hex.EncodeToString(nonce))
}
```

## API 说明

### GenerateNonce

```go
func GenerateNonce(length int) ([]byte, error)
```

生成指定长度的随机字节切片。

参数：
- `length`：指定生成随机字节的长度。

返回：
- `[]byte`：生成的随机字节切片。
- `error`：如果生成过程中出现错误，则返回相应的错误。

示例：
```go
// 生成32字节的随机数据
token, err := bytes.GenerateNonce(32)
if err != nil {
    // 处理错误
}
```

注意：
- 长度参数不能为负数，否则会返回错误。
- 生成的随机字节使用加密安全的随机数生成器，适合用于安全关键场景。

## 常见用途

1. 生成加密操作的nonce（随机数）
2. 生成盐值（salt）用于密码哈希
3. 生成会话令牌
4. 生成临时密钥
5. 生成随机ID或UUID的基础

## 最佳实践

1. 安全考虑
   - 使用本包生成的随机字节用于安全场景，如密码学应用
   - 对于密码学应用，建议使用足够长的随机字节（如16字节或更长）

2. 错误处理
   - 始终检查返回的错误值
   - 随机数生成失败可能表示系统熵不足，应当妥善处理

3. 性能考虑
   - 加密安全的随机数生成比伪随机数生成慢，在性能关键场景中应谨慎使用
   - 对于高频调用，考虑生成一次较长的随机字节然后分段使用

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../LICENSE) 文件。 