// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import "encoding/binary"

type (
	// Message 定义协议消息的类型、封包和解包契约。
	//
	// Pack 和 Unpack 只处理 payload，不包含协议头中的 MessageType 和长度字段。
	Message interface {
		// MessageType 返回消息的协议类型。
		//
		// 参数：无。
		//
		// 返回：
		//   - MessageType: 当前消息的协议类型。
		MessageType() MessageType

		// Pack 将消息编码为 payload。
		//
		// 返回结果不包含协议头中的 MessageType 和长度字段。
		//
		// 参数：无。
		//
		// 返回：
		//   - []byte: 当前消息编码后的 payload。
		//   - error: 编码失败时返回错误。
		Pack() ([]byte, error)

		// Unpack 使用 payload 还原消息内容。
		//
		// payload 不包含协议头中的 MessageType 和长度字段。
		//
		// 参数：
		//   - payload: 需要解码的消息 payload。
		//
		// 返回：
		//   - error: 解码失败时返回错误。
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
	// Generator 根据消息类型和 payload 生成具体消息实例。
	Generator interface {
		// GenerateMessage 根据消息类型和 payload 生成消息实例。
		//
		// payload 不包含协议头中的 MessageType 和长度字段。
		//
		// 参数：
		//   - messageType: 目标消息类型。
		//   - payload: 待解码的消息 payload。
		//
		// 返回：
		//   - Message: 生成的消息实例。
		//   - error: 生成失败时返回错误。
		GenerateMessage(messageType MessageType, payload []byte) (Message, error)
	}

	// GenerateMessageFunc 适配普通函数为 [Generator] 实现。
	//
	// 适配函数收到的 payload 不包含协议头中的 MessageType 和长度字段。
	//
	// 参数：
	//   - messageType: 目标消息类型。
	//   - payload: 待解码的消息 payload。
	//
	// 返回：
	//   - Message: 生成的消息实例。
	//   - error: 生成失败时返回错误。
	GenerateMessageFunc func(MessageType, []byte) (Message, error)
)

// GenerateMessage 调用底层函数生成消息实例。
//
// 参数：
//   - messageType: 目标消息类型。
//   - payload: 待解码的消息 payload。
//
// 返回：
//   - Message: 生成的消息实例。
//   - error: 底层适配函数返回的错误。
func (f GenerateMessageFunc) GenerateMessage(messageType MessageType, payload []byte) (Message, error) {
	return f(messageType, payload)
}
