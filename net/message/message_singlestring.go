// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import (
	"math"

	cockroachdberrors "github.com/cockroachdb/errors"
)

var (
	// 断言 singleStringMessage 实现 Message 和 SingleStringMessage 接口。
	_ Message             = (*singleStringMessage)(nil)
	_ SingleStringMessage = (*singleStringMessage)(nil)
)

type (
	// SingleStringMessage 表示仅携带一个字符串 payload 的消息。
	SingleStringMessage interface {
		// Message 返回消息中的字符串内容。
		//
		// 参数：无。
		//
		// 返回：
		//   - string: 当前消息中的字符串内容。
		Message() string
	}

	// singleStringMessage 是 [SingleStringMessage] 的默认实现。
	singleStringMessage struct {
		messageType MessageType // 消息类型。
		message     string      // 字符串消息内容。
	}
)

// MessageType 返回消息类型。
//
// 参数：无。
//
// 返回：
//   - MessageType: 当前消息的协议类型。
func (m *singleStringMessage) MessageType() MessageType {
	return m.messageType
}

// Pack 将字符串内容编码为 payload。
//
// 编码后的 payload 长度不能超过 uint16 最大值，因为协议头只用 2 字节记录长度。
//
// 参数：无。
//
// 返回：
//   - []byte: 当前消息编码后的字符串 payload。
//   - error: payload 超过协议长度上限或发生 panic 恢复时返回错误。
func (m *singleStringMessage) Pack() (msg []byte, err error) {
	defer func() {
		if r := recover(); nil != r {
			err = cockroachdberrors.Newf("封包过程发生异常：%[1]v。", r)
		}
	}()

	msg = []byte(m.message)

	// 如果 msg 长度超过 uint16 最大值，则返回错误。
	if messageLength := len(msg); messageLength > math.MaxUint16 {
		msg = nil
		err = cockroachdberrors.Newf("字符串消息长度 %[1]d 超过 uint16 最大值 %[2]d。", messageLength, math.MaxUint16)
	}

	return msg, err
}

// Unpack 从 payload 还原字符串内容。
//
// 参数：
//   - payload: 待解码的字符串消息 payload。
//
// 返回：
//   - error: 当前实现仅在发生 panic 恢复时返回错误。
func (m *singleStringMessage) Unpack(payload []byte) (err error) {
	defer func() {
		if r := recover(); nil != r {
			err = cockroachdberrors.Newf("解包过程发生异常：%[1]v。", r)
		}
	}()

	m.message = string(payload)

	return err
}

// Message 返回消息中的字符串内容。
//
// 参数：无。
//
// 返回：
//   - string: 当前消息中的字符串内容。
func (m *singleStringMessage) Message() string {
	return m.message
}

// NewSingleStringMessage 创建单字符串消息。
//
// 参数：
//   - message: 要写入消息 payload 的字符串内容。
//
// 返回：
//   - *singleStringMessage: 新创建的单字符串消息实例。
func NewSingleStringMessage(message string) *singleStringMessage {
	m := &singleStringMessage{
		messageType: SingleStringMessageType,
		message:     message,
	}

	return m
}

// GenerateSingleStringMessage 根据消息类型和 payload 生成单字符串消息。
//
// messageType 必须等于 [SingleStringMessageType]；nil payload 会被拒绝，非 nil 的空 payload 表示合法空字符串。
//
// 参数：
//   - messageType: 目标消息类型。
//   - payload: 待解码的字符串消息 payload。
//
// 返回：
//   - Message: 生成的单字符串消息实例。
//   - error: messageType 不匹配、payload 为 nil 或解码失败时返回错误。
func GenerateSingleStringMessage(messageType MessageType, payload []byte) (Message, error) {
	var m *singleStringMessage
	var err error

	if messageType != SingleStringMessageType {
		err = cockroachdberrors.Newf("消息类型 %[1]d 与目标消息类型 %[2]d 不匹配。", messageType, SingleStringMessageType)
	} else if nil == payload {
		err = cockroachdberrors.Newf("有效负载不能为空。")
	} else {
		m = &singleStringMessage{
			messageType: SingleStringMessageType,
		}

		err = m.Unpack(payload)
	}

	return m, err
}
