// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
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
			c := &client{}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			_ = kitgoroutine.Submit(func() { _ = c.Start(ctx) })

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGILL, syscall.SIGTERM)

			<-sigChan
			_ = c.Stop(ctx)

			return nil
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

func (c *client) Start(ctx context.Context) error {
	if conn, err := net.Dial("tcp", addr); nil != err {
		fmt.Printf("client dial %s failed: %v\n", addr, err)
		return err
	} else {
		fmt.Printf("client dial %s success: %v\n", addr, conn.LocalAddr())
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
				fmt.Printf("client receive heartbeat: %d\n", hm.SerialNumber())
			} else {
				fmt.Printf("client receive: %v\n", message)
			}
		}
	}
}

func (c *client) Stop(ctx context.Context) error {
	return c.conn.Close()
}
