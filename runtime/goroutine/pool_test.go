// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package goroutine

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewGoroutinePool 验证协程池创建时会应用默认配置和显式 Option。
//
// 该测试通过表驱动用例覆盖默认、自定义和最小配置，确保 WithSize、WithPreAlloc、
// WithNonBlocking、WithName 和 WithMetrics 等配置具有可观察结果。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewGoroutinePool(t *testing.T) {
	const giveCustomPoolName = "test-new-goroutine-pool-custom"
	giveCustomExpiry := 25 * time.Millisecond

	tests := []struct {
		name        string
		description string
		opts        []Option
		assert      func(t *testing.T, pool GoroutinePool)
	}{
		{
			name:        "success/default-config",
			description: "验证默认配置会创建未关闭的协程池，并保留默认容量和指标开关。",
			assert: func(t *testing.T, pool GoroutinePool) {
				t.Helper()

				concretePool, ok := pool.(*goroutinePool)
				require.True(t, ok)
				assert.Equal(t, sizeDefault, pool.Cap())
				assert.False(t, pool.IsClosed())
				assert.True(t, concretePool.metrics)
			},
		},
		{
			name:        "success/custom-observable-options",
			description: "验证自定义大小、预分配、非阻塞、名称和关闭指标配置会应用到协程池。",
			opts: []Option{
				WithSize(10),
				WithExpiry(giveCustomExpiry),
				WithPreAlloc(true),
				WithNonBlocking(true),
				WithMaxBlocking(100),
				WithName(giveCustomPoolName),
				WithMetrics(false),
			},
			assert: func(t *testing.T, pool GoroutinePool) {
				t.Helper()

				concretePool, ok := pool.(*goroutinePool)
				require.True(t, ok)
				assert.Equal(t, 10, pool.Cap())
				assert.Equal(t, 10, pool.Free())
				assert.Equal(t, giveCustomExpiry, concretePool.expiry)
				assert.True(t, concretePool.preAlloc)
				assert.True(t, concretePool.nonBlocking)
				assert.Equal(t, 100, concretePool.maxBlocking)
				assert.Equal(t, giveCustomPoolName, concretePool.name)
				assert.False(t, concretePool.metrics)
			},
		},
		{
			name:        "success/minimal-single-worker",
			description: "验证最小配置会创建单 worker、阻塞模式且关闭指标采集的协程池。",
			opts: []Option{
				WithSize(1),
				WithExpiry(time.Millisecond),
				WithPreAlloc(false),
				WithNonBlocking(false),
				WithMaxBlocking(0),
				WithMetrics(false),
			},
			assert: func(t *testing.T, pool GoroutinePool) {
				t.Helper()

				concretePool, ok := pool.(*goroutinePool)
				require.True(t, ok)
				assert.Equal(t, 1, pool.Cap())
				assert.False(t, concretePool.preAlloc)
				assert.False(t, concretePool.nonBlocking)
				assert.Equal(t, 0, concretePool.maxBlocking)
				assert.False(t, concretePool.metrics)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			pool, cleanup, err := NewGoroutinePool(tt.opts...)
			require.NoError(t, err)
			require.NotNil(t, pool)
			require.NotNil(t, cleanup)
			t.Cleanup(cleanup)

			tt.assert(t, pool)
		})
	}
}

