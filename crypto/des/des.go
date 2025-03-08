// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package des 实现了 DES 加密算法相关的功能。
package des

import (
	"bytes"
)

// PKCS7Padding 使用 PKCS7 标准对数据进行填充。
//
// 参数：
//   - data：需要填充的原始数据。
//   - blockSize：加密块的大小（以字节为单位）。
//
// 返回：
//   - []byte：完成填充后的数据。
func PKCS7Padding(data []byte, blockSize int) []byte {
	// PKCS7 填充规则说明：
	// 1. 如果数据长度小于块大小，填充的值为"缺少的字节数"。
	// 2. 如果数据长度是块大小的整数倍，填充一个完整块，每个字节的值为块大小。
	// 3. 填充值的范围是 1-255，保证能放入一个字节中。

	// 计算需要填充的字节数：
	// - 如果数据长度是块大小的整数倍，填充一个完整块。
	// - 如果不是整数倍，填充到下一个块大小。
	padding := blockSize - len(data)%blockSize

	// 创建填充数据：
	// - 使用 bytes.Repeat 函数生成指定数量的填充字节。
	// - 每个填充字节的值都等于填充的字节数。
	padData := bytes.Repeat([]byte{byte(padding)}, padding)

	// 将填充数据追加到原始数据后面。
	d := append(data, padData...)

	return d
}

// PKCS7UnPadding 对使用 PKCS7 标准填充的数据进行去填充处理。
//
// 参数：
//   - data：已经填充过的数据。
//
// 返回：
//   - []byte：去除填充后的原始数据。
func PKCS7UnPadding(data []byte) []byte {
	// 获取数据的总长度。
	length := len(data)

	// 获取填充值：
	// - 最后一个字节的值就是填充的字节数。
	unPadding := int(data[length-1])

	// 返回去除填充后的数据：
	// - 从数据开始到（总长度-填充字节数）的部分就是原始数据。
	d := data[:(length - unPadding)]

	return d
}
