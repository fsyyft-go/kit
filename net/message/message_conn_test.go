// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWrapConn_DelegatesNetConnBehavior 验证包装连接对 net.Conn 基础行为的透传与关闭状态管理。
//
// 该测试覆盖地址查询、截止时间设置、读写透传和幂等关闭，确保包装层不会破坏底层连接契约。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestWrapConn_DelegatesNetConnBehavior(t *testing.T) {
	tests := []struct {
		name        string
		description string
		setup       func() (*scriptedConn, *conn)
		assert      func(t *testing.T, base *scriptedConn, wrapped *conn)
	}{
		{
			name:        "success/address-and-deadline-delegation",
			description: "验证包装连接会透传地址查询并记录底层截止时间设置。",
			setup: func() (*scriptedConn, *conn) {
				base := newScriptedConn(nil)
				return base, WrapConn(base, 0)
			},
			assert: func(t *testing.T, base *scriptedConn, wrapped *conn) {
				t.Helper()
				deadline := time.Unix(100, 0)
				readDeadline := time.Unix(101, 0)
				writeDeadline := time.Unix(102, 0)

				assert.Equal(t, base.localAddr, wrapped.LocalAddr())
				assert.Equal(t, base.remoteAddr, wrapped.RemoteAddr())
				require.NoError(t, wrapped.SetDeadline(deadline))
				require.NoError(t, wrapped.SetReadDeadline(readDeadline))
				require.NoError(t, wrapped.SetWriteDeadline(writeDeadline))

				_, _, _, gotDeadline, gotReadDeadline, gotWriteDeadline := base.snapshot()
				assert.Equal(t, deadline, gotDeadline)
				assert.Equal(t, readDeadline, gotReadDeadline)
				assert.Equal(t, writeDeadline, gotWriteDeadline)
			},
		},
		{
			name:        "success/read-write-delegation",
			description: "验证包装连接的 Read 和 Write 会直接使用底层连接。",
			setup: func() (*scriptedConn, *conn) {
				base := newScriptedConn([]byte("abc"))
				return base, WrapConn(base, 0)
			},
			assert: func(t *testing.T, base *scriptedConn, wrapped *conn) {
				t.Helper()
				buf := make([]byte, 2)
				n, err := wrapped.Read(buf)
				require.NoError(t, err)
				assert.Equal(t, 2, n)
				assert.Equal(t, []byte("ab"), buf)

				n, err = wrapped.Write([]byte("xy"))
				require.NoError(t, err)
				assert.Equal(t, 2, n)
				writes, _, _, _, _, _ := base.snapshot()
				assert.Equal(t, [][]byte{[]byte("xy")}, writes)
			},
		},
		{
			name:        "success/idempotent-close",
			description: "验证包装连接关闭后会关闭消息通道并且重复关闭不会重复调用底层 Close。",
			setup: func() (*scriptedConn, *conn) {
				base := newScriptedConn(nil)
				return base, WrapConn(base, 0)
			},
			assert: func(t *testing.T, base *scriptedConn, wrapped *conn) {
				t.Helper()
				assert.False(t, wrapped.Closed())
				require.NoError(t, wrapped.Close())
				assert.True(t, wrapped.Closed())
				require.NoError(t, wrapped.Close())

				_, closed, closeCount, _, _, _ := base.snapshot()
				assert.True(t, closed)
				assert.Equal(t, 1, closeCount)
				_, readOK := <-wrapped.messageRead
				assert.False(t, readOK)
				select {
				case <-wrapped.closedNotify:
				default:
					require.Fail(t, "expected close notification")
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			base, wrapped := tt.setup()
			tt.assert(t, base, wrapped)
		})
	}
}

// TestWrapConn_CloseReturnsUnderlyingError 验证关闭包装连接时会返回底层 Close 错误。
//
// 该测试覆盖底层连接关闭失败的边界，确保调用方仍能感知关闭错误且包装连接状态进入 closed。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestWrapConn_CloseReturnsUnderlyingError(t *testing.T) {
	// 验证底层 Close 返回错误时，包装连接仍标记为关闭并关闭内部通道。
	base := newScriptedConn(nil)
	base.closeErr = errTestClose
	wrapped := WrapConn(base, 0)

	err := wrapped.Close()

	require.ErrorIs(t, err, errTestClose)
	assert.True(t, wrapped.Closed())
	_, closed, closeCount, _, _, _ := base.snapshot()
	assert.True(t, closed)
	assert.Equal(t, 1, closeCount)
}

