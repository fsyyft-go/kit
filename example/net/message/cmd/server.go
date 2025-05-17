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
			s := &server{}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			_ = kitgoroutine.Submit(func() {
				if err := s.Start(ctx); nil != err {
					fmt.Printf("server start failed: %v\n", err)
				}
			})

			time.Sleep(300 * time.Millisecond)
			if nil != s.listen {
				fmt.Printf("server start success: %v\n", s.listen.Addr())
			} else {
				return nil
			}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGILL, syscall.SIGTERM)

			<-sigChan
			_ = s.Stop(ctx)

			return nil
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
			}()

			conn := kitmessage.WrapConn(conn, 3*time.Second)

			conn.Start(ctx)

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
