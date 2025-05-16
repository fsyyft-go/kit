package message

import (
	"bytes"
	"encoding/binary"

	cockroachdbErrors "github.com/cockroachdb/errors"
)

var (
	_ Message          = (*heartbeatMessage)(nil)
	_ HeartbeatMessage = (*heartbeatMessage)(nil)
)

type (
	// HeartbeatMessage 心跳消息包。
	HeartbeatMessage interface {
		// SerialNumber 心跳包序列号。
		SerialNumber() uint64
	}

	// heartbeatMessage 心跳消息包，实现接口 Message 和 HeartbeatMessage。
	heartbeatMessage struct {
		messageType  uint16 // 消息类型；
		serialNumber uint64 // 心跳包序列号。
	}
)

// MessageType 消息类型；
func (m *heartbeatMessage) MessageType() uint16 {
	return m.messageType
}

// Pack 封包；
//
// 返回消息对应的 payload 的字节数组表示形式（不包含消息类型和长度）。
func (m *heartbeatMessage) Pack() ([]byte, error) {
	var msg []byte
	var err error

	defer func() {
		if r := recover(); nil != r {
			err = cockroachdbErrors.Newf("封包过程发生异常：%[1]v。", r)
		}
	}()

	buf := &bytes.Buffer{}
	if errWriteSerialNumber := binary.Write(buf, binary.BigEndian, m.serialNumber); nil != errWriteSerialNumber {
		err = cockroachdbErrors.Wrap(errWriteSerialNumber, "封包过程发生异常。")
	} else {
		msg = buf.Bytes()
	}

	return msg, err
}

// Unpack 拆包；
//
// payload 消息对应的字节数组的表示形式（不包含消息类型和长度）。
func (m *heartbeatMessage) Unpack(payload []byte) error {
	var err error

	defer func() {
		if r := recover(); nil != r {
			err = cockroachdbErrors.Newf("解包过程发生异常：%[1]v。", r)
		}
	}()

	buf := bytes.NewBuffer(payload)

	if errRead := binary.Read(buf, binary.BigEndian, &m.serialNumber); nil != errRead {
		err = cockroachdbErrors.Wrap(errRead, "解包过程发生异常。")
	}

	return err
}

// SerialNumber 心跳包序列号。
func (m *heartbeatMessage) SerialNumber() uint64 {
	return m.serialNumber
}

// NewHeartbeatMessage 新创建一个心跳消息包；
//
// serialNumber 心跳包序列号；
func NewHeartbeatMessage(serialNumber uint64) *heartbeatMessage {
	m := &heartbeatMessage{
		messageType:  HeartbeatMessageType,
		serialNumber: serialNumber,
	}

	return m
}

// GenerateHeartbeatMessage 生成心跳消息包结构体的方法 ；
//
// messageType 消息类型；
// payload 消息对应的字节数组的表示形式（不包含消息类型和长度）；
//
// 返回心跳消息包。
func GenerateHeartbeatMessage(messageType uint16, payload []byte) (Message, error) {
	var m *heartbeatMessage
	var err error

	if messageType != HeartbeatMessageType {
		err = cockroachdbErrors.Newf("消息类型 %[1]d 与目标消息类型 %[2]d 不匹配。", messageType, HeartbeatMessageType)
	} else if nil == payload {
		err = cockroachdbErrors.Newf("有效负载不能为空。")
	} else {
		m = &heartbeatMessage{
			messageType: HeartbeatMessageType,
		}

		err = m.Unpack(payload)
	}

	return m, err
}
