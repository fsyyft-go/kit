// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	kitmessage "github.com/fsyyft-go/kit/net/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRootCommand_Topology 验证网络消息示例命令树的拓扑契约。
//
// 该测试覆盖 root、client 和 server 命令的稳定结构，避免通过执行 root 默认流程访问固定本地端口。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRootCommand_Topology(t *testing.T) {
	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "success/root-metadata",
			description: "验证 root 命令保留 network 用法、说明和默认执行函数。",
			assert: func(t *testing.T) {
				assert.Equal(t, "network", rootCmd.Use)
				assert.Equal(t, "基本网络消息处理", rootCmd.Short)
				assert.Contains(t, rootCmd.Long, "TCP 服务端和客户端")
				require.NotNil(t, rootCmd.RunE)
			},
		},
		{
			name:        "success/server-command-registered",
			description: "验证 server 子命令已注册并提供可执行入口。",
			assert: func(t *testing.T) {
				gotCommand, _, err := rootCmd.Find([]string{"server"})
				require.NoError(t, err)
				require.NotNil(t, gotCommand)
				assert.Same(t, serverCmd, gotCommand)
				assert.Equal(t, "server", gotCommand.Use)
				require.NotNil(t, gotCommand.RunE)
			},
		},
		{
			name:        "success/client-command-registered",
			description: "验证 client 子命令已注册并提供可执行入口。",
			assert: func(t *testing.T) {
				gotCommand, _, err := rootCmd.Find([]string{"client"})
				require.NoError(t, err)
				require.NotNil(t, gotCommand)
				assert.Same(t, clientCmd, gotCommand)
				assert.Equal(t, "client", gotCommand.Use)
				require.NotNil(t, gotCommand.RunE)
			},
		},
		{
			name:        "success/default-address-contract",
			description: "验证示例默认地址保持历史文档中的本地端口，但测试不会直接使用该固定端口。",
			assert: func(t *testing.T) {
				assert.Equal(t, "127.0.0.1:44444", addr)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			require.NotNil(t, tt.assert)
			tt.assert(t)
		})
	}
}

// TestRootCommand_DefaultRunUsesBoundedLifecycle 验证 root 默认流程可在受控生命周期内完成。
//
// 该测试将监听地址改为随机端口并缩短默认生命周期，覆盖 root 默认启动 server/client 的流程，同时避免依赖固定本地端口。
//
// 参数：
//   - t: 测试上下文，用于设置临时状态和报告断言失败。
func TestRootCommand_DefaultRunUsesBoundedLifecycle(t *testing.T) {
	// 该用例验证无参数执行 root 命令时会进入默认示例流程，并由超时上下文稳定退出。
	setMessageExampleAddr(t, "127.0.0.1:0")
	setRootLifecycleDelays(t, 30*time.Millisecond, time.Millisecond)

	gotErr := rootCmd.RunE(rootCmd, nil)

	require.NoError(t, gotErr)
}

// TestExecute_ErrorExitsProcess 验证 Execute 在 Cobra 执行失败时退出进程。
//
// 该测试通过子进程触发未知命令错误，避免 os.Exit 终止当前测试进程，同时覆盖 Execute 的错误处理契约。
//
// 参数：
//   - t: 测试上下文，用于启动子进程和报告断言失败。
func TestExecute_ErrorExitsProcess(t *testing.T) {
	if os.Getenv("KIT_TEST_NET_MESSAGE_EXECUTE_ERROR") == "1" {
		rootCmd.SetArgs([]string{"missing"})
		rootCmd.SetOut(io.Discard)
		rootCmd.SetErr(io.Discard)
		Execute()
		return
	}

	output, err := runNetMessageCommandSubprocess(t, "TestExecute_ErrorExitsProcess", "KIT_TEST_NET_MESSAGE_EXECUTE_ERROR")

	require.Error(t, err)
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 1, exitErr.ExitCode())
	assert.Contains(t, output, "unknown command")
}

