// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"sync"
	"time"

	cockroachdberrors "github.com/cockroachdb/errors"
)

var (
	_ Conn     = (*conn)(nil)
	_ net.Conn = (*conn)(nil)
)

type (
	// Conn 自定义消息包传输时使用的网络连接接口。
	// 多个 goroutine 可以同时调用 Conn 上的方法。
	Conn interface {
		// Close 关闭连接。
		//
		// 返回值：
		//   - error: 错误信息。
		Close() error
		// LocalAddr 返回本地网络地址。
		//
		// 返回值：
		//   - net.Addr: 本地网络地址。
		LocalAddr() net.Addr
		// RemoteAddr 返回远程网络地址。
		//
		// 返回值：
		//   - net.Addr: 远程网络地址。
		RemoteAddr() net.Addr
		// SetDeadline 设置读写相关的截止时间。
		//
		// 参数：
		//   - t: 截止时间。
		//
		// 返回值：
		//   - error: 错误信息。
		SetDeadline(time.Time) error
		// SetReadDeadline 设置读截止时间。
		//
		// 参数：
		//   - t: 截止时间。
		//
		// 返回值：
		//   - error: 错误信息。
		SetReadDeadline(time.Time) error
		// SetWriteDeadline 设置写截止时间。
		//
		// 参数：
		//   - t: 截止时间。
		//
		// 返回值：
		//   - error: 错误信息。
		SetWriteDeadline(time.Time) error

		// Closed 返回连接是否已经关闭。
		//
		// 返回值：
		//   - bool: 连接是否关闭。
		Closed() bool
		// Start 启动消息读写 goroutine。
		//
		// 参数：
		//   - ctx: 上下文，用于控制 goroutine 生命周期。
		Start(context.Context)
		// SendMessage 发送消息。
		//
		// 参数：
		//   - message: 待发送的消息。
		//
		// 返回值：
		//   - error: 错误信息。
		SendMessage(Message) error
		// Message 返回只读消息通道。
		//
		// 返回值：
		//   - <-chan Message: 只读消息通道。
		Message() <-chan Message
	}
	// conn 自定义消息包传输时使用的网络连接，实现接口 net.Conn 和 Conn。
	conn struct {
		conn net.Conn // 底层网络连接。

		closed       bool        // 连接关闭标志。
		closedLocker sync.Locker // 关闭操作互斥锁，保证并发安全。

		messageRead  chan Message // 读取到的消息队列通道。
		messageWrite chan Message // 待发送的消息队列通道。

		heartbeatInterval time.Duration // 心跳包发送间隔。
	}
)

// Closed 返回连接是否已经关闭。
//
// 返回值：
//   - bool: 连接是否关闭。
func (c *conn) Closed() bool {
	return c.closed
}

// Start 启动消息读写 goroutine。
//
// 参数：
//   - ctx: 上下文，用于控制 goroutine 生命周期。
func (c *conn) Start(ctx context.Context) {
	go c.send(ctx)    // 启动发送消息的 goroutine。
	go c.receive(ctx) // 启动接收消息的 goroutine。

	if c.heartbeatInterval > 0 {
		ticker := time.NewTicker(c.heartbeatInterval)
		go c.sendHeartbeat(ctx, ticker) // 启动定时发送心跳包的 goroutine。
	}
}

// SendMessage 发送消息。
//
// 参数：
//   - message: 待发送的消息。
//
// 返回值：
//   - error: 错误信息。
func (c *conn) SendMessage(message Message) error {
	var err error

	if c.closed {
		err = cockroachdberrors.Newf("连接已经关闭。")
	} else {
		c.messageWrite <- message // 将消息写入发送队列。
	}

	return err
}

// Message 返回只读消息通道。
//
// 返回值：
//   - <-chan Message: 只读消息通道。
func (c *conn) Message() <-chan Message {
	return c.messageRead
}

// Read 从连接中读取数据。
//
// 参数：
//   - b: 读取缓冲区。
//
// 返回值：
//   - int: 实际读取的字节数。
//   - error: 错误信息。
func (c *conn) Read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}

// Write 向连接中写数据。
//
// 参数：
//   - b: 写入缓冲区。
//
// 返回值：
//   - int: 实际写入的字节数。
//   - error: 错误信息。
func (c *conn) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

// Close 关闭连接。
//
// 返回值：
//   - error: 错误信息。
func (c *conn) Close() error {
	var err error

	// 可能出现不同的 goroutine 同时调用方法，需要加锁操作。
	c.closedLocker.Lock()
	defer c.closedLocker.Unlock()

	if !c.closed {
		err = c.conn.Close()

		c.closed = true

		close(c.messageRead) // 关闭消息读取通道。
	}

	return err
}

