// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package des

import (
	"bytes"
	"errors"
)

var (
	// defaultDESKey 是默认的 DES 密钥。
	defaultDESKey = "go-kit-k"
)

// GetDefaultDESKey 返回包内置的默认 DES 密钥。
//
// 该密钥仅供历史兼容包装函数复用，不应被视为新的安全默认配置。
//
// 参数：无。
//
// 返回：
//   - string: 供历史兼容包装函数复用的默认 DES 密钥。
func GetDefaultDESKey() string {
	return defaultDESKey
}

// PKCS7Padding 使用 PKCS7 标准对 data 进行填充。
//
// blockSize 必须大于 0，且填充长度需要能放入单字节；函数不会对该约束做显式校验，
// 传入非法 blockSize 可能导致 panic 或产生不可逆的填充数据。返回切片通过 append 构造，
// 在容量允许时可能与 data 共享底层数组。
//
// 参数：
//   - data: 需要填充的原始数据，可为空。
//   - blockSize: 加密块大小，单位为字节；用于计算需要追加的 PKCS7 填充长度。
//
// 返回：
//   - []byte: 追加 PKCS7 padding 后的数据。
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

// PKCS7UnPadding 对使用 PKCS7 标准填充的 data 进行去填充处理。
//
// 函数通过最后一个字节判断 padding 长度，并校验 padding 字节是否全部等于该长度。
// 由于没有 blockSize 参数，函数不会校验 padding 长度是否不超过加密块大小。
// 返回的切片是 data 的子切片，会与输入共享底层数组。
//
// 参数：
//   - data: 已经填充过的数据，长度必须大于 0，最后一个字节用于表示待移除的 padding 长度。
//
// 返回：
//   - []byte: 去除 PKCS7 padding 后的原始数据。
//   - error: data 为空、padding 长度为 0、padding 长度超过 data 长度或 padding 字节不一致时返回错误。
func PKCS7UnPadding(data []byte) ([]byte, error) {
	// 获取数据的总长度。
	length := len(data)
	// 如果数据为空，返回错误。
	if length == 0 {
		return nil, errors.New("empty data")
	}

	// 获取填充值：
	// - 最后一个字节的值就是填充的字节数。
	unPadding := int(data[length-1])
	// 验证填充值的合法性：
	// - 填充值不能为 0。
	// - 填充值不能大于数据总长度。
	if unPadding == 0 || unPadding > length {
		return nil, errors.New("invalid padding value")
	}

	// 验证所有填充字节是否相同：
	// - PKCS7 要求所有填充字节的值必须相同。
	// - 填充字节的值必须等于填充的字节数。
	for i := length - unPadding; i < length; i++ {
		if int(data[i]) != unPadding {
			return nil, errors.New("invalid padding")
		}
	}

	// 返回去除填充后的数据：
	// - 从数据开始到（总长度-填充字节数）的部分就是原始数据。
	return data[:(length - unPadding)], nil
}
