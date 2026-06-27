# rsa

## 简介

`rsa` 包提供了 RSA 加密算法的实现，支持新代码推荐使用的 RSA-OAEP 公钥加密/私钥解密，并保留 PKCS#1 v1.5 加解密 API 用于兼容历史密文格式或既有协议。该包封装了 Go 标准库中的 crypto/rsa 包，提供更简便的 API 接口，适用于需要非对称加密的各种应用场景；其中私钥加密/公钥解密能力仅用于兼容历史数字签名场景。

### 主要特性

- 支持 RSA-OAEP 公钥加密与私钥解密，新代码优先使用
- 默认 OAEP 使用 SHA-256 + nil label，可通过 `*OAEPWithHash` 函数指定 hash 和 label
- 保留 PKCS#1 v1.5 公钥加密/私钥解密 API，仅用于兼容历史密文格式或既有协议
- 支持历史私钥加密/公钥解密场景，用于兼容旧数字签名协议
- 提供 PEM 格式 RSA 私钥解析和公钥导出功能
- 完整的错误处理
- 简洁易用的 API

### 设计理念

该包的设计理念是简化 RSA 加密/解密操作，使开发者能够轻松集成到应用中而无需深入了解 RSA 算法的复杂细节。通过提供清晰的 API 接口和完善的错误处理，降低了使用非对称加密的门槛。包的实现注重安全性和可靠性，使用标准的密钥格式和加密方式，确保与其他系统的兼容性。

## 安装

### 前置条件

- Go 版本要求：Go 1.26 或更高版本
- 依赖要求：
  - Go 标准库的 crypto/rsa
  - Go 标准库的 crypto/x509
  - Go 标准库的 encoding/pem

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/crypto/rsa
```

## 快速开始

### 基础用法

PKCS#1 v1.5 API 仅用于兼容历史密文格式或既有协议，新代码应优先使用 OAEP。

```go
package main

import (
    "fmt"
    "github.com/fsyyft-go/kit/crypto/rsa"
)

