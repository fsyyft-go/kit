// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package des 实现了 DES 加密算法相关的功能。
package des

import (
	"crypto/cipher"
	"crypto/des"
	"encoding/hex"
	"strings"
)

// EncryptStringCBCPkCS7PaddingStringHex 使用 CBC 模式、PKCS7 填充进行 DES 加密。
//
// 密钥使用 UTF-8 编码的字符串；
// 明文数据为 UTF-8 编码的字符串；
// 加密数据使用 16 进制字符串表示形式；
func EncryptStringCBCPkCS7PaddingStringHex(key, data string) (string, error) {
	keyHex := hex.EncodeToString([]byte(key))
	return EncryptStringCBCPkCS7PaddingHex(keyHex, data)
}

// EncryptStringCBCPkCS7PaddingHex 使用 CBC 模式、PKCS7 填充进行 DES 加密。
//
// 密钥使用 16 进制字符串表示形式；
// 明文数据为 UTF-8 编码的字符串；
// 加密数据使用 16 进制字符串表示形式；
func EncryptStringCBCPkCS7PaddingHex(keyHex, data string) (string, error) {
	var result string
	var err error

	if key, errKey := hex.DecodeString(keyHex); nil != errKey {
		err = errKey
	} else if tmpResult, errEncrypt := EncryptCBCPkCS7Padding(key, []byte(data)); nil != errEncrypt {
		err = errEncrypt
	} else {
		result = hex.EncodeToString(tmpResult)
		result = strings.ToUpper(result)
	}

	return result, err
}

// EncryptCBCPkCS7Padding 使用 CBC 模式、PKCS7 填充进行 DES 加密。
//
// key 与 iv 使用相同的值；
func EncryptCBCPkCS7Padding(key, data []byte) ([]byte, error) {
	return EncryptCBCPkCS7PaddingAloneIV(key, key, data)
}

// EncryptCBCPkCS7PaddingAloneIV 使用 CBC 模式、PKCS7 填充进行 DES 加密。
func EncryptCBCPkCS7PaddingAloneIV(key, iv, data []byte) ([]byte, error) {
	var result []byte
	var err error

	if block, errBlock := des.NewCipher(key); nil != errBlock { //nolint:gosec
		err = errBlock
	} else {
		dataPadded := PKCS7Padding(data, block.BlockSize())
		mode := cipher.NewCBCEncrypter(block, iv)
		result = make([]byte, len(dataPadded))
		mode.CryptBlocks(result, dataPadded)
	}

	return result, err
}

// DecryptStringCBCPkCS7PaddingStringHex 使用 CBC 模式、PKCS7 填充进行 DES 解密。
//
// 密钥使用 UTF-8 编码的字符串；
// 加密数据使用 16 进制字符串表示形式；
// 解密数据使用 UTF-8 编码；
func DecryptStringCBCPkCS7PaddingStringHex(key, dataHex string) (string, error) {
	keyHex := hex.EncodeToString([]byte(key))
	return DecryptStringCBCPkCS7PaddingHex(keyHex, dataHex)
}

// DecryptStringCBCPkCS7PaddingHex 使用 CBC 模式、PKCS7 填充进行 DES 解密。
//
// 密钥使用 16 进制字符串表示形式；
// 加密数据使用 16 进制字符串表示形式；
// 解密数据为 UTF-8 编码的字符串；
func DecryptStringCBCPkCS7PaddingHex(keyHex, dataHex string) (string, error) {
	var result string
	var err error

	if key, errKey := hex.DecodeString(keyHex); nil != errKey {
		err = errKey
	} else if data, errData := hex.DecodeString(dataHex); nil != errData {
		err = errData
	} else if tmpResult, errDecrypt := DecryptCBCPkCS7Padding(key, data); nil != errDecrypt {
		err = errDecrypt
	} else {
		result = string(tmpResult)
	}

	return result, err
}

// DecryptCBCPkCS7Padding 使用 CBC 模式、PKCS7 填充进行 DES 解密。
//
// key 与 iv 使用相同的值；
func DecryptCBCPkCS7Padding(key, data []byte) ([]byte, error) {
	return DecryptCBCPkCS7PaddingAloneIV(key, key, data)
}

// DecryptCBCPkCS7PaddingAloneIV 使用 CBC 模式、PKCS7 填充进行 DES 解密。
func DecryptCBCPkCS7PaddingAloneIV(key, iv, data []byte) ([]byte, error) {
	var result []byte
	var err error

	if block, errBlock := des.NewCipher(key); nil != errBlock { //nolint:gosec
		err = errBlock
	} else {
		mode := cipher.NewCBCDecrypter(block, iv)
		dataPadded := make([]byte, len(data))
		mode.CryptBlocks(dataPadded, data)
		result = PKCS7UnPadding(dataPadded)
	}

	return result, err
}