// TestClientCommand_RunEStopsOnSignal 验证 client 子命令响应进程信号并结束运行。
//
// 该测试在子进程中发送 SIGTERM，覆盖 Cobra 子命令创建信号上下文并将取消结果传递给 client.Run 的契约。
//
// 参数：
//   - t: 测试上下文，用于启动子进程和报告断言失败。
func TestClientCommand_RunEStopsOnSignal(t *testing.T) {
	// 该用例验证 client 子命令建立连接后可以通过 SIGTERM 取消上下文并返回 context.Canceled。
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()
	setMessageExampleAddr(t, listener.Addr().String())
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr == nil {
			defer func() {
				_ = conn.Close()
			}()
			time.Sleep(200 * time.Millisecond)
		}
	}()
	signalDone := make(chan error, 1)
	setClientConnectedNotifier(t, func(net.Conn) {
		signalDone <- syscall.Kill(os.Getpid(), syscall.SIGTERM)
	})
	defer signal.Reset(syscall.SIGILL, syscall.SIGTERM)

	gotErr := clientCmd.RunE(clientCmd, nil)

	require.NoError(t, <-signalDone)
	require.ErrorIs(t, gotErr, context.Canceled)
}

// TestServerCommand_RunEStopsOnSignal 验证 server 子命令响应进程信号并结束运行。
//
// 该测试在子进程中使用随机监听端口并发送 SIGTERM，覆盖 Cobra 子命令创建信号上下文并将取消结果传递给 server.Run 的契约。
//
// 参数：
//   - t: 测试上下文，用于启动子进程和报告断言失败。
func TestServerCommand_RunEStopsOnSignal(t *testing.T) {
	// 该用例验证 server 子命令启动监听后可以通过 SIGTERM 取消上下文并返回 context.Canceled。
	setMessageExampleAddr(t, "127.0.0.1:0")
	signalDone := make(chan error, 1)
	setServerListenReadyNotifier(t, func(net.Listener) {
		signalDone <- syscall.Kill(os.Getpid(), syscall.SIGTERM)
	})
	defer signal.Reset(syscall.SIGILL, syscall.SIGTERM)

	gotErr := serverCmd.RunE(serverCmd, nil)

	require.NoError(t, <-signalDone)
	require.ErrorIs(t, gotErr, context.Canceled)
}

