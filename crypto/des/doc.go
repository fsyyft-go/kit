// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package des 提供基于 DES-CBC 与 PKCS7 padding 的兼容性加解密工具。
//
// 本包包含 PKCS7 padding 辅助函数、使用独立 IV 的 DES-CBC 加解密函数，
// 以及将 key 兼作 IV 的历史包装函数和字符串、十六进制辅助函数。
// 加密函数返回的密文不携带 IV、认证标签或 MAC；调用方需要自行管理 IV 传递、
// 完整性校验和密文存储格式。GetDefaultDESKey 返回历史兼容包装层复用的默认 key；
// 新代码不应把它视为安全默认配置。
// DES 以及“key 作为 IV”的用法都不适合新的安全设计；新代码应优先使用更现代的算法和随机独立 IV。
package des
