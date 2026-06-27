// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	kitbytes "github.com/fsyyft-go/kit/bytes"
)

// EncryptStringGCMBase64 使用 Base64 密钥加密字符串，并返回 Base64 编码的组合密文。
//
// 本函数将 data 按原始字节传入 AES-GCM，不额外附加 AAD。返回内容是
// nonce || ciphertextAndTag 的 Base64 编码，可交给 DecryptStringGCMBase64 解密。
//
// 参数：
//   - keyBase64：Base64 格式的 AES 密钥，解码后长度必须符合 aes.NewCipher 要求。
//   - nonceLength：随机生成 nonce 的字节长度，必须与 GCM nonce 长度一致。
//   - data：待加密的字符串明文；调用方负责保证其文本编码符合业务约定。
//
// 返回：
//   - string：Base64 编码的 nonce || ciphertextAndTag；失败时为空字符串。
//   - error：keyBase64 解码失败、nonce 生成失败、密钥非法或 nonceLength 不匹配时返回错误。
func EncryptStringGCMBase64(keyBase64 string, nonceLength int, data string) (string, error) {
	// 声明返回值变量。
	var result string
	var err error

	// 将 Base64 格式的密钥解码为字节数组。
	if key, errKey := base64.StdEncoding.DecodeString(keyBase64); nil != errKey {
		// 如果密钥解码失败，保存错误。
		err = errKey
	} else if tmpResult, errEncrypt := EncryptGCMNonceLength(key, nonceLength, []byte(data)); nil != errEncrypt {
		// 如果加密失败，保存错误。
		err = errEncrypt
	} else {
		// 加密成功，将结果编码为 Base64 格式。
		result = base64.StdEncoding.EncodeToString(tmpResult)
	}

	// 返回加密结果和可能的错误。
	return result, err
}

// EncryptStringGCMHex 使用 Hex 密钥加密字符串，并返回 Hex 编码的组合密文。
//
// 本函数将 data 按原始字节传入 AES-GCM，不额外附加 AAD。返回内容是
// nonce || ciphertextAndTag 的小写 Hex 编码，可交给 DecryptStringGCMHex 解密。
//
// 参数：
//   - keyHex：Hex 格式的 AES 密钥，解码后长度必须符合 aes.NewCipher 要求。
//   - nonceLength：随机生成 nonce 的字节长度，必须与 GCM nonce 长度一致。
//   - data：待加密的字符串明文；调用方负责保证其文本编码符合业务约定。
//
// 返回：
//   - string：小写 Hex 编码的 nonce || ciphertextAndTag；失败时为空字符串。
//   - error：keyHex 解码失败、nonce 生成失败、密钥非法或 nonceLength 不匹配时返回错误。
func EncryptStringGCMHex(keyHex string, nonceLength int, data string) (string, error) {
	// 声明返回值变量。
	var result string
	var err error

	// 将十六进制格式的密钥解码为字节数组。
	if key, errKey := hex.DecodeString(keyHex); nil != errKey {
		// 如果密钥解码失败，保存错误。
		err = errKey
	} else if tmpResult, errEncrypt := EncryptGCMNonceLength(key, nonceLength, []byte(data)); nil != errEncrypt {
		// 如果加密失败，保存错误。
		err = errEncrypt
	} else {
		// 加密成功，将结果编码为十六进制格式。
		result = hex.EncodeToString(tmpResult)
	}

	// 返回加密结果和可能的错误。
	return result, err
}

// EncryptGCMBase64 解码 Base64 密钥和明文后执行 AES-GCM 加密。
//
// 本函数先解码 dataBase64，再按 nonce || ciphertextAndTag 组合加密结果，最后返回该组合密文的 Base64 编码。
// 不额外附加 AAD。
//
// 参数：
//   - keyBase64：Base64 格式的 AES 密钥，解码后长度必须符合 aes.NewCipher 要求。
//   - nonceLength：随机生成 nonce 的字节长度，必须与 GCM nonce 长度一致。
//   - dataBase64：Base64 格式的明文字节数据。
//
// 返回：
//   - string：Base64 编码的 nonce || ciphertextAndTag；失败时为空字符串。
//   - error：keyBase64 或 dataBase64 解码失败、nonce 生成失败、密钥非法或 nonceLength 不匹配时返回错误。
func EncryptGCMBase64(keyBase64 string, nonceLength int, dataBase64 string) (string, error) {
	// 声明返回值变量。
	var result string
	var err error

	// 将 Base64 格式的密钥解码为字节数组。
	if key, errKey := base64.StdEncoding.DecodeString(keyBase64); nil != errKey {
		// 如果密钥解码失败，保存错误。
		err = errKey
	} else if data, errData := base64.StdEncoding.DecodeString(dataBase64); nil != errData {
		// 如果数据解码失败，保存错误。
		err = errData
	} else if tmpResult, errEncrypt := EncryptGCMNonceLength(key, nonceLength, data); nil != errEncrypt {
		// 如果加密失败，保存错误。
		err = errEncrypt
	} else {
		// 加密成功，将结果编码为 Base64 格式。
		result = base64.StdEncoding.EncodeToString(tmpResult)
	}

	// 返回加密结果和可能的错误。
	return result, err
}