// TestWrapConn_ConcurrentCloseAndSend 验证连接关闭与发送并发执行时没有数据竞争或 channel panic。
//
// 该测试并发调用 Closed、Close 和 SendMessage，覆盖审核发现的关闭标志竞态与向已关闭通道发送风险。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestWrapConn_ConcurrentCloseAndSend(t *testing.T) {
	// 验证并发关闭和发送消息时，关闭只执行一次且发送方只返回成功或已关闭错误，不发生 panic。
	base := newScriptedConn(nil)
	wrapped := WrapConn(base, 0)
	const workerCount = 32
	done := make(chan struct{}, workerCount*3)
	errCh := make(chan error, workerCount)

	for i := 0; i < workerCount; i++ {
		go func() {
			defer func() {
				if recovered := recover(); nil != recovered {
					errCh <- fmt.Errorf("unexpected panic during concurrent send: %v", recovered)
				}
				done <- struct{}{}
			}()
			err := wrapped.SendMessage(NewSingleStringMessage("concurrent"))
			if nil != err && !strings.Contains(err.Error(), "连接已经关闭") {
				errCh <- err
			}
		}()
		go func() {
			_ = wrapped.Closed()
			done <- struct{}{}
		}()
		go func() {
			errCh <- wrapped.Close()
			done <- struct{}{}
		}()
	}

	for i := 0; i < workerCount*3; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			require.Fail(t, "timed out waiting for concurrent close/send workers")
		}
	}
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	_, closed, closeCount, _, _, _ := base.snapshot()
	assert.True(t, wrapped.Closed())
	assert.True(t, closed)
	assert.Equal(t, 1, closeCount)
}

// TestConn_Pack 验证连接封包会生成符合协议的完整字节包。
//
// 该测试通过表驱动用例覆盖空 payload、普通 payload、最大 payload、payload 封包失败和超长 payload，确保协议头与错误边界稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestConn_Pack(t *testing.T) {
	maxPayload := bytes.Repeat([]byte{'a'}, 1<<16-1)
	tests := []struct {
		name            string
		description     string
		giveMessage     Message
		wantPacket      []byte
		wantErr         bool
		wantErrContains string
	}{
		{
			name:        "success/empty-payload",
			description: "验证空 payload 会被封包为仅包含协议头的合法消息包。",
			giveMessage: NewSingleStringMessage(""),
			wantPacket:  buildTestPacket(t, SingleStringMessageType, []byte{}),
		},
		{
			name:        "success/heartbeat-payload",
			description: "验证心跳消息会封包为类型、长度和 8 字节序列号组成的完整协议包。",
			giveMessage: NewHeartbeatMessage(7),
			wantPacket:  buildTestPacket(t, HeartbeatMessageType, buildHeartbeatPayload(7)),
		},
		{
			name:        "boundary/max-payload",
			description: "验证 uint16 最大长度 payload 可以被连接封包。",
			giveMessage: &testMessage{messageType: MessageType(0x3001), payload: maxPayload},
			wantPacket:  buildTestPacket(t, MessageType(0x3001), maxPayload),
		},
		{
			name:            "error/payload-pack-error",
			description:     "验证消息自身 Pack 失败时连接封包会包装并返回该错误。",
			giveMessage:     &testMessage{messageType: MessageType(0x3002), packErr: errTestMessagePack},
			wantErr:         true,
			wantErrContains: "消息负载封包出现错误",
		},
		{
			name:            "error/payload-too-large",
			description:     "验证超过 uint16 最大长度的 payload 会被连接封包拒绝。",
			giveMessage:     &testMessage{messageType: MessageType(0x3003), payload: bytes.Repeat([]byte{'b'}, 1<<16)},
			wantErr:         true,
			wantErrContains: "超过 uint16 最大值",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			wrapped := WrapConn(newScriptedConn(nil), 0)
			packet, err := wrapped.pack(tt.giveMessage)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, packet)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantPacket, packet)
		})
	}
}