func main() {
    // 准备测试数据
    pubKeyPEM := []byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvxqfCMefeTjArTX338LJ
KF1p4QHzk0XA7/twfgsaVBLQiqn4Rg1j7hP5sE5NnD/RgX8XJG6S/WSNTPPLyQ/M
0eYuzI/SC5sH5HWXxS4juHjBmwozqjDqxwlS96XH7tHaSlqxbr61TdYq8M9wYZuG
Ny+uNRvXoJmQ6zNrssB7V4KHtR0Z4iB6Jys6jQ5QmNzDGCeJvQnkzEBidgkZ+kYt
MKBgVW/KFfHZk9slzkWeZJxB1ptHGUPYJOLdQHkwKQxmNu+3oR0E5gFSQJbWF16M
YIfNBx8R29MN4ZOKiY/Gae+S0dCnoHOQG7hfHVxTDSZdMgkwBMcQKRvo//NEYGUj
SwIDAQAB
-----END PUBLIC KEY-----`)

    privKeyPEM := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAvxqfCMefeTjArTX338LJKF1p4QHzk0XA7/twfgsaVBLQiqn4
Rg1j7hP5sE5NnD/RgX8XJG6S/WSNTPPLyQ/M0eYuzI/SC5sH5HWXxS4juHjBmwoz
qjDqxwlS96XH7tHaSlqxbr61TdYq8M9wYZuGNy+uNRvXoJmQ6zNrssB7V4KHtR0Z
4iB6Jys6jQ5QmNzDGCeJvQnkzEBidgkZ+kYtMKBgVW/KFfHZk9slzkWeZJxB1ptH
GUPYJOLdQHkwKQxmNu+3oR0E5gFSQJbWF16MYIfNBx8R29MN4ZOKiY/Gae+S0dCn
oHOQG7hfHVxTDSZdMgkwBMcQKRvo//NEYGUjSwIDAQABAoIBAQCTDLuKgzT9rY4h
vCkz9LjYXjt6qWGvgUwxvTfp04o9pTzxhSuiLZV8AW/5/h1Q0k+el/OoZ60H03/q
+mxnZO8frFZej/zYDZ9RdVYUZOSyXZBkxOBqzZ3N2OiKYZxQ5Cp5EDhrPkm/Z/gR
1gChmmdQfJIQR3mKSIwRLnTyF+UZYlZAExfM2NxbwUn8FTwm5RzLX0FQhiWgLOEJ
8FCQaLny1ckvZqUzRzNPZ8ViUagfMxUcHDUUi7T1jdHuJ4IRSUbU+mPnEINBYOCM
imoxeehwE0B9SuJeJmQYEq5rHxvFlLdzeDRkcK/BVY6P+/LNL0r1WoRnM7CvB2Sz
XPrh+jWBAoGBAPJxjKJRRKz2YyEVe+KJ9+iGJEXnWCU91G0AmUhTQUMbLfCNwCwN
sR0AwutNQwK2J5v8MN6PdpSN8fFFBcMeCVBQ3RNZQKpU5kCnbtM5YI5+rQS9fKcs
2/+WNZMaItOYrQxlXjV5z5oi23UE+CV/WYY/CHbPNv9QM4KyCKciHs/jAoGBAMmo
pJleYKv3pWGO6iTu68vZuwxbENmsfJZZVrRZODdLLHPigIGRM8xKpvAQZ7xgNYj/
iQPSdp0yC5fP8WZrZXrH77hgGzTcRnuOLVrDtJRMC4Y9GxRwZ3/nExGYG1n0Ya1S
GENlJRHN1s1xCP26p78tLMbLZnB4FLUFmXFsijVpAoGAKMHO1d/ONT+FW3qh701d
s5wNpFGwF8WhYPOBGE1PBRHDJqXc0DV2xEd5YXjLKONYjrGYUcj0PblmTDzww/7R
7VDHO8JY2KFpVigEy7SjhZ5MQZ6JhFP0jRJJGVVCOYIpLUJr6I+MKjVVTGYcmqLG
M3JFcImofBhmvXmvmYZeHQ0CgYBfcQmYYmJFRDbGGWxhZGJo2GGUjkDvYYOQYnFY
b9OxGcdnczWVBCBVnCxEbwvnJ+ZFrEN9U8Xl8+FgOIJM+XcOzwCL4t3ykEfQZ/Vl
yvdV9bPOYKoKVoLiX3jJO98iaPzXZVyV5M+V2cJGnqF2kSEJKp8vgIVLCpmI/Sm/
aMkPqQKBgQDG1EVRuEF3xGCJP2zEMTSzhD2xj89Pr0v2UKcpLcRu7YXoRY/LVksd
caHZUZG3CXTtKI/UhzN5/LNYOkWzr0hCYjnkUlXvpTrO3XOV6Pk05gWi5mi7SLY/
SxkMa7QWMOLxGLDFNdrMiGKBe+Hy5CgRvU9QHdmJIJFYnRjz/8dkJA==
-----END RSA PRIVATE KEY-----`)

    // 要加密的数据
    plaintext := []byte("Hello, RSA encryption!")
    
    // 使用 RSA-OAEP 公钥加密；默认使用 SHA-256 + nil label
    ciphertext, err := rsa.EncryptPubKeyOAEP(pubKeyPEM, plaintext)
    if err != nil {
        panic(err)
    }
    fmt.Printf("加密后的数据长度: %d 字节\n", len(ciphertext))
    
    // 使用 RSA-OAEP 私钥解密；hash 和 label 必须与加密时一致
    decrypted, err := rsa.DecryptPrivKeyOAEP(privKeyPEM, ciphertext)
    if err != nil {
        panic(err)
    }
    fmt.Printf("解密后的数据: %s\n", string(decrypted))
}
```

### 配置选项

```go
// RSA 包设计简单，大部分功能通过直接调用函数实现
// 以下示例展示如何手动处理密钥对象

// 将 PEM 格式的私钥转换为 RSA 私钥对象
privKey, err := rsa.ConvertPrivateKey(privKeyPEM)
if err != nil {
    panic(err)
}

// 结构体 API 适用于已经取得 *rsa.PublicKey 的场景；
// 例如可以从已解析的私钥中取得对应公钥。
pubKey := &privKey.PublicKey

// 使用公钥对象按 RSA-OAEP 直接加密
encrypted, err := rsa.EncryptPublicKeyOAEP(pubKey, data)
if err != nil {
    panic(err)
}

// 使用私钥对象按 RSA-OAEP 直接解密
decrypted, err := rsa.DecryptPrivateKeyOAEP(privKey, encrypted)
if err != nil {
    panic(err)
}