// EncryptGCMHex 解码 Hex 密钥和明文后执行 AES-GCM 加密。
//
// 本函数先解码 dataHex，再按 nonce || ciphertextAndTag 组合加密结果，最后返回该组合密文的小写 Hex 编码。
// 不额外附加 AAD。
//
// 参数：
//   - keyHex：Hex 格式的 AES 密钥，解码后长度必须符合 aes.NewCipher 要求。
//   - nonceLength：随机生成 nonce 的字节长度，必须与 GCM nonce 长度一致。
//   - dataHex：Hex 格式的明文字节数据。
//
// 返回：
//   - string：小写 Hex 编码的 nonce || ciphertextAndTag；失败时为空字符串。
//   - error：keyHex 或 dataHex 解码失败、nonce 生成失败、密钥非法或 nonceLength 不匹配时返回错误。
func EncryptGCMHex(keyHex string, nonceLength int, dataHex string) (string, error) {
	// 声明返回值变量。
	var result string
	var err error

	// 将十六进制格式的密钥解码为字节数组。
	if key, errKey := hex.DecodeString(keyHex); nil != errKey {
		// 如果密钥解码失败，保存错误。
		err = errKey
	} else if data, errData := hex.DecodeString(dataHex); nil != errData {
		// 如果数据解码失败，保存错误。
		err = errData
	} else if tmpResult, errEncrypt := EncryptGCMNonceLength(key, nonceLength, data); nil != errEncrypt {
		// 如果加密失败，保存错误。
		err = errEncrypt
	} else {
		// 加密成功，将结果编码为十六进制格式。
		result = hex.EncodeToString(tmpResult)
	}

	// 返回加密结果和可能的错误。
	return result, err
}

// EncryptGCMNonceLength 生成指定长度的随机 nonce，并返回 nonce || ciphertextAndTag。
//
// 本函数通过 crypto/rand 生成 nonce，并以 nil AAD 调用 AES-GCM 加密。
// nonceLength 必须与 GCM nonce 长度一致，否则返回错误。
//
// 参数：
//   - key：AES 密钥字节切片，长度必须符合标准库 aes.NewCipher 要求。
//   - nonceLength：待生成 nonce 的字节长度，必须与 GCM nonce 长度一致。
//   - data：待加密的明文字节切片，可为空。
//
// 返回：
//   - []byte：按 nonce || ciphertextAndTag 组合后的加密结果；失败时为 nil。
//   - error：nonceLength 为负数、随机源读取失败、密钥非法或 nonceLength 与 GCM 要求不一致时返回错误。
func EncryptGCMNonceLength(key []byte, nonceLength int, data []byte) ([]byte, error) {
	// 声明返回值变量。
	var result []byte
	var err error

	// 生成指定长度的随机 nonce（混淆值）。
	if nonce, errNonce := kitbytes.GenerateNonce(nonceLength); nil != errNonce {
		// 如果 nonce 生成失败，保存错误。
		err = errNonce
	} else {
		// 使用生成的 nonce 进行 GCM 加密。
		result, err = EncryptGCM(key, nonce, data)
	}

	// 返回加密结果和可能的错误。
	return result, err
}

