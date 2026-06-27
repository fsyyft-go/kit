// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package bytes

import (
	"crypto/rand"
	"fmt"
	"io"
)

// GenerateNonce 返回指定长度的密码学安全随机字节切片。
//
// length 必须大于等于 0。返回值可用作 nonce、IV、salt 或 token 的原始随机材料；
// 上层协议要求的唯一性、长度、重放防护和编码格式由调用方保证。
//
// 参数：
//   - length: 指定生成随机字节的长度，必须大于等于 0；为 0 时返回非 nil 的空切片。
//
// 返回：
//   - []byte: length 为负数时返回 nil；否则返回长度为 length 的缓冲区。若随机源读取失败，
//     该缓冲区可能只包含部分已读取随机字节，调用方不应在 err 非 nil 时继续将其作为完整随机值使用。
//   - error: length 为负数时返回参数错误；底层随机源失败或短读时返回 io.ReadFull 产生的错误，
//     例如 io.ErrUnexpectedEOF 或随机源自身错误。
func GenerateNonce(length int) ([]byte, error) {
	// 先拒绝非法长度，避免 make 在负数长度上触发运行时 panic。
	if length < 0 {
		return nil, fmt.Errorf("长度不能为负数：%d", length)
	}

	// 预先分配目标长度的缓冲区；读取失败时仍会返回该缓冲区以保留已写入内容。
	var nonce = make([]byte, length)
	// 保留显式 err 变量，便于在读取随机源后统一返回底层错误。
	var err error

	// 使用加密安全的随机数生成器填充 nonce 切片。
	// io.ReadFull 确保读取足够的随机字节以填满整个切片，短读会转换为明确错误。
	_, err = io.ReadFull(rand.Reader, nonce)

	// 返回缓冲区和读取结果；调用方必须先检查 err 再使用 nonce。
	return nonce, err
}