// TestConn_PackBinaryWriteErrors 验证连接封包会处理二进制写入错误。
//
// 该测试通过临时替换二进制写入函数覆盖常规 bytes.Buffer 下不可达的错误分支，确保错误会被包装为对应阶段的诊断信息。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestConn_PackBinaryWriteErrors(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		failCall        int
		wantErrContains string
	}{
		{
			name:            "error/write-message-type",
			description:     "验证消息类型字段写入失败时返回消息类型封包错误。",
			failCall:        1,
			wantErrContains: "消息类型封包出现错误",
		},
		{
			name:            "error/write-payload-length",
			description:     "验证 payload 长度字段写入失败时返回消息负载长度封包错误。",
			failCall:        2,
			wantErrContains: "消息负载长度封包出现错误",
		},
		{
			name:            "error/write-payload",
			description:     "验证 payload 数据写入失败时返回消息负载封包错误。",
			failCall:        3,
			wantErrContains: "消息负载封包出现错误",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			callCount := 0
			replaceBinaryWrite(t, func(w io.Writer, order binary.ByteOrder, data any) error {
				callCount++
				if callCount == tt.failCall {
					return errTestBinaryWrite
				}
				return binary.Write(w, order, data)
			})
			wrapped := WrapConn(newScriptedConn(nil), 0)

			packet, err := wrapped.pack(NewSingleStringMessage("payload"))

			require.Error(t, err)
			assert.Nil(t, packet)
			assert.Contains(t, err.Error(), tt.wantErrContains)
			require.ErrorIs(t, err, errTestBinaryWrite)
		})
	}
}

// TestConn_GenerateMessageBinaryReadError 验证消息类型字段读取错误会被包装返回。
//
// 该测试通过临时替换二进制读取函数覆盖常规完整 token 下不可达的读取错误分支。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestConn_GenerateMessageBinaryReadError(t *testing.T) {
	// 验证解析消息类型时的二进制读取错误会被转换为可诊断错误。
	replaceBinaryRead(t, func(io.Reader, binary.ByteOrder, any) error {
		return errTestBinaryRead
	})
	wrapped := WrapConn(newScriptedConn(nil), 0)
	packet := buildTestPacket(t, SingleStringMessageType, []byte("payload"))
	scanner := bufio.NewScanner(bytes.NewReader(packet))
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		if len(data) == 0 && atEOF {
			return 0, nil, nil
		}
		return len(data), data, nil
	})

	got, err := wrapped.generateMessage(scanner)

	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "解包数据类型发生异常")
	require.ErrorIs(t, err, errTestBinaryRead)
}

// TestConn_SendMessage 验证发送消息入队与关闭状态错误。
//
// 该测试通过表驱动用例覆盖连接打开和已关闭两种状态，确保 SendMessage 不会在关闭后接受新消息。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestConn_SendMessage(t *testing.T) {
	tests := []struct {
		name        string
		description string
		setup       func(t *testing.T) *conn
		wantErr     bool
	}{
		{
			name:        "success/enqueue-when-open",
			description: "验证连接打开时 SendMessage 会将消息放入待发送队列。",
			setup: func(t *testing.T) *conn {
				t.Helper()
				return WrapConn(newScriptedConn(nil), 0)
			},
		},
		{
			name:        "error/reject-when-closed",
			description: "验证连接关闭后 SendMessage 返回错误且不会写入已关闭通道。",
			setup: func(t *testing.T) *conn {
				t.Helper()
				wrapped := WrapConn(newScriptedConn(nil), 0)
				require.NoError(t, wrapped.Close())
				return wrapped
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			wrapped := tt.setup(t)
			message := NewSingleStringMessage("queued")
			err := wrapped.SendMessage(message)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "连接已经关闭")
				return
			}

			require.NoError(t, err)
			select {
			case got := <-wrapped.messageWrite:
				assert.Same(t, message, got)
			default:
				require.Fail(t, "expected queued message")
			}
		})
	}
}