// EncryptGCM 使用给定 key 和 nonce 执行 AES-GCM 加密，并返回 nonce || ciphertextAndTag。
//
// 本函数不使用附加认证数据（AAD）。调用方必须保证同一 key 下 nonce 不复用，否则会破坏 GCM 的安全性。
// 返回切片由 append(nonce, ciphertextAndTag...) 生成，可能与 nonce 共享底层数组；调用方在使用返回值期间不应修改 nonce 的底层数组。
//
// 参数：
//   - key：AES 密钥字节切片，长度必须符合标准库 aes.NewCipher 要求。
//   - nonce：本次加密使用的 nonce，长度必须与当前 GCM 实例要求一致。
//   - data：待加密的明文字节切片，可为空。
//
// 返回：
//   - []byte：按 nonce || ciphertextAndTag 组合后的加密结果；失败时为 nil。
//   - error：密钥非法或 nonce 长度不匹配时返回错误。
func EncryptGCM(key, nonce, data []byte) ([]byte, error) {
	// 声明返回值变量。
	var result []byte
	var err error

	// 使用密钥创建 AES 密码块。
	if block, errBlock := aes.NewCipher(key); nil != errBlock {
		// 如果密码块创建失败，保存错误。
		err = errBlock
	} else if aead, errAead := cipher.NewGCM(block); nil != errAead {
		// 如果 GCM 认证加密模式创建失败，保存错误。
		err = errAead
	} else if len(nonce) != aead.NonceSize() {
		// 如果 nonce 长度不符合 GCM 要求，返回错误，避免 Seal 触发 panic。
		err = fmt.Errorf("invalid nonce length: got %d, want %d", len(nonce), aead.NonceSize())
	} else {
		// 使用 GCM 模式加密数据，nil 表示不使用附加认证数据（AAD）。
		tmpResult := aead.Seal(nil, nonce, data, nil)
		// 将 nonce 拼接在加密结果前面，以便解密时使用。
		result = append(nonce, tmpResult...)
	}

	// 返回加密结果和可能的错误。
	return result, err
}

// DecryptStringGCMBase64 解码 Base64 组合密文，并返回 nonce 字符串和明文字符串。
//
// dataBase64 必须是 EncryptStringGCMBase64 返回的 nonce || ciphertextAndTag 的 Base64 编码。
// 本函数不使用 AAD，返回的 nonce 直接按原始 nonce 字节转换为 string。
//
// 参数：
//   - keyBase64：Base64 格式的 AES 密钥，解码后长度必须符合 aes.NewCipher 要求。
//   - nonceLength：dataBase64 解码后前缀中 nonce 的字节长度，必须与 GCM nonce 长度一致。
//   - dataBase64：Base64 编码的 nonce || ciphertextAndTag 组合密文。
//
// 返回：
//   - string：从组合密文前缀提取的 nonce 原始字节字符串；失败时为空字符串。
//   - string：解密得到的明文字符串；失败时为空字符串。
//   - error：keyBase64 或 dataBase64 解码失败、nonceLength 非法、密文长度不足、密钥非法、nonce 长度不匹配或认证失败时返回错误。
func DecryptStringGCMBase64(keyBase64 string, nonceLength int, dataBase64 string) (string, string, error) {
	// 声明返回值变量。
	var nonce string
	var result string
	var err error

	// 将 Base64 格式的密钥解码为字节数组。
	if key, errKey := base64.StdEncoding.DecodeString(keyBase64); nil != errKey {
		// 如果密钥解码失败，保存错误。
		err = errKey
	} else if data, errData := base64.StdEncoding.DecodeString(dataBase64); nil != errData {
		// 如果数据解码失败，保存错误。
		err = errData
	} else if tmpNonce, tmpResult, errDecrypt := DecryptGCMNonceLength(key, nonceLength, data); nil != errDecrypt {
		// 如果解密失败，保存错误。
		err = errDecrypt
	} else {
		// 解密成功，将 nonce 和结果转换为字符串。
		nonce = string(tmpNonce)
		result = string(tmpResult)
	}

	// 返回 nonce 字符串、解密结果和可能的错误。
	return nonce, result, err
}

// DecryptStringGCMHex 解码 Hex 组合密文，并返回 nonce 字符串和明文字符串。
//
// dataHex 必须是 EncryptStringGCMHex 返回的 nonce || ciphertextAndTag 的 Hex 编码。
// 本函数不使用 AAD，返回的 nonce 直接按原始 nonce 字节转换为 string。
//
// 参数：
//   - keyHex：Hex 格式的 AES 密钥，解码后长度必须符合 aes.NewCipher 要求。
//   - nonceLength：dataHex 解码后前缀中 nonce 的字节长度，必须与 GCM nonce 长度一致。
//   - dataHex：Hex 编码的 nonce || ciphertextAndTag 组合密文，大小写均可。
//
// 返回：
//   - string：从组合密文前缀提取的 nonce 原始字节字符串；失败时为空字符串。
//   - string：解密得到的明文字符串；失败时为空字符串。
//   - error：keyHex 或 dataHex 解码失败、nonceLength 非法、密文长度不足、密钥非法、nonce 长度不匹配或认证失败时返回错误。
func DecryptStringGCMHex(keyHex string, nonceLength int, dataHex string) (string, string, error) {
	// 声明返回值变量。
	var nonce string
	var result string
	var err error

	// 将十六进制格式的密钥解码为字节数组。
	if key, errKey := hex.DecodeString(keyHex); nil != errKey {
		// 如果密钥解码失败，保存错误。
		err = errKey
	} else if data, errData := hex.DecodeString(dataHex); nil != errData {
		// 如果数据解码失败，保存错误。
		err = errData
	} else if tmpNonce, tmpResult, errDecrypt := DecryptGCMNonceLength(key, nonceLength, data); nil != errDecrypt {
		// 如果解密失败，保存错误。
		err = errDecrypt
	} else {
		// 解密成功，将 nonce 和结果转换为字符串。
		nonce = string(tmpNonce)
		result = string(tmpResult)
	}

	// 返回 nonce 字符串、解密结果和可能的错误。
	return nonce, result, err
}

