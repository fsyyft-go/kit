// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSingleStringMessage_PackUnpack 验证简单字符串消息的封包与解包契约。
//
// 该测试通过表驱动用例覆盖普通文本、空文本、Unicode 文本和最大允许长度，确保 payload 与字符串内容按字节等价转换。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestSingleStringMessage_PackUnpack(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveMessage string
	}{
		{
			name:        "success/basic-ascii",
			description: "验证普通 ASCII 字符串会被封包为等价字节并可恢复原文。",
			giveMessage: "hello world",
		},
		{
			name:        "boundary/empty-string",
			description: "验证空字符串是合法 payload，封包结果为空字节切片并可恢复为空文本。",
			giveMessage: "",
		},
		{
			name:        "success/unicode-and-control-characters",
			description: "验证 Unicode 与控制字符按 UTF-8 字节序列稳定封包和解包。",
			giveMessage: "你好，世界！\n\t",
		},
		{
			name:        "boundary/max-uint16-payload-length",
			description: "验证长度等于 uint16 最大值的字符串仍可作为合法 payload 封包。",
			giveMessage: strings.Repeat("a", math.MaxUint16),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			msg := NewSingleStringMessage(tt.giveMessage)
			payload, err := msg.Pack()

			require.NoError(t, err)
			assert.Equal(t, []byte(tt.giveMessage), payload)
			assert.Equal(t, SingleStringMessageType, msg.MessageType())

			unpacked := NewSingleStringMessage("placeholder")
			require.NoError(t, unpacked.Unpack(payload))
			assert.Equal(t, SingleStringMessageType, unpacked.MessageType())
			assert.Equal(t, tt.giveMessage, unpacked.Message())
		})
	}
}

// TestSingleStringMessage_PackErrors 验证简单字符串消息封包的长度限制。
//
// 该测试覆盖超过 uint16 最大 payload 长度的错误分支，确保协议长度字段无法表达的字符串不会被封包。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestSingleStringMessage_PackErrors(t *testing.T) {
	// 验证超过 uint16 最大值的字符串会被拒绝，且错误信息保留真实超长长度。
	const wantLength = math.MaxUint16 + 1
	msg := NewSingleStringMessage(strings.Repeat("a", wantLength))

	payload, err := msg.Pack()

	require.Error(t, err)
	assert.Nil(t, payload)
	assert.Contains(t, err.Error(), "超过 uint16 最大值")
	assert.Contains(t, err.Error(), "字符串消息长度 65536")
}

// TestSingleStringMessage_Unpack 验证简单字符串消息解包对 payload 的直接映射行为。
//
// 该测试通过表驱动用例覆盖 nil、空切片、普通字节和非 UTF-8 字节，确保解包遵循 Go 字符串转换语义。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestSingleStringMessage_Unpack(t *testing.T) {
	tests := []struct {
		name        string
		description string
		givePayload []byte
		wantMessage string
	}{
		{
			name:        "boundary/nil-payload",
			description: "验证 nil payload 会解包为空字符串，与空字节序列语义一致。",
			givePayload: nil,
			wantMessage: "",
		},
		{
			name:        "boundary/empty-payload",
			description: "验证空 payload 会解包为空字符串。",
			givePayload: []byte{},
			wantMessage: "",
		},
		{
			name:        "success/ascii-payload",
			description: "验证普通字节 payload 会按字符串内容保存。",
			givePayload: []byte("abc"),
			wantMessage: "abc",
		},
		{
			name:        "compatibility/non-utf8-payload",
			description: "验证非 UTF-8 字节也按 Go 字符串转换语义原样保存。",
			givePayload: []byte{0xff, 0xfe, 0xfd},
			wantMessage: string([]byte{0xff, 0xfe, 0xfd}),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			msg := NewSingleStringMessage("placeholder")

			require.NoError(t, msg.Unpack(tt.givePayload))
			assert.Equal(t, tt.wantMessage, msg.Message())
		})
	}
}

// TestSingleStringMessage_RecoverNilReceiver 验证简单字符串消息方法对 nil receiver panic 的错误化处理。
//
// 该测试覆盖 Pack 和 Unpack 内置的 recover 契约，确保异常场景不会向调用方传播 panic，而是返回可诊断错误。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestSingleStringMessage_RecoverNilReceiver(t *testing.T) {
	tests := []struct {
		name        string
		description string
		act         func() error
	}{
		{
			name:        "panic/pack-nil-receiver",
			description: "验证 nil 字符串消息 receiver 调用 Pack 时会被 recover 并返回封包错误。",
			act: func() error {
				var msg *singleStringMessage
				payload, err := msg.Pack()
				assert.Nil(t, payload)
				return err
			},
		},
		{
			name:        "panic/unpack-nil-receiver",
			description: "验证 nil 字符串消息 receiver 调用 Unpack 时会被 recover 并返回解包错误。",
			act: func() error {
				var msg *singleStringMessage
				return msg.Unpack([]byte("payload"))
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

// TestNewSingleStringMessage 验证简单字符串消息构造函数会设置固定消息类型并保留内容。
//
// 该测试覆盖构造函数的公开属性契约，确保后续封包和工厂逻辑能够依赖正确的消息类型。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestNewSingleStringMessage(t *testing.T) {
	// 验证构造函数会保留调用方传入的字符串，并设置简单字符串消息类型。
	const giveMessage = "abc"

	msg := NewSingleStringMessage(giveMessage)

	require.NotNil(t, msg)
	assert.Equal(t, SingleStringMessageType, msg.MessageType())
	assert.Equal(t, giveMessage, msg.Message())
}

// TestGenerateSingleStringMessage 验证简单字符串消息生成器的类型和 payload 校验行为。
//
// 该测试通过表驱动用例覆盖生成成功、消息类型不匹配、nil payload 和空 payload，确保工厂调用时字符串消息语义稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGenerateSingleStringMessage(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		giveMessageType MessageType
		givePayload     []byte
		wantMessage     string
		wantErr         bool
	}{
		{
			name:            "success/ascii-payload",
			description:     "验证匹配的字符串消息类型和普通 payload 可以生成字符串消息。",
			giveMessageType: SingleStringMessageType,
			givePayload:     []byte("hello"),
			wantMessage:     "hello",
		},
		{
			name:            "success/empty-payload",
			description:     "验证非 nil 的空 payload 是合法字符串消息并生成空字符串。",
			giveMessageType: SingleStringMessageType,
			givePayload:     []byte{},
			wantMessage:     "",
		},
		{
			name:            "error/type-mismatch",
			description:     "验证非字符串消息类型会被生成器拒绝。",
			giveMessageType: 0x99,
			givePayload:     []byte("hello"),
			wantErr:         true,
		},
		{
			name:            "error/nil-payload",
			description:     "验证 nil payload 会被生成器拒绝，避免与合法空 payload 混淆。",
			giveMessageType: SingleStringMessageType,
			givePayload:     nil,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			msg, err := GenerateSingleStringMessage(tt.giveMessageType, tt.givePayload)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, msg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, msg)
			singleString, ok := msg.(SingleStringMessage)
			require.True(t, ok)
			assert.Equal(t, SingleStringMessageType, msg.MessageType())
			assert.Equal(t, tt.wantMessage, singleString.Message())
		})
	}
}