// TestConn_SendMessageClosedNotifyBranch 验证发送消息会在关闭通知已发出时失败。
//
// 该测试直接覆盖 SendMessage 的关闭通知分支，确保关闭过程中不会继续向发送队列写入消息。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestConn_SendMessageClosedNotifyBranch(t *testing.T) {
	// 验证发送队列暂不可写且 closedNotify 已关闭时，防御分支会拒绝发送。
	wrapped := WrapConn(newScriptedConn(nil), 0)
	for i := 0; i < cap(wrapped.messageWrite); i++ {
		wrapped.messageWrite <- NewSingleStringMessage("queued")
	}
	close(wrapped.closedNotify)

	err := wrapped.SendMessage(NewSingleStringMessage("late-message"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "连接已经关闭")
}

// TestConn_GenerateMessage 验证连接从 Scanner token 生成消息的行为。
//
// 该测试通过表驱动用例覆盖 nil scanner、扫描错误、数据流结束、内置消息成功解析和未知类型错误，确保解析失败信息可诊断。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestConn_GenerateMessage(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		setupScanner    func(t *testing.T) *bufio.Scanner
		assertMessage   func(t *testing.T, got Message)
		wantErr         bool
		wantErrContains string
	}{
		{
			name:        "error/nil-scanner",
			description: "验证 nil Scanner 会被拒绝并返回明确错误。",
			setupScanner: func(t *testing.T) *bufio.Scanner {
				t.Helper()
				return nil
			},
			wantErr:         true,
			wantErrContains: "扫描器不能为空",
		},
		{
			name:        "error/scanner-error",
			description: "验证 Scanner 内部读取错误会被包装为扫描错误。",
			setupScanner: func(t *testing.T) *bufio.Scanner {
				t.Helper()
				scanner := NewScanner(errorReader{err: errTestRead})
				assert.False(t, scanner.Scan())
				require.ErrorIs(t, scanner.Err(), errTestRead)
				return scanner
			},
			wantErr:         true,
			wantErrContains: "扫描出错",
		},
		{
			name:        "error/no-more-message",
			description: "验证数据流结束且无 token 时会返回无更多消息错误。",
			setupScanner: func(t *testing.T) *bufio.Scanner {
				t.Helper()
				return NewScanner(bytes.NewReader(nil))
			},
			wantErr:         true,
			wantErrContains: "数据流已结束或无更多消息",
		},
		{
			name:        "success/heartbeat-message",
			description: "验证连接可以从完整心跳协议包生成心跳消息。",
			setupScanner: func(t *testing.T) *bufio.Scanner {
				t.Helper()
				packet := buildTestPacket(t, HeartbeatMessageType, buildHeartbeatPayload(321))
				return NewScanner(bytes.NewReader(packet))
			},
			assertMessage: func(t *testing.T, got Message) {
				t.Helper()
				heartbeat, ok := got.(HeartbeatMessage)
				require.True(t, ok)
				assert.Equal(t, uint64(321), heartbeat.SerialNumber())
			},
		},
		{
			name:        "error/unregistered-message-type",
			description: "验证未知消息类型会在工厂生成阶段返回可诊断错误。",
			setupScanner: func(t *testing.T) *bufio.Scanner {
				t.Helper()
				packet := buildTestPacket(t, MessageType(0x7777), []byte("payload"))
				return NewScanner(bytes.NewReader(packet))
			},
			wantErr:         true,
			wantErrContains: "数据包转消息发生异常",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			wrapped := WrapConn(newScriptedConn(nil), 0)
			gotMessage, err := wrapped.generateMessage(tt.setupScanner(t))

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, gotMessage)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, gotMessage)
			tt.assertMessage(t, gotMessage)
		})
	}
}

