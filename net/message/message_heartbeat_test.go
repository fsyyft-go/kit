// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import (
	"encoding/binary"
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHeartbeatMessage_PackUnpack 验证心跳消息的封包与解包契约。
//
// 该测试通过表驱动用例覆盖普通值、零值和最大值序列号，确保心跳 payload 始终使用 8 字节大端序编码。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestHeartbeatMessage_PackUnpack(t *testing.T) {
	tests := []struct {
		name             string
		description      string
		giveSerialNumber uint64
	}{
		{
			name:             "success/regular-serial-number",
			description:      "验证普通心跳序列号可以按大端序封包并被等价解包。",
			giveSerialNumber: 123456789,
		},
		{
			name:             "boundary/zero-serial-number",
			description:      "验证零值心跳序列号作为有效边界值被保留。",
			giveSerialNumber: 0,
		},
		{
			name:             "boundary/max-serial-number",
			description:      "验证 uint64 最大心跳序列号作为有效边界值被保留。",
			giveSerialNumber: math.MaxUint64,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			msg := NewHeartbeatMessage(tt.giveSerialNumber)
			payload, err := msg.Pack()

			require.NoError(t, err)
			require.Len(t, payload, 8)
			assert.Equal(t, tt.giveSerialNumber, binary.BigEndian.Uint64(payload))
			assert.Equal(t, HeartbeatMessageType, msg.MessageType())

			unpacked := NewHeartbeatMessage(0)
			require.NoError(t, unpacked.Unpack(payload))
			assert.Equal(t, HeartbeatMessageType, unpacked.MessageType())
			assert.Equal(t, tt.giveSerialNumber, unpacked.SerialNumber())
		})
	}
}

// TestHeartbeatMessage_Unpack 验证心跳消息解包对 payload 长度的处理。
//
// 该测试覆盖无效长度、有效长度和携带额外字节的 payload，确保解包错误与保留前 8 字节序列号的行为稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestHeartbeatMessage_Unpack(t *testing.T) {
	tests := []struct {
		name             string
		description      string
		givePayload      []byte
		wantSerialNumber uint64
		wantErr          bool
	}{
		{
			name:        "error/nil-payload",
			description: "验证 nil payload 无法解包为完整心跳序列号。",
			givePayload: nil,
			wantErr:     true,
		},
		{
			name:        "error/short-payload",
			description: "验证长度不足 8 字节的 payload 会返回解包错误。",
			givePayload: []byte{0x01, 0x02, 0x03},
			wantErr:     true,
		},
		{
			name:             "success/exact-length-payload",
			description:      "验证 8 字节 payload 可以解包为对应心跳序列号。",
			givePayload:      buildHeartbeatPayload(42),
			wantSerialNumber: 42,
		},
		{
			name:             "compatibility/extra-trailing-bytes",
			description:      "验证解包仅消费前 8 字节序列号并容忍后续扩展字节。",
			givePayload:      append(buildHeartbeatPayload(99), 0xAA, 0xBB),
			wantSerialNumber: 99,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			msg := NewHeartbeatMessage(0)
			err := msg.Unpack(tt.givePayload)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantSerialNumber, msg.SerialNumber())
		})
	}
}

// TestHeartbeatMessage_RecoverNilReceiver 验证心跳消息方法对 nil receiver panic 的错误化处理。
//
// 该测试覆盖 Pack 和 Unpack 内置的 recover 契约，确保异常场景不会向调用方传播 panic，而是返回可诊断错误。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestHeartbeatMessage_RecoverNilReceiver(t *testing.T) {
	tests := []struct {
		name        string
		description string
		act         func() error
	}{
		{
			name:        "panic/pack-nil-receiver",
			description: "验证 nil 心跳消息 receiver 调用 Pack 时会被 recover 并返回封包错误。",
			act: func() error {
				var msg *heartbeatMessage
				payload, err := msg.Pack()
				assert.Nil(t, payload)
				return err
			},
		},
		{
			name:        "panic/unpack-nil-receiver",
			description: "验证 nil 心跳消息 receiver 调用 Unpack 时会被 recover 并返回解包错误。",
			act: func() error {
				var msg *heartbeatMessage
				return msg.Unpack(buildHeartbeatPayload(1))
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			err := tt.act()

			require.Error(t, err)
			assert.Contains(t, err.Error(), "发生异常")
		})
	}
}