// 如需导出 PEM 格式公钥，可使用 ConvertPubKey
pubKeyPEM, err := rsa.ConvertPubKey(pubKey)
if err != nil {
    panic(err)
}
```

## 详细指南

### 核心概念

RSA 是一种非对称加密算法，使用一对密钥：公钥和私钥。公钥可以公开，用于加密数据；
私钥保密，用于解密数据。这种非对称性使得 RSA 适用于需要安全通信但无法预先共享密钥的场景。

本包实现了标准的 RSA 加密和解密功能，支持推荐用于新代码的 RSA-OAEP 填充方式，并保留 PKCS#1 v1.5 填充 API 用于兼容历史密文格式或既有协议。
通过封装底层的复杂操作，使开发者能够简单地集成 RSA 加密功能。

### 常见用例

#### 1. 安全传输敏感数据

```go
// 接收方提供公钥
publicKeyPEM := []byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvxqfCMefeTjArTX338LJ
...省略部分内容...
-----END PUBLIC KEY-----`)

// 发送方使用 RSA-OAEP 公钥加密敏感数据，默认使用 SHA-256 + nil label
sensitiveData := []byte("密码: 123456")
encrypted, err := rsa.EncryptPubKeyOAEP(publicKeyPEM, sensitiveData)
if err != nil {
    panic(err)
}

// 将加密后的数据发送给接收方
// ...传输过程...

// 接收方使用 RSA-OAEP 私钥解密
decrypted, err := rsa.DecryptPrivKeyOAEP(privateKeyPEM, encrypted)
if err != nil {
    panic(err)
}
fmt.Printf("解密后的敏感数据: %s\n", string(decrypted))
```

#### 2. 使用经过转换的密钥对象

```go
// 转换私钥为对象，可以重复使用；如果已有 *rsa.PublicKey，
// 也可以直接使用 EncryptPublicKeyOAEP。
privKey, err := rsa.ConvertPrivateKey(privateKeyPEM)
if err != nil {
    panic(err)
}
pubKey := &privKey.PublicKey

// 多次使用同一密钥对象按 RSA-OAEP 加密不同数据
message1 := []byte("第一条消息")
encrypted1, err := rsa.EncryptPublicKeyOAEP(pubKey, message1)
if err != nil {
    panic(err)
}

message2 := []byte("第二条消息")
encrypted2, err := rsa.EncryptPublicKeyOAEP(pubKey, message2)
if err != nil {
    panic(err)
}
```

### 最佳实践

- 算法选择
  - 新代码优先使用 `EncryptPubKeyOAEP` / `DecryptPrivKeyOAEP`，默认参数为 SHA-256 + nil label
  - 需要指定 OAEP hash 或 label 时，使用 `EncryptPubKeyOAEPWithHash` / `DecryptPrivKeyOAEPWithHash`，并确保加密和解密参数完全一致
  - 旧 `EncryptPubKey` / `DecryptPrivKey` 使用 PKCS#1 v1.5，仅用于兼容历史密文格式或既有协议
  - 私钥加密/公钥解密仅用于兼容历史数字签名场景，新签名逻辑建议使用标准签名 API

- 密钥管理
  - 安全地存储私钥，避免泄露
  - 使用至少 2048 位的密钥长度以确保安全性
  - 考虑定期轮换密钥对

- 加密限制
  - RSA 加密的明文长度有限制，取决于密钥长度和填充方式
  - 对于大型数据，使用混合加密方案：使用 AES 等对称算法加密数据，再用 RSA-OAEP 加密对称密钥

- 性能考虑
  - RSA 操作计算密集，不适合频繁加密大量数据
  - 对于重复使用的密钥，预先转换为密钥对象可以提高性能

- 错误处理
  - 始终检查加密和解密函数返回的错误
  - 正确处理密钥格式错误和解密失败情况

## API 文档

### 主要类型

```go
// 本包主要使用 Go 标准库中的类型：
// *rsa.PublicKey - RSA 公钥对象
// *rsa.PrivateKey - RSA 私钥对象
```

### 关键函数

#### EncryptPubKeyOAEP

使用 PEM 格式的公钥按 RSA-OAEP 加密数据。默认使用 SHA-256 + nil label。

```go
func EncryptPubKeyOAEP(publicKey, dataClear []byte) ([]byte, error)
```

示例：
```go
encrypted, err := rsa.EncryptPubKeyOAEP(publicKeyPEM, []byte("加密数据"))
if err != nil {
    panic(err)
}
```

