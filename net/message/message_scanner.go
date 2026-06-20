// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
)

const (
	// messageHeaderLength 表示消息类型和 payload 长度字段占用的固定头部长度。
	messageHeaderLength = 4
	// maxMessagePacketLength 表示协议允许的最大完整消息包长度，即 4 字节头部加 uint16 最大 payload。
	maxMessagePacketLength = messageHeaderLength + 1<<16 - 1
)

// scanMessage 供 [bufio.Scanner] 按自定义协议分割完整消息包。
//
// 参数：
//   - data: 当前缓冲区中的原始字节数据。
//   - atEOF: 是否已经到达输入流末尾。
//
// 返回：
//   - int: 已消费的字节数。
//   - []byte: 当前解析出的完整消息包；数据不足时返回 nil。
//   - error: 读取长度字段失败时返回错误。
func scanMessage(data []byte, atEOF bool) (int, []byte, error) {
	// advance：已消费的字节数。
	var advance int
	// token：完整消息包。
	var token []byte
	// err：错误信息。
	var err error

	// 1. 优先检查数据长度，至少需要 4 字节（2 字节类型 + 2 字节长度）才能判断消息包长度。
	if len(data) >= messageHeaderLength {
		messageLength := uint16(0)

		// 2. 读取消息长度字段（第 3-4 字节），采用大端序。
		if errReadLength := binaryRead(bytes.NewReader(data[2:messageHeaderLength]), binary.BigEndian, &messageLength); nil != errReadLength {
			// 长度字段解包失败，返回错误。
			err = errReadLength
		} else if packetLength := int(messageLength) + messageHeaderLength; packetLength <= len(data) {
			// 3. 即使 atEOF 为 true，也必须先返回已经完整到达的最后一个消息包。
			advance = packetLength
			token = data[:packetLength]
		}
		// 若数据不足完整包长度，Scanner 会自动读取更多数据后重试。
	}
	// 若数据不足 4 字节，Scanner 会自动读取更多数据后重试。

	if nil == token && atEOF {
		// 官方推荐 SplitFunc 在 atEOF 时，如果没有 token，返回 (0, nil, nil)，而不是返回 io.EOF。
		// 返回 io.EOF 会被 Scanner 视为错误，导致 scanner.Err() 返回该错误。
		return 0, nil, nil
	}

	// 返回：已消费字节数、完整消息包（或 nil）、错误信息（或 nil）。
	return advance, token, err
}

// NewScanner 创建按本包协议拆分消息包的 [bufio.Scanner]。
//
// 返回的 Scanner 使用 [scanMessage] 作为 SplitFunc，并将 token 缓冲区上限设置为协议允许的最大完整包长度。
//
// 参数：
//   - r: 提供协议字节流的输入源。
//
// 返回：
//   - *bufio.Scanner: 按本包协议拆分消息包的 Scanner。
func NewScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, messageHeaderLength), maxMessagePacketLength)

	scanner.Split(scanMessage)

	return scanner
}
