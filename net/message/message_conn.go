// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"

	cockroachdberrors "github.com/cockroachdb/errors"

	kitgoroutine "github.com/fsyyft-go/kit/runtime/goroutine"
)

var (
	_ Conn     = (*conn)(nil)
	_ net.Conn = (*conn)(nil)
)

type (
	// Conn 定义按本包协议收发消息的连接契约。
	//
	// Close、Closed 和 SendMessage 可以并发调用。Start 不会自行去重，
	// 调用方只应调用一次；重复调用会额外启动读写循环和可选心跳循环，
	// 并竞争同一底层连接。Message 返回单个共享接收 channel，多个消费者
	// 同时读取时会竞争消费消息；连接关闭时该 channel 会被关闭。
	// 地址查询方法直接委托给底层 net.Conn；截止时间设置会同步影响
	// [Conn.Start] 启动的内部收发流程。
	Conn interface {
		// Close 关闭连接并通知内部 goroutine 退出。
		//
		// Close 可与 Closed 和 SendMessage 并发调用；重复调用只会在首次关闭时关闭底层连接，
		// 并关闭 [Conn.Message] 返回的共享接收 channel。
		//
		// 参数：无。
		//
		// 返回：
		//   - error: 首次关闭底层连接时返回的错误。
		Close() error
		// LocalAddr 返回底层连接的本地网络地址。
		//
		// 参数：无。
		//
		// 返回：
		//   - net.Addr: 本地网络地址。
		LocalAddr() net.Addr
		// RemoteAddr 返回底层连接的远程网络地址。
		//
		// 参数：无。
		//
		// 返回：
		//   - net.Addr: 远程网络地址。
		RemoteAddr() net.Addr
		// SetDeadline 设置底层连接的读写截止时间。
		//
		// 调用后会同时影响 [Conn.Start] 启动的内部收发流程。
		//
		// 参数：
		//   - time.Time: 截止时间；零值表示取消已设置的读写截止时间。
		//
		// 返回：
		//   - error: 底层连接设置截止时间失败时返回错误。
		SetDeadline(time.Time) error
		// SetReadDeadline 设置底层连接的读截止时间。
		//
		// 该设置会影响 [Conn.Start] 启动的内部接收流程。
		//
		// 参数：
		//   - time.Time: 截止时间；零值表示取消已设置的读截止时间。
		//
		// 返回：
		//   - error: 底层连接设置读截止时间失败时返回错误。
		SetReadDeadline(time.Time) error
		// SetWriteDeadline 设置底层连接的写截止时间。
		//
		// 该设置会影响 [Conn.Start] 启动的内部发送流程。
		//
		// 参数：
		//   - time.Time: 截止时间；零值表示取消已设置的写截止时间。
		//
		// 返回：
		//   - error: 底层连接设置写截止时间失败时返回错误。
		SetWriteDeadline(time.Time) error

		// Closed 返回连接是否已经关闭。
		//
		// Closed 可与 Close 和 SendMessage 并发调用。
		//
		// 参数：无。
		//
		// 返回：
		//   - bool: 连接是否已经进入关闭状态。
		Closed() bool
		// Start 启动内部发送、接收以及可选心跳 goroutine。
		//
		// Start 不会自行去重，调用方只应调用一次。传入的上下文结束或连接关闭后，
		// Start 启动的 goroutine 会退出。
		//
		// 参数：
		//   - context.Context: 控制内部 goroutine 生命周期的上下文，不能为空。
		Start(context.Context)
		// SendMessage 将消息放入内部发送队列。
		//
		// SendMessage 可与 Close 和 Closed 并发调用。返回 nil 仅表示消息已入队，
		// 不表示已经写入底层连接；当发送队列已满时会阻塞，直到队列腾出空间或连接关闭。
		//
		// 参数：
		//   - Message: 待异步发送的消息；调用方应保证其非 nil。
		//
		// 返回：
		//   - error: 连接已关闭，或消息在入队前因收到关闭通知而被拒绝时返回错误。
		SendMessage(Message) error
		// Message 返回连接的共享接收 channel。
		//
		// 该 channel 只创建一次；多个消费者同时读取时会竞争消费消息。
		// 连接关闭时，该 channel 也会被关闭。
		//
		// 参数：无。
		//
		// 返回：
		//   - <-chan Message: 共享的只读消息通道。
		Message() <-chan Message
	}
	// conn 将底层 net.Conn 包装为按本包协议异步收发消息的连接，
	// 同时实现 [Conn] 和 [net.Conn]。
	//
	// conn 的零值不可用，应通过 [WrapConn] 创建。Close、Closed 和 SendMessage
	// 可并发调用；Start 只应调用一次。
	conn struct {
		conn net.Conn // 承载原始协议字节流的底层网络连接，必须非 nil。

		closed       atomic.Bool   // 标记连接是否已进入关闭状态；true 后不会恢复。
		closedLocker sync.Locker   // 串行化首次关闭流程，避免并发重复关闭底层连接和通道。
		closedNotify chan struct{} // 首次关闭时关闭，用于通知内部 goroutine 退出。

		messageRead       chan Message // 供 Message 返回的共享接收 channel，首次 Close 时关闭。
		messageReadLocker sync.RWMutex // 协调接收 goroutine 投递消息与 Close 关闭 messageRead。
		messageWrite      chan Message // 内部异步发送队列，由 SendMessage 入队、send 出队；Close 不关闭该通道。

		heartbeatInterval time.Duration // 大于 0 时，Start 会按该间隔额外启动心跳发送循环；小于等于 0 时禁用心跳。
	}
)