// TestConn_SendLoop 验证发送循环会从队列取消息并写入底层连接。
//
// 该测试通过表驱动用例覆盖成功写入、封包失败和底层写入失败，确保发送循环在终止时关闭连接。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestConn_SendLoop(t *testing.T) {
	tests := []struct {
		name        string
		description string
		setup       func(t *testing.T, base *scriptedConn) Message
		wantWrite   []byte
	}{
		{
			name:        "success/write-packed-message",
			description: "验证发送循环会将队列中的消息封包后写入底层连接。",
			setup: func(t *testing.T, base *scriptedConn) Message {
				t.Helper()
				return NewSingleStringMessage("send")
			},
			wantWrite: buildTestPacket(t, SingleStringMessageType, []byte("send")),
		},
		{
			name:        "error/pack-failure-closes-connection",
			description: "验证消息封包失败时发送循环会关闭连接且不会写入底层连接。",
			setup: func(t *testing.T, base *scriptedConn) Message {
				t.Helper()
				return &testMessage{messageType: MessageType(0x4001), packErr: errTestMessagePack}
			},
		},
		{
			name:        "error/write-failure-closes-connection",
			description: "验证底层写入失败时发送循环会关闭连接并保留尝试写入的数据。",
			setup: func(t *testing.T, base *scriptedConn) Message {
				t.Helper()
				base.writeErr = errTestWrite
				return NewSingleStringMessage("send")
			},
			wantWrite: buildTestPacket(t, SingleStringMessageType, []byte("send")),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			base := newScriptedConn(nil)
			base.writeSignal = make(chan []byte, 1)
			wrapped := WrapConn(base, 0)
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			go func() {
				defer close(done)
				wrapped.send(ctx)
			}()

			wrapped.messageWrite <- tt.setup(t, base)

			if nil != tt.wantWrite {
				select {
				case got := <-base.writeSignal:
					assert.Equal(t, tt.wantWrite, got)
				case <-time.After(time.Second):
					require.Fail(t, "timed out waiting for write")
				}
			}

			cancel()
			select {
			case <-done:
			case <-time.After(time.Second):
				require.Fail(t, "timed out waiting for send loop to stop")
			}

			writes, closed, _, _, _, _ := base.snapshot()
			assert.True(t, closed)
			if nil == tt.wantWrite {
				assert.Empty(t, writes)
			} else {
				assert.Contains(t, writes, tt.wantWrite)
			}
		})
	}
}

// TestConn_ReceiveLoop 验证接收循环可以读取消息并处理错误关闭。
//
// 该测试通过表驱动用例覆盖成功接收内置消息和未知消息类型错误，确保接收循环对消息通道和关闭状态的处理稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestConn_ReceiveLoop(t *testing.T) {
	tests := []struct {
		name          string
		description   string
		giveReadData  []byte
		wantMessage   bool
		assertMessage func(t *testing.T, got Message)
	}{
		{
			name:         "success/read-single-string-message",
			description:  "验证接收循环会从底层连接读取完整协议包并发送到消息通道。",
			giveReadData: buildTestPacket(t, SingleStringMessageType, []byte("receive")),
			wantMessage:  true,
			assertMessage: func(t *testing.T, got Message) {
				t.Helper()
				singleString, ok := got.(SingleStringMessage)
				require.True(t, ok)
				assert.Equal(t, "receive", singleString.Message())
			},
		},
		{
			name:         "error/unknown-message-type-closes-connection",
			description:  "验证接收循环遇到未知消息类型时关闭连接且不投递消息。",
			giveReadData: buildTestPacket(t, MessageType(0x5555), []byte("receive")),
			wantMessage:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			base := newScriptedConn(tt.giveReadData)
			wrapped := WrapConn(base, 0)
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			go func() {
				defer close(done)
				wrapped.receive(ctx)
			}()

			if tt.wantMessage {
				select {
				case got := <-wrapped.Message():
					require.NotNil(t, got)
					tt.assertMessage(t, got)
				case <-time.After(time.Second):
					require.Fail(t, "timed out waiting for received message")
				}
				cancel()
			} else {
				select {
				case got, ok := <-wrapped.Message():
					assert.False(t, ok)
					assert.Nil(t, got)
				case <-time.After(time.Second):
					require.Fail(t, "timed out waiting for receive loop to close")
				}
				cancel()
			}

			select {
			case <-done:
			case <-time.After(time.Second):
				require.Fail(t, "timed out waiting for receive loop to stop")
			}

			_, closed, _, _, _, _ := base.snapshot()
			assert.True(t, closed)
		})
	}
}

