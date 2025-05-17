// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import (
	"bytes"
	"encoding/binary"

	cockroachdberrors "github.com/cockroachdb/errors"
)

var (
	// 断言 heartbeatMessage 实现 Message 和 HeartbeatMessage 接口。
	_ Message          = (*heartbeatMessage)(nil)
	_ HeartbeatMessage = (*heartbeatMessage)(nil)
)

type (
	// HeartbeatMessage 心跳消息包接口。
	HeartbeatMessage interface {
		// SerialNumber 返回心跳包序列号。
		//
		// 返回值：
		//   - uint64: 心跳包序列号。
		SerialNumber() uint64
	}

	// heartbeatMessage 心跳消息包，实现接口 Message 和 HeartbeatMessage。
	heartbeatMessage struct {
		messageType  MessageType // 消息类型。
		serialNumber uint64      // 心跳包序列号。
	}
)

// MessageType 返回消息类型。
//
// 返回值：
//   - uint16: 消息类型。
func (m *heartbeatMessage) MessageType() MessageType {
	return m.messageType
}

// Pack 将心跳包序列号转换为 payload 字节数组。
//
// 返回值：
//   - []byte: 心跳包序列号的字节数组。
//   - error: 错误信息。
func (m *heartbeatMessage) Pack() ([]byte, error) {
	var msg []byte
	var err error

	defer func() {
		if r := recover(); nil != r {
			err = cockroachdberrors.Newf("封包过程发生异常：%[1]v。", r)
		}
	}()

	buf := &bytes.Buffer{}
	if errWriteSerialNumber := binary.Write(buf, binary.BigEndian, m.serialNumber); nil != errWriteSerialNumber {
		err = cockroachdberrors.Wrap(errWriteSerialNumber, "封包过程发生异常。")
	} else {
		msg = buf.Bytes()
	}

	return msg, err
}

// Unpack 从 payload 字节数组还原心跳包序列号。
//
// 参数：
//   - payload: 心跳包序列号的字节数组。
//
// 返回值：
//   - error: 错误信息。
func (m *heartbeatMessage) Unpack(payload []byte) error {
	var err error

	defer func() {
		if r := recover(); nil != r {
			err = cockroachdberrors.Newf("解包过程发生异常：%[1]v。", r)
		}
	}()

	buf := bytes.NewBuffer(payload)

	if errRead := binary.Read(buf, binary.BigEndian, &m.serialNumber); nil != errRead {
		err = cockroachdberrors.Wrap(errRead, "解包过程发生异常。")
	}

	return err
}

// SerialNumber 返回心跳包序列号。
//
// 返回值：
//   - uint64: 心跳包序列号。
func (m *heartbeatMessage) SerialNumber() uint64 {
	return m.serialNumber
}

// NewHeartbeatMessage 创建一个心跳消息包。
//
// 参数：
//   - serialNumber: 心跳包序列号。
//
// 返回值：
//   - *heartbeatMessage: 新建的心跳消息包。
func NewHeartbeatMessage(serialNumber uint64) *heartbeatMessage {
	m := &heartbeatMessage{
		messageType:  HeartbeatMessageType,
		serialNumber: serialNumber,
	}

	return m
}

// GenerateHeartbeatMessage 生成心跳消息包结构体。
//
// 参数：
//   - messageType: 消息类型。
//   - payload: 心跳包序列号的字节数组。
//
// 返回值：
//   - Message: 生成的心跳消息包。
//   - error: 错误信息。
func GenerateHeartbeatMessage(messageType MessageType, payload []byte) (Message, error) {
	var m *heartbeatMessage
	var err error

	if messageType != HeartbeatMessageType {
		err = cockroachdberrors.Newf("消息类型 %[1]d 与目标消息类型 %[2]d 不匹配。", messageType, HeartbeatMessageType)
	} else if nil == payload {
		err = cockroachdberrors.Newf("有效负载不能为空。")
	} else {
		m = &heartbeatMessage{
			messageType: HeartbeatMessageType,
		}

		err = m.Unpack(payload)
	}

	return m, err
}
