// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
// 包 sha 提供 SHA256 哈希算法相关的工具函数。
package sha

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
)

// SHA256HashStringWithoutError 对输入字符串进行 SHA256 哈希计算。
//
// 参数：
//   - source：待哈希的字符串。
//
// 返回值：
//   - string：哈希后的十六进制字符串，不返回错误信息，若哈希过程中发生错误，返回空字符串。
func SHA256HashStringWithoutError(source string) string {
	// 调用 SHA256HashString 获取哈希值和错误信息。
	result, _ := SHA256HashString(source)
	// 仅返回哈希值字符串。
	return result
}

// SHA256HashString 对输入字符串 message 进行 SHA256 哈希计算。
// 返回哈希后的十六进制字符串和可能出现的错误。
//
// 参数：
//   - message：待哈希的字符串。
//
// 返回值：
//   - string：哈希后的十六进制字符串。
//   - error：哈希过程中可能出现的错误。
func SHA256HashString(message string) (string, error) {
	// 定义哈希结果字符串。
	var hashCode string
	// 定义错误变量。
	var err error
	// 创建一个基于 SHA256 算法的 hash.Hash 接口对象。
	hash := sha256.New()
	// 将输入字符串转换为字节切片并写入哈希对象。
	if _, err = hash.Write([]byte(message)); err == nil {
		// 计算哈希值，返回字节切片。
		bytes := hash.Sum(nil)
		// 将字节切片编码为十六进制字符串。
		hashCode = hex.EncodeToString(bytes)
	} else {
		// 若写入数据时发生错误，返回空字符串。
		hashCode = ""
	}
	// 返回哈希结果和错误信息。
	return hashCode, err
}

// SHA1HashStringWithoutError 对输入字符串进行 SHA1 哈希计算。
//
// 参数：
//   - source：待哈希的字符串。
//
// 返回值：
//   - string：哈希后的十六进制字符串，不返回错误信息，若哈希过程中发生错误，返回空字符串。
func SHA1HashStringWithoutError(source string) string {
	// 调用 SHA1HashString 获取哈希值和错误信息。
	result, _ := SHA1HashString(source)
	// 仅返回哈希值字符串。
	return result
}

// SHA1HashString 对输入字符串 message 进行 SHA1 哈希计算。
// 返回哈希后的十六进制字符串和可能出现的错误。
//
// 参数：
//   - message：待哈希的字符串。
//
// 返回值：
//   - string：哈希后的十六进制字符串。
//   - error：哈希过程中可能出现的错误。
func SHA1HashString(message string) (string, error) {
	// 定义哈希结果字符串。
	var hashCode string
	// 定义错误变量。
	var err error
	// 创建一个基于 SHA1 算法的 hash.Hash 接口对象。
	hash := sha1.New()
	// 将输入字符串转换为字节切片并写入哈希对象。
	if _, err = hash.Write([]byte(message)); err == nil {
		// 计算哈希值，返回字节切片。
		bytes := hash.Sum(nil)
		// 将字节切片编码为十六进制字符串。
		hashCode = hex.EncodeToString(bytes)
	} else {
		// 若写入数据时发生错误，返回空字符串。
		hashCode = ""
	}
	// 返回哈希结果和错误信息。
	return hashCode, err
}