// TestGoroutinePool_Submit 验证协程池提交任务的执行、panic 处理和后续可用性。
//
// 该测试通过表驱动用例覆盖普通任务、最小可观测任务和 panic 任务，确保 Submit 不仅返回成功，
// 还会实际执行任务；panic 任务通过自定义 panic handler 证明 panic 已被捕获且池仍可继续使用。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGoroutinePool_Submit(t *testing.T) {
	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "success/normal-task-executes",
			description: "验证普通任务提交成功后会被协程池实际执行。",
			assert: func(t *testing.T) {
				t.Helper()

				pool := newGoroutinePoolForTest(t, WithSize(2))
				executed := make(chan struct{}, 1)

				err := pool.Submit(func() {
					executed <- struct{}{}
				})

				require.NoError(t, err)
				receiveWithin(t, executed, "normal task execution")
			},
		},
		{
			name:        "success/minimal-task-executes",
			description: "验证只包含可观测完成信号的最小任务提交成功后会被协程池执行。",
			assert: func(t *testing.T) {
				t.Helper()

				pool := newGoroutinePoolForTest(t, WithSize(2))
				executed := make(chan struct{}, 1)

				err := pool.Submit(func() {
					executed <- struct{}{}
				})

				require.NoError(t, err)
				receiveWithin(t, executed, "minimal task execution")
			},
		},
		{
			name:        "panic/handler-captures-and-pool-remains-usable",
			description: "验证任务 panic 会被 panic handler 捕获，并且协程池仍可继续执行后续任务。",
			assert: func(t *testing.T) {
				t.Helper()

				panicCaught := make(chan interface{}, 1)
				pool := newGoroutinePoolForTest(t,
					WithSize(1),
					WithPanicHandler(func(r interface{}) {
						panicCaught <- r
					}),
				)

				err := pool.Submit(func() {
					panic("test panic")
				})
				require.NoError(t, err)
				assert.Equal(t, "test panic", receiveWithin(t, panicCaught, "panic handler invocation"))

				executed := make(chan struct{}, 1)
				err = pool.Submit(func() {
					executed <- struct{}{}
				})
				require.NoError(t, err)
				receiveWithin(t, executed, "task execution after panic")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			tt.assert(t)
		})
	}
}

// TestGoroutinePool_SubmitAfterClose 验证 cleanup 关闭协程池后 Submit 返回错误。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestGoroutinePool_SubmitAfterClose(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "error/submit-after-cleanup",
			description: "验证 cleanup 关闭协程池后，继续提交任务会返回错误。",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			pool, cleanup, err := NewGoroutinePool()
			require.NoError(t, err)
			cleanup()

			err = pool.Submit(func() {})
			assert.Error(t, err, "向已关闭的池提交任务应该返回错误")
		})
	}
}

// TestGoroutinePool_NonBlocking 验证非阻塞模式在 worker 被占用时拒绝新任务。
//
// 该测试使用 started/release 通道精确控制首个任务占用唯一 worker，避免依赖固定 sleep 或调度时序。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGoroutinePool_NonBlocking(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "error/rejects-when-worker-busy",
			description: "验证唯一 worker 被阻塞占用后，非阻塞提交会立即返回错误。",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			pool := newGoroutinePoolForTest(t,
				WithSize(1),
				WithNonBlocking(true),
			)
			started := make(chan struct{}, 1)
			release := make(chan struct{})

			err := pool.Submit(func() {
				started <- struct{}{}
				<-release
			})
			require.NoError(t, err)
			receiveWithin(t, started, "nonblocking worker occupation")

			err = pool.Submit(func() {})
			assert.Error(t, err, "非阻塞模式下，当没有可用协程时应该返回错误")
			close(release)
		})
	}
}

// TestGoroutinePool_MaxBlocking 验证最大阻塞数限制。
//
// 该测试用阻塞通道占用唯一 worker，并在一个提交调用进入等待队列后验证后续提交会被拒绝，
// 避免使用固定 sleep 推测调度状态。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGoroutinePool_MaxBlocking(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "error/rejects-when-max-blocking-reached",
			description: "验证首个任务占用唯一 worker 且第二个提交进入等待队列后，第三个提交超过最大阻塞数会返回错误。",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			pool := newGoroutinePoolForTest(t,
				WithSize(1),
				WithMaxBlocking(1),
			)
			started := make(chan struct{}, 1)
			releaseFirst := make(chan struct{})
			releaseFirstOnce := sync.Once{}
			t.Cleanup(func() {
				releaseFirstOnce.Do(func() {
					close(releaseFirst)
				})
			})

			err := pool.Submit(func() {
				started <- struct{}{}
				<-releaseFirst
			})
			require.NoError(t, err)
			receiveWithin(t, started, "max-blocking worker occupation")

			secondSubmitted := make(chan error, 1)
			secondExecuted := make(chan struct{}, 1)
			go func() {
				secondSubmitted <- pool.Submit(func() {
					secondExecuted <- struct{}{}
				})
			}()

			require.Eventually(t, func() bool {
				return pool.Waiting() == 1
			}, goroutineTestTimeout, time.Millisecond, "第二个提交应该进入等待队列")

			err = pool.Submit(func() {})
			assert.Error(t, err, "超过最大阻塞数时应该返回错误")

			releaseFirstOnce.Do(func() {
				close(releaseFirst)
			})
			assert.NoError(t, receiveWithin(t, secondSubmitted, "blocked Submit completion"))
			receiveWithin(t, secondExecuted, "queued task execution")
		})
	}
}