// LocalAddr 返回本地网络地址。
//
// 返回值：
//   - net.Addr: 本地网络地址。
func (c *conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr 返回远程网络地址。
//
// 返回值：
//   - net.Addr: 远程网络地址。
func (c *conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline 设置读写相关的截止时间。
//
// 参数：
//   - t: 截止时间。
//
// 返回值：
//   - error: 错误信息。
func (c *conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// SetReadDeadline 设置读截止时间。
//
// 参数：
//   - t: 截止时间。
//
// 返回值：
//   - error: 错误信息。
func (c *conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline 设置写截止时间。
//
// 参数：
//   - t: 截止时间。
//
// 返回值：
//   - error: 错误信息。
func (c *conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

// pack 封包，将消息类型、长度和 payload 组装为完整数据包。
//
// 参数：
//   - message: 待封包的消息。
//
// 返回值：
//   - []byte: 完整数据包字节数组。
//   - error: 错误信息。
func (c *conn) pack(message Message) ([]byte, error) {
	var data []byte
	var err error

	buf := &bytes.Buffer{}
	if payload, errPayload := message.Pack(); nil != errPayload {
		err = cockroachdberrors.Wrap(errPayload, "消息负载封包出现错误。")
	} else if errWriteType := binary.Write(buf, binary.BigEndian, message.MessageType()); nil != errWriteType {
		err = cockroachdberrors.Wrap(errWriteType, "消息类型封包出现错误。")
	} else if errWriteLen := binary.Write(buf, binary.BigEndian, uint16(len(payload))); nil != errWriteLen { //nolint:gosec
		// 注意：payload 的数据包长度不能超过 65535。
		err = cockroachdberrors.Wrap(errWriteLen, "消息负载长度封包出现错误。")
	} else if errWritePayload := binary.Write(buf, binary.BigEndian, payload); nil != errWritePayload {
		err = cockroachdberrors.Wrap(errWritePayload, "消息负载封包出现错误。")
	} else {
		data = buf.Bytes()
	}

	return data, err
}

// send 向网络连接发送消息。
//
// 参数：
//   - ctx: 上下文，用于控制 goroutine 生命周期。
func (c *conn) send(ctx context.Context) {
LoopSend:
	for {
		select {
		case <-ctx.Done():
			// 可能出现还有没消费完的信息。
			_ = c.Close()
			break LoopSend
		case tmp, ok := <-c.messageWrite:
			if !ok {
				_ = c.Close()
				break LoopSend
			} else if pack, errPack := c.pack(tmp); nil != errPack {
				_ = c.Close()
				break LoopSend
			} else if _, errWrite := c.Write(pack); nil != errWrite {
				_ = c.Close()
				break LoopSend
			}
		}
	}
}

// generateMessage 生成消息，从 bufio.Scanner 解析出完整消息并还原为 Message 实例。
//
// 参数：
//   - scanner: 消息包扫描器。
//
// 返回值：
//   - Message: 生成的消息实例。
//   - error: 错误信息。
func (c *conn) generateMessage(scanner *bufio.Scanner) (Message, error) {
	var message Message
	var err error

	if nil == scanner {
		err = cockroachdberrors.Newf("扫描器不能为空。")
	} else if errScanner := scanner.Err(); nil != errScanner {
		err = cockroachdberrors.Wrap(errScanner, "扫描出错。")
	} else if scanner.Scan() {
		data := scanner.Bytes()
		messageType := uint16(0)
		if errReadType := binary.Read(bytes.NewReader(data[:2]), binary.BigEndian, &messageType); nil != errReadType {
			err = cockroachdberrors.Wrap(errReadType, "解包数据类型发生异常。")
		} else {
			payload := data[4:]
			if msg, errGenerate := FactoryGenerate(messageType, payload); nil != errGenerate {
				err = cockroachdberrors.Wrap(errGenerate, "数据包转消息发生异常。")
			} else {
				message = msg
			}
		}
	}

	return message, err
}

// receive 从网络连接接收消息。
//
// 参数：
//   - ctx: 上下文，用于控制 goroutine 生命周期。
func (c *conn) receive(ctx context.Context) {
	scanner := NewScanner(c)
	lastReceived := time.Now()

	// 默认超时时间为 5 秒。
	timeoutDuration := int64(5000)
	if c.heartbeatInterval > 0 {
		// 如果有发送心跳，则默认超时时间为发送心跳的 2 倍时长。
		// 心跳是双向的，两倍足够了。
		timeoutDuration = (c.heartbeatInterval * 2).Milliseconds()
	}

LoopReceive:
	for {
		select {
		case <-ctx.Done():
			_ = c.Close()
			break LoopReceive
		default:
			if tmp, errGenerate := c.generateMessage(scanner); nil != errGenerate {
				_ = c.Close()
				break LoopReceive
			} else if s := time.Since(lastReceived).Milliseconds(); s > timeoutDuration {
				_ = c.Close()
				break LoopReceive
			} else if nil != tmp {
				c.messageRead <- tmp
				lastReceived = time.Now()
			}
		}
	}
}

// sendHeartbeat 定时发送心跳消息。
//
// 参数：
//   - ctx: 上下文，用于控制 goroutine 生命周期。
//   - ticker: 定时器。
func (c *conn) sendHeartbeat(ctx context.Context, ticker *time.Ticker) {
	serialNumber := uint64(0)
LoopHeartbeat:
	for {
		select {
		case <-ctx.Done():
			_ = c.Close()
			break LoopHeartbeat
		case <-ticker.C:
			serialNumber = serialNumber + 1
			hm := NewHeartbeatMessage(serialNumber)
			_ = c.SendMessage(hm)
		}
	}
}

// WrapConn 将 net.Conn 包装成自定义消息包传输时使用的网络连接。
//
// 参数：
//   - c: 底层网络连接。
//   - heartbeatInterval: 心跳包发送间隔。
//
// 返回值：
//   - *conn: 自定义连接实例。
func WrapConn(c net.Conn, heartbeatInterval time.Duration) *conn {
	newConn := &conn{
		conn:              c,
		closedLocker:      &sync.Mutex{},
		messageRead:       make(chan Message, 5120), // 读取消息通道，缓冲区 5120。
		messageWrite:      make(chan Message, 5120), // 发送消息通道，缓冲区 5120。
		heartbeatInterval: heartbeatInterval,
	}

	return newConn
}
