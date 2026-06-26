// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package rsa 提供 RSA-OAEP、PEM 密钥转换和历史兼容的 RSA 包装函数。
//
// 本包接受 PKCS#1 RSA PRIVATE KEY PEM 私钥和 PKIX PUBLIC KEY PEM 公钥，
// 可在 PEM 字节与标准库 RSA key 类型之间转换。OAEP 入口默认使用 SHA-256 和 nil label，
// 自定义 hash 或 label 时，加密与解密必须使用完全一致的参数。
//
// PKCS#1 v1.5 encryption 以及“私钥加密、公钥解密”函数仅为兼容历史密文格式、
// 旧协议或迁移场景保留，不提供分块、大消息处理、签名验签或协议级认证策略；
// 新代码应优先使用 OAEP 或标准库 crypto/rsa 的签名 API。
package rsa
