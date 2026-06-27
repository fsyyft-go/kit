// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package otp 提供基于 TOTP 的一次性密码工具。
//
// NewOneTimePassword 会解码 Base32 secret，并应用 hash、digits、period、window、issuer
// 和 label 等可选项。生成出的实例可返回当前口令、窗口内可接受口令，并生成
// otpauth://totp/ URL；包级 VeryfyPassword 和 GenerateURL 是便捷包装。
// 本包不提供 HOTP、密钥生成、状态持久化或重放检测；重复校验后的消费语义由调用方负责。
// 当前实现也不会在构建实例时校验 period 必须大于 0，调用方需要保证相关选项有效，
// 否则后续生成或校验口令时可能因除零而 panic。
package otp
