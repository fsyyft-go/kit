// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package aes 提供 AES-GCM 加解密及 Base64、Hex 字符串封装。
//
// 当前 API 围绕标准库 cipher.NewGCM 工作：EncryptGCM 和 DecryptGCM 处理单个 nonce
// 对应的密文片段，EncryptGCMNonceLength 和 DecryptGCMNonceLength 处理
// nonce || ciphertextAndTag 的组合格式。
// String、Base64 和 Hex 辅助函数负责在编码文本、原始字节和组合密文格式之间转换；
// 所有 GCM 调用都以 nil AAD 运行。
// 认证失败、nonce 长度不匹配或密钥长度非法时返回错误。本包不实现 AAD 管理、nonce
// 去重或重放检测；这些约束由调用方或上层协议负责。
package aes
