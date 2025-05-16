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
	var advance int
	var token []byte
	var err error

	defer func() {
		if r := recover(); nil != r {
			err = cockroachdberrors.Newf("解包过程发生异常：$[1]v。", r)
		}
	}()

	if atEOF {
		// 输入流已结束，返回 io.EOF。
		err = io.EOF
	} else {
		// 至少需要 4 字节（2 字节类型 + 2 字节长度）才能判断消息包长度。
		if len(data) > 4 {
			messageLength := uint16(0)

			// 读取消息长度字段（第 3-4 字节）。
			if errReadLength := binary.Read(bytes.NewReader(data[2:4]), binary.BigEndian, &messageLength); nil != errReadLength {
				err = errReadLength
			} else if int(messageLength+4) <= len(data) {
				// 如果缓冲区足够长，返回完整消息包。
				advance = int(messageLength + 4)
				token = data[:int(messageLength+4)]
			}
		}
	}

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
