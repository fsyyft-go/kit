// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package bytes 提供了字节操作相关的工具函数，包括随机字节生成、字节序列处理等功能。

主要特性：

  - 提供安全的随机字节生成
  - 支持指定长度的字节序列生成
  - 提供字节序列的常用操作
  - 支持字节序列的编码和解码
  - 内置常用的字节处理函数

基本功能：

1. 随机字节生成：

	// 生成指定长度的随机字节序列
	bytes, err := bytes.GenerateRandomBytes(16)
	if err != nil {
	    panic(err)
	}

	// 生成用于加密的 nonce
	nonce, err := bytes.GenerateNonce(12)
	if err != nil {
	    panic(err)
	}

2. 字节序列操作：

	// 复制字节序列
	src := []byte("Hello, World!")
	dst := bytes.Clone(src)

	// 比较字节序列
	if bytes.Equal(src, dst) {
	    fmt.Println("字节序列相等")
	}

	// 填充字节序列
	zeros := bytes.Repeat(0x00, 10)

3. 编码转换：

	data := []byte("Hello, World!")

	// Base64 编码
	encoded := bytes.ToBase64(data)

	// 十六进制编码
	hex := bytes.ToHex(data)

4. 字节序列处理：

	// 截取字节序列
	data := []byte("Hello, World!")
	sub := bytes.Sub(data, 0, 5)  // "Hello"

	// 连接字节序列
	parts := [][]byte{
	    []byte("Hello"),
	    []byte(", "),
	    []byte("World!"),
	}
	joined := bytes.Join(parts, nil)

安全性考虑：

1. 随机性：
  - 使用密码学安全的随机数生成器
  - 避免使用不安全的随机源
  - 注意随机数的熵源质量

2. 内存处理：
  - 及时清理敏感数据
  - 避免内存泄露
  - 使用安全的内存操作

3. 并发安全：
  - 注意字节切片的并发访问
  - 避免数据竞争
  - 使用适当的同步机制

性能优化：

1. 内存分配：
  - 预分配足够的空间
  - 重用字节切片
  - 避免不必要的复制

2. 批量操作：
  - 使用批量处理提高效率
  - 合理设置缓冲区大小
  - 注意内存使用效率

3. 算法选择：
  - 选择合适的算法
  - 权衡时间和空间复杂度
  - 考虑数据规模的影响

使用建议：

1. 错误处理：
  - 始终检查错误返回值
  - 合理处理异常情况
  - 提供有意义的错误信息

2. 资源管理：
  - 及时释放不再使用的内存
  - 使用 defer 确保资源清理
  - 避免资源泄露

3. 代码组织：
  - 保持函数简单明确
  - 提供清晰的文档
  - 使用有意义的命名

更多示例和最佳实践请参考 example/bytes 目录。
*/
package bytes
