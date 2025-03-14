# AES 包

`aes` 包提供了 AES 加密和解密的实用函数，支持 GCM 模式并可处理多种数据格式。

## 特性

- GCM 模式的 AES 加密/解密
- 支持多种输入格式（字节数组、字符串、Base64、Hex）
- 自动随机 nonce 生成
- 线程安全
- 完整的错误处理
- 简洁易用的 API

## 快速开始

### 基本使用

```go
package main

import (
    "encoding/base64"
    "fmt"
    "github.com/fsyyft-go/kit/crypto/aes"
)

func main() {
    // 准备测试数据
    key := []byte("01234567890123456789012345678901") // 32字节密钥适用于AES-256
    plaintext := []byte("Hello, World! This is a test.")
    
    // 使用随机生成的 nonce 加密（推荐方式）
    encrypted, err := aes.EncryptGCMNonceLength(key, 12, plaintext)
    if err != nil {
        panic(err)
    }
    
    // 解密数据
    nonce, decrypted, err := aes.DecryptGCMNonceLength(key, 12, encrypted)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("解密后的数据: %s\n", string(decrypted))
    fmt.Printf("使用的 nonce 长度: %d\n", len(nonce))
}
```

### 字符串加密与解密（Base64 格式）

```go
// 使用 Base64 编码的密钥和字符串明文进行加密
keyBase64 := "MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDE="
plaintext := "Hello, World!"
nonceLength := 12

encryptedBase64, err := aes.EncryptStringGCMBase64(keyBase64, nonceLength, plaintext)
if err != nil {
    panic(err)
}

// 解密 Base64 编码的密文
nonce, decrypted, err := aes.DecryptStringGCMBase64(keyBase64, nonceLength, encryptedBase64)
if err != nil {
    panic(err)
}

fmt.Printf("解密后的数据: %s\n", decrypted)
```

### 十六进制格式操作

```go
// 使用十六进制编码的密钥和明文进行加密
keyHex := "3031323334353637383930313233343536373839303132333435363738393031"
dataHex := "48656c6c6f2c20576f726c6421" // "Hello, World!" 的十六进制表示

encryptedHex, err := aes.EncryptGCMHex(keyHex, 12, dataHex)
if err != nil {
    panic(err)
}

// 解密十六进制编码的密文
nonce, decrypted, err := aes.DecryptGCMHex(keyHex, 12, encryptedHex)
if err != nil {
    panic(err)
}

fmt.Printf("解密后的数据 (十六进制): %s\n", decrypted)
```

## 函数说明

### 加密函数

- `EncryptStringGCMBase64`: 使用 Base64 格式的密钥加密 UTF-8 字符串，返回 Base64 格式的密文
- `EncryptStringGCMHex`: 使用十六进制格式的密钥加密 UTF-8 字符串，返回十六进制格式的密文
- `EncryptGCMBase64`: 使用 Base64 格式的密钥加密 Base64 格式的明文，返回 Base64 格式的密文
- `EncryptGCMHex`: 使用十六进制格式的密钥加密十六进制格式的明文，返回十六进制格式的密文
- `EncryptGCMNonceLength`: 使用指定长度的随机生成 nonce 进行加密
- `EncryptGCM`: 使用指定的密钥和 nonce 进行加密的底层函数

### 解密函数

- `DecryptStringGCMBase64`: 使用 Base64 格式的密钥解密 Base64 格式的密文，返回 UTF-8 字符串
- `DecryptStringGCMHex`: 使用十六进制格式的密钥解密十六进制格式的密文，返回 UTF-8 字符串
- `DecryptGCMBase64`: 使用 Base64 格式的密钥解密 Base64 格式的密文，返回 Base64 格式的明文
- `DecryptGCMHex`: 使用十六进制格式的密钥解密十六进制格式的密文，返回十六进制格式的明文
- `DecryptGCMNonceLength`: 从密文中提取指定长度的 nonce 并进行解密
- `DecryptGCM`: 使用指定的密钥和 nonce 进行解密的底层函数

## 最佳实践

1. 密钥管理
   - 使用安全的方式存储和传输密钥
   - 密钥长度应为 16、24 或 32 字节（对应 AES-128、AES-192、AES-256）
   - 避免硬编码密钥在源代码中

2. Nonce 使用
   - 推荐使用 12 字节长度的 nonce（GCM 模式的推荐值）
   - 永远不要为相同的密钥重用相同的 nonce
   - 默认情况下，nonce 会自动随机生成并附加到密文前面

3. 错误处理
   - 总是检查加密和解密函数返回的错误
   - 处理因密钥格式不正确或长度无效可能引发的错误
   - 注意处理解密失败（数据被篡改或使用错误的密钥）的情况

4. 性能考虑
   - 对于频繁的加密/解密操作，考虑重用相同的密钥以减少解析开销
   - 对于大型数据，请考虑分块处理以减少内存使用

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../../LICENSE) 文件。 