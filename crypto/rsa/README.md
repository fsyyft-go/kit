# RSA 加密包

`rsa` 包提供了 RSA 加密算法相关功能的实现，支持公钥加密/私钥解密以及私钥加密/公钥解密操作。

## 特性

- 公钥加密和私钥解密
- 私钥加密（签名）和公钥解密（验证）
- PEM 格式密钥处理
- 支持 PKCS#1 v1.5 填充方案
- 错误处理和 Panic 恢复机制
- 丰富的类型转换工具函数

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/fsyyft-go/kit/crypto/rsa"
)

func main() {
    // 假设已有 PEM 格式的公钥和私钥
    var publicKeyPEM, privateKeyPEM []byte
    // 从文件或其他来源加载密钥...
    
    // 原始数据
    data := []byte("Hello, RSA encryption!")
    
    // 使用公钥加密
    encrypted, err := rsa.EncryptPubKey(publicKeyPEM, data)
    if err != nil {
        panic(err)
    }
    
    // 使用私钥解密
    decrypted, err := rsa.DecryptPrivKey(privateKeyPEM, encrypted)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("解密后: %s\n", decrypted)
}
```

### 数字签名

```go
// 使用私钥加密（签名）
signature, err := rsa.EncryptPrivKey(privateKeyPEM, data)
if err != nil {
    panic(err)
}

// 使用公钥解密（验证签名）
verified, err := rsa.DecryptPubKey(publicKeyPEM, signature)
if err != nil {
    panic(err)
}

fmt.Printf("验证签名: %s\n", verified)
```

### 直接使用 RSA 结构

如果已经有 `rsa.PublicKey` 和 `rsa.PrivateKey` 对象，可以直接使用：

```go
// 已有 RSA 密钥对象
var pubKey *rsa.PublicKey
var privKey *rsa.PrivateKey

// 使用公钥对象加密
encrypted, err := rsa.EncryptPublicKey(pubKey, data)
if err != nil {
    panic(err)
}

// 使用私钥对象解密
decrypted, err := rsa.DecryptPrivateKey(privKey, encrypted)
if err != nil {
    panic(err)
}
```

### 密钥转换

```go
// PEM 格式私钥转换为 RSA 私钥对象
privKey, err := rsa.ConvertPrivateKey(privateKeyPEM)
if err != nil {
    panic(err)
}

// RSA 公钥对象转换为 PEM 格式
pubKeyPEM, err := rsa.ConvertPubKey(pubKey)
if err != nil {
    panic(err)
}
```

## 最佳实践

1. 安全性考虑
   - 妥善保管私钥，避免泄露
   - 加密数据大小不应超过密钥长度减去填充长度
   - 尽量使用 2048 位或以上长度的密钥

2. 性能优化
   - RSA 操作计算密集，不适合大量数据加密
   - 对于大数据可使用混合加密方案（如 RSA 加密对称密钥）
   - 重用密钥对象避免重复解析

3. 错误处理
   - 始终检查并处理返回的错误
   - 包含适当的错误恢复机制

4. 密钥管理
   - 采用标准 PEM 格式存储密钥
   - 实现密钥轮换机制
   - 考虑使用密钥管理系统

## 注意事项

- 本模块仅用于基本的 RSA 加密解密操作，如需更高级的功能，请考虑使用专业的加密库
- 默认使用 PKCS#1 v1.5 填充，该填充方案可能面临某些攻击风险
- RSA 加密存在明显的大小限制，单次加密的数据量不能超过密钥长度减去填充长度

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../../LICENSE) 文件。 