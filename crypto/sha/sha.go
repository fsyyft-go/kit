// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package sha

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
)

// SHA256HashStringWithoutError 返回 source 的 SHA256 摘要小写十六进制字符串。
//
// 参数：
//   - source: 待计算摘要的字符串，函数按其原始字节序列处理，不做字符集转换或规范化。
//
// 返回：
//   - string: source 对应的 SHA256 摘要小写十六进制编码。
func SHA256HashStringWithoutError(source string) string {
	// 调用 SHA256HashString 获取哈希值并忽略兼容性错误返回值；当前实现固定返回 nil。
	result, _ := SHA256HashString(source)
	// 仅返回哈希值字符串，保持无错误返回版本只暴露摘要字符串的历史签名。
	return result
}

// SHA256HashString 返回 message 的 SHA256 摘要小写十六进制字符串。
//
// 参数：
//   - message: 待计算摘要的字符串，函数按其原始字节序列处理，不做字符集转换或规范化。
//
// 返回：
//   - string: message 对应的 SHA256 摘要小写十六进制编码。
//   - error: 兼容既有 API 的错误返回值；当前实现使用 crypto/sha256.Sum256，固定返回 nil。
func SHA256HashString(message string) (string, error) {
	// 直接使用标准库 Sum API，避免保留 hash.Write 不会触发的错误分支。
	bytes := sha256.Sum256([]byte(message))
	// 十六进制编码使用小写字母，与 encoding/hex.EncodeToString 保持一致。
	hashCode := hex.EncodeToString(bytes[:])
	// 错误返回值保留给依赖双返回值签名的调用方，当前没有失败路径。
	return hashCode, nil
}

// SHA1HashStringWithoutError 返回 source 的 SHA1 摘要小写十六进制字符串。
//
// 参数：
//   - source: 待计算摘要的字符串，函数按其原始字节序列处理，不做字符集转换或规范化。
//
// 返回：
//   - string: source 对应的 SHA1 摘要小写十六进制编码。
func SHA1HashStringWithoutError(source string) string {
	// 调用 SHA1HashString 获取哈希值并忽略兼容性错误返回值；当前实现固定返回 nil。
	result, _ := SHA1HashString(source)
	// 仅返回哈希值字符串，保持无错误返回版本只暴露摘要字符串的历史签名。
	return result
}

// SHA1HashString 返回 message 的 SHA1 摘要小写十六进制字符串。
//
// 参数：
//   - message: 待计算摘要的字符串，函数按其原始字节序列处理，不做字符集转换或规范化。
//
// 返回：
//   - string: message 对应的 SHA1 摘要小写十六进制编码。
//   - error: 兼容既有 API 的错误返回值；当前实现使用 crypto/sha1.Sum，固定返回 nil。
func SHA1HashString(message string) (string, error) {
	// 直接使用标准库 Sum API，避免保留 hash.Write 不会触发的错误分支。
	bytes := sha1.Sum([]byte(message))
	// 十六进制编码使用小写字母，与 encoding/hex.EncodeToString 保持一致。
	hashCode := hex.EncodeToString(bytes[:])
	// 错误返回值保留给依赖双返回值签名的调用方，当前没有失败路径。
	return hashCode, nil
}