// TestConn_StartWithNetPipe 验证包装连接可通过本地内存连接完成端到端消息收发。
//
// 该测试使用 net.Pipe 间接创建的内存连接路径，不访问外部网络，确保 Start、SendMessage、send 和 receive 能协同传输消息。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestConn_StartWithNetPipe(t *testing.T) {
	// 验证两个包装连接启动后，可以通过内存连接端到端传输简单字符串消息。
	leftRaw, rightRaw := netPipe(t)
	left := WrapConn(leftRaw, 0)
	right := WrapConn(rightRaw, 0)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	t.Cleanup(func() { _ = left.Close() })
	t.Cleanup(func() { _ = right.Close() })

	left.Start(ctx)
	right.Start(ctx)

	require.NoError(t, left.SendMessage(NewSingleStringMessage("pipe-message")))

	select {
	case got := <-right.Message():
		require.NotNil(t, got)
		singleString, ok := got.(SingleStringMessage)
		require.True(t, ok)
		assert.Equal(t, "pipe-message", singleString.Message())
	case <-time.After(time.Second):
		require.Fail(t, "timed out waiting for pipe message")
	}
}

// TestConn_SendLoopStopsWhenQueueClosed 验证发送队列关闭时发送循环会退出。
//
// 该测试覆盖连接已关闭后发送队列关闭的路径，确保发送循环不会阻塞或重复关闭底层连接。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestConn_SendLoopStopsWhenQueueClosed(t *testing.T) {
	// 验证已关闭连接的发送队列关闭后，发送循环可以直接退出。
	base := newScriptedConn(nil)
	wrapped := WrapConn(base, 0)
	require.NoError(t, wrapped.Close())

	done := make(chan struct{})
	go func() {
		defer close(done)
		wrapped.send(context.Background())
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		require.Fail(t, "timed out waiting for send loop to stop")
	}

	_, closed, closeCount, _, _, _ := base.snapshot()
	assert.True(t, closed)
	assert.Equal(t, 1, closeCount)
}

// TestConn_SendLoopStopsWhenClosedNotify 验证发送循环收到关闭通知时退出。
//
// 该测试覆盖发送循环的 closedNotify 分支，确保连接被其他路径关闭时发送 goroutine 可以稳定退出。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestConn_SendLoopStopsWhenClosedNotify(t *testing.T) {
	// 验证关闭通知关闭后，发送循环无需读取发送队列即可退出。
	wrapped := WrapConn(newScriptedConn(nil), 0)
	close(wrapped.closedNotify)
	done := make(chan struct{})
	go func() {
		defer close(done)
		wrapped.send(context.Background())
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		require.Fail(t, "timed out waiting for send loop to stop")
	}
}

// TestConn_SendLoopStopsWhenMessageWriteClosed 验证发送队列关闭时发送循环会关闭连接并退出。
//
// 该测试覆盖发送循环读取发送队列返回 ok=false 的兼容路径，确保异常关闭发送队列时连接状态仍一致。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestConn_SendLoopStopsWhenMessageWriteClosed(t *testing.T) {
	// 验证发送队列被外部关闭时，发送循环会主动关闭连接并退出。
	base := newScriptedConn(nil)
	wrapped := WrapConn(base, 0)
	close(wrapped.messageWrite)
	done := make(chan struct{})
	go func() {
		defer close(done)
		wrapped.send(context.Background())
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		require.Fail(t, "timed out waiting for send loop to stop")
	}

	_, closed, closeCount, _, _, _ := base.snapshot()
	assert.True(t, wrapped.Closed())
	assert.True(t, closed)
	assert.Equal(t, 1, closeCount)
}

// TestConn_ReceiveLoopStopsWhenClosedNotify 验证接收循环收到关闭通知时退出。
//
// 该测试覆盖接收循环的 closedNotify 分支，确保连接被其他路径关闭时接收 goroutine 可以稳定退出。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestConn_ReceiveLoopStopsWhenClosedNotify(t *testing.T) {
	// 验证关闭通知关闭后，接收循环无需读取底层连接即可退出。
	wrapped := WrapConn(newScriptedConn(nil), 0)
	close(wrapped.closedNotify)

	wrapped.receive(context.Background())

	assert.False(t, wrapped.Closed())
}

// TestConn_ReceiveLoopStopsWhenContextCancelled 验证接收循环在上下文取消时退出。
//
// 该测试覆盖接收循环的 ctx.Done 分支，确保取消信号会关闭连接并结束循环。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestConn_ReceiveLoopStopsWhenContextCancelled(t *testing.T) {
	// 验证上下文预先取消时，接收循环无需读取底层连接即可关闭并返回。
	base := newScriptedConn(nil)
	wrapped := WrapConn(base, 0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	wrapped.receive(ctx)

	_, closed, closeCount, _, _, _ := base.snapshot()
	assert.True(t, wrapped.Closed())
	assert.True(t, closed)
	assert.Equal(t, 1, closeCount)
}

