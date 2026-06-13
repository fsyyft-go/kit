// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package goroutine

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/stretchr/testify/require"
)

// goroutineTestTimeout 定义异步测试等待事件完成的最大时间。
const goroutineTestTimeout = 2 * time.Second

// newGoroutinePoolForTest 创建默认关闭指标采集的协程池测试夹具。
//
// 该辅助函数集中创建可自动清理的协程池，避免单元测试遗留 worker 或指标采集 goroutine。
// 调用方传入的配置会覆盖默认配置，因此可以显式开启 metrics 或其他选项。
//
// 参数：
//   - t: 测试上下文，用于报告夹具创建失败并注册清理逻辑。
//   - opts: 协程池配置选项，会追加在默认的 WithMetrics(false) 之后。
//
// 返回：
//   - GoroutinePool: 已创建并会在测试结束时释放的协程池实例。
func newGoroutinePoolForTest(t *testing.T, opts ...Option) GoroutinePool {
	t.Helper()

	allOpts := append([]Option{WithMetrics(false)}, opts...)
	pool, cleanup, err := NewGoroutinePool(allOpts...)
	require.NoError(t, err)
	require.NotNil(t, pool)
	require.NotNil(t, cleanup)
	t.Cleanup(cleanup)

	return pool
}

// receiveWithin 在限定时间内从通道接收一个值。
//
// 该辅助函数用于验证异步任务已经执行，避免测试依赖固定 sleep 或无限期阻塞。
//
// 参数：
//   - t: 测试上下文，用于报告超时失败并标记辅助函数调用栈。
//   - ch: 待读取的只读通道。
//   - description: 当前等待行为的业务语义，用于生成可诊断的失败信息。
//
// 返回：
//   - T: 通道中接收到的值。
func receiveWithin[T any](t *testing.T, ch <-chan T, description string) T {
	t.Helper()

	select {
	case got := <-ch:
		return got
	case <-time.After(goroutineTestTimeout):
		require.FailNowf(t, "timed out waiting for asynchronous event", "%s did not complete within %s", description, goroutineTestTimeout)
	}

	var zero T
	return zero
}

// requireWaitGroupWithin 在限定时间内等待 WaitGroup 完成。
//
// 该辅助函数为并发测试提供超时保护，避免测试在任务未完成时无限阻塞。
//
// 参数：
//   - t: 测试上下文，用于报告超时失败并标记辅助函数调用栈。
//   - wg: 需要等待完成的 WaitGroup。
//   - timeout: 等待 WaitGroup 完成的最大时间。
//   - description: 当前等待行为的业务语义，用于生成可诊断的失败信息。
func requireWaitGroupWithin(t *testing.T, wg *sync.WaitGroup, timeout time.Duration, description string) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		require.FailNowf(t, "timed out waiting for wait group", "%s did not complete within %s", description, timeout)
	}
}

// closeGoroutinePoolForTest 释放测试期间可能残留的具体协程池实例。
//
// 该辅助函数用于全局默认池隔离场景，尽力通知指标 goroutine 退出并释放 ants 池资源。
//
// 参数：
//   - t: 测试上下文，用于标记辅助函数调用栈。
//   - pool: 需要释放的具体协程池实例，nil 时不执行任何操作。
func closeGoroutinePoolForTest(t *testing.T, pool *goroutinePool) {
	t.Helper()

	if pool == nil {
		return
	}

	select {
	case pool.closed <- struct{}{}:
	default:
	}
	if pool.pool != nil {
		err := pool.pool.ReleaseTimeout(goroutineTestTimeout)
		if err != nil && !errors.Is(err, ants.ErrPoolClosed) {
			require.NoError(t, err)
		}
	}
}

// isolateDefaultPoolForTest 隔离并恢复默认协程池相关全局状态。
//
// 该辅助函数用于测试包级 Submit，确保 sizeDefault、metricsDefault 等默认配置在用例结束后恢复，
// 同时关闭测试期间创建的默认池，避免全局 worker 或指标采集 goroutine 污染后续测试。
//
// 参数：
//   - t: 测试上下文，用于注册全局状态恢复和资源清理逻辑。
func isolateDefaultPoolForTest(t *testing.T) {
	t.Helper()

	poolDefaultLocker.Lock()
	previousPool := poolDefault
	previousSizeDefault := sizeDefault
	previousExpiryDefault := expiryDefault
	previousPreAllocDefault := preAllocDefault
	previousNonBlockingDefault := nonBlockingDefault
	previousMaxBlockingDefault := maxBlockingDefault
	previousPanicHandlerDefault := panicHandlerDefault
	previousMetricsDefault := metricsDefault
	poolDefault = nil
	poolDefaultLocker.Unlock()

	closeGoroutinePoolForTest(t, previousPool)

	t.Cleanup(func() {
		poolDefaultLocker.Lock()
		currentPool := poolDefault
		poolDefault = nil
		sizeDefault = previousSizeDefault
		expiryDefault = previousExpiryDefault
		preAllocDefault = previousPreAllocDefault
		nonBlockingDefault = previousNonBlockingDefault
		maxBlockingDefault = previousMaxBlockingDefault
		panicHandlerDefault = previousPanicHandlerDefault
		metricsDefault = previousMetricsDefault
		poolDefaultLocker.Unlock()

		closeGoroutinePoolForTest(t, currentPool)
	})
}
