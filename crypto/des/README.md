# DES 包

`des` 包提供了 DES 加密算法的实现，支持 CBC 模式和 PKCS7 填充。

## 特性

- 支持 DES-CBC 加密/解密
- 支持 PKCS7 填充/去填充
- 支持多种输入格式（字节数组、字符串、16 进制字符串）
- 支持独立设置 IV（初始化向量）
- 线程安全
- 完整的错误处理

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/fsyyft-go/kit/crypto/des"
)

func main() {
    // 使用字符串密钥和数据
    key := "12345678"           // 8 字节密钥
    data := "Hello, World!"     // 待加密数据
    
    // 加密
    encrypted, err := des.EncryptStringCBCPkCS7PaddingStringHex(key, data)
    if err != nil {
        panic(err)
    }
    fmt.Printf("加密结果（16 进制）：%s\n", encrypted)
    
    // 解密
    decrypted, err := des.DecryptStringCBCPkCS7PaddingStringHex(key, encrypted)
    if err != nil {
        panic(err)
    }
    fmt.Printf("解密结果：%s\n", decrypted)
}
```

### 使用 16 进制密钥

```go
// 使用 16 进制字符串表示的密钥
keyHex := "1234567890ABCDEF"   // 16 进制表示的 8 字节密钥
data := "Hello, World!"

// 加密
encrypted, err := des.EncryptStringCBCPkCS7PaddingHex(keyHex, data)
if err != nil {
    panic(err)
}

// 解密
decrypted, err := des.DecryptStringCBCPkCS7PaddingHex(keyHex, encrypted)
if err != nil {
    panic(err)
}
```

### 使用字节数组和独立 IV

```go
key := []byte("12345678")      // 8 字节密钥
iv := []byte("87654321")       // 8 字节 IV
data := []byte("Hello, World!")

// 加密
encrypted, err := des.EncryptCBCPkCS7PaddingAloneIV(key, iv, data)
if err != nil {
    panic(err)
}

// 解密
decrypted, err := des.DecryptCBCPkCS7PaddingAloneIV(key, iv, encrypted)
if err != nil {
    panic(err)
}
```

## API 说明

### 加密函数

1. `EncryptStringCBCPkCS7PaddingStringHex(key, data string) (string, error)`
   - 使用 UTF-8 编码的字符串密钥
   - 返回 16 进制字符串形式的加密结果

2. `EncryptStringCBCPkCS7PaddingHex(keyHex, data string) (string, error)`
   - 使用 16 进制字符串表示的密钥
   - 返回 16 进制字符串形式的加密结果

3. `EncryptCBCPkCS7Padding(key, data []byte) ([]byte, error)`
   - 使用字节数组形式的密钥（同时用作 IV）
   - 返回字节数组形式的加密结果

4. `EncryptCBCPkCS7PaddingAloneIV(key, iv, data []byte) ([]byte, error)`
   - 使用独立的密钥和 IV
   - 返回字节数组形式的加密结果

### 解密函数

1. `DecryptStringCBCPkCS7PaddingStringHex(key, dataHex string) (string, error)`
   - 使用 UTF-8 编码的字符串密钥
   - 接受 16 进制字符串形式的加密数据

2. `DecryptStringCBCPkCS7PaddingHex(keyHex, dataHex string) (string, error)`
   - 使用 16 进制字符串表示的密钥
   - 接受 16 进制字符串形式的加密数据

3. `DecryptCBCPkCS7Padding(key, data []byte) ([]byte, error)`
   - 使用字节数组形式的密钥（同时用作 IV）
   - 接受字节数组形式的加密数据

4. `DecryptCBCPkCS7PaddingAloneIV(key, iv, data []byte) ([]byte, error)`
   - 使用独立的密钥和 IV
   - 接受字节数组形式的加密数据

### 填充函数

1. `PKCS7Padding(data []byte, blockSize int) []byte`
   - 对数据进行 PKCS7 标准填充
   - 返回填充后的数据

2. `PKCS7UnPadding(data []byte) ([]byte, error)`
   - 移除 PKCS7 标准填充
   - 返回原始数据

## 注意事项

1. 密钥长度
   - DES 密钥必须是 8 字节（64 位）
   - 16 进制表示的密钥长度必须是 16 个字符

2. IV 长度
   - IV 长度必须等于块大小（8 字节）
   - 如果不单独指定 IV，将使用密钥作为 IV

3. 安全性
   - DES 算法已不再被认为是安全的，建议在新项目中使用 AES
   - 仅在需要兼容旧系统时使用 DES

## 最佳实践

1. 密钥管理
   - 使用安全的方式生成和存储密钥
   - 避免硬编码密钥
   - 定期更换密钥

2. IV 处理
   - 建议使用随机生成的 IV
   - 每次加密使用不同的 IV
   - 安全传输 IV

3. 错误处理
   - 始终检查返回的错误
   - 避免在错误信息中泄露敏感信息
   - 合理处理填充错误

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../../LICENSE) 文件。 