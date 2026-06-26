// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package md5 提供字符串 MD5 摘要计算函数。
//
// 本包封装标准库 crypto/md5 的字符串输入场景，返回结果统一为小写十六进制摘要。
// HashString 会保留底层写入错误，HashStringWithoutError 在兼容只需要摘要字符串的场景中忽略该错误，
// 并以空字符串表示失败。
//
// MD5 不具备抗碰撞安全性，仅适用于历史协议兼容、非安全校验或普通散列场景；
// 密码存储、签名和完整性保护等安全场景应选择更合适的算法。
package md5
