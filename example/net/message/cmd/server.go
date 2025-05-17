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
	"time"

	"github.com/spf13/cobra"

	kitmessage "github.com/fsyyft-go/kit/net/message"
	kitruntime "github.com/fsyyft-go/kit/runtime"
	kitgoroutine "github.com/fsyyft-go/kit/runtime/goroutine"
)

var (
	serverCmd = &cobra.Command{
		Use: "server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGILL, syscall.SIGTERM)

			_ = kitgoroutine.Submit(func() {
				<-sigChan
				cancel()
			})

			s := &server{}
			return s.Run(ctx)
		},
	}
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var (
	_ kitruntime.Runner = (*server)(nil)
)

type (
	server struct {
		listen net.Listener
	}
)

func (s *server) Run(ctx context.Context) error {
	_ = kitgoroutine.Submit(func() {
		if err := s.Start(ctx); nil != err {
			fmt.Printf("服务启动失败：%v\n", err)
		}
	})

	time.Sleep(50 * time.Millisecond)
	if nil != s.listen {
		fmt.Printf("服务启动成功：%v\n", s.listen.Addr())
	} else {
		return nil
	}

	<-ctx.Done()
	_ = s.Stop(ctx)

	return ctx.Err()
}

func (s *server) Start(ctx context.Context) error {
	if listen, err := net.Listen("tcp", addr); nil != err {
		return err
	} else {
		s.listen = listen
	}

	for {
		conn, err := s.listen.Accept()
		if nil != err {
			return err
		}

		_ = kitgoroutine.Submit(func() {
			defer func() {
				_ = conn.Close()
				fmt.Printf("服务断开连接：%s\n", conn.RemoteAddr())
			}()

			conn := kitmessage.WrapConn(conn, 2*time.Second)

			conn.Start(ctx)
			if err := conn.SendMessage(kitmessage.NewSingleStringMessage(fmt.Sprintf("Hello, %s", conn.RemoteAddr()))); nil != err {
				fmt.Printf("发送握手消息失败：%s\n", err)
				return
			}
			fmt.Printf("服务接收连接：%s\n", conn.RemoteAddr())

			for {
				select {
				case <-ctx.Done():
					return
				case message, ok := <-conn.Message():
					if !ok {
						return
					}
					fmt.Println(message)
				}
			}
		})
	}
}

func (s *server) Stop(ctx context.Context) error {
	return s.listen.Close()
}
