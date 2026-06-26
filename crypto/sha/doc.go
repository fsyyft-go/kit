// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package sha 提供字符串 SHA1 和 SHA256 摘要的十六进制编码辅助函数。
//
// SHA1HashString 和 SHA256HashString 使用输入字符串的原始字节序列计算标准摘要，
// 返回小写十六进制编码结果，并保留一个兼容既有 API 的 error 返回值。当前实现直接
// 使用 crypto/sha1 和 crypto/sha256 的 Sum API，因此 error 始终为 nil。对应的
// WithoutError 变体仅返回摘要字符串，适合不需要双返回值签名的调用场景。
package sha
