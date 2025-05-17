// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import (
	"math"

	cockroachdbErrors "github.com/cockroachdb/errors"
)

var (
	// 断言 singleStringMessage 实现 Message 和 SingleStringMessage 接口。
	_ Message             = (*singleStringMessage)(nil)
	_ SingleStringMessage = (*singleStringMessage)(nil)
)

type (
	// SingleStringMessage 简单的字符串消息包接口。
	SingleStringMessage interface {
		// Message 返回字符串消息内容。
		//
		// 返回值：
		//   - string: 字符串消息内容。
		Message() string
	}

	// singleStringMessage 简单的字符串消息包，实现接口 Message 和 SingleStringMessage。
	singleStringMessage struct {
		messageType MessageType // 消息类型。
		message     string      // 字符串消息内容。
	}
)

// MessageType 返回消息类型。
//
// 返回值：
//   - uint16: 消息类型。
func (m *singleStringMessage) MessageType() MessageType {
	return m.messageType
}

// Pack 将字符串消息内容转换为 payload 字节数组。
//
// 返回值：
//   - []byte: 字符串消息的字节数组。
//   - error: 错误信息。
func (m *singleStringMessage) Pack() ([]byte, error) {
	var msg []byte
	var err error

	defer func() {
		if r := recover(); nil != r {
			err = cockroachdbErrors.Newf("封包过程发生异常：%[1]v。", r)
		}
	}()

	msg = []byte(m.message)

	// 如果 msg 长度超过 uint16 最大值，则返回错误。
	if len(msg) > math.MaxUint16 {
		msg = nil
		err = cockroachdbErrors.Newf("字符串消息长度 %[1]d 超过 uint16 最大值 %[2]d。", len(msg), math.MaxUint16)
	}

	return msg, err
}

// Unpack 从 payload 字节数组还原字符串消息内容。
//
// 参数：
//   - payload: 字符串消息的字节数组。
//
// 返回值：
//   - error: 错误信息。
func (m *singleStringMessage) Unpack(payload []byte) error {
	var err error

	defer func() {
		if r := recover(); nil != r {
			err = cockroachdbErrors.Newf("解包过程发生异常：%[1]v。", r)
		}
	}()

	m.message = string(payload)

	return err
}

// Message 返回字符串消息内容。
//
// 返回值：
//   - string: 字符串消息内容。
func (m *singleStringMessage) Message() string {
	return m.message
}

// NewSingleStringMessage 创建一个简单的字符串消息包。
//
// 参数：
//   - message: 字符串消息内容。
//
// 返回值：
//   - *singleStringMessage: 新建的字符串消息包。
func NewSingleStringMessage(message string) *singleStringMessage {
	m := &singleStringMessage{
		messageType: SingleStringMessageType,
		message:     message,
	}

	return m
}

// GenerateSingleStringMessage 生成简单的字符串消息包结构体。
//
// 参数：
//   - messageType: 消息类型。
//   - payload: 字符串消息的字节数组。
//
// 返回值：
//   - Message: 生成的字符串消息包。
//   - error: 错误信息。
func GenerateSingleStringMessage(messageType MessageType, payload []byte) (Message, error) {
	var m *singleStringMessage
	var err error

	if messageType != SingleStringMessageType {
		err = cockroachdbErrors.Newf("消息类型 %[1]d 与目标消息类型 %[2]d 不匹配。", messageType, SingleStringMessageType)
	} else if nil == payload {
		err = cockroachdbErrors.Newf("有效负载不能为空。")
	} else {
		m = &singleStringMessage{
			messageType: SingleStringMessageType,
		}

		err = m.Unpack(payload)
	}

	return m, err
}