// TestServer_RunBranches 验证 server.Run 的监听失败与正常取消分支。
//
// 该测试通过表驱动用例覆盖监听地址非法时的空操作返回，以及随机端口监听成功后随上下文取消关闭监听器的生命周期语义。
//
// 参数：
//   - t: 测试上下文，用于运行子测试、设置临时状态和报告断言失败。
func TestServer_RunBranches(t *testing.T) {
	tests := []struct {
		name        string
		description string
		setup       func(t *testing.T) context.Context
		wantErrIs   error
	}{
		{
			name:        "error/listen-failure-returns-nil",
			description: "验证 server.Run 在监听地址非法导致 Start 失败时保持示例命令的兼容空返回。",
			setup: func(t *testing.T) context.Context {
				setMessageExampleAddr(t, "127.0.0.1:not-a-port")
				ctx, cancel := context.WithCancel(context.Background())
				t.Cleanup(cancel)
				return ctx
			},
		},
		{
			name:        "success/context-cancel-stops-listener",
			description: "验证 server.Run 在随机端口监听成功后会随上下文取消关闭监听器并返回取消错误。",
			setup: func(t *testing.T) context.Context {
				setMessageExampleAddr(t, "127.0.0.1:0")
				ctx, cancel := context.WithCancel(context.Background())
				t.Cleanup(cancel)
				setServerListenReadyNotifier(t, func(net.Listener) {
					cancel()
				})
				return ctx
			},
			wantErrIs: context.Canceled,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			ctx := tt.setup(t)
			gotErr := (&server{}).Run(ctx)

			if tt.wantErrIs != nil {
				assert.ErrorIs(t, gotErr, tt.wantErrIs)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

// TestClient_StartConnectionFailure 验证 client.Start 在连接失败时返回错误。
//
// 该测试使用不可拨号的 127.0.0.1:0 地址覆盖连接失败分支，避免访问固定本地服务。
//
// 参数：
//   - t: 测试上下文，用于设置临时状态和报告断言失败。
func TestClient_StartConnectionFailure(t *testing.T) {
	// 该用例验证 TCP 连接无法建立时 client.Start 会返回底层拨号错误。
	setMessageExampleAddr(t, "127.0.0.1:0")

	gotErr := (&client{}).Start(context.Background())

	require.Error(t, gotErr)
}

// TestServer_StartStopWithEphemeralPort 验证 server 可在随机本地端口启动、发送握手消息并停止。
//
// 该测试将示例监听地址临时设置为 127.0.0.1:0，使用系统分配端口完成本地 loopback 生命周期验证，避免依赖固定端口。
//
// 参数：
//   - t: 测试上下文，用于创建上下文、运行生命周期和报告断言失败。
func TestServer_StartStopWithEphemeralPort(t *testing.T) {
	// 该用例验证服务端使用随机本地端口时可以接受连接，并向客户端发送包含远端地址的字符串握手消息。
	setMessageExampleAddr(t, "127.0.0.1:0")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	listenReady := make(chan net.Listener, 1)
	setServerListenReadyNotifier(t, func(listener net.Listener) {
		listenReady <- listener
	})
	serverReceived := make(chan kitmessage.Message, 1)
	setServerMessageReceivedNotifier(t, func(message kitmessage.Message) {
		serverReceived <- message
	})

	giveServer := &server{}
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- giveServer.Start(ctx)
	}()

	listener := requireServerListener(t, listenReady)

	rawClient, err := net.Dial("tcp", listener.Addr().String())
	require.NoError(t, err)
	clientConn := kitmessage.WrapConn(rawClient, 0)
	clientConn.Start(ctx)

	gotMessage := requireMessageFromConn(t, clientConn)
	gotSingleString, ok := gotMessage.(kitmessage.SingleStringMessage)
	require.True(t, ok)
	assert.Equal(t, "Hello, "+rawClient.LocalAddr().String(), gotSingleString.Message())

	require.NoError(t, clientConn.SendMessage(kitmessage.NewSingleStringMessage("hello server")))
	gotServerMessage := requireMessageFromChannel(t, serverReceived)
	gotServerSingleString, ok := gotServerMessage.(kitmessage.SingleStringMessage)
	require.True(t, ok)
	assert.Equal(t, "hello server", gotServerSingleString.Message())

	cancel()
	assert.NoError(t, clientConn.Close())
	assert.NoError(t, giveServer.Stop(ctx))
	assert.Error(t, requireAsyncError(t, serverDone))
}

// TestClient_RunAgainstIsolatedListener 验证 client 可连接随机本地监听器并随上下文取消退出。
//
// 该测试使用测试创建的本地监听器替代固定示例端口，覆盖 client.Run 的成功连接和取消退出语义。
//
// 参数：
//   - t: 测试上下文，用于创建监听器、协调 goroutine 和报告断言失败。
func TestClient_RunAgainstIsolatedListener(t *testing.T) {
	// 该用例验证客户端读取可覆盖的 addr 变量，并在连接成功后由 Run 随上下文取消返回 context.Canceled。
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() {
		_ = listener.Close()
	}()
	setMessageExampleAddr(t, listener.Addr().String())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	accepted := make(chan net.Conn, 1)
	acceptDone := make(chan error, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			acceptDone <- acceptErr
			return
		}
		accepted <- conn
		<-ctx.Done()
		acceptDone <- conn.Close()
	}()

	clientConnected := make(chan net.Conn, 1)
	setClientConnectedNotifier(t, func(conn net.Conn) {
		clientConnected <- conn
	})

	giveClient := &client{}
	clientDone := make(chan error, 1)
	go func() {
		clientDone <- giveClient.Run(ctx)
	}()

	gotClientConn := requireClientConnectedConn(t, clientConnected)
	select {
	case gotConn := <-accepted:
		require.NotNil(t, gotConn)
		assert.Equal(t, listener.Addr().String(), gotClientConn.RemoteAddr().String())
	case <-time.After(time.Second):
		require.Fail(t, "等待服务端接受客户端连接超时")
	}

	cancel()
	assert.ErrorIs(t, requireAsyncError(t, clientDone), context.Canceled)
	assert.NoError(t, requireAsyncError(t, acceptDone))
}

// TestClient_StartReceivesMessageTypes 验证 client.Start 对不同消息类型的读取分支。
//
// 该测试使用本地随机端口和消息协议连接，向客户端发送心跳和字符串消息后关闭连接，覆盖客户端的消息分类处理分支。
//
// 参数：
//   - t: 测试上下文，用于运行子测试、协调本地连接和报告断言失败。
func TestClient_StartReceivesMessageTypes(t *testing.T) {
	tests := []struct {
		name        string
		description string
		setup       func(t *testing.T)
		giveMessage kitmessage.Message
	}{
		{
			name:        "success/heartbeat-message",
			description: "验证 client.Start 接收到心跳消息时进入心跳处理分支。",
			giveMessage: kitmessage.NewHeartbeatMessage(12345),
		},
		{
			name:        "success/single-string-message",
			description: "验证 client.Start 接收到字符串消息时进入字符串处理分支。",
			giveMessage: kitmessage.NewSingleStringMessage("hello client"),
		},
		{
			name:        "success/unknown-message",
			description: "验证 client.Start 接收到非心跳且非字符串消息时进入通用消息处理分支。",
			setup: func(t *testing.T) {
				registerUnknownMessageForClientStart(t)
			},
			giveMessage: unknownMessageForClientStart{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			if tt.setup != nil {
				tt.setup(t)
			}
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			defer func() {
				_ = listener.Close()
			}()
			setMessageExampleAddr(t, listener.Addr().String())
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			serverDone := make(chan error, 1)
			go func() {
				rawConn, acceptErr := listener.Accept()
				if acceptErr != nil {
					serverDone <- acceptErr
					return
				}
				serverConn := kitmessage.WrapConn(rawConn, 0)
				serverConn.Start(ctx)
				if sendErr := serverConn.SendMessage(tt.giveMessage); sendErr != nil {
					serverDone <- sendErr
					return
				}
				time.Sleep(20 * time.Millisecond)
				serverDone <- serverConn.Close()
			}()

			gotErr := (&client{}).Start(ctx)

			require.NoError(t, gotErr)
			assert.NoError(t, requireAsyncError(t, serverDone))
		})
	}
}

var (
	unknownMessageForClientStartType     = kitmessage.MessageType(0xfffe)
	registerUnknownMessageForClientOnce  sync.Once
	registerUnknownMessageForClientError error
)

type unknownMessageForClientStart struct{}

// MessageType 返回测试专用未知消息类型。
//
// 返回：
//   - kitmessage.MessageType: 不与内置消息类型冲突的测试消息类型。
func (unknownMessageForClientStart) MessageType() kitmessage.MessageType {
	return unknownMessageForClientStartType
}

// Pack 返回测试专用未知消息的空负载。
//
// 返回：
//   - []byte: 空负载字节切片。
//   - error: 固定为 nil，表示封包成功。
func (unknownMessageForClientStart) Pack() ([]byte, error) {
	return []byte{}, nil
}

// Unpack 接收测试专用未知消息负载。
//
// 参数：
//   - payload: 消息负载，测试实现不读取该参数。
//
// 返回：
//   - error: 固定为 nil，表示解包成功。
func (unknownMessageForClientStart) Unpack(payload []byte) error {
	return nil
}

// registerUnknownMessageForClientStart 注册 client.Start 通用消息分支所需的测试消息类型。
//
// 该辅助函数通过 sync.Once 避免重复注册默认消息工厂，并在注册失败时立即报告夹具错误。
//
// 参数：
//   - t: 测试上下文，用于报告注册失败并标记辅助函数调用栈。
func registerUnknownMessageForClientStart(t *testing.T) {
	t.Helper()

	registerUnknownMessageForClientOnce.Do(func() {
		registerUnknownMessageForClientError = kitmessage.FactoryRegister(
			unknownMessageForClientStartType,
			func(messageType kitmessage.MessageType, payload []byte) (kitmessage.Message, error) {
				return unknownMessageForClientStart{}, nil
			},
		)
	})
	require.NoError(t, registerUnknownMessageForClientError)
}

// setMessageExampleAddr 临时覆盖网络消息示例使用的地址。
//
// 该辅助函数恢复包级 addr 变量，确保使用随机端口的测试不会影响后续用例和示例默认值断言。
//
// 参数：
//   - t: 测试上下文，用于注册清理逻辑并标记辅助函数调用栈。
//   - giveAddr: 本用例需要使用的监听或连接地址。
func setMessageExampleAddr(t *testing.T, giveAddr string) {
	t.Helper()

	oldAddr := addr
	addr = giveAddr
	t.Cleanup(func() {
		addr = oldAddr
	})
}

// setServerListenReadyNotifier 临时覆盖服务端监听就绪通知函数。
//
// 该辅助函数恢复包级 notifyServerListenReady 变量，使生命周期测试可以通过 channel 同步监听器地址，避免读取业务 goroutine 写入的 server.listen 字段。
//
// 参数：
//   - t: 测试上下文，用于注册清理逻辑并标记辅助函数调用栈。
//   - giveNotify: 本用例使用的监听器就绪通知函数。
func setServerListenReadyNotifier(t *testing.T, giveNotify func(net.Listener)) {
	t.Helper()
	require.NotNil(t, giveNotify)

	oldNotify := notifyServerListenReady
	notifyServerListenReady = giveNotify
	t.Cleanup(func() {
		notifyServerListenReady = oldNotify
	})
}

// setServerMessageReceivedNotifier 临时覆盖服务端消息接收通知函数。
//
// 该辅助函数恢复包级 notifyServerMessageReceived 变量，使服务端生命周期测试可以稳定断言服务端已收到客户端消息。
//
// 参数：
//   - t: 测试上下文，用于注册清理逻辑并标记辅助函数调用栈。
//   - giveNotify: 本用例使用的服务端消息接收通知函数。
func setServerMessageReceivedNotifier(t *testing.T, giveNotify func(kitmessage.Message)) {
	t.Helper()
	require.NotNil(t, giveNotify)

	oldNotify := notifyServerMessageReceived
	notifyServerMessageReceived = giveNotify
	t.Cleanup(func() {
		notifyServerMessageReceived = oldNotify
	})
}

// setClientConnectedNotifier 临时覆盖客户端连接成功通知函数。
//
// 该辅助函数恢复包级 notifyClientConnected 变量，使生命周期测试可以通过 channel 同步客户端连接，避免取消后读取 client.conn 时与赋值并发。
//
// 参数：
//   - t: 测试上下文，用于注册清理逻辑并标记辅助函数调用栈。
//   - giveNotify: 本用例使用的客户端连接成功通知函数。
func setClientConnectedNotifier(t *testing.T, giveNotify func(net.Conn)) {
	t.Helper()
	require.NotNil(t, giveNotify)

	oldNotify := notifyClientConnected
	notifyClientConnected = giveNotify
	t.Cleanup(func() {
		notifyClientConnected = oldNotify
	})
}

// setRootLifecycleDelays 临时覆盖 root 默认示例流程的生命周期时长。
//
// 该辅助函数恢复包级超时和启动等待配置，确保 root 默认流程测试不会影响其他命令生命周期用例。
//
// 参数：
//   - t: 测试上下文，用于注册清理逻辑并标记辅助函数调用栈。
//   - giveRunTimeout: root 默认流程使用的上下文超时时间。
//   - giveStartupWaitDelay: 启动客户端前等待服务端的时间。
func setRootLifecycleDelays(t *testing.T, giveRunTimeout, giveStartupWaitDelay time.Duration) {
	t.Helper()

	oldRunTimeout := rootRunTimeout
	oldStartupWaitDelay := startupWaitDelay
	rootRunTimeout = giveRunTimeout
	startupWaitDelay = giveStartupWaitDelay
	t.Cleanup(func() {
		rootRunTimeout = oldRunTimeout
		startupWaitDelay = oldStartupWaitDelay
	})
}

// runNetMessageCommandSubprocess 在子进程中运行指定测试分支。
//
// 该辅助函数通过环境变量激活当前测试函数的子进程分支，用于验证 os.Exit 和进程信号流程而不终止父测试进程。
//
// 参数：
//   - t: 测试上下文，用于报告子进程启动错误并标记辅助函数调用栈。
//   - giveTestName: 需要在子进程中执行的测试函数名称。
//   - giveEnvName: 激活子进程分支的环境变量名称。
//
// 返回：
//   - string: 子进程合并后的标准输出和标准错误。
//   - error: 子进程退出错误；退出码非零时返回 *exec.ExitError。
func runNetMessageCommandSubprocess(t *testing.T, giveTestName, giveEnvName string) (string, error) {
	t.Helper()

	cmd := exec.Command(os.Args[0], "-test.run=^"+giveTestName+"$")
	cmd.Env = append(os.Environ(), giveEnvName+"=1")
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// requireServerListener 等待服务端监听器就绪并返回该监听器。
//
// 该辅助函数为服务端启动同步点提供超时保护，确保测试使用同步信号获取监听地址。
//
// 参数：
//   - t: 测试上下文，用于报告等待失败并标记辅助函数调用栈。
//   - ready: 承载服务端监听器的同步 channel。
//
// 返回：
//   - net.Listener: 已经完成监听的服务端监听器。
func requireServerListener(t *testing.T, ready <-chan net.Listener) net.Listener {
	t.Helper()

	select {
	case listener := <-ready:
		require.NotNil(t, listener)
		return listener
	case <-time.After(time.Second):
		require.Fail(t, "等待服务端监听器就绪超时")
		return nil
	}
}

// requireClientConnectedConn 等待客户端连接成功并返回该连接。
//
// 该辅助函数为客户端启动同步点提供超时保护，确保测试在取消 Run 上下文前已观察到连接赋值完成。
//
// 参数：
//   - t: 测试上下文，用于报告等待失败并标记辅助函数调用栈。
//   - connected: 承载客户端连接的同步 channel。
//
// 返回：
//   - net.Conn: 已经建立的客户端连接。
func requireClientConnectedConn(t *testing.T, connected <-chan net.Conn) net.Conn {
	t.Helper()

	select {
	case conn := <-connected:
		require.NotNil(t, conn)
		return conn
	case <-time.After(time.Second):
		require.Fail(t, "等待客户端连接成功超时")
		return nil
	}
}

// requireMessageFromConn 等待并返回消息连接中的下一条消息。
//
// 该辅助函数为异步网络消息读取提供超时保护，避免测试在连接异常时永久阻塞。
//
// 参数：
//   - t: 测试上下文，用于报告等待失败并标记辅助函数调用栈。
//   - conn: 待读取的消息连接。
//
// 返回：
//   - kitmessage.Message: 连接中读取到的消息。
func requireMessageFromConn(t *testing.T, conn kitmessage.Conn) kitmessage.Message {
	t.Helper()

	return requireMessageFromChannel(t, conn.Message())
}

// requireMessageFromChannel 等待并返回消息通道中的下一条消息。
//
// 该辅助函数为异步消息接收通知提供超时保护，避免测试在服务端未收到消息时永久阻塞。
//
// 参数：
//   - t: 测试上下文，用于报告等待失败并标记辅助函数调用栈。
//   - messageCh: 待读取的消息通道。
//
// 返回：
//   - kitmessage.Message: 通道中读取到的消息。
func requireMessageFromChannel(t *testing.T, messageCh <-chan kitmessage.Message) kitmessage.Message {
	t.Helper()

	select {
	case gotMessage, ok := <-messageCh:
		require.True(t, ok)
		require.NotNil(t, gotMessage)
		return gotMessage
	case <-time.After(time.Second):
		require.Fail(t, "等待消息超时")
		return nil
	}
}

// requireAsyncError 等待异步 goroutine 返回错误。
//
// 该辅助函数统一为 server/client 生命周期 goroutine 提供超时保护，并返回其结果供调用方断言。
//
// 参数：
//   - t: 测试上下文，用于报告等待失败并标记辅助函数调用栈。
//   - done: 承载 goroutine 返回错误的 channel。
//
// 返回：
//   - error: goroutine 返回的错误。
func requireAsyncError(t *testing.T, done <-chan error) error {
	t.Helper()

	select {
	case err := <-done:
		if errors.Is(err, net.ErrClosed) {
			return err
		}
		return err
	case <-time.After(time.Second):
		require.Fail(t, "等待异步生命周期结束超时")
		return nil
	}
}
