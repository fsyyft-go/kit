// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errTestGenerateMessage = errors.New("test generate message failed")

// TestGenerateMessageFunc_GenerateMessage 验证函数适配器会透传消息类型、payload 和返回值。
//
// 该测试通过表驱动用例覆盖成功返回与错误返回，确保 GenerateMessageFunc 满足 Generator 接口的行为稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGenerateMessageFunc_GenerateMessage(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		giveMessageType MessageType
		givePayload     []byte
		giveErr         error
		wantErr         bool
	}{
		{
			name:            "success/delegates-to-function",
			description:     "验证适配器会将消息类型和 payload 原样传递给底层函数并返回消息。",
			giveMessageType: MessageType(0x1201),
			givePayload:     []byte("payload"),
		},
		{
			name:            "error/delegates-error",
			description:     "验证底层函数返回错误时适配器会原样向调用方返回该错误。",
			giveMessageType: MessageType(0x1202),
			givePayload:     []byte("payload"),
			giveErr:         errTestGenerateMessage,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			called := false
			wantMessage := &testMessage{messageType: tt.giveMessageType, payload: tt.givePayload}
			generator := GenerateMessageFunc(func(messageType MessageType, payload []byte) (Message, error) {
				called = true
				assert.Equal(t, tt.giveMessageType, messageType)
				assert.Equal(t, tt.givePayload, payload)
				if nil != tt.giveErr {
					return nil, tt.giveErr
				}
				return wantMessage, nil
			})

			gotMessage, err := generator.GenerateMessage(tt.giveMessageType, tt.givePayload)

			assert.True(t, called)
			if tt.wantErr {
				require.ErrorIs(t, err, tt.giveErr)
				assert.Nil(t, gotMessage)
				return
			}

			require.NoError(t, err)
			assert.Same(t, wantMessage, gotMessage)
		})
	}
}

// TestMessageFactory_RegisterAndGenerate 验证独立消息工厂的注册与生成行为。
//
// 该测试通过表驱动用例覆盖注册成功、重复注册、nil 生成器、未知类型、nil payload 和生成器错误，确保工厂边界稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestMessageFactory_RegisterAndGenerate(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		setup           func(t *testing.T, factory MessageFactory) MessageType
		givePayload     []byte
		wantMessage     bool
		wantErr         bool
		wantErrContains string
	}{
		{
			name:        "success/register-and-generate",
			description: "验证成功注册的消息类型会调用对应生成器并返回生成消息。",
			setup: func(t *testing.T, factory MessageFactory) MessageType {
				t.Helper()
				messageType := MessageType(0x2001)
				err := factory.Register(messageType, func(gotMessageType MessageType, gotPayload []byte) (Message, error) {
					assert.Equal(t, messageType, gotMessageType)
					assert.Equal(t, []byte("ok"), gotPayload)
					return &testMessage{messageType: gotMessageType, payload: gotPayload}, nil
				})
				require.NoError(t, err)
				return messageType
			},
			givePayload: []byte("ok"),
			wantMessage: true,
		},
		{
			name:        "error/duplicate-register",
			description: "验证重复注册同一消息类型会返回错误并保留首次注册结果。",
			setup: func(t *testing.T, factory MessageFactory) MessageType {
				t.Helper()
				messageType := MessageType(0x2002)
				first := GenerateMessageFunc(func(gotMessageType MessageType, gotPayload []byte) (Message, error) {
					return &testMessage{messageType: gotMessageType, payload: []byte("first")}, nil
				})
				second := GenerateMessageFunc(func(gotMessageType MessageType, gotPayload []byte) (Message, error) {
					return &testMessage{messageType: gotMessageType, payload: []byte("second")}, nil
				})
				require.NoError(t, factory.Register(messageType, first))
				err := factory.Register(messageType, second)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "已经存在")
				return messageType
			},
			givePayload: []byte("ok"),
			wantMessage: true,
		},
		{
			name:        "error/nil-generator-register",
			description: "验证 nil 生成器不能注册到消息工厂。",
			setup: func(t *testing.T, factory MessageFactory) MessageType {
				t.Helper()
				messageType := MessageType(0x2003)
				err := factory.Register(messageType, nil)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "方法不允许为空")
				return messageType
			},
			givePayload:     []byte("ok"),
			wantErr:         true,
			wantErrContains: "不存在",
		},
		{
			name:        "error/unregistered-type",
			description: "验证未注册消息类型生成消息时返回不存在错误。",
			setup: func(t *testing.T, factory MessageFactory) MessageType {
				t.Helper()
				return MessageType(0x2004)
			},
			givePayload:     []byte("ok"),
			wantErr:         true,
			wantErrContains: "不存在",
		},
		{
			name:        "error/nil-payload",
			description: "验证工厂拒绝 nil payload，避免生成器收到无效负载。",
			setup: func(t *testing.T, factory MessageFactory) MessageType {
				t.Helper()
				messageType := MessageType(0x2005)
				require.NoError(t, factory.Register(messageType, func(gotMessageType MessageType, gotPayload []byte) (Message, error) {
					return &testMessage{messageType: gotMessageType, payload: gotPayload}, nil
				}))
				return messageType
			},
			givePayload:     nil,
			wantErr:         true,
			wantErrContains: "有效负载不能为空",
		},
		{
			name:        "error/generator-error",
			description: "验证已注册生成器返回错误时工厂会将该错误返回给调用方。",
			setup: func(t *testing.T, factory MessageFactory) MessageType {
				t.Helper()
				messageType := MessageType(0x2006)
				require.NoError(t, factory.Register(messageType, func(gotMessageType MessageType, gotPayload []byte) (Message, error) {
					return nil, errTestGenerateMessage
				}))
				return messageType
			},
			givePayload: []byte("ok"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			factory := NewMessageFactory()
			messageType := tt.setup(t, factory)

			gotMessage, err := factory.Generate(messageType, tt.givePayload)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, gotMessage)
				if tt.wantErrContains != "" {
					assert.Contains(t, err.Error(), tt.wantErrContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, gotMessage)
			assert.Equal(t, messageType, gotMessage.MessageType())
			if tt.wantMessage {
				payload, packErr := gotMessage.Pack()
				require.NoError(t, packErr)
				assert.NotNil(t, payload)
			}
		})
	}
}