// Closed 返回连接是否已经关闭。
//
// Closed 可与 Close 和 SendMessage 并发调用。
//
// 参数：无。
//
// 返回：
//   - bool: 连接是否已经进入关闭状态。
func (c *conn) Closed() bool {
	return c.closed.Load()
}

// Start 启动内部发送、接收以及可选心跳 goroutine。
//
// Start 不会自行去重，调用方只应调用一次。传入的上下文结束或连接关闭后，
// Start 启动的 goroutine 会退出；heartbeatInterval 大于 0 时，
// Start 会额外启动定时心跳发送 goroutine。
//
// 参数：
//   - ctx: 控制内部 goroutine 生命周期的上下文，不能为空。
func (c *conn) Start(ctx context.Context) {
	_ = kitgoroutine.Submit(func() { c.send(ctx) })    // 启动发送消息的 goroutine。
	_ = kitgoroutine.Submit(func() { c.receive(ctx) }) // 启动接收消息的 goroutine。

	if c.heartbeatInterval > 0 {
		ticker := time.NewTicker(c.heartbeatInterval)
		_ = kitgoroutine.Submit(func() { c.sendHeartbeat(ctx, ticker) }) // 启动定时发送心跳包的 goroutine。
	}
}

// SendMessage 将消息放入内部发送队列。
//
// SendMessage 可与 Close 和 Closed 并发调用。返回 nil 仅表示消息已入队，
// 不表示消息已经写入底层连接；当发送队列已满时会阻塞，直到队列腾出空间或连接关闭。
//
// 参数：
//   - message: 待异步发送的消息；调用方应保证其非 nil。
//
// 返回：
//   - error: 连接已关闭，或消息在入队前因收到关闭通知而被拒绝时返回错误。
func (c *conn) SendMessage(message Message) error {
	var err error

	if c.Closed() {
		err = cockroachdberrors.Newf("连接已经关闭。")
	} else {
		select {
		case <-c.closedNotify:
			err = cockroachdberrors.Newf("连接已经关闭。")
		case c.messageWrite <- message:
			// 将消息写入发送队列。
		}
	}

	return err
}

// Message 返回连接的共享接收 channel。
//
// 该 channel 只创建一次；多个消费者同时读取时会竞争消费消息。
// 连接关闭时，该 channel 也会被关闭。
//
// 参数：无。
//
// 返回：
//   - <-chan Message: 共享的只读消息通道。
func (c *conn) Message() <-chan Message {
	return c.messageRead
}

// Read 从底层连接读取原始协议字节流。
//
// 该方法直接委托给底层 net.Conn，不参与本类型的消息拆包流程。
//
// 参数：
//   - b: 读取缓冲区。
//
// 返回：
//   - int: 实际读取的字节数。
//   - error: 底层连接读取失败时返回错误。
func (c *conn) Read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}

// Write 向底层连接写入原始协议字节流。
//
// 该方法直接委托给底层 net.Conn，不参与本类型的消息封包流程。
//
// 参数：
//   - b: 待写入的原始字节数据。
//
// 返回：
//   - int: 实际写入的字节数。
//   - error: 底层连接写入失败时返回错误。
func (c *conn) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