#### DecryptPrivKeyOAEP

使用 PEM 格式的私钥按 RSA-OAEP 解密数据。默认使用 SHA-256 + nil label。

```go
func DecryptPrivKeyOAEP(privateKey, dataCipher []byte) ([]byte, error)
```

示例：
```go
decrypted, err := rsa.DecryptPrivKeyOAEP(privateKeyPEM, encrypted)
if err != nil {
    panic(err)
}
```

#### EncryptPubKey / DecryptPrivKey

使用 PKCS#1 v1.5 加解密数据，仅用于兼容历史密文格式或既有协议；新代码优先使用 OAEP API。

```go
func EncryptPubKey(publicKey, dataClear []byte) ([]byte, error)
func DecryptPrivKey(privateKey, dataCipher []byte) ([]byte, error)
```

#### EncryptPubKeyOAEPWithHash / DecryptPrivKeyOAEPWithHash

使用指定 hash 和 label 的 RSA-OAEP 加解密。加密和解密必须使用完全一致的 hash 与 label。

```go
func EncryptPubKeyOAEPWithHash(publicKey, dataClear []byte, hash hash.Hash, label []byte) ([]byte, error)
func DecryptPrivKeyOAEPWithHash(privateKey, dataCipher []byte, hash hash.Hash, label []byte) ([]byte, error)
```

#### ConvertPubKey / ConvertPrivateKey

`ConvertPrivateKey` 将 PEM 格式私钥转换为私钥对象；`ConvertPubKey` 将公钥对象导出为 PEM。若已有 PEM 公钥并只需加密，直接使用 `EncryptPubKeyOAEP` 即可。

```go
func ConvertPubKey(publicKey *rsa.PublicKey) ([]byte, error)
func ConvertPrivateKey(privateKey []byte) (*rsa.PrivateKey, error)
```

示例：
```go
privKey, err := rsa.ConvertPrivateKey(privateKeyPEM)
publicKeyPEM, err := rsa.ConvertPubKey(&privKey.PublicKey)
```

### 错误处理

本包返回以下类型的错误：
- 密钥格式错误：当 PEM 格式的密钥无法正确解码或解析时
- 加密/解密错误：当加密/解密操作失败时
- 数据长度错误：当明文数据超过 RSA 加密的长度限制时

建议始终检查所有函数返回的错误，并妥善处理密钥解析和加解密操作中可能出现的异常情况。

## 性能指标

| 操作 | 性能指标 | 说明 |
|------|----------|------|
| 密钥解析 | ~100µs | PEM 格式转换为密钥对象 |
| 公钥加密 (2048位密钥) | ~500µs | 加密短消息 |
| 私钥解密 (2048位密钥) | ~2ms | 解密短消息 |

## 测试覆盖率

| 包 | 覆盖率 |
|------|--------|
| rsa | 不低于 95% |

## 调试指南

### 常见问题排查

#### 加密/解密失败

- 确认 PEM 格式的密钥格式正确
- 检查公钥和私钥是否匹配
- 验证明文数据长度是否超过 RSA 加密的限制
- 对于 2048 位密钥，默认 RSA-OAEP（SHA-256）明文通常应不超过 190 字节；PKCS#1 v1.5 兼容 API 应不超过 245 字节

#### 密钥解析错误

- 检查 PEM 格式的正确性，包括开始和结束标记
- 确认公钥使用 `-----BEGIN PUBLIC KEY-----` 标记
- 确认私钥使用 `-----BEGIN RSA PRIVATE KEY-----` 标记
- 验证 PEM 数据未被破坏或修改

## 相关文档

- [RSA 加密算法](https://en.wikipedia.org/wiki/RSA_%28cryptosystem%29)
- [PKCS#1 规范](https://tools.ietf.org/html/rfc8017)
- [Go crypto/rsa 包文档](https://pkg.go.dev/crypto/rsa)
- [PEM 格式规范](https://tools.ietf.org/html/rfc1421)
- [RFC 5280 - X.509 公钥基础设施](https://tools.ietf.org/html/rfc5280)
- [RFC 3447 - PKCS #1: RSA 加密规范 2.1 版](https://tools.ietf.org/html/rfc3447)

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考我们的[贡献指南](../../CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT 许可证。查看 [LICENSE](../../LICENSE) 文件了解更多信息。