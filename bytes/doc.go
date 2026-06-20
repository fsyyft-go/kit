// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package bytes 提供基于 crypto/rand 的随机字节生成工具。
//
// 当前导出 API 仅包含 GenerateNonce，用于生成指定长度的随机字节切片。
// 该函数适用于需要随机 nonce、IV 或 token 原始字节的场景；协议层面的唯一性、
// 重放防护和结果编码由调用方负责。
package bytes
