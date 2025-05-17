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

type (
	client struct {
		conn net.Conn
	}
)

func (c *client) Run(ctx context.Context) error {
	if err := kitgoroutine.Submit(func() { _ = c.Start(ctx) }); nil != err {
		return err
	}

	<-ctx.Done()
	_ = c.Stop(ctx)

	return ctx.Err()
}

func (c *client) Start(ctx context.Context) error {
	if conn, err := net.Dial("tcp", addr); nil != err {
		fmt.Printf("客户端连接失败：%[1]s\n", err.Error())
		return err
	} else {
		fmt.Printf("客户端连接成功：%[1]s -> %[2]s\n", conn.LocalAddr(), addr)
		c.conn = conn
	}

	conn := kitmessage.WrapConn(c.conn, 0)

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
	return c.conn.Close()
}
