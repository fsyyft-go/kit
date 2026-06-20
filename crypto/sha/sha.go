// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

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
//   - string：哈希后的十六进制字符串。
func SHA256HashStringWithoutError(source string) string {
	// 调用 SHA256HashString 获取哈希值并忽略兼容性错误返回值。
	result, _ := SHA256HashString(source)
	// 仅返回哈希值字符串。
	return result
}

// SHA256HashString 对输入字符串 message 进行 SHA256 哈希计算。
// 返回哈希后的十六进制字符串；错误返回值为兼容既有 API 保留，标准库 Sum 计算不会产生错误。
//
// 参数：
//   - message：待哈希的字符串。
//
// 返回值：
//   - string：哈希后的十六进制字符串。
//   - error：兼容既有 API 的错误返回值，固定为 nil。
func SHA256HashString(message string) (string, error) {
	// 直接使用标准库 Sum API，避免保留 hash.Write 不会触发的错误分支。
	bytes := sha256.Sum256([]byte(message))
	// 将字节切片编码为十六进制字符串。
	hashCode := hex.EncodeToString(bytes[:])
	// 返回哈希结果和兼容性错误信息。
	return hashCode, nil
}

// SHA1HashStringWithoutError 对输入字符串进行 SHA1 哈希计算。
//
// 参数：
//   - source：待哈希的字符串。
//
// 返回值：
//   - string：哈希后的十六进制字符串。
func SHA1HashStringWithoutError(source string) string {
	// 调用 SHA1HashString 获取哈希值并忽略兼容性错误返回值。
	result, _ := SHA1HashString(source)
	// 仅返回哈希值字符串。
	return result
}

// SHA1HashString 对输入字符串 message 进行 SHA1 哈希计算。
// 返回哈希后的十六进制字符串；错误返回值为兼容既有 API 保留，标准库 Sum 计算不会产生错误。
//
// 参数：
//   - message：待哈希的字符串。
//
// 返回值：
//   - string：哈希后的十六进制字符串。
//   - error：兼容既有 API 的错误返回值，固定为 nil。
func SHA1HashString(message string) (string, error) {
	// 直接使用标准库 Sum API，避免保留 hash.Write 不会触发的错误分支。
	bytes := sha1.Sum([]byte(message))
	// 将字节切片编码为十六进制字符串。
	hashCode := hex.EncodeToString(bytes[:])
	// 返回哈希结果和兼容性错误信息。
	return hashCode, nil
}