// TestMessageFactory_RegisterConcurrent 验证消息工厂并发注册时保持一致性。
//
// 该测试覆盖注册锁保护的并发路径，确保不同消息类型可以并发注册并在之后成功生成消息。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestMessageFactory_RegisterConcurrent(t *testing.T) {
	// 验证并发注册不同消息类型不会丢失生成器，并且每个类型都能生成对应消息。
	factory := NewMessageFactory()
	messageTypes := []MessageType{0x2101, 0x2102, 0x2103, 0x2104}
	errCh := make(chan error, len(messageTypes))
	doneCh := make(chan struct{}, len(messageTypes))

	for _, messageType := range messageTypes {
		messageType := messageType
		go func() {
			errCh <- factory.Register(messageType, func(gotMessageType MessageType, gotPayload []byte) (Message, error) {
				return &testMessage{messageType: gotMessageType, payload: gotPayload}, nil
			})
			doneCh <- struct{}{}
		}()
	}

	for range messageTypes {
		select {
		case <-doneCh:
		case <-time.After(time.Second):
			require.Fail(t, "timed out waiting for concurrent registration")
		}
	}
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	for _, messageType := range messageTypes {
		got, err := factory.Generate(messageType, []byte("payload"))
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, messageType, got.MessageType())
	}
}

// TestFactoryRegister_DefaultFactoryProxy 验证默认工厂注册代理会更新全局默认工厂。
//
// 该测试临时替换 defaultFactory 并在清理阶段恢复，确保 FactoryRegister 与 FactoryGenerate 组合可用于外部扩展消息类型。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestFactoryRegister_DefaultFactoryProxy(t *testing.T) {
	// 验证通过包级注册函数添加的类型可以立即通过包级生成函数创建消息。
	originalFactory := defaultFactory
	defaultFactory = NewMessageFactory()
	t.Cleanup(func() { defaultFactory = originalFactory })

	const giveMessageType = MessageType(0x2201)
	require.NoError(t, FactoryRegister(giveMessageType, func(gotMessageType MessageType, gotPayload []byte) (Message, error) {
		return &testMessage{messageType: gotMessageType, payload: gotPayload}, nil
	}))

	gotMessage, err := FactoryGenerate(giveMessageType, []byte("proxy"))

	require.NoError(t, err)
	require.NotNil(t, gotMessage)
	assert.Equal(t, giveMessageType, gotMessage.MessageType())
	payload, packErr := gotMessage.Pack()
	require.NoError(t, packErr)
	assert.Equal(t, []byte("proxy"), payload)
}

// TestFactoryGenerate_DefaultFactory 验证默认消息工厂已注册内置消息类型。
//
// 该测试通过表驱动用例覆盖心跳消息和简单字符串消息，确保包初始化注册对外部调用者可用。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestFactoryGenerate_DefaultFactory(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		giveMessageType MessageType
		givePayload     []byte
		assertMessage   func(t *testing.T, got Message)
	}{
		{
			name:            "success/default-heartbeat-generator",
			description:     "验证默认工厂可以根据内置心跳类型生成心跳消息。",
			giveMessageType: HeartbeatMessageType,
			givePayload:     buildHeartbeatPayload(88),
			assertMessage: func(t *testing.T, got Message) {
				t.Helper()
				heartbeat, ok := got.(HeartbeatMessage)
				require.True(t, ok)
				assert.Equal(t, uint64(88), heartbeat.SerialNumber())
			},
		},
		{
			name:            "success/default-single-string-generator",
			description:     "验证默认工厂可以根据内置字符串类型生成简单字符串消息。",
			giveMessageType: SingleStringMessageType,
			givePayload:     []byte("factory"),
			assertMessage: func(t *testing.T, got Message) {
				t.Helper()
				singleString, ok := got.(SingleStringMessage)
				require.True(t, ok)
				assert.Equal(t, "factory", singleString.Message())
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			gotMessage, err := FactoryGenerate(tt.giveMessageType, tt.givePayload)

			require.NoError(t, err)
			require.NotNil(t, gotMessage)
			assert.Equal(t, tt.giveMessageType, gotMessage.MessageType())
			tt.assertMessage(t, gotMessage)
		})
	}
}