// TestConn_SendHeartbeat 验证心跳循环会按 ticker 生成心跳消息并响应取消。
//
// 该测试直接覆盖心跳发送循环，确保生成的消息类型和序列号有效，且上下文取消后连接会关闭。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestConn_SendHeartbeat(t *testing.T) {
	// 验证 ticker 触发后会向发送队列写入心跳消息，且取消上下文会关闭连接。
	base := newScriptedConn(nil)
	wrapped := WrapConn(base, time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(time.Millisecond)
	t.Cleanup(ticker.Stop)
	done := make(chan struct{})
	go func() {
		defer close(done)
		wrapped.sendHeartbeat(ctx, ticker)
	}()

	select {
	case got := <-wrapped.messageWrite:
		require.NotNil(t, got)
		heartbeat, ok := got.(HeartbeatMessage)
		require.True(t, ok)
		assert.Equal(t, HeartbeatMessageType, got.MessageType())
		assert.Positive(t, heartbeat.SerialNumber())
	case <-time.After(time.Second):
		require.Fail(t, "timed out waiting for heartbeat message")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		require.Fail(t, "timed out waiting for heartbeat loop to stop")
	}

	_, closed, _, _, _, _ := base.snapshot()
	assert.True(t, closed)
}

// TestConn_SendHeartbeatStopsWhenClosedNotify 验证心跳循环收到关闭通知时退出。
//
// 该测试覆盖心跳发送循环的 closedNotify 分支，确保连接被其他路径关闭时心跳 goroutine 可以稳定退出。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestConn_SendHeartbeatStopsWhenClosedNotify(t *testing.T) {
	// 验证关闭通知关闭后，心跳循环无需等待 ticker 即可退出。
	wrapped := WrapConn(newScriptedConn(nil), time.Millisecond)
	close(wrapped.closedNotify)
	ticker := time.NewTicker(time.Hour)
	t.Cleanup(ticker.Stop)
	done := make(chan struct{})
	go func() {
		defer close(done)
		wrapped.sendHeartbeat(context.Background(), ticker)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		require.Fail(t, "timed out waiting for heartbeat loop to stop")
	}
}

// TestConn_ReceiveLoopTimeoutWithNilGeneratedMessage 验证接收循环会在无消息投递且超时后关闭连接。
//
// 该测试使用临时默认工厂生成 nil 消息，覆盖接收循环的心跳超时分支，并在测试结束后恢复全局工厂状态。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestConn_ReceiveLoopTimeoutWithNilGeneratedMessage(t *testing.T) {
	// 验证生成器返回 nil 消息且心跳超时时，接收循环会主动关闭连接。
	originalFactory := defaultFactory
	defaultFactory = NewMessageFactory()
	t.Cleanup(func() { defaultFactory = originalFactory })

	const giveMessageType = MessageType(0x6600)
	require.NoError(t, FactoryRegister(giveMessageType, func(MessageType, []byte) (Message, error) {
		time.Sleep(2 * time.Millisecond)
		return nil, nil
	}))

	base := newScriptedConn(buildTestPacket(t, giveMessageType, []byte("timeout")))
	wrapped := WrapConn(base, time.Nanosecond)
	done := make(chan struct{})
	go func() {
		defer close(done)
		wrapped.receive(context.Background())
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		require.Fail(t, "timed out waiting for receive timeout")
	}

	_, closed, _, _, _, _ := base.snapshot()
	assert.True(t, wrapped.Closed())
	assert.True(t, closed)
}

