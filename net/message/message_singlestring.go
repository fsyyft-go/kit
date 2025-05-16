package message

import (
	cockroachdbErrors "github.com/cockroachdb/errors"
)

var (
	_ Message             = (*singleStringMessage)(nil)
	_ SingleStringMessage = (*singleStringMessage)(nil)
)

type (
	// SingleStringMessage 简单的字符串消息包。
	SingleStringMessage interface {
		// Message 字符串消息。
		Message() string
	}

	// singleStringMessage 简单的字符串消息包，实现接口 Message 和 SingleStringMessage。
	singleStringMessage struct {
		messageType uint16 // 消息类型；
		message     string // 字符串消息。
	}
)

// MessageType 消息类型；
func (m *singleStringMessage) MessageType() uint16 {
	return m.messageType
}

// Pack 封包；
//
// 返回消息对应的 payload 的字节数组表示形式（不包含消息类型和长度）。
func (m *singleStringMessage) Pack() ([]byte, error) {
	var msg []byte
	var err error

	defer func() {
		if r := recover(); nil != r {
			err = cockroachdbErrors.Newf("封包过程发生异常：%[1]v。", r)
		}
	}()

	msg = []byte(m.message)

	return msg, err
}

// Unpack 拆包；
//
// payload 消息对应的字节数组的表示形式（不包含消息类型和长度）。
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

// Message 字符串消息。
func (m *singleStringMessage) Message() string {
	return m.message
}

// NewSingleStringMessage 新创建一个简单的字符串消息包；
//
// message 字符串消息。
func NewSingleStringMessage(message string) *singleStringMessage {
	m := &singleStringMessage{
		messageType: SingleStringMessageType,
		message:     message,
	}

	return m
}

// GenerateSingleStringMessage 生成简单的字符串消息包结构体的方法 ；
//
// messageType 消息类型；
// message 字符串消息。
//
// 返回简单的字符串消息包。
func GenerateSingleStringMessage(messageType uint16, payload []byte) (Message, error) {
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
