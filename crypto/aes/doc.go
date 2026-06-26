// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package aes 提供基于标准库 AES-GCM 的加解密工具。
//
// 本包围绕 cipher.NewGCM 封装 nonce 生成、nonce || ciphertextAndTag 组合格式，
// 以及字符串、Base64 和 Hex 编码转换。EncryptGCM 使用调用方提供的 nonce 生成
// nonce || ciphertextAndTag，DecryptGCM 使用调用方提供的 nonce 解密不含 nonce 前缀的
// ciphertextAndTag；EncryptGCMNonceLength 与 DecryptGCMNonceLength 处理带 nonce 前缀的
// 组合密文。所有 GCM 调用都以 nil AAD 运行。
//
// AES 密钥长度必须满足标准库 aes.NewCipher 的要求。默认 GCM nonce 长度来自
// cipher.AEAD.NonceSize，当前标准库 NewGCM 为 12 字节；同一密钥下 nonce 不得复用。
// 本包不负责 AAD 管理、nonce 去重或重放检测，这些安全约束由调用方或上层协议保证。
package aes
