// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// 本测试文件针对 net/message/message_singlestring.go 的所有核心功能进行单元测试。
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
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSingleStringMessage_PackUnpack 测试 Pack 和 Unpack 的互操作性。
func TestSingleStringMessage_PackUnpack(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"普通字符串", "hello world", false},
		{"空字符串", "", false},
		{"特殊字符", "你好，世界！\n\t", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			msg := NewSingleStringMessage(c.input)
			payload, err := msg.Pack()
			assert.Equal(t, c.wantErr, err != nil, "Pack 错误断言")
			if err != nil {
				return
			}
			msg2 := NewSingleStringMessage("")
			err = msg2.Unpack(payload)
			assert.Equal(t, c.wantErr, err != nil, "Unpack 错误断言")
			if err == nil {
				assert.Equal(t, c.input, msg2.Message(), "字符串内容应一致")
			}
		})
	}
}

// TestSingleStringMessage_Pack_异常测试
func TestSingleStringMessage_Pack_Error(t *testing.T) {
	msg := NewSingleStringMessage("test")
	payload, err := msg.Pack()
	assert.NoError(t, err, "正常情况下 Pack 不应报错")
	assert.Equal(t, []byte("test"), payload, "payload 内容应一致")
}

// TestSingleStringMessage_Unpack_异常测试
func TestSingleStringMessage_Unpack_Error(t *testing.T) {
	cases := []struct {
		name    string
		input   []byte
		want    string
		wantErr bool
	}{
		{"空 payload", nil, "", false},
		{"正常字符串", []byte("abc"), "abc", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			msg := NewSingleStringMessage("")
			err := msg.Unpack(c.input)
			assert.Equal(t, c.wantErr, err != nil, "Unpack 错误断言")
			if err == nil {
				assert.Equal(t, c.want, msg.Message(), "解包后内容应一致")
			}
		})
	}
}

// TestNewSingleStringMessage_属性测试
func TestNewSingleStringMessage(t *testing.T) {
	str := "abc"
	msg := NewSingleStringMessage(str)
	assert.Equal(t, SingleStringMessageType, msg.MessageType(), "消息类型应为 SingleStringMessageType")
	assert.Equal(t, str, msg.Message(), "内容应一致")
}

// TestGenerateSingleStringMessage_工厂函数测试
func TestGenerateSingleStringMessage(t *testing.T) {
	cases := []struct {
		name    string
		msgType MessageType
		input   string
		wantErr bool
	}{
		{"类型匹配", SingleStringMessageType, "hello", false},
		{"类型不匹配", 0x99, "hello", true},
		{"空 payload", SingleStringMessageType, "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var payload []byte
			if !c.wantErr || c.msgType != SingleStringMessageType || c.input != "" {
				payload = []byte(c.input)
			}
			msg, err := GenerateSingleStringMessage(c.msgType, payload)
			assert.Equal(t, c.wantErr, err != nil, "工厂函数错误断言")
			if !c.wantErr {
				assert.NotNil(t, msg, "消息不应为 nil")
				ss, ok := msg.(SingleStringMessage)
				assert.True(t, ok, "应实现 SingleStringMessage 接口")
				assert.Equal(t, c.input, ss.Message(), "内容应一致")
			}
		})
	}
}

// TestSingleStringMessage_Pack_超长字符串测试
func TestSingleStringMessage_Pack_TooLong(t *testing.T) {
	// 构造长度超过 uint16 最大值的字符串。
	longStr := make([]byte, 1+0xFFFF) // math.MaxUint16 == 65535
	for i := range longStr {
		longStr[i] = 'a'
	}
	msg := NewSingleStringMessage(string(longStr))
	payload, err := msg.Pack()
	// 断言应返回错误，payload 为空。
	assert.Error(t, err, "超长字符串应返回错误")
	assert.Empty(t, payload, "payload 应为空")
}
