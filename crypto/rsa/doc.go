// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package rsa 提供了 RSA 加密算法的实现，支持公钥加密、私钥解密以及数字签名功能。

主要功能：

1. 密钥操作：
  - 支持公钥加密和解密
  - 支持私钥加密和解密
  - 支持密钥格式转换
  - 自动密钥长度验证

2. 加密功能：
  - 公钥加密：用于数据加密
  - 私钥加密：用于数字签名
  - 支持多种数据格式
  - 自动填充处理

3. 解密功能：
  - 私钥解密：解密加密数据
  - 公钥解密：验证数字签名
  - 错误恢复机制
  - 自动移除填充

基本用法：

1. 公钥加密：

	// 使用公钥加密数据
	ciphertext, err := rsa.EncryptPubKey(publicKey, plaintext)
	if err != nil {
	    // 处理错误
	}

	// 使用私钥解密数据
	plaintext, err := rsa.DecryptPrivKey(privateKey, ciphertext)

2. 数字签名：

	// 使用私钥签名数据
	signature, err := rsa.EncryptPrivKey(privateKey, data)
	if err != nil {
	    // 处理错误
	}

	// 使用公钥验证签名
	data, err := rsa.DecryptPubKey(publicKey, signature)

安全特性：

1. 密钥管理：
  - 支持标准 PEM 格式
  - 密钥长度验证
  - 密钥完整性检查

2. 填充机制：
  - PKCS#1 v1.5 填充
  - 防止填充 Oracle 攻击
  - 安全的填充验证

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
  - 避免直接加密敏感数据
  - 结合对称加密使用
  - 注意并发安全
*/
package rsa
