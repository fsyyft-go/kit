// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// 本测试文件针对 net/message/message_heartbeat.go 的所有核心功能进行单元测试。
// 设计思路：
//   - 采用表格驱动法，覆盖正常、异常、边界等多种情况。
//   - 断言使用 stretchr/testify 包，保证断言表达力和可读性。
//   - 注释详细，便于理解每个测试用例的目的和预期。
//   - 主要测试点包括：Pack/Unpack、构造函数、工厂函数、接口实现。
//
// 使用方法：
//   - 直接 go test 运行本文件。
//   - 可结合覆盖率工具查看测试完整性。
package message

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHeartbeatMessage_PackUnpack 测试 Pack 和 Unpack 的互操作性。
func TestHeartbeatMessage_PackUnpack(t *testing.T) {
	cases := []struct {
		name         string
		serialNumber uint64
		wantErr      bool
	}{
		{"正常序列号", 123456789, false},
		{"零序列号", 0, false},
		{"最大序列号", ^uint64(0), false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			msg := NewHeartbeatMessage(c.serialNumber)
			payload, err := msg.Pack()
			assert.Equal(t, c.wantErr, err != nil, "Pack 错误断言")
			if err != nil {
				return
			}

			msg2 := NewHeartbeatMessage(0)
			err = msg2.Unpack(payload)
			assert.Equal(t, c.wantErr, err != nil, "Unpack 错误断言")
			if err == nil {
				assert.Equal(t, c.serialNumber, msg2.SerialNumber(), "序列号应一致")
			}
		})
	}
}

// TestHeartbeatMessage_Pack_异常测试
func TestHeartbeatMessage_Pack_Error(t *testing.T) {
	// 由于 Pack 只涉及内存操作，理论上不会出错，除非内存溢出等极端情况。
	// 这里主要测试 recover 机制。
	msg := NewHeartbeatMessage(1)
	// 通过反射或其他手段制造异常场景较难，略过。
	// 只验证正常情况下 err 为 nil。
	payload, err := msg.Pack()
	assert.NoError(t, err, "正常情况下 Pack 不应报错")
	assert.Len(t, payload, 8, "payload 长度应为 8 字节")
}

// TestHeartbeatMessage_Unpack_异常测试
func TestHeartbeatMessage_Unpack_Error(t *testing.T) {
	cases := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{"空 payload", nil, true},
		{"长度不足", []byte{1, 2, 3}, true},
		{"正常 8 字节", make([]byte, 8), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			msg := NewHeartbeatMessage(0)
			err := msg.Unpack(c.input)
			assert.Equal(t, c.wantErr, err != nil, "Unpack 错误断言")
		})
	}
}

// TestNewHeartbeatMessage_属性测试
func TestNewHeartbeatMessage(t *testing.T) {
	serial := uint64(42)
	msg := NewHeartbeatMessage(serial)
	assert.Equal(t, HeartbeatMessageType, msg.MessageType(), "消息类型应为 HeartbeatMessageType")
	assert.Equal(t, serial, msg.SerialNumber(), "序列号应一致")
}

// TestGenerateHeartbeatMessage_工厂函数测试
func TestGenerateHeartbeatMessage(t *testing.T) {
	cases := []struct {
		name    string
		msgType MessageType
		sn      uint64
		wantErr bool
	}{
		{"类型匹配", HeartbeatMessageType, 123, false},
		{"类型不匹配", 0x99, 123, true},
		{"空 payload", HeartbeatMessageType, 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var payload []byte
			if !c.wantErr {
				payload = make([]byte, 8)
				binary.BigEndian.PutUint64(payload, c.sn)
			}
			msg, err := GenerateHeartbeatMessage(c.msgType, payload)
			assert.Equal(t, c.wantErr, err != nil, "工厂函数错误断言")
			if !c.wantErr {
				assert.NotNil(t, msg, "消息不应为 nil")
				hb, ok := msg.(HeartbeatMessage)
				assert.True(t, ok, "应实现 HeartbeatMessage 接口")
				assert.Equal(t, c.sn, hb.SerialNumber(), "序列号应一致")
			}
		})
	}
}