// TestConn_StartWithHeartbeat 验证 Start 会启动心跳发送协程。
//
// 该测试使用 net.Pipe 建立本地内存连接，覆盖 Start 在配置心跳间隔时启动心跳发送路径的行为。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestConn_StartWithHeartbeat(t *testing.T) {
	// 验证配置心跳间隔后，Start 会通过本地内存连接发送心跳协议包。
	leftRaw, rightRaw := netPipe(t)
	wrapped := WrapConn(leftRaw, time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	t.Cleanup(func() { _ = wrapped.Close() })

	wrapped.Start(ctx)
	require.NoError(t, rightRaw.SetReadDeadline(time.Now().Add(time.Second)))
	scanner := NewScanner(rightRaw)
	require.True(t, scanner.Scan())
	packet := append([]byte(nil), scanner.Bytes()...)
	require.NoError(t, scanner.Err())

	assert.Len(t, packet, messageHeaderLength+8)
	assert.Equal(t, buildTestPacket(t, HeartbeatMessageType, packet[messageHeaderLength:]), packet)

	cancel()
}

// TestConn_ReceiveLoopClosedWhileDelivering 验证接收循环投递消息时会响应并发关闭。
//
// 该测试覆盖接收循环在解析消息后、投递消息前观察到连接关闭的防御分支，确保不会向已关闭通道发送消息。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestConn_ReceiveLoopClosedWhileDelivering(t *testing.T) {
	const giveMessageType = MessageType(0x6601)
	tests := []struct {
		name        string
		description string
		setup       func(t *testing.T, wrapped *conn, generated chan struct{})
		trigger     func(t *testing.T, wrapped *conn)
	}{
		{
			name:        "success/closed-state-before-delivery",
			description: "验证接收循环解析出消息后，如果连接已关闭，会在投递前退出。",
			setup: func(t *testing.T, wrapped *conn, generated chan struct{}) {
				t.Helper()
				wrapped.messageReadLocker.Lock()
			},
			trigger: func(t *testing.T, wrapped *conn) {
				t.Helper()
				done := make(chan error, 1)
				go func() { done <- wrapped.Close() }()
				select {
				case <-wrapped.closedNotify:
				case <-time.After(time.Second):
					require.Fail(t, "timed out waiting for close notification")
				}
				wrapped.messageReadLocker.Unlock()
				select {
				case err := <-done:
					require.NoError(t, err)
				case <-time.After(time.Second):
					require.Fail(t, "timed out waiting for close to finish")
				}
			},
		},
		{
			name:        "success/closed-notify-before-delivery",
			description: "验证接收循环解析出消息后，如果收到关闭通知，会在 select 分支退出。",
			setup: func(t *testing.T, wrapped *conn, generated chan struct{}) {
				t.Helper()
				for i := 0; i < cap(wrapped.messageRead); i++ {
					wrapped.messageRead <- NewSingleStringMessage("queued")
				}
			},
			trigger: func(t *testing.T, wrapped *conn) {
				t.Helper()
				close(wrapped.closedNotify)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			originalFactory := defaultFactory
			defaultFactory = NewMessageFactory()
			t.Cleanup(func() { defaultFactory = originalFactory })
			generated := make(chan struct{})
			require.NoError(t, FactoryRegister(giveMessageType, func(MessageType, []byte) (Message, error) {
				close(generated)
				return NewSingleStringMessage("blocked"), nil
			}))

			base := newScriptedConn(buildTestPacket(t, giveMessageType, []byte("blocked")))
			wrapped := WrapConn(base, 0)
			tt.setup(t, wrapped, generated)

			done := make(chan struct{})
			go func() {
				defer close(done)
				wrapped.receive(context.Background())
			}()

			select {
			case <-generated:
			case <-time.After(time.Second):
				require.Fail(t, "timed out waiting for message generation")
			}
			tt.trigger(t, wrapped)

			select {
			case <-done:
			case <-time.After(time.Second):
				require.Fail(t, "timed out waiting for receive loop to stop")
			}
		})
	}
}

// netPipe 创建本地内存双向连接。
//
// 该辅助函数封装标准库 net.Pipe 调用并注册清理逻辑，确保端到端连接测试不依赖外部网络且资源可恢复。
//
// 参数：
//   - t: 测试上下文，用于报告连接创建失败并注册清理逻辑。
//
// 返回：
//   - net.Conn: 内存连接左端。
//   - net.Conn: 内存连接右端。
func netPipe(t *testing.T) (net.Conn, net.Conn) {
	t.Helper()

	left, right := net.Pipe()
	t.Cleanup(func() { _ = left.Close() })
	t.Cleanup(func() { _ = right.Close() })

	return left, right
}
