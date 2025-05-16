// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"

	cockroachdberrors "github.com/cockroachdb/errors"
)

// scanMessage 用于 bufio.Scanner 按照自定义协议分割消息包。
//
// 参数：
//   - data: 输入数据缓冲区。
//   - atEOF: 是否到达输入结尾。
//
// 返回值：
//   - int: 已消费字节数。
//   - []byte: 完整消息包。
//   - error: 错误信息。
func scanMessage(data []byte, atEOF bool) (int, []byte, error) {
	// advance：已消费的字节数。
	var advance int
	// token：完整消息包。
	var token []byte
	// err：错误信息。
	var err error

	// 使用 defer 捕获 panic，保证异常时返回详细错误信息。
	defer func() {
		if r := recover(); nil != r {
			// 捕获异常，返回 cockroachdb 格式的错误。
			err = cockroachdberrors.Newf("解包过程发生异常：%[1]v。", r)
		}
	}()

	// 1. 检查是否到达输入流结尾。
	if atEOF {
		// 如果输入流已结束，返回 io.EOF，通知 Scanner 停止扫描。
		err = io.EOF
	} else {
		// 2. 检查数据长度，至少需要 4 字节（2 字节类型 + 2 字节长度）才能判断消息包长度。
		if len(data) > 4 {
			messageLength := uint16(0)

			// 3. 读取消息长度字段（第 3-4 字节），采用大端序。
			if errReadLength := binary.Read(bytes.NewReader(data[2:4]), binary.BigEndian, &messageLength); nil != errReadLength {
				// 长度字段解包失败，返回错误。
				err = errReadLength
			} else if int(messageLength+4) <= len(data) {
				// 4. 判断缓冲区是否包含完整消息包（类型+长度+payload）。
				//    若足够，设置 advance 为完整包长度，token 为完整包字节切片。
				advance = int(messageLength + 4)
				token = data[:int(messageLength+4)]
			}
			// 若数据不足完整包长度，Scanner 会自动读取更多数据后重试。
		}
		// 若数据不足 4 字节，Scanner 会自动读取更多数据后重试。
	}

	// 返回：已消费字节数、完整消息包（或 nil）、错误信息（或 nil）。
	return advance, token, err
}

// NewScanner 创建自定义消息包分割的 bufio.Scanner。
//
// 参数：
//   - r: 输入流。
//
// 返回值：
//   - *bufio.Scanner: 自定义分割的 Scanner。
func NewScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)

	scanner.Split(scanMessage)

	return scanner
}