// Close 关闭连接并通知内部 goroutine 退出。
//
// Close 可与 Closed 和 SendMessage 并发调用；重复调用只会在首次关闭时关闭底层连接，
// 并关闭 [Conn.Message] 返回的共享接收 channel。
//
// 参数：无。
//
// 返回：
//   - error: 首次关闭底层连接时返回的错误；连接已关闭时返回 nil。
func (c *conn) Close() error {
	var err error

	if !c.Closed() {
		// 可能出现不同的 goroutine 同时调用方法，需要加锁操作。
		c.closedLocker.Lock()
		defer c.closedLocker.Unlock()

		if !c.Closed() {
			c.closed.Store(true)
			close(c.closedNotify) // 通知发送、接收和心跳 goroutine 退出，避免关闭 messageWrite 后并发发送 panic。

			err = c.conn.Close()

			c.messageReadLocker.Lock()
			close(c.messageRead) // 等待接收 goroutine 完成可能的投递后，再关闭消息读取通道。
			c.messageReadLocker.Unlock()
		}
	}

	return err
}

// LocalAddr 返回底层连接的本地网络地址。
//
// 参数：无。
//
// 返回：
//   - net.Addr: 本地网络地址。
func (c *conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr 返回底层连接的远程网络地址。
//
// 参数：无。
//
// 返回：
//   - net.Addr: 远程网络地址。
func (c *conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline 设置底层连接的读写截止时间。
//
// 该设置会同时影响直接调用 Read 和 Write，以及 Start 启动的内部收发流程。
//
// 参数：
//   - t: 截止时间；零值表示取消已设置的读写截止时间。
//
// 返回：
//   - error: 底层连接设置截止时间失败时返回错误。
func (c *conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// SetReadDeadline 设置底层连接的读截止时间。
//
// 该设置会影响直接调用 Read，以及 Start 启动的内部接收流程。
//
// 参数：
//   - t: 截止时间；零值表示取消已设置的读截止时间。
//
// 返回：
//   - error: 底层连接设置读截止时间失败时返回错误。
func (c *conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline 设置底层连接的写截止时间。
//
// 该设置会影响直接调用 Write，以及 Start 启动的内部发送流程。
//
// 参数：
//   - t: 截止时间；零值表示取消已设置的写截止时间。
//
// 返回：
//   - error: 底层连接设置写截止时间失败时返回错误。
func (c *conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

// pack 将消息编码为本包协议定义的完整数据包。
//
// 返回结果依次包含 2 字节消息类型、2 字节 payload 长度和 payload 本体，
// 其中消息类型与长度字段均使用大端序编码。payload 长度超过 uint16 上限时返回错误。
//
// 参数：
//   - message: 待封包的消息；必须非 nil，否则调用 Pack 或 MessageType 时会 panic。
//
// 返回：
//   - []byte: 封包成功后的完整协议数据包。
//   - error: message.Pack 失败、payload 长度超限，或协议头与 payload 写入失败时返回错误。
func (c *conn) pack(message Message) ([]byte, error) {
	// 定义最终返回的数据包字节数组和错误变量。
	var data []byte
	var err error

	// 创建一个字节缓冲区，用于顺序写入消息各字段。
	buf := &bytes.Buffer{}

	// 步骤 1：调用消息的 Pack 方法获取 payload 数据。
	// 若 payload 封包失败，则直接返回错误。
	if payload, errPayload := message.Pack(); nil != errPayload {
		// 封包 payload 失败，进行错误包装并返回。
		err = cockroachdberrors.Wrap(errPayload, "消息负载封包出现错误。")
		// 步骤 2：独立校验 payload 长度是否超过 uint16 最大值（65535）。
		// 若超出限制，直接返回错误，不再进行后续写入操作。
	} else if payLoadLength := uint64(len(payload)); payLoadLength > math.MaxUint16 {
		// payload 长度超限，返回详细错误信息。
		err = cockroachdberrors.Newf("消息负载长度 %[1]d 超过 uint16 最大值 %[2]d。", payLoadLength, math.MaxUint16)
		// 步骤 3：写入消息类型字段（2 字节，uint16，BigEndian）。
		// 若写入失败，则返回错误。
	} else if errWriteType := binaryWrite(buf, binary.BigEndian, message.MessageType()); nil != errWriteType {
		// 写入消息类型失败，进行错误包装并返回。
		err = cockroachdberrors.Wrap(errWriteType, "消息类型封包出现错误。")
		// 步骤 4：写入 payload 长度字段（2 字节，uint16，BigEndian）。
		// 此时 payload 长度已保证不超限。
		// 若写入失败，则返回错误。
	} else if errWriteLen := binaryWrite(buf, binary.BigEndian, uint16(payLoadLength)); nil != errWriteLen { //nolint:gosec
		// 写入 payload 长度失败，进行错误包装并返回。
		err = cockroachdberrors.Wrap(errWriteLen, "消息负载长度封包出现错误。")
		// 步骤 5：写入 payload 数据本体。
		// 若写入失败，则返回错误。
	} else if errWritePayload := binaryWrite(buf, binary.BigEndian, payload); nil != errWritePayload {
		// 写入 payload 数据失败，进行错误包装并返回。
		err = cockroachdberrors.Wrap(errWritePayload, "消息负载封包出现错误。")
		// 步骤 6：所有字段写入成功，将缓冲区内容作为最终数据包返回。
	} else {
		data = buf.Bytes()
	}

	// 返回完整的数据包字节数组和错误信息。
	return data, err
}

// send 持续消费内部发送队列并将消息写入底层连接。
//
// send 会先调用 pack 生成协议数据包，再通过 Write 写入底层连接。
// ctx 结束、连接收到关闭通知、发送队列被关闭，或封包与写入失败时，
// send 会退出；其中除收到关闭通知外，其余异常路径都会主动关闭连接。
//
// 参数：
//   - ctx: 控制发送循环生命周期的上下文，不能为空。
func (c *conn) send(ctx context.Context) {
LoopSend:
	for {
		select {
		case <-ctx.Done():
			// 可能出现还有没消费完的信息。
			_ = c.Close()
			break LoopSend
		case <-c.closedNotify:
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

// generateMessage 从 scanner 读取一个完整协议包并还原为消息实例。
//
// scanner 应由 [NewScanner] 创建，以保证每次 Scan 返回的 token 都是完整协议包。
// 当 scanner 为 nil、scanner 已记录错误、数据流结束，或消息类型与 payload 还原失败时，
// generateMessage 返回错误。
//
// 参数：
//   - scanner: 按本包协议拆分消息包的扫描器；为 nil 时无法继续解析消息。
//
// 返回：
//   - Message: 成功还原的消息实例；解析失败时返回 nil。
//   - error: 扫描失败、数据流结束、消息类型读取失败，或调用 FactoryGenerate 还原消息失败时返回错误。
func (c *conn) generateMessage(scanner *bufio.Scanner) (Message, error) {
	// 定义返回的消息对象和错误变量。
	var message Message
	var err error

	// 检查传入的 scanner 是否为 nil，防止空指针异常。
	if nil == scanner {
		// 如果 scanner 为空，返回错误。
		err = cockroachdberrors.Newf("扫描器不能为空。")
		// 检查 scanner 是否已经发生错误。
	} else if errScanner := scanner.Err(); nil != errScanner {
		// 如果 scanner 内部有错误，进行错误包装并返回。
		err = cockroachdberrors.Wrap(errScanner, "扫描出错。")
		// 调用 scanner.Scan()，尝试扫描下一个 token（即一条完整消息包）。
	} else if !scanner.Scan() {
		// 如果 scanner.Scan() 返回 false，说明数据流已结束或无更多消息，返回明确错误。
		err = cockroachdberrors.Newf("数据流已结束或无更多消息。")
	} else {
		// 获取扫描到的字节数据。
		data := scanner.Bytes()
		// 定义消息类型变量，初始为 0。
		messageType := MessageType(uint16(0))
		// 从数据包前两个字节读取消息类型，采用大端序。
		if errReadType := binaryRead(bytes.NewReader(data[:2]), binary.BigEndian, &messageType); nil != errReadType {
			// 如果读取类型失败，进行错误包装并返回。
			err = cockroachdberrors.Wrap(errReadType, "解包数据类型发生异常。")
		} else {
			// 跳过前 4 字节（2 字节类型 + 2 字节长度），获取 payload 部分。
			payload := data[4:]
			// 调用工厂方法，根据消息类型和 payload 生成具体的消息对象。
			if msg, errGenerate := FactoryGenerate(messageType, payload); nil != errGenerate {
				// 如果生成消息失败，进行错误包装并返回。
				err = cockroachdberrors.Wrap(errGenerate, "数据包转消息发生异常。")
			} else {
				// 生成成功，赋值给返回变量。
				message = msg
			}
		}
	}

	// 返回生成的消息对象和错误信息。
	return message, err
}

// receive 持续从底层连接读取协议包并投递到共享消息通道。
//
// receive 使用 NewScanner 拆分完整协议包，并通过 generateMessage 还原消息。
// ctx 结束、连接收到关闭通知、消息解析失败、投递前观察到连接关闭，
// 或完成一次扫描后发现距离上次投递消息已超过超时阈值时，receive 会退出；
// 其中除收到关闭通知外，其余异常路径都会主动关闭连接。
//
// 参数：
//   - ctx: 控制接收循环生命周期的上下文，不能为空。
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
		case <-c.closedNotify:
			break LoopReceive
		default:
			if tmp, errGenerate := c.generateMessage(scanner); nil != errGenerate {
				_ = c.Close()
				break LoopReceive
			} else if s := time.Since(lastReceived).Milliseconds(); s > timeoutDuration {
				_ = c.Close()
				break LoopReceive
			} else if nil != tmp {
				c.messageReadLocker.RLock()
				if c.Closed() {
					c.messageReadLocker.RUnlock()
					break LoopReceive
				}
				select {
				case <-c.closedNotify:
					c.messageReadLocker.RUnlock()
					break LoopReceive
				case c.messageRead <- tmp:
					c.messageReadLocker.RUnlock()
					lastReceived = time.Now()
				}
			}
		}
	}
}

// sendHeartbeat 按 ticker 周期生成并发送心跳消息。
//
// 每次触发都会生成新的心跳序列号并调用 SendMessage 入队。ctx 结束或连接收到关闭通知时，
// sendHeartbeat 会退出；ctx 结束时还会主动关闭连接。本方法不会停止传入的 ticker。
//
// 参数：
//   - ctx: 控制心跳循环生命周期的上下文，不能为空。
//   - ticker: 心跳触发定时器，不能为空。
func (c *conn) sendHeartbeat(ctx context.Context, ticker *time.Ticker) {
	generateSerialNumber := func(serialNumberSingle uint16) (uint16, uint64) {
		const serialNumberSingleMax = 10000
		// 每次加 1。
		serialNumberSingle = serialNumberSingle + 1
		// 取模 10000。
		serialNumberSingle = serialNumberSingle % serialNumberSingleMax
		// 取格林威治时间戳。
		s1 := time.Now().UTC().Unix() * serialNumberSingleMax
		// 生成序列号。
		s := uint64(s1 + int64(serialNumberSingle))
		return serialNumberSingle, s
	}

	// 序列号单次增量的计数器。
	serialNumberSingle := uint16(0)
LoopHeartbeat:
	for {
		select {
		case <-ctx.Done():
			_ = c.Close()
			break LoopHeartbeat
		case <-c.closedNotify:
			break LoopHeartbeat
		case <-ticker.C:
			var serialNumber uint64
			serialNumberSingle, serialNumber = generateSerialNumber(serialNumberSingle)
			hm := NewHeartbeatMessage(serialNumber)
			_ = c.SendMessage(hm)
		}
	}
}

// WrapConn 将底层 net.Conn 包装为按本包协议异步收发消息的连接。
//
// 返回的连接会创建固定容量的接收与发送队列，但不会自动启动后台 goroutine；
// 调用方需要显式调用 [Conn.Start] 启动读写循环，且 Start 只应调用一次。
// heartbeatInterval 大于 0 时，Start 会额外启动定时心跳发送 goroutine。
//
// 参数：
//   - c: 待包装的底层网络连接，必须非 nil；调用方负责保证其满足所需的 net.Conn 语义。
//   - heartbeatInterval: 心跳发送间隔；小于等于 0 时不会启动心跳 goroutine。
//
// 返回：
//   - *conn: 包装后的协议连接实例，初始处于未关闭状态。
func WrapConn(c net.Conn, heartbeatInterval time.Duration) *conn {
	newConn := &conn{
		conn:              c,
		closedLocker:      &sync.Mutex{},
		closedNotify:      make(chan struct{}),
		messageRead:       make(chan Message, 5120), // 读取消息通道，缓冲区 5120。
		messageWrite:      make(chan Message, 5120), // 发送消息通道，缓冲区 5120。
		heartbeatInterval: heartbeatInterval,
	}

	return newConn
}
