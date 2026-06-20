// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package bytes

import (
	"crypto/rand"
	"fmt"
	"io"
)

// GenerateNonce 返回指定长度的密码学安全随机字节切片。
//
// length 必须大于等于 0。返回值可用作 nonce、IV 或其他需要随机原始字节的场景；
// 结果是否满足上层协议对唯一性、长度和编码的要求由调用方负责。
//
// 参数：
//   - length：指定生成随机字节的长度，必须大于等于 0。
//
// 返回：
//   - []byte：生成的随机字节切片。
//   - error：当 length 为负数或底层随机源读取失败时返回错误。
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