// TestGoroutinePool_PanicHandler 验证 panic 处理器会接收任务 panic。
//
// 该测试使用 channel 等待 panic handler 回调，避免通过固定 sleep 推测异步回调完成时间。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGoroutinePool_PanicHandler(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "panic/custom-handler-receives-value",
			description: "验证提交的 panic 任务会触发自定义 panic handler 并传递 panic 值。",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			panicCaught := make(chan interface{}, 1)
			pool := newGoroutinePoolForTest(t,
				WithPanicHandler(func(i interface{}) {
					panicCaught <- i
				}),
			)

			err := pool.Submit(func() {
				panic("test panic")
			})
			require.NoError(t, err)

			assert.Equal(t, "test panic", receiveWithin(t, panicCaught, "panic handler invocation"))
		})
	}
}

// TestGoroutinePool_Expiry 验证配置过期时间后协程池仍可继续执行新任务。
//
// 该测试使用 channel 证明首个任务执行完成，并通过 Eventually 在过期窗口后提交新任务，
// 验证协程池在 worker 过期清理周期后仍保持可用。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGoroutinePool_Expiry(t *testing.T) {
	pool := newGoroutinePoolForTest(t,
		WithSize(1),
		WithExpiry(20*time.Millisecond),
	)

	firstExecuted := make(chan struct{}, 1)
	err := pool.Submit(func() {
		firstExecuted <- struct{}{}
	})
	require.NoError(t, err)
	receiveWithin(t, firstExecuted, "first task before worker expiry")

	deadline := time.Now().Add(20 * time.Millisecond)
	secondExecuted := make(chan struct{}, 1)
	assert.Eventually(t, func() bool {
		if time.Now().Before(deadline) {
			return false
		}
		err = pool.Submit(func() {
			secondExecuted <- struct{}{}
		})
		return err == nil
	}, goroutineTestTimeout, time.Millisecond, "过期窗口后协程池应该接受新任务")
	require.NoError(t, err)
	receiveWithin(t, secondExecuted, "task execution after worker expiry")
}

// TestGoroutinePool_PreAlloc 验证预分配配置会使容量和空闲 worker 数与配置一致。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestGoroutinePool_PreAlloc(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveSize    int
	}{
		{
			name:        "success/prealloc-free-workers",
			description: "验证开启预分配后，协程池容量和空闲 worker 数与配置大小一致。",
			giveSize:    5,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			pool, cleanup, err := NewGoroutinePool(
				WithSize(tt.giveSize),
				WithPreAlloc(true),
			)
			require.NoError(t, err)
			defer cleanup()

			assert.Equal(t, tt.giveSize, pool.Cap(), "池容量应该等于配置值")
			assert.Equal(t, tt.giveSize, pool.Free(), "空闲协程数应该等于配置值")
		})
	}
}

