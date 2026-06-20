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

	kitgoroutine "github.com/fsyyft-go/kit/runtime/goroutine"
)

var (
	// rootCmd 表示在没有调用任何子命令时的基础命令。
	rootCmd = &cobra.Command{
		// 指定命令的名称。
		Use: "network",
		// 简短的命令描述。
		Short: "基本网络消息处理",
		// 详细的命令描述。
		Long: `提供一个简易的 TCP 服务端和客户端，用于测试消息通过网络传输后的封包与拆包。`,
		// RunE 函数定义了命令的执行逻辑。
		RunE: func(cmd *cobra.Command, args []string) error {
			// 当没有提供子命令时，运行示例函数。
			if len(args) == 0 {
				return runRootDefaultExample()
			}
			return nil
		},
	}
)

var (
	addr             = "127.0.0.1:44444"
	rootRunTimeout   = 5 * time.Second
	startupWaitDelay = 50 * time.Millisecond
)

// runRootDefaultExample 运行 root 命令的默认网络消息示例。
//
// 该函数在启动时快照当前地址，并等待服务端与客户端 goroutine 完成，避免示例退出后继续读取包级状态。
func runRootDefaultExample() error {
	ctx, cancel := context.WithTimeout(context.Background(), rootRunTimeout)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGILL, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	signalDone := make(chan struct{})
	if err := kitgoroutine.Submit(func() {
		defer close(signalDone)
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
		}
	}); err != nil {
		return err
	}

	var waitGroup sync.WaitGroup
	serverReady := make(chan net.Listener, 1)
	serverDone := make(chan error, 1)
	serverAddr := addr
	s := &server{addr: serverAddr}
	s.notifyListenReady = func(listener net.Listener) {
		select {
		case serverReady <- listener:
		default:
		}
	}

	cleanup := func() {
		cancel()
		_ = s.Stop(ctx)
		waitGroup.Wait()
		<-signalDone
	}
	defer cleanup()

	waitGroup.Add(1)
	if err := kitgoroutine.Submit(func() {
		defer waitGroup.Done()
		serverDone <- s.Run(ctx)
	}); err != nil {
		waitGroup.Done()
		return err
	}

	var listener net.Listener
	select {
	case listener = <-serverReady:
	case <-serverDone:
		return nil
	case <-ctx.Done():
		return nil
	}

	c := &client{addr: listener.Addr().String()}
	waitGroup.Add(1)
	if err := kitgoroutine.Submit(func() {
		defer waitGroup.Done()
		_ = c.Run(ctx)
	}); err != nil {
		waitGroup.Done()
		return err
	}
	defer func() {
		_ = c.Stop(ctx)
	}()

	<-ctx.Done()

	return nil
}

// Execute 将所有子命令添加到根命令并适当设置标志。
// 这个函数由 main.main() 调用，只需要对 rootCmd 执行一次。
func Execute() {
	// 执行根命令，如果出现错误则打印错误信息并退出程序。
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
