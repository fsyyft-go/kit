// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package aes 提供了 AES 加密算法的实现，特别是 GCM 模式的加密和解密功能。

主要功能：

1. GCM 模式加密：
  - 支持字节数组加密
  - 支持字符串加密
  - 支持 Base64 和 Hex 编码格式
  - 自动生成和管理 nonce

2. 编码格式支持：
  - Base64 编码：适用于文本传输
  - Hex 编码：适用于调试和日志
  - 原始字节数组：适用于二进制传输

3. 字符串操作：
  - UTF-8 编码的字符串加密
  - Base64/Hex 格式的字符串处理
  - 自动编码转换

基本用法：

1. 字符串加密（Base64）：

	// 加密 UTF-8 字符串
	ciphertext, err := aes.EncryptStringGCMBase64(keyBase64, nonceLength, plaintext)
	if err != nil {
	    // 处理错误
	}

	// 解密获取原文
	nonce, plaintext, err := aes.DecryptStringGCMBase64(keyBase64, nonceLength, ciphertext)

2. 字符串加密（Hex）：

	// 加密 UTF-8 字符串
	ciphertext, err := aes.EncryptStringGCMHex(keyHex, nonceLength, plaintext)
	if err != nil {
	    // 处理错误
	}

	// 解密获取原文
	nonce, plaintext, err := aes.DecryptStringGCMHex(keyHex, nonceLength, ciphertext)

3. 字节数组加密：

	// 使用指定的 nonce 长度加密
	ciphertext, err := aes.EncryptGCMNonceLength(key, nonceLength, data)
	if err != nil {
	    // 处理错误
	}

	// 使用已知的 nonce 加密
	ciphertext, err := aes.EncryptGCM(key, nonce, data)

安全特性：

1. GCM 模式：
  - 提供认证加密（AEAD）
  - 防止重放攻击
  - 保证数据完整性

2. Nonce 处理：
  - 自动生成随机 nonce
  - 支持自定义 nonce 长度
  - nonce 与密文一起存储

3. 错误处理：
  - 安全的错误信息
  - 完整的错误检查
  - panic 保护机制

性能优化：

1. 内存使用：
  - 避免不必要的内存分配
  - 使用适当的缓冲区大小
  - 及时释放资源

2. 编码处理：
  - 最小化编码转换
  - 直接处理字节数组
  - 避免重复编码

注意事项：

1. 密钥管理：
  - 使用安全的密钥生成方法
  - 妥善保管密钥
  - 避免重用 nonce

2. 数据处理：
  - 检查输入数据大小
  - 验证密钥长度
  - 处理编码错误

3. 安全使用：
  - 不要修改 nonce 长度
  - 验证解密结果
  - 注意并发安全
*/
package aes