// TestGoroutinePool_Tune 验证协程池调容对容量查询的影响。
//
// 该测试通过表驱动用例覆盖扩容、缩容以及零值和负值边界，确保 Tune 后容量保持有效。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGoroutinePool_Tune(t *testing.T) {
	pool, cleanup, err := NewGoroutinePool(WithSize(2))
	require.NoError(t, err)
	defer cleanup()

	tests := []struct {
		name        string
		description string
		size        int
	}{
		{
			name:        "success/increase-size",
			description: "验证 Tune 扩大容量后 Cap 返回新的正向容量。",
			size:        5,
		},
		{
			name:        "success/decrease-size",
			description: "验证 Tune 缩小容量后 Cap 返回新的正向容量。",
			size:        1,
		},
		{
			name:        "boundary/zero-size",
			description: "验证 Tune 接收零值时协程池容量仍保持为有效正数。",
			size:        0,
		},
		{
			name:        "boundary/negative-size",
			description: "验证 Tune 接收负值时协程池容量仍保持为有效正数。",
			size:        -1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			pool.Tune(tt.size)
			if tt.size > 0 {
				assert.Equal(t, tt.size, pool.Cap())
			} else {
				assert.Greater(t, pool.Cap(), 0, "池容量不应该小于等于 0")
			}
		})
	}
}

// TestGoroutinePool_Status 验证协程池状态查询。
//
// 该测试使用 started/release 通道让任务稳定停留在运行状态，再断言 Running、Free、Waiting 和关闭状态，
// 避免提交后立即断言造成调度相关的不稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGoroutinePool_Status(t *testing.T) {
	pool, cleanup, err := NewGoroutinePool(WithSize(2), WithMetrics(false))
	require.NoError(t, err)

	// 测试初始状态
	assert.Equal(t, 2, pool.Cap())
	assert.Equal(t, 0, pool.Running())
	assert.Equal(t, 2, pool.Free())
	assert.Equal(t, 0, pool.Waiting())
	assert.False(t, pool.IsClosed())

	started := make(chan struct{}, 1)
	release := make(chan struct{})
	done := make(chan struct{}, 1)

	// 验证任务被通道阻塞时，协程池状态稳定反映一个运行 worker 和一个空闲 worker。
	err = pool.Submit(func() {
		started <- struct{}{}
		<-release
		done <- struct{}{}
	})
	require.NoError(t, err)
	receiveWithin(t, started, "status task start")

	assert.Equal(t, 2, pool.Cap())
	assert.Equal(t, 1, pool.Running())
	assert.Equal(t, 1, pool.Free())
	assert.Equal(t, 0, pool.Waiting())
	assert.False(t, pool.IsClosed())

	close(release)
	receiveWithin(t, done, "status task completion")

	// 关闭后测试状态
	cleanup()
	assert.True(t, pool.IsClosed())
}

// TestSubmit 验证默认池的任务提交行为。
//
// 该测试隔离默认池全局状态，并通过 channel 证明普通任务确实执行；panic 任务由包级 Submit 恢复，
// 随后继续提交普通任务验证默认池仍保持可用。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestSubmit(t *testing.T) {
	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "success/executes-task",
			description: "验证包级 Submit 会使用隔离后的默认池执行普通任务。",
			assert: func(t *testing.T) {
				t.Helper()

				executed := make(chan struct{}, 1)
				err := Submit(func() {
					executed <- struct{}{}
				})
				require.NoError(t, err)
				receiveWithin(t, executed, "default Submit task execution")
			},
		},
		{
			name:        "panic/recovers-and-remains-usable",
			description: "验证包级 Submit 会恢复任务 panic，并且默认池仍可继续执行后续任务。",
			assert: func(t *testing.T) {
				t.Helper()

				err := Submit(func() {
					panic("test panic")
				})
				require.NoError(t, err)

				executed := make(chan struct{}, 1)
				err = Submit(func() {
					executed <- struct{}{}
				})
				require.NoError(t, err)
				receiveWithin(t, executed, "default Submit task execution after panic")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			isolateDefaultPoolForTest(t)
			metricsDefault = false
			tt.assert(t)
		})
	}
}

