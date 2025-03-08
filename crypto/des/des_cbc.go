// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package des 实现了 DES 加密算法相关的功能。
package des

import (
	"crypto/cipher"
	"crypto/des"
	"encoding/hex"
	"fmt"
	"strings"
)

// EncryptStringCBCPkCS7PaddingStringHex 使用 CBC 模式、PKCS7 填充进行 DES 加密。
//
// 参数：
//   - key：UTF-8 编码的字符串密钥。
//   - data：UTF-8 编码的待加密字符串。
//
// 返回：
//   - string：16 进制表示的加密结果。
//   - error：加密过程中可能发生的错误。
func EncryptStringCBCPkCS7PaddingStringHex(key, data string) (string, error) {
	// 将 UTF-8 编码的密钥转换为 16 进制字符串。
	keyHex := hex.EncodeToString([]byte(key))
	// 调用 16 进制密钥版本的加密函数。
	return EncryptStringCBCPkCS7PaddingHex(keyHex, data)
}

// EncryptStringCBCPkCS7PaddingHex 使用 CBC 模式、PKCS7 填充进行 DES 加密。
//
// 参数：
//   - keyHex：16 进制字符串表示的密钥。
//   - data：UTF-8 编码的待加密字符串。
//
// 返回：
//   - string：16 进制表示的加密结果。
//   - error：加密过程中可能发生的错误。
func EncryptStringCBCPkCS7PaddingHex(keyHex, data string) (string, error) {
	var result string
	var err error

	// 将 16 进制密钥转换为字节切片。
	if key, errKey := hex.DecodeString(keyHex); nil != errKey {
		err = errKey
	} else if tmpResult, errEncrypt := EncryptCBCPkCS7Padding(key, []byte(data)); nil != errEncrypt {
		// 使用转换后的密钥进行加密。
		err = errEncrypt
	} else {
		// 将加密结果转换为大写的 16 进制字符串。
		result = hex.EncodeToString(tmpResult)
		result = strings.ToUpper(result)
	}

	return result, err
}

// EncryptCBCPkCS7Padding 使用 CBC 模式、PKCS7 填充进行 DES 加密。
//
// 参数：
//   - key：密钥字节切片（同时用作 IV）。
//   - data：待加密数据。
//
// 返回：
//   - []byte：加密后的数据。
//   - error：加密过程中可能发生的错误。
func EncryptCBCPkCS7Padding(key, data []byte) ([]byte, error) {
	// 使用相同的值作为密钥和 IV。
	return EncryptCBCPkCS7PaddingAloneIV(key, key, data)
}

// EncryptCBCPkCS7PaddingAloneIV 使用 CBC 模式、PKCS7 填充进行 DES 加密。
//
// 参数：
//   - key：密钥字节切片。
//   - iv：初始化向量。
//   - data：待加密数据。
//
// 返回：
//   - []byte：加密后的数据。
//   - error：加密过程中可能发生的错误。
func EncryptCBCPkCS7PaddingAloneIV(key, iv, data []byte) ([]byte, error) {
	var result []byte
	var err error

	// 创建 DES 加密块。
	if block, errBlock := des.NewCipher(key); nil != errBlock { //nolint:gosec
		err = errBlock
	} else if len(iv) != block.BlockSize() {
		// 验证 IV 长度是否等于块大小。
		err = fmt.Errorf("IV length must equal block size")
	} else {
		// 对数据进行 PKCS7 填充。
		dataPadded := PKCS7Padding(data, block.BlockSize())
		// 创建 CBC 加密器。
		mode := cipher.NewCBCEncrypter(block, iv)
		// 分配结果缓冲区。
		result = make([]byte, len(dataPadded))
		// 执行加密操作。
		mode.CryptBlocks(result, dataPadded)
	}

	return result, err
}

// DecryptStringCBCPkCS7PaddingStringHex 使用 CBC 模式、PKCS7 填充进行 DES 解密。
//
// 参数：
//   - key：UTF-8 编码的字符串密钥。
//   - dataHex：16 进制字符串表示的加密数据。
//
// 返回：
//   - string：UTF-8 编码的解密结果。
//   - error：解密过程中可能发生的错误。
func DecryptStringCBCPkCS7PaddingStringHex(key, dataHex string) (string, error) {
	// 将 UTF-8 编码的密钥转换为 16 进制字符串。
	keyHex := hex.EncodeToString([]byte(key))
	// 调用 16 进制密钥版本的解密函数。
	return DecryptStringCBCPkCS7PaddingHex(keyHex, dataHex)
}

// DecryptStringCBCPkCS7PaddingHex 使用 CBC 模式、PKCS7 填充进行 DES 解密。
//
// 参数：
//   - keyHex：16 进制字符串表示的密钥。
//   - dataHex：16 进制字符串表示的加密数据。
//
// 返回：
//   - string：UTF-8 编码的解密结果。
//   - error：解密过程中可能发生的错误。
func DecryptStringCBCPkCS7PaddingHex(keyHex, dataHex string) (string, error) {
	var result string
	var err error

	// 将 16 进制密钥转换为字节切片。
	if key, errKey := hex.DecodeString(keyHex); nil != errKey {
		err = errKey
	} else if data, errData := hex.DecodeString(dataHex); nil != errData {
		// 将 16 进制加密数据转换为字节切片。
		err = errData
	} else if tmpResult, errDecrypt := DecryptCBCPkCS7Padding(key, data); nil != errDecrypt {
		// 使用转换后的密钥和数据进行解密。
		err = errDecrypt
	} else {
		// 将解密结果转换为 UTF-8 字符串。
		result = string(tmpResult)
	}

	return result, err
}

// DecryptCBCPkCS7Padding 使用 CBC 模式、PKCS7 填充进行 DES 解密。
//
// 参数：
//   - key：密钥字节切片（同时用作 IV）。
//   - data：待解密数据。
//
// 返回：
//   - []byte：解密后的数据。
//   - error：解密过程中可能发生的错误。
func DecryptCBCPkCS7Padding(key, data []byte) ([]byte, error) {
	// 使用相同的值作为密钥和 IV。
	return DecryptCBCPkCS7PaddingAloneIV(key, key, data)
}

// DecryptCBCPkCS7PaddingAloneIV 使用 CBC 模式、PKCS7 填充进行 DES 解密。
//
// 参数：
//   - key：密钥字节切片。
//   - iv：初始化向量。
//   - data：待解密数据。
//
// 返回：
//   - []byte：解密后的数据。
//   - error：解密过程中可能发生的错误。
func DecryptCBCPkCS7PaddingAloneIV(key, iv, data []byte) ([]byte, error) {
	var result []byte
	var err error

	// 创建 DES 解密块。
	if block, errBlock := des.NewCipher(key); nil != errBlock { //nolint:gosec
		err = errBlock
	} else if len(iv) != block.BlockSize() {
		// 验证 IV 长度是否等于块大小。
		err = fmt.Errorf("IV length must equal block size")
	} else {
		// 创建 CBC 解密器。
		mode := cipher.NewCBCDecrypter(block, iv)
		// 分配结果缓冲区。
		dataPadded := make([]byte, len(data))
		// 执行解密操作。
		mode.CryptBlocks(dataPadded, data)
		// 移除 PKCS7 填充。
		result, err = PKCS7UnPadding(dataPadded)
	}

	return result, err
}
