// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package des 提供了 DES（Data Encryption Standard）加密算法的实现，包括标准 DES 和三重 DES（3DES）加密。

主要功能：

1. DES 加密：
  - 标准 DES 加密和解密
  - ECB 和 CBC 工作模式
  - 支持多种数据格式
  - 自动填充处理

2. 3DES 加密：
  - 三重 DES 加密和解密
  - 支持 EDE2 和 EDE3 模式
  - 增强的安全性
  - 兼容标准 DES

3. 工作模式：
  - ECB：电子密码本模式
  - CBC：密码分组链接模式
  - 自动 IV 管理
  - 块对齐处理

基本用法：

1. 标准 DES 加密：

	// 使用 ECB 模式加密
	ciphertext, err := des.EncryptECB(key, plaintext)
	if err != nil {
	    // 处理错误
	}

	// 使用 ECB 模式解密
	plaintext, err := des.DecryptECB(key, ciphertext)

2. 3DES 加密：

	// 使用 CBC 模式加密
	ciphertext, err := des.EncryptTripleDESCBC(key, iv, plaintext)
	if err != nil {
	    // 处理错误
	}

	// 使用 CBC 模式解密
	plaintext, err := des.DecryptTripleDESCBC(key, iv, ciphertext)

安全特性：

1. 密钥管理：
  - 密钥长度验证
  - 密钥完整性检查
  - 安全的密钥处理

2. IV 处理：
  - 随机 IV 生成
  - IV 长度验证
  - 安全的 IV 管理

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
  - 高效的块处理
  - 优化的填充处理
  - 减少内存拷贝

注意事项：

1. 安全考虑：
  - DES 已不再推荐用于新系统
  - 优先使用 3DES 或 AES
  - 注意密钥强度

2. 数据处理：
  - 检查数据块大小
  - 验证密钥长度
  - 处理填充错误

3. 最佳实践：
  - 避免使用 ECB 模式
  - 使用随机 IV
  - 注意并发安全
*/
package des
