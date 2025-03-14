// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package bytes 提供了字节操作相关的工具函数。
package bytes

import (
	"crypto/rand"
	"fmt"
	"io"
)

// GenerateNonce 获取一次性随机字节数组。
//
// 参数：
//   - length：指定生成随机字节的长度。
//
// 返回：
//   - []byte：生成的随机字节切片。
//   - error：如果生成过程中出现错误，则返回相应的错误。
func GenerateNonce(length int) ([]byte, error) {
	// 检查长度参数是否合法。
	if length < 0 {
		return nil, fmt.Errorf("长度不能为负数：%d", length)
	}

	// 创建指定长度的字节切片用于存储随机数。
	var nonce = make([]byte, length)
	// 声明错误变量。
	var err error

	// 使用加密安全的随机数生成器填充 nonce 切片。
	// io.ReadFull 确保读取足够的随机字节以填满整个切片。
	_, err = io.ReadFull(rand.Reader, nonce)

	// 返回生成的随机字节切片和可能的错误。
	return nonce, err
}
