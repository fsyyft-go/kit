// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package bytes 提供基于 crypto/rand 的随机字节生成工具。
//
// 本包面向需要密码学安全随机原始字节的场景，例如 nonce、IV、salt 或 token
// 原始材料生成。GenerateNonce 会按请求长度读取随机源；长度合法性、随机源错误、
// 协议要求的唯一性、重放防护和结果编码由调用方在使用处处理。
package bytes
