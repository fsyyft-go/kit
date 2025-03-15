// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package md5 提供了 MD5 消息摘要算法的实现，用于生成数据的固定长度哈希值。

主要功能：

1. 哈希计算：
  - 支持字符串哈希
  - 支持文件哈希
  - 支持字节数组哈希
  - 多种输出格式

2. 输出格式：
  - 十六进制字符串
  - 字节数组
  - 大写/小写转换
  - Base64 编码

3. 特殊功能：
  - 流式处理
  - 增量更新
  - 并发安全
  - 性能优化

基本用法：

1. 字符串哈希：

	// 计算字符串的 MD5 值
	hash := md5.String("hello world")

	// 计算并转换为大写
	hashUpper := md5.StringUpper("hello world")

2. 文件哈希：

	// 计算文件的 MD5 值
	hash, err := md5.File("path/to/file")
	if err != nil {
	    // 处理错误
	}

	// 计算并转换为大写
	hashUpper, err := md5.FileUpper("path/to/file")

3. 字节数组哈希：

	// 计算字节数组的 MD5 值
	hash := md5.Bytes(data)

	// 计算并转换为大写
	hashUpper := md5.BytesUpper(data)

性能优化：

1. 内存使用：
  - 避免不必要的内存分配
  - 使用适当的缓冲区大小
  - 及时释放资源

2. 计算优化：
  - 高效的块处理
  - 优化的字符串处理
  - 减少内存拷贝

注意事项：

1. 安全考虑：
  - MD5 不适用于安全场景
  - 存在碰撞风险
  - 建议用于校验用途

2. 使用建议：
  - 适用于数据完整性校验
  - 适用于缓存键生成
  - 不适用于密码存储

3. 最佳实践：
  - 选择合适的输出格式
  - 注意大小写敏感性
  - 考虑并发安全性
*/
package md5
