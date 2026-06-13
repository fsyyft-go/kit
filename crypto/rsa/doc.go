// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package rsa 提供 RSA 加密相关工具，支持 RSA-OAEP 公钥加密/私钥解密、
PKCS#1 v1.5 兼容加解密、PEM 密钥转换，以及历史私钥加密/公钥解密场景。

主要功能：

1. 密钥操作：
  - 支持 PEM 私钥解析和公钥导出
  - 支持密钥格式转换
  - 依赖标准库解析与加解密错误返回
  - 密钥完整性检查

2. 加密功能：
  - 公钥加密：新代码推荐 RSA-OAEP，优先使用 EncryptPubKeyOAEP 或 EncryptPublicKeyOAEP
  - 私钥加密：用于兼容历史“私钥加密、公钥解密”的数字签名场景
  - 支持多种数据格式
  - 自动填充处理

3. 解密功能：
  - 私钥解密：新代码推荐 RSA-OAEP，优先使用 DecryptPrivKeyOAEP 或 DecryptPrivateKeyOAEP
  - 公钥解密：用于兼容历史数字签名验证场景
  - 错误恢复机制
  - 自动移除填充

基本用法：

1. 公钥加密：

	// 使用 RSA-OAEP 公钥加密数据，默认使用 SHA-256 和 nil label
	ciphertext, err := rsa.EncryptPubKeyOAEP(publicKey, plaintext)
	if err != nil {
	    // 处理错误
	}

	// 使用 RSA-OAEP 私钥解密数据
	plaintext, err := rsa.DecryptPrivKeyOAEP(privateKey, ciphertext)

2. 历史数字签名兼容：

	// 使用私钥加密数据，仅用于兼容历史数字签名场景
	signature, err := rsa.EncryptPrivKey(privateKey, data)
	if err != nil {
	    // 处理错误
	}

	// 使用公钥验证历史签名数据
	data, err := rsa.DecryptPubKey(publicKey, signature)

安全特性：

1. 密钥管理：
  - 支持标准 PEM 格式
  - 依赖标准库解析与加解密错误返回
  - 密钥完整性检查

2. 填充机制：
  - RSA-OAEP 是新代码推荐方案
  - PKCS#1 v1.5 仅用于兼容历史密文格式或既有协议
  - OAEP 使用标准库填充处理；PKCS#1 v1.5 仅作兼容用途

3. 错误处理：
  - 安全的错误信息
  - 完整的错误检查
  - panic 保护机制

性能优化：

1. 内存使用：
  - 避免不必要的内存分配
  - 使用适当的缓冲区大小
  - 及时释放资源

2. 计算优化：
  - 高效的大数运算
  - 优化的填充处理
  - 减少内存拷贝

注意事项：

1. 密钥安全：
  - 使用足够长的密钥
  - 安全存储私钥
  - 定期更新密钥

2. 数据处理：
  - 检查数据大小限制
  - 验证密钥有效性
  - 处理填充错误

3. 安全使用：
  - 新代码优先使用 EncryptPubKeyOAEP / DecryptPrivKeyOAEP
  - 大型数据使用混合加密，例如用对称加密处理数据，再用 RSA-OAEP 加密对称密钥
  - 注意并发安全
*/
package rsa
