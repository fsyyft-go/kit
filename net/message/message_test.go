// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	errTestMessagePack = errors.New("test message pack failed")
	errTestRead        = errors.New("test read failed")
	errTestWrite       = errors.New("test write failed")
	errTestClose       = errors.New("test close failed")
	errTestBinaryRead  = errors.New("test binary read failed")
	errTestBinaryWrite = errors.New("test binary write failed")
)

// replaceBinaryRead 临时替换二进制读取函数并在测试结束后恢复。
//
// 参数：
//   - t: 测试上下文，用于注册清理逻辑。
//   - fn: 测试期间使用的二进制读取函数。
func replaceBinaryRead(t *testing.T, fn func(io.Reader, binary.ByteOrder, any) error) {
	t.Helper()
	original := binaryRead
	binaryRead = fn
	t.Cleanup(func() { binaryRead = original })
}

// replaceBinaryWrite 临时替换二进制写入函数并在测试结束后恢复。
//
// 参数：
//   - t: 测试上下文，用于注册清理逻辑。
//   - fn: 测试期间使用的二进制写入函数。
func replaceBinaryWrite(t *testing.T, fn func(io.Writer, binary.ByteOrder, any) error) {
	t.Helper()
	original := binaryWrite
	binaryWrite = fn
	t.Cleanup(func() { binaryWrite = original })
}

type testMessage struct {
	messageType MessageType
	payload     []byte
	packErr     error
}

// MessageType 返回测试消息的协议类型。
//
// 该辅助方法用于让 testMessage 满足 Message 接口，并在测试中验证封包逻辑会读取消息类型。
//
// 返回：
//   - MessageType: 测试消息携带的协议类型。
func (m *testMessage) MessageType() MessageType {
	return m.messageType
}

// Pack 返回测试消息预设的 payload 或错误。
//
// 该辅助方法用于让 testMessage 满足 Message 接口，并在测试中稳定模拟封包成功、封包失败和超长 payload 分支。
//
// 返回：
//   - []byte: 测试消息 payload 的防御性拷贝。
//   - error: 预设的封包错误。
func (m *testMessage) Pack() ([]byte, error) {
	if nil != m.packErr {
		return nil, m.packErr
	}

	return append([]byte(nil), m.payload...), nil
}

// Unpack 将 payload 保存到测试消息中。
//
// 该辅助方法用于让 testMessage 满足 Message 接口，并在测试中验证工厂生成器接收的 payload 内容。
//
// 参数：
//   - payload: 需要保存的消息负载。
//
// 返回：
//   - error: 当前辅助实现始终返回 nil。
func (m *testMessage) Unpack(payload []byte) error {
	m.payload = append([]byte(nil), payload...)
	return nil
}

type testAddr string

// Network 返回测试地址的网络名称。
//
// 该辅助方法用于让 testAddr 满足 net.Addr 接口，并验证连接包装对象会透传底层地址。
//
// 返回：
//   - string: 测试网络名称。
func (a testAddr) Network() string {
	return "test"
}

// String 返回测试地址的文本表示。
//
// 该辅助方法用于让 testAddr 满足 net.Addr 接口，并验证连接包装对象会透传底层地址。
//
// 返回：
//   - string: 测试地址文本。
func (a testAddr) String() string {
	return string(a)
}

type scriptedConn struct {
	mu sync.Mutex

	readBuffer *bytes.Reader
	readErr    error
	writeErr   error
	closeErr   error

	localAddr  net.Addr
	remoteAddr net.Addr

	deadline      time.Time
	readDeadline  time.Time
	writeDeadline time.Time

	writes      [][]byte
	writeSignal chan []byte
	closed      bool
	closeCount  int
}

// newScriptedConn 构造可脚本化读写行为的内存连接。
//
// 该辅助函数集中创建 net.Conn 测试替身，避免单元测试依赖外部网络或固定端口。
//
// 参数：
//   - giveReadData: 连接读取端预置的数据；为 nil 时读取端默认返回 io.EOF。
//
// 返回：
//   - *scriptedConn: 可用于 WrapConn 和收发逻辑测试的连接替身。
func newScriptedConn(giveReadData []byte) *scriptedConn {
	return &scriptedConn{
		readBuffer: bytes.NewReader(giveReadData),
		localAddr:  testAddr("local-address"),
		remoteAddr: testAddr("remote-address"),
	}
}

// Read 从预置读取缓冲区复制数据。
//
// 该辅助方法用于让 scriptedConn 满足 net.Conn 接口，并在接收测试中提供确定性的协议字节流。
//
// 参数：
//   - b: 调用方提供的读取缓冲区。
//
// 返回：
//   - int: 实际读取的字节数。
//   - error: 读取完成或预设的读取错误。
func (c *scriptedConn) Read(b []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if nil != c.readBuffer {
		n, err := c.readBuffer.Read(b)
		if n > 0 {
			return n, nil
		}
		if nil != c.readErr {
			return 0, c.readErr
		}
		return n, err
	}

	if nil != c.readErr {
		return 0, c.readErr
	}

	return 0, io.EOF
}

// Write 记录写入的数据并返回预设写入结果。
//
// 该辅助方法用于让 scriptedConn 满足 net.Conn 接口，并在发送测试中断言协议字节输出。
//
// 参数：
//   - b: 需要写入底层连接的数据。
//
// 返回：
//   - int: 成功写入时返回输入数据长度，失败时返回 0。
//   - error: 预设的写入错误。
func (c *scriptedConn) Write(b []byte) (int, error) {
	copied := append([]byte(nil), b...)

	c.mu.Lock()
	c.writes = append(c.writes, copied)
	writeErr := c.writeErr
	writeSignal := c.writeSignal
	c.mu.Unlock()

	if nil != writeSignal {
		select {
		case writeSignal <- copied:
		default:
		}
	}

	if nil != writeErr {
		return 0, writeErr
	}

	return len(b), nil
}

