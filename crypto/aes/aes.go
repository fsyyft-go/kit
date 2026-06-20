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

// EncryptStringGCMBase64 已知混淆值字节数组长度和 Base64 格式的密钥和 UTF-8 编码的字符串明文，使用 GCM 模式加密，获得 Base64 格式的字符串密文。
//
// 参数：
//   - keyBase64：Base64 格式的密钥字符串。
//   - nonceLength：混淆值（nonce）的字节长度。
//   - data：UTF-8 编码的字符串明文。
//
// 返回值：
//   - string：Base64 格式的加密结果。
//   - error：加密过程中可能发生的错误，成功时为 nil。
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

// EncryptStringGCMHex 已知混淆值字节数组长度和 16 进制格式的密钥和 UTF-8 编码的字符串明文，使用 GCM 模式加密，获得 16 进制格式的字符串密文。
//
// 参数：
//   - keyHex：16 进制格式的密钥字符串。
//   - nonceLength：混淆值（nonce）的字节长度。
//   - data：UTF-8 编码的字符串明文。
//
// 返回值：
//   - string：16 进制格式的加密结果。
//   - error：加密过程中可能发生的错误，成功时为 nil。
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

// EncryptGCMBase64 已知混淆值字节数组长度和 Base64 格式的密钥和明文，使用 GCM 模式加密，获得 Base64 格式的字符串密文。
//
// 参数：
//   - keyBase64：Base64 格式的密钥字符串。
//   - nonceLength：混淆值（nonce）的字节长度。
//   - dataBase64：Base64 格式的明文数据。
//
// 返回值：
//   - string：Base64 格式的加密结果。
//   - error：加密过程中可能发生的错误，成功时为 nil。
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

// EncryptGCMHex 已知混淆值字节数组长度和 16 进制格式的密钥和明文，使用 GCM 模式加密，获得 16 进制格式的字符串密文。
//
// 参数：
//   - keyHex：16 进制格式的密钥字符串。
//   - nonceLength：混淆值（nonce）的字节长度。
//   - dataHex：16 进制格式的明文数据。
//
// 返回值：
//   - string：16 进制格式的加密结果。
//   - error：加密过程中可能发生的错误，成功时为 nil。
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
// 本函数内部使用 crypto/rand 生成 nonce，并以 nil AAD 调用 GCM 加密。
// nonceLength 必须与 cipher.NewGCM 要求的 nonce 长度一致，否则返回错误。
//
// 参数：
//   - key：AES 密钥字节切片，长度必须符合标准库 aes.NewCipher 要求。
//   - nonceLength：待生成 nonce 的字节长度，必须与 GCM 要求的 nonce 长度一致。
//   - data：待加密的明文字节切片。
//
// 返回值：
//   - []byte：按 nonce || ciphertextAndTag 组合后的加密结果。
//   - error：nonce 生成失败、密钥非法或 nonceLength 不匹配时返回错误。
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
// 本函数不使用附加认证数据（AAD）。调用方必须保证同一 key 下 nonce 不复用；
// 否则会破坏 GCM 的安全性。nonce 长度不匹配或密钥长度非法时返回错误。
//
// 参数：
//   - key：AES 密钥字节切片，长度必须符合标准库 aes.NewCipher 要求。
//   - nonce：本次加密使用的 nonce，长度必须与当前 GCM 实例要求一致。
//   - data：待加密的明文字节切片。
//
// 返回值：
//   - []byte：按 nonce || ciphertextAndTag 组合后的加密结果。
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

// DecryptStringGCMBase64 已知混淆值字节数组长度和 Base64 格式的密钥和密文，使用 GCM 模式解密，获得 UTF-8 编码的字符串明文。
//
// 参数：
//   - keyBase64：Base64 格式的密钥字符串。
//   - nonceLength：混淆值（nonce）的字节长度。
//   - dataBase64：Base64 格式的密文数据。
//
// 返回值：
//   - string：解密得到的 nonce 字符串。
//   - string：UTF-8 编码的解密结果。
//   - error：解密过程中可能发生的错误，成功时为 nil。
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

// DecryptStringGCMHex 已知混淆值字节数组长度和 16 进制格式的密钥和密文，使用 GCM 模式解密，获得 UTF-8 编码的字符串明文。
//
// 参数：
//   - keyHex：16 进制格式的密钥字符串。
//   - nonceLength：混淆值（nonce）的字节长度。
//   - dataHex：16 进制格式的密文数据。
//
// 返回值：
//   - string：解密得到的 nonce 字符串。
//   - string：UTF-8 编码的解密结果。
//   - error：解密过程中可能发生的错误，成功时为 nil。
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

// DecryptGCMBase64 已知混淆值字节数组长度和 Base64 格式的密钥和密文，使用 GCM 模式解密，获得 Base64 格式的字符串明文。
//
// 参数：
//   - keyBase64：Base64 格式的密钥字符串。
//   - nonceLength：混淆值（nonce）的字节长度。
//   - dataBase64：Base64 格式的密文数据。
//
// 返回值：
//   - string：Base64 格式的 nonce 字符串。
//   - string：Base64 格式的解密结果。
//   - error：解密过程中可能发生的错误，成功时为 nil。
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

// DecryptGCMHex 已知混淆值字节数组长度和 16 进制格式的密钥和密文，使用 GCM 模式解密，获得 16 进制格式的字符串明文。
//
// 参数：
//   - keyHex：16 进制格式的密钥字符串。
//   - nonceLength：混淆值（nonce）的字节长度。
//   - dataHex：16 进制格式的密文数据。
//
// 返回值：
//   - string：16 进制格式的 nonce 字符串（大写）。
//   - string：16 进制格式的解密结果（大写）。
//   - error：解密过程中可能发生的错误，成功时为 nil。
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
// data 必须以前缀 nonce 开头，且总长度必须大于 nonceLength。
// 返回值中的第一个切片是提取出的 nonce，第二个切片是认证通过后的明文。
//
// 参数：
//   - key：AES 密钥字节切片，长度必须符合标准库 aes.NewCipher 要求。
//   - nonceLength：data 前缀中 nonce 的字节长度，必须与 GCM 要求一致。
//   - data：按 nonce || ciphertextAndTag 组合的输入密文。
//
// 返回值：
//   - []byte：从 data 前缀提取出的 nonce。
//   - []byte：认证通过后解出的明文字节切片。
//   - error：nonceLength 非法、data 长度不足、密钥非法、nonce 长度不匹配或认证失败时返回错误。
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
// data 参数不包含 nonce 前缀。本函数同样以 nil AAD 调用 GCM Open；
// nonce 长度不匹配、密钥非法或认证失败时返回错误。
//
// 参数：
//   - key：AES 密钥字节切片，长度必须符合标准库 aes.NewCipher 要求。
//   - nonce：与 data 对应的 GCM nonce，长度必须与当前 GCM 实例要求一致。
//   - data：不含 nonce 前缀的 ciphertextAndTag 字节切片。
//
// 返回值：
//   - []byte：认证通过后解出的明文字节切片。
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
