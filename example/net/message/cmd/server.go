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
	"sync"
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

var (
	// notifyServerListenReady 在服务端监听器完成赋值后发送同步通知，生产路径保持空操作。
	notifyServerListenReady = func(net.Listener) {}
	// notifyServerMessageReceived 在服务端收到消息后发送同步通知，生产路径保持空操作。
	notifyServerMessageReceived = func(kitmessage.Message) {}
)

type (
	server struct {
		addr              string
		listen            net.Listener
		listenMu          sync.RWMutex
		notifyListenReady func(net.Listener)
	}
)

func (s *server) Run(ctx context.Context) error {
	s.ensureAddr()

	listenReady := make(chan net.Listener, 1)
	startDone := make(chan error, 1)
	oldNotifyListenReady := s.notifyListenReady
	s.notifyListenReady = func(listener net.Listener) {
		select {
		case listenReady <- listener:
		default:
		}
		if oldNotifyListenReady != nil {
			oldNotifyListenReady(listener)
		}
	}

	if err := kitgoroutine.Submit(func() {
		startDone <- s.Start(ctx)
	}); nil != err {
		return err
	}

	select {
	case listener := <-listenReady:
		fmt.Printf("服务启动成功：%v\n", listener.Addr())
	case err := <-startDone:
		if nil != err {
			fmt.Printf("服务启动失败：%v\n", err)
		}
		return nil
	case <-ctx.Done():
		_ = s.Stop(ctx)
		<-startDone
		return ctx.Err()
	}

	<-ctx.Done()
	_ = s.Stop(ctx)
	<-startDone

	return ctx.Err()
}

func (s *server) Start(ctx context.Context) error {
	listen, err := net.Listen("tcp", s.address())
	if nil != err {
		return err
	}
	s.setListen(listen)
	s.notifyReady(listen)
	select {
	case <-ctx.Done():
		_ = listen.Close()
		return ctx.Err()
	default:
	}

	for {
		conn, err := listen.Accept()
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
					notifyServerMessageReceived(message)
					fmt.Println(message)
				}
			}
		})
	}
}

func (s *server) Stop(ctx context.Context) error {
	listen := s.listener()
	if listen == nil {
		return nil
	}
	return listen.Close()
}

func (s *server) ensureAddr() {
	if s.addr == "" {
		s.addr = addr
	}
}

func (s *server) address() string {
	s.ensureAddr()
	return s.addr
}

func (s *server) setListen(listen net.Listener) {
	s.listenMu.Lock()
	defer s.listenMu.Unlock()

	s.listen = listen
}

func (s *server) listener() net.Listener {
	s.listenMu.RLock()
	defer s.listenMu.RUnlock()

	return s.listen
}

func (s *server) notifyReady(listen net.Listener) {
	if s.notifyListenReady != nil {
		s.notifyListenReady(listen)
	}
	notifyServerListenReady(listen)
}