// DecryptGCMBase64 解码 Base64 组合密文，并返回 Base64 编码的 nonce 和明文。
//
// dataBase64 必须是 EncryptGCMBase64 返回的 nonce || ciphertextAndTag 的 Base64 编码。
// 本函数不使用 AAD，认证通过后会分别 Base64 编码提取出的 nonce 和明文字节。
//
// 参数：
//   - keyBase64：Base64 格式的 AES 密钥，解码后长度必须符合 aes.NewCipher 要求。
//   - nonceLength：dataBase64 解码后前缀中 nonce 的字节长度，必须与 GCM nonce 长度一致。
//   - dataBase64：Base64 编码的 nonce || ciphertextAndTag 组合密文。
//
// 返回：
//   - string：Base64 编码的 nonce；失败时为空字符串。
//   - string：Base64 编码的明文字节；失败时为空字符串。
//   - error：keyBase64 或 dataBase64 解码失败、nonceLength 非法、密文长度不足、密钥非法、nonce 长度不匹配或认证失败时返回错误。
func DecryptGCMBase64(keyBase64 string, nonceLength int, dataBase64 string) (string, string, error) {
	// 声明返回值变量。
	var nonce string
	var result string
	var err error

	// 将 Base64 格式的密钥解码为字节数组。
	if key, errKey := base64.StdEncoding.DecodeString(keyBase64); nil != errKey {
		// 如果密钥解码失败，保存错误。
		err = errKey
	} else if data, errData := base64.StdEncoding.DecodeString(dataBase64); nil != errData {
		// 如果数据解码失败，保存错误。
		err = errData
	} else if tmpNonce, tmpResult, errDecrypt := DecryptGCMNonceLength(key, nonceLength, data); nil != errDecrypt {
		// 如果解密失败，保存错误。
		err = errDecrypt
	} else {
		// 解密成功，将 nonce 和结果编码为 Base64 格式。
		nonce = base64.StdEncoding.EncodeToString(tmpNonce)
		result = base64.StdEncoding.EncodeToString(tmpResult)
	}

	// 返回 nonce 字符串、解密结果和可能的错误。
	return nonce, result, err
}

// DecryptGCMHex 解码 Hex 组合密文，并返回大写 Hex 编码的 nonce 和明文。
//
// dataHex 必须是 EncryptGCMHex 返回的 nonce || ciphertextAndTag 的 Hex 编码，大小写均可。
// 本函数不使用 AAD，认证通过后会分别 Hex 编码提取出的 nonce 和明文字节，并转换为大写。
//
// 参数：
//   - keyHex：Hex 格式的 AES 密钥，解码后长度必须符合 aes.NewCipher 要求。
//   - nonceLength：dataHex 解码后前缀中 nonce 的字节长度，必须与 GCM nonce 长度一致。
//   - dataHex：Hex 编码的 nonce || ciphertextAndTag 组合密文。
//
// 返回：
//   - string：大写 Hex 编码的 nonce；失败时为空字符串。
//   - string：大写 Hex 编码的明文字节；失败时为空字符串。
//   - error：keyHex 或 dataHex 解码失败、nonceLength 非法、密文长度不足、密钥非法、nonce 长度不匹配或认证失败时返回错误。
func DecryptGCMHex(keyHex string, nonceLength int, dataHex string) (string, string, error) {
	// 声明返回值变量。
	var nonce string
	var result string
	var err error

	// 将十六进制格式的密钥解码为字节数组。
	if key, errKey := hex.DecodeString(keyHex); nil != errKey {
		// 如果密钥解码失败，保存错误。
		err = errKey
	} else if data, errData := hex.DecodeString(dataHex); nil != errData {
		// 如果数据解码失败，保存错误。
		err = errData
	} else if tmpNonce, tmpResult, errDecrypt := DecryptGCMNonceLength(key, nonceLength, data); nil != errDecrypt {
		// 如果解密失败，保存错误。
		err = errDecrypt
	} else {
		// 解密成功，将 nonce 和结果编码为十六进制格式。
		nonce = hex.EncodeToString(tmpNonce)
		result = hex.EncodeToString(tmpResult)

		// 将十六进制字符串转换为大写。
		nonce = strings.ToUpper(nonce)
		result = strings.ToUpper(result)
	}

	// 返回 nonce 字符串、解密结果和可能的错误。
	return nonce, result, err
}