// Close 标记脚本化连接已关闭。
//
// 该辅助方法用于让 scriptedConn 满足 net.Conn 接口，并在关闭相关测试中验证关闭次数和错误透传。
//
// 返回：
//   - error: 预设的关闭错误。
func (c *scriptedConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
	c.closeCount++

	return c.closeErr
}

// LocalAddr 返回脚本化连接的本地地址。
//
// 该辅助方法用于让 scriptedConn 满足 net.Conn 接口，并验证 WrapConn 对地址查询的透传行为。
//
// 返回：
//   - net.Addr: 测试本地地址。
func (c *scriptedConn) LocalAddr() net.Addr {
	return c.localAddr
}

// RemoteAddr 返回脚本化连接的远端地址。
//
// 该辅助方法用于让 scriptedConn 满足 net.Conn 接口，并验证 WrapConn 对地址查询的透传行为。
//
// 返回：
//   - net.Addr: 测试远端地址。
func (c *scriptedConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

// SetDeadline 记录读写截止时间。
//
// 该辅助方法用于让 scriptedConn 满足 net.Conn 接口，并验证 WrapConn 对截止时间设置的透传行为。
//
// 参数：
//   - t: 需要设置的读写截止时间。
//
// 返回：
//   - error: 当前辅助实现始终返回 nil。
func (c *scriptedConn) SetDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.deadline = t

	return nil
}

// SetReadDeadline 记录读取截止时间。
//
// 该辅助方法用于让 scriptedConn 满足 net.Conn 接口，并验证 WrapConn 对读取截止时间设置的透传行为。
//
// 参数：
//   - t: 需要设置的读取截止时间。
//
// 返回：
//   - error: 当前辅助实现始终返回 nil。
func (c *scriptedConn) SetReadDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.readDeadline = t

	return nil
}

// SetWriteDeadline 记录写入截止时间。
//
// 该辅助方法用于让 scriptedConn 满足 net.Conn 接口，并验证 WrapConn 对写入截止时间设置的透传行为。
//
// 参数：
//   - t: 需要设置的写入截止时间。
//
// 返回：
//   - error: 当前辅助实现始终返回 nil。
func (c *scriptedConn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.writeDeadline = t

	return nil
}

// snapshot 返回脚本化连接当前的可断言状态。
//
// 该辅助方法在持锁状态下复制连接内部状态，避免测试直接读取可变字段。
//
// 返回：
//   - [][]byte: 已写入数据的防御性拷贝。
//   - bool: 底层连接是否已关闭。
//   - int: 底层连接关闭次数。
//   - time.Time: 最近一次读写截止时间。
//   - time.Time: 最近一次读取截止时间。
//   - time.Time: 最近一次写入截止时间。
func (c *scriptedConn) snapshot() ([][]byte, bool, int, time.Time, time.Time, time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	writes := make([][]byte, 0, len(c.writes))
	for _, write := range c.writes {
		writes = append(writes, append([]byte(nil), write...))
	}

	return writes, c.closed, c.closeCount, c.deadline, c.readDeadline, c.writeDeadline
}

type errorReader struct {
	err error
}

// Read 始终返回预设错误。
//
// 该辅助方法用于构造已经进入错误状态的 bufio.Scanner，从而验证扫描错误分支。
//
// 参数：
//   - _ : 调用方提供的读取缓冲区，本辅助实现不会写入该缓冲区。
//
// 返回：
//   - int: 始终返回 0。
//   - error: 预设读取错误。
func (r errorReader) Read(_ []byte) (int, error) {
	return 0, r.err
}

// buildTestPacket 构造符合消息协议的完整数据包。
//
// 该辅助函数使用大端序写入消息类型、payload 长度和 payload 内容，确保各测试共享一致的协议夹具。
//
// 参数：
//   - t: 测试上下文，用于报告夹具构造失败并标记辅助函数调用栈。
//   - giveMessageType: 需要写入数据包头部的消息类型。
//   - givePayload: 需要写入数据包正文的 payload。
//
// 返回：
//   - []byte: 完整协议数据包。
func buildTestPacket(t *testing.T, giveMessageType MessageType, givePayload []byte) []byte {
	t.Helper()

	require.LessOrEqual(t, len(givePayload), math.MaxUint16)

	buf := &bytes.Buffer{}
	require.NoError(t, binary.Write(buf, binary.BigEndian, giveMessageType))
	require.NoError(t, binary.Write(buf, binary.BigEndian, uint16(len(givePayload)))) //nolint:gosec
	_, err := buf.Write(givePayload)
	require.NoError(t, err)

	return buf.Bytes()
}

// buildHeartbeatPayload 构造心跳消息 payload。
//
// 该辅助函数按协议规定使用大端序编码 uint64 序列号，供心跳消息和连接收发测试复用。
//
// 参数：
//   - giveSerialNumber: 需要编码的心跳序列号。
//
// 返回：
//   - []byte: 心跳消息 payload。
func buildHeartbeatPayload(giveSerialNumber uint64) []byte {
	payload := make([]byte, 8)
	binary.BigEndian.PutUint64(payload, giveSerialNumber)

	return payload
}
