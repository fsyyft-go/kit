// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"math"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"

	kitmessage "github.com/fsyyft-go/kit/net/message"
	kitruntime "github.com/fsyyft-go/kit/runtime"
	kitgoroutine "github.com/fsyyft-go/kit/runtime/goroutine"
)

var (
	clientCmd = &cobra.Command{
		Use: "client",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGILL, syscall.SIGTERM)

			_ = kitgoroutine.Submit(func() {
				<-sigChan
				cancel()
			})

			c := &client{}
			return c.Run(ctx)
		},
	}
)

func init() {
	rootCmd.AddCommand(clientCmd)
}

var (
	_ kitruntime.Runner = (*client)(nil)
)

var (
	// notifyClientConnected 在客户端连接完成赋值后发送同步通知，生产路径保持空操作。
	notifyClientConnected = func(net.Conn) {}
)

type (
	client struct {
		addr   string
		conn   net.Conn
		connMu sync.RWMutex
	}
)

func (c *client) Run(ctx context.Context) error {
	c.ensureAddr()

	startDone := make(chan error, 1)
	if err := kitgoroutine.Submit(func() { startDone <- c.Start(ctx) }); nil != err {
		return err
	}

	select {
	case <-startDone:
		<-ctx.Done()
	case <-ctx.Done():
		_ = c.Stop(ctx)
		<-startDone
	}

	return ctx.Err()
}

func (c *client) Start(ctx context.Context) error {
	remoteAddr := c.address()
	if conn, err := net.Dial("tcp", remoteAddr); nil != err {
		fmt.Printf("客户端连接失败：%[1]s\n", err.Error())
		return err
	} else {
		fmt.Printf("客户端连接成功：%[1]s -> %[2]s\n", conn.LocalAddr(), remoteAddr)
		c.setConn(conn)
		notifyClientConnected(conn)
	}

	conn := kitmessage.WrapConn(c.connection(), 0)

	conn.Start(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case message, ok := <-conn.Message():
			if !ok {
				return nil
			}
			if hm, ok := message.(kitmessage.HeartbeatMessage); ok {
				serialNumber := hm.SerialNumber()
				fmt.Printf("客户端收到心跳数据包: %d %d\n", serialNumber, math.MaxUint64-serialNumber)
			} else if sm, ok := message.(kitmessage.SingleStringMessage); ok {
				fmt.Printf("客户端收到字符串数据包: %s\n", sm.Message())
			} else {
				fmt.Printf("客户端收到数据包: %v\n", message)
			}
		}
	}
}

func (c *client) Stop(ctx context.Context) error {
	conn := c.connection()
	if conn == nil {
		return nil
	}
	return conn.Close()
}

func (c *client) ensureAddr() {
	if c.addr == "" {
		c.addr = addr
	}
}

func (c *client) address() string {
	c.ensureAddr()
	return c.addr
}

func (c *client) setConn(conn net.Conn) {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	c.conn = conn
}

func (c *client) connection() net.Conn {
	c.connMu.RLock()
	defer c.connMu.RUnlock()

	return c.conn
}
