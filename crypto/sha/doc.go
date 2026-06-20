// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package sha 提供字符串 SHA1 和 SHA256 摘要的十六进制编码辅助函数。
//
// SHA1HashString 和 SHA256HashString 返回标准库摘要结果及一个为兼容既有 API 保留的
// error 值。当前实现直接使用 crypto/sha1 和 crypto/sha256 的 Sum API，因此错误返回
// 始终为 nil。对应的 WithoutError 变体仅返回摘要字符串，适合不需要双返回值签名的调
// 用场景。
package sha