// DecryptGCMNonceLength 按给定 nonceLength 解析 nonce || ciphertextAndTag，并执行 AES-GCM 解密。
//
// data 必须以前缀 nonce 开头，且总长度必须大于 nonceLength。本函数不使用 AAD。
// 返回的 nonce 是 data 的前缀切片，会与 data 共享底层数组。
//
// 参数：
//   - key：AES 密钥字节切片，长度必须符合标准库 aes.NewCipher 要求。
//   - nonceLength：data 前缀中 nonce 的字节长度，必须大于等于 0 且与 GCM nonce 长度一致。
//   - data：按 nonce || ciphertextAndTag 组合的输入密文。
//
// 返回：
//   - []byte：从 data 前缀提取出的 nonce；nonceLength 为负数或 data 长度不足时为 nil，解密阶段失败时可能随 error 返回非 nil nonce。
//   - []byte：认证通过后解出的明文字节切片；失败时为 nil。
//   - error：nonceLength 为负数、data 长度不足、密钥非法、nonce 长度不匹配或认证失败时返回错误。
func DecryptGCMNonceLength(key []byte, nonceLength int, data []byte) ([]byte, []byte, error) {
	// 声明返回值变量。
	var nonce []byte
	var result []byte
	var err error

	// 检查 nonce 长度是否为有效的非负值，避免公开 API 因负数切片边界触发 panic。
	if nonceLength < 0 {
		err = fmt.Errorf("invalid nonce length: %d", nonceLength)
	} else if len(data) > nonceLength {
		// 从数据中提取 nonce 部分。
		nonce = data[:nonceLength]
		// 提取实际的加密数据部分。
		tmpData := data[nonceLength:]
		// 使用 nonce 和密钥解密数据。
		result, err = DecryptGCM(key, nonce, tmpData)
	} else {
		// 数据长度不足，返回错误。
		err = fmt.Errorf("数据长度不足，无法提取 nonce。")
	}

	// 返回 nonce、解密结果和可能的错误。
	return nonce, result, err
}

// DecryptGCM 使用给定 key、nonce 和 ciphertextAndTag 执行 AES-GCM 解密。
//
// data 参数不包含 nonce 前缀，本函数以 nil AAD 调用 GCM Open。
// 解密会验证认证标签；密钥、nonce、密文或标签任一不匹配都会返回错误。
//
// 参数：
//   - key：AES 密钥字节切片，长度必须符合标准库 aes.NewCipher 要求。
//   - nonce：与 data 对应的 GCM nonce，长度必须与当前 GCM 实例要求一致。
//   - data：不含 nonce 前缀的 ciphertextAndTag 字节切片。
//
// 返回：
//   - []byte：认证通过后解出的明文字节切片；失败时为 nil。
//   - error：密钥非法、nonce 长度不匹配或认证失败时返回错误。
func DecryptGCM(key, nonce, data []byte) ([]byte, error) {
	// 声明返回值变量。
	var result []byte
	var err error

	// 使用密钥创建 AES 密码块。
	if block, errBlock := aes.NewCipher(key); nil != errBlock {
		// 如果密码块创建失败，保存错误。
		err = errBlock
	} else if aead, errAead := cipher.NewGCM(block); nil != errAead {
		// 如果 GCM 认证加密模式创建失败，保存错误。
		err = errAead
	} else if len(nonce) != aead.NonceSize() {
		// 如果 nonce 长度不符合 GCM 要求，返回错误，避免 Open 触发 panic。
		err = fmt.Errorf("invalid nonce length: got %d, want %d", len(nonce), aead.NonceSize())
	} else if tmpResult, errOpen := aead.Open(nil, nonce, data, nil); nil != errOpen {
		// 如果解密或认证失败，保存错误。
		err = errOpen
	} else {
		// 解密成功，保存结果。
		result = tmpResult
	}

	// 返回解密结果和可能的错误。
	return result, err
}