// TestHeartbeatMessage_PackBinaryWriteError 验证心跳消息封包会处理二进制写入错误。
//
// 该测试通过临时替换二进制写入函数覆盖常规 bytes.Buffer 下不可达的错误分支，确保错误会被包装为封包错误。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestHeartbeatMessage_PackBinaryWriteError(t *testing.T) {
	// 验证序列号写入失败时，Pack 返回封包错误而不是静默成功。
	replaceBinaryWrite(t, func(io.Writer, binary.ByteOrder, any) error {
		return errTestBinaryWrite
	})
	msg := NewHeartbeatMessage(1)

	payload, err := msg.Pack()

	require.Error(t, err)
	assert.Nil(t, payload)
	assert.Contains(t, err.Error(), "封包过程发生异常")
	require.ErrorIs(t, err, errTestBinaryWrite)
}

// TestNewHeartbeatMessage 验证心跳消息构造函数会设置固定消息类型并保留序列号。
//
// 该测试覆盖构造函数的公开属性契约，确保后续封包和工厂逻辑能够依赖正确的消息类型。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestNewHeartbeatMessage(t *testing.T) {
	// 验证构造函数会保留调用方传入的心跳序列号，并设置心跳消息类型。
	const giveSerialNumber = uint64(42)

	msg := NewHeartbeatMessage(giveSerialNumber)

	require.NotNil(t, msg)
	assert.Equal(t, HeartbeatMessageType, msg.MessageType())
	assert.Equal(t, giveSerialNumber, msg.SerialNumber())
}

// TestGenerateHeartbeatMessage 验证心跳消息生成器的类型和 payload 校验行为。
//
// 该测试通过表驱动用例覆盖生成成功、消息类型不匹配、nil payload 和无效 payload 长度，确保工厂调用时错误可诊断。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGenerateHeartbeatMessage(t *testing.T) {
	tests := []struct {
		name             string
		description      string
		giveMessageType  MessageType
		givePayload      []byte
		wantSerialNumber uint64
		wantMessage      bool
		wantErr          bool
	}{
		{
			name:             "success/valid-heartbeat-payload",
			description:      "验证匹配的心跳类型和 8 字节 payload 可以生成心跳消息。",
			giveMessageType:  HeartbeatMessageType,
			givePayload:      buildHeartbeatPayload(123),
			wantSerialNumber: 123,
			wantMessage:      true,
		},
		{
			name:            "error/type-mismatch",
			description:     "验证非心跳消息类型会被生成器拒绝。",
			giveMessageType: 0x99,
			givePayload:     buildHeartbeatPayload(123),
			wantErr:         true,
		},
		{
			name:            "error/nil-payload",
			description:     "验证 nil payload 会被生成器拒绝，避免生成不完整消息。",
			giveMessageType: HeartbeatMessageType,
			givePayload:     nil,
			wantErr:         true,
		},
		{
			name:            "error/short-payload",
			description:     "验证长度不足的 payload 会返回解包错误并暴露部分构造消息。",
			giveMessageType: HeartbeatMessageType,
			givePayload:     []byte{0x01, 0x02},
			wantMessage:     true,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			msg, err := GenerateHeartbeatMessage(tt.giveMessageType, tt.givePayload)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantMessage {
					assert.NotNil(t, msg)
				} else {
					assert.Nil(t, msg)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, msg)
			heartbeat, ok := msg.(HeartbeatMessage)
			require.True(t, ok)
			assert.Equal(t, HeartbeatMessageType, msg.MessageType())
			assert.Equal(t, tt.wantSerialNumber, heartbeat.SerialNumber())
		})
	}
}
