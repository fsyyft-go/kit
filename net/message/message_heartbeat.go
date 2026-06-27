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
	// HeartbeatMessage 表示携带心跳序列号的消息。
	HeartbeatMessage interface {
		// SerialNumber 返回心跳消息中的序列号。
		//
		// 参数：无。
		//
		// 返回：
		//   - uint64: 心跳消息中的序列号。
		SerialNumber() uint64
	}

	// heartbeatMessage 是 [HeartbeatMessage] 的默认实现。
	heartbeatMessage struct {
		messageType  MessageType // 消息类型。
		serialNumber uint64      // 心跳包序列号。
	}
)

// MessageType 返回消息类型。
//
// 参数：无。
//
// 返回：
//   - MessageType: 当前消息的协议类型。
func (m *heartbeatMessage) MessageType() MessageType {
	return m.messageType
}

// Pack 将心跳序列号编码为 8 字节大端序 payload。
//
// 参数：无。
//
// 返回：
//   - []byte: 按大端序编码后的心跳 payload。
//   - error: 编码失败或发生 panic 恢复时返回错误。
func (m *heartbeatMessage) Pack() (msg []byte, err error) {
	defer func() {
		if r := recover(); nil != r {
			err = cockroachdberrors.Newf("封包过程发生异常：%[1]v。", r)
		}
	}()

	buf := &bytes.Buffer{}
	if errWriteSerialNumber := binaryWrite(buf, binary.BigEndian, m.serialNumber); nil != errWriteSerialNumber {
		err = cockroachdberrors.Wrap(errWriteSerialNumber, "封包过程发生异常。")
	} else {
		msg = buf.Bytes()
	}

	return msg, err
}

// Unpack 从 payload 的前 8 字节还原心跳序列号。
//
// payload 少于 8 字节时返回错误；多余字节会被忽略。
//
// 参数：
//   - payload: 待解码的心跳消息 payload。
//
// 返回：
//   - error: 解码失败或发生 panic 恢复时返回错误。
func (m *heartbeatMessage) Unpack(payload []byte) (err error) {
	defer func() {
		if r := recover(); nil != r {
			err = cockroachdberrors.Newf("解包过程发生异常：%[1]v。", r)
		}
	}()

	buf := bytes.NewBuffer(payload)

	if errRead := binaryRead(buf, binary.BigEndian, &m.serialNumber); nil != errRead {
		err = cockroachdberrors.Wrap(errRead, "解包过程发生异常。")
	}

	return err
}

// SerialNumber 返回心跳消息中的序列号。
//
// 参数：无。
//
// 返回：
//   - uint64: 心跳消息中的序列号。
func (m *heartbeatMessage) SerialNumber() uint64 {
	return m.serialNumber
}

// NewHeartbeatMessage 创建心跳消息。
//
// 参数：
//   - serialNumber: 要写入消息的心跳序列号。
//
// 返回：
//   - *heartbeatMessage: 新创建的心跳消息实例。
func NewHeartbeatMessage(serialNumber uint64) *heartbeatMessage {
	m := &heartbeatMessage{
		messageType:  HeartbeatMessageType,
		serialNumber: serialNumber,
	}

	return m
}

// GenerateHeartbeatMessage 根据消息类型和 payload 生成心跳消息。
//
// messageType 必须等于 [HeartbeatMessageType]，payload 不能为空。
//
// 参数：
//   - messageType: 目标消息类型。
//   - payload: 待解码的心跳消息 payload。
//
// 返回：
//   - Message: 生成的心跳消息实例。
//   - error: messageType 不匹配、payload 为 nil 或解码失败时返回错误。
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
