// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import "encoding/binary"

// Message 消息包接口，定义消息类型、封包与拆包方法。
type (
	Message interface {
		// MessageType 返回消息类型。
		//
		// 返回值：
		//   - MessageType: 消息类型。
		MessageType() MessageType

		// Pack 将消息内容转换为 payload 字节数组（不包含消息类型和长度）。
		//
		// 返回值：
		//   - []byte: 消息内容的字节数组。
		//   - error: 错误信息。
		Pack() ([]byte, error)

		// Unpack 将 payload 字节数组（不包含消息类型和长度）还原为消息内容。
		//
		// 参数：
		//   - payload: 消息内容的字节数组。
		//
		// 返回值：
		//   - error: 错误信息。
		Unpack(payload []byte) error
	}
)

var (
	// 断言 GenerateMessageFunc 实现 Generator 接口。
	_ Generator = (GenerateMessageFunc)(nil)
)

var (
	// binaryRead 指向二进制读取函数，生产路径默认使用 encoding/binary.Read。
	//
	// 该变量作为包内测试 seam，用于在单元测试中注入读取错误，覆盖标准库在正常输入下难以触发的错误分支。
	binaryRead = binary.Read
	// binaryWrite 指向二进制写入函数，生产路径默认使用 encoding/binary.Write。
	//
	// 该变量作为包内测试 seam，用于在单元测试中注入写入错误，覆盖 bytes.Buffer 在正常输入下难以触发的错误分支。
	binaryWrite = binary.Write
)

type (
	// Generator 消息生成器接口，定义生成消息包结构体的方法。
	Generator interface {
		// GenerateMessage 生成消息包结构体。
		//
		// 参数：
		//   - messageType: 消息类型。
		//   - payload: 消息对应的字节数组的表示形式（不包含消息类型和长度）。
		//
		// 返回值：
		//   - Message: 生成的消息包。
		//   - error: 错误信息。
		GenerateMessage(messageType MessageType, payload []byte) (Message, error)
	}

	// GenerateMessageFunc 生成消息包结构体的方法类型，实现 Generator 接口。
	//
	// 参数：
	//   - messageType: 消息类型。
	//   - payload: 消息对应的字节数组的表示形式（不包含消息类型和长度）。
	//
	// 返回值：
	//   - Message: 生成的消息包。
	//   - error: 错误信息。
	GenerateMessageFunc func(MessageType, []byte) (Message, error)
)

// GenerateMessage 调用函数生成消息包结构体。
//
// 参数：
//   - messageType: 消息类型。
//   - payload: 消息对应的字节数组的表示形式（不包含消息类型和长度）。
//
// 返回值：
//   - Message: 生成的消息包。
//   - error: 错误信息。
func (f GenerateMessageFunc) GenerateMessage(messageType MessageType, payload []byte) (Message, error) {
	return f(messageType, payload)
}