// TestGoroutinePool_Concurrent 验证协程池可以稳定处理并发提交。
//
// 该测试通过统一启动信号、错误通道和 waitgroup 超时保护验证并发任务全部提交并执行，
// 避免在 goroutine 内直接断言或通过 sleep 推测任务完成。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGoroutinePool_Concurrent(t *testing.T) {
	pool := newGoroutinePoolForTest(t, WithSize(5))

	const giveTaskCount = 100
	start := make(chan struct{})
	errCh := make(chan error, giveTaskCount)
	var submitWg sync.WaitGroup
	var taskWg sync.WaitGroup
	submitWg.Add(giveTaskCount)
	taskWg.Add(giveTaskCount)

	// 验证所有提交 goroutine 同时开始后，任务均能被协程池执行完成。
	for range giveTaskCount {
		go func() {
			defer submitWg.Done()
			<-start
			errCh <- pool.Submit(func() {
				defer taskWg.Done()
			})
		}()
	}

	close(start)
	requireWaitGroupWithin(t, &submitWg, goroutineTestTimeout, "concurrent task submission")
	close(errCh)
	for err := range errCh {
		assert.NoError(t, err)
	}
	requireWaitGroupWithin(t, &taskWg, goroutineTestTimeout, "concurrent task execution")

	assert.GreaterOrEqual(t, pool.Cap(), 5, "池容量应该大于等于初始大小")
	assert.Equal(t, 0, pool.Waiting(), "所有任务都应该执行完成")
}

// TestGoroutinePool_ConcurrentTune 验证协程池可在并发调容和提交任务时保持可用。
//
// 该测试使用错误通道收集异步提交结果，并用 waitgroup 超时保护并发流程，避免 goroutine 内断言和固定 sleep。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGoroutinePool_ConcurrentTune(t *testing.T) {
	pool := newGoroutinePoolForTest(t, WithSize(5))

	const giveOperationCount = 20
	start := make(chan struct{})
	errCh := make(chan error, giveOperationCount)
	taskDone := make(chan struct{}, giveOperationCount)
	var wg sync.WaitGroup
	wg.Add(2)

	// 验证并发调容与任务提交同时发生时，提交结果由主 goroutine 统一断言且所有任务可完成。
	go func() {
		defer wg.Done()
		<-start
		for i := range giveOperationCount {
			pool.Tune(5 + i%5)
		}
	}()

	go func() {
		defer wg.Done()
		<-start
		for range giveOperationCount {
			errCh <- pool.Submit(func() {
				taskDone <- struct{}{}
			})
		}
	}()

	close(start)
	requireWaitGroupWithin(t, &wg, goroutineTestTimeout, "concurrent tune and submit operations")
	close(errCh)
	for err := range errCh {
		assert.NoError(t, err)
	}
	for range giveOperationCount {
		receiveWithin(t, taskDone, "task execution during concurrent tuning")
	}
	assert.GreaterOrEqual(t, pool.Cap(), 5, "池容量应该大于等于初始大小")
}

// TestGoroutinePool_Cleanup 验证清理函数会关闭协程池并拒绝后续提交。
//
// 该测试通过通道控制任务阻塞与释放，确保 cleanup 在任务运行期间调用，并明确断言任务完成、
// 池关闭以及关闭后的提交失败。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGoroutinePool_Cleanup(t *testing.T) {
	pool, cleanup, err := NewGoroutinePool(WithSize(2), WithMetrics(false))
	require.NoError(t, err)

	const giveTaskCount = 2
	started := make(chan struct{}, giveTaskCount)
	release := make(chan struct{})
	done := make(chan struct{}, giveTaskCount)

	// 验证 cleanup 调用前存在受控运行任务，释放后清理函数能完成关闭。
	for range giveTaskCount {
		err := pool.Submit(func() {
			started <- struct{}{}
			<-release
			done <- struct{}{}
		})
		require.NoError(t, err)
	}
	for range giveTaskCount {
		receiveWithin(t, started, "cleanup task start")
	}

	cleanupDone := make(chan struct{}, 1)
	go func() {
		cleanup()
		cleanupDone <- struct{}{}
	}()

	close(release)
	for range giveTaskCount {
		receiveWithin(t, done, "cleanup task completion")
	}
	receiveWithin(t, cleanupDone, "pool cleanup completion")

	assert.True(t, pool.IsClosed())
	err = pool.Submit(func() {})
	assert.Error(t, err, "向已清理的池提交任务应该返回错误")
}
