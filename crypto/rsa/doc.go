// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package rsa 提供 RSA-OAEP、PEM 转换，以及若干历史兼容 RSA 包装函数。
//
// ConvertPrivateKey 和 ConvertPubKey 负责在 PEM 与标准库 RSA key 类型之间转换。
// 新代码应优先使用 EncryptPubKeyOAEP、EncryptPublicKeyOAEP、DecryptPrivKeyOAEP 和
// DecryptPrivateKeyOAEP。包内仍保留 PKCS#1 v1.5 encryption 与“私钥加密/公钥解密”
// 兼容函数，用于历史协议或旧数据格式迁移。
// 这些兼容函数不会替调用方补充哈希、验签、分块或大消息处理策略；相关协议责任仍由调用方承担。
package rsa
