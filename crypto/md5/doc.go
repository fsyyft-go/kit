// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package md5 提供字符串 MD5 摘要计算函数。
//
// HashString 返回小写十六进制摘要和底层写入错误，便于调用方显式处理异常。
// HashStringWithoutError 适用于已经接受“写入异常时返回空字符串”语义的简化场景。
// MD5 仅适用于兼容历史协议、非抗碰撞校验或普通散列场景，不适合密码学安全用途。
package md5
