// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package goroutine 提供了协程池的测试实现。
// 本测试文件主要测试 GoroutinePool 接口及其实现 goroutinePool 的功能。
// 测试用例采用表格驱动的方式组织，使用 testify 包进行断言。
// 测试覆盖了协程池的主要功能点，包括创建、任务提交、容量调整和状态查询等。
// 每个测试用例都包含详细的注释说明，便于理解测试目的和预期结果。

package goroutine

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewGoroutinePool 测试创建新的协程池。
func TestNewGoroutinePool(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "使用默认配置创建协程池",
			opts:    nil,
			wantErr: false,
		},
		{
			name: "使用自定义配置创建协程池",
			opts: []Option{
				WithSize(10),
				WithExpiry(time.Second),
				WithPreAlloc(true),
				WithNonBlocking(true),
				WithMaxBlocking(100),
				WithName("test-pool"),
				WithMetrics(true),
			},
			wantErr: false,
		},
		{
			name: "使用最小配置创建协程池",
			opts: []Option{
				WithSize(1),
				WithExpiry(time.Millisecond),
				WithPreAlloc(false),
				WithNonBlocking(false),
				WithMaxBlocking(0),
				WithMetrics(false),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, cleanup, err := NewGoroutinePool(tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, pool)
			assert.NotNil(t, cleanup)
			cleanup()
		})
	}
}

// TestGoroutinePool_Submit 测试提交任务到协程池。
func TestGoroutinePool_Submit(t *testing.T) {
	pool, cleanup, err := NewGoroutinePool(WithSize(2))
	require.NoError(t, err)
	defer cleanup()

	tests := []struct {
		name    string
		task    func()
		wantErr bool
	}{
		{
			name: "提交正常任务",
			task: func() {
				time.Sleep(10 * time.Millisecond)
			},
			wantErr: false,
		},
		{
			name: "提交 panic 任务",
			task: func() {
				panic("test panic")
			},
			wantErr: false,
		},
		{
			name:    "提交空任务",
			task:    func() {},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pool.Submit(tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGoroutinePool_SubmitAfterClose 测试关闭后提交任务。
func TestGoroutinePool_SubmitAfterClose(t *testing.T) {
	pool, cleanup, err := NewGoroutinePool()
	require.NoError(t, err)
	cleanup()

	err = pool.Submit(func() {})
	assert.Error(t, err, "向已关闭的池提交任务应该返回错误")
}

// TestGoroutinePool_NonBlocking 测试非阻塞模式。
func TestGoroutinePool_NonBlocking(t *testing.T) {
	pool, cleanup, err := NewGoroutinePool(
		WithSize(1),
		WithNonBlocking(true),
	)
	require.NoError(t, err)
	defer cleanup()

	// 提交一个长时间运行的任务
	err = pool.Submit(func() {
		time.Sleep(100 * time.Millisecond)
	})
	require.NoError(t, err)

	// 立即提交另一个任务，应该被拒绝
	err = pool.Submit(func() {})
	assert.Error(t, err, "非阻塞模式下，当没有可用协程时应该返回错误")
}

// TestGoroutinePool_MaxBlocking 测试最大阻塞数限制。
func TestGoroutinePool_MaxBlocking(t *testing.T) {
	pool, cleanup, err := NewGoroutinePool(
		WithSize(1),
		WithMaxBlocking(1),
	)
	require.NoError(t, err)
	defer cleanup()

	// 提交一个长时间运行的任务
	err = pool.Submit(func() {
		time.Sleep(100 * time.Millisecond)
	})
	require.NoError(t, err)

	// 提交第二个任务，应该被阻塞
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := pool.Submit(func() {})
		assert.NoError(t, err)
	}()

	// 提交第三个任务，应该返回错误
	time.Sleep(10 * time.Millisecond) // 等待第二个任务被阻塞
	err = pool.Submit(func() {})
	assert.Error(t, err, "超过最大阻塞数时应该返回错误")

	wg.Wait()
}

// TestGoroutinePool_PanicHandler 测试 panic 处理器。
func TestGoroutinePool_PanicHandler(t *testing.T) {
	var panicCount int32
	pool, cleanup, err := NewGoroutinePool(
		WithPanicHandler(func(i interface{}) {
			atomic.AddInt32(&panicCount, 1)
		}),
	)
	require.NoError(t, err)
	defer cleanup()

	// 提交会 panic 的任务
	err = pool.Submit(func() {
		panic("test panic")
	})
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond) // 等待 panic 处理器执行
	assert.Equal(t, int32(1), atomic.LoadInt32(&panicCount), "panic 处理器应该被调用一次")
}

// TestGoroutinePool_Expiry 测试协程过期。
func TestGoroutinePool_Expiry(t *testing.T) {
	pool, cleanup, err := NewGoroutinePool(
		WithSize(1),
		WithExpiry(50*time.Millisecond),
	)
	require.NoError(t, err)
	defer cleanup()

	// 提交一个任务
	err = pool.Submit(func() {})
	require.NoError(t, err)

	assert.Equal(t, 1, pool.Running()+pool.Free(), "应该有一个协程")

	// 等待协程过期
	time.Sleep(100 * time.Millisecond)

	// 提交新任务，确保可以正常工作
	err = pool.Submit(func() {})
	assert.NoError(t, err)
}

// TestGoroutinePool_PreAlloc 测试预分配。
func TestGoroutinePool_PreAlloc(t *testing.T) {
	pool, cleanup, err := NewGoroutinePool(
		WithSize(5),
		WithPreAlloc(true),
	)
	require.NoError(t, err)
	defer cleanup()

	assert.Equal(t, 5, pool.Cap(), "池容量应该为 5")
	assert.Equal(t, 5, pool.Free(), "应该有 5 个空闲协程")
}

// TestGoroutinePool_Tune 测试调整协程池大小。
func TestGoroutinePool_Tune(t *testing.T) {
	pool, cleanup, err := NewGoroutinePool(WithSize(2))
	require.NoError(t, err)
	defer cleanup()

	tests := []struct {
		name string
		size int
	}{
		{
			name: "增加池大小",
			size: 5,
		},
		{
			name: "减少池大小",
			size: 1,
		},
		{
			name: "设置为零",
			size: 0,
		},
		{
			name: "设置为负数",
			size: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool.Tune(tt.size)
			if tt.size > 0 {
				assert.Equal(t, tt.size, pool.Cap())
			} else {
				assert.Greater(t, pool.Cap(), 0, "池容量不应该小于等于 0")
			}
		})
	}
}

// TestGoroutinePool_Status 测试协程池状态查询。
func TestGoroutinePool_Status(t *testing.T) {
	pool, cleanup, err := NewGoroutinePool(WithSize(2))
	require.NoError(t, err)
	defer cleanup()

	// 测试初始状态
	assert.Equal(t, 2, pool.Cap())
	assert.Equal(t, 0, pool.Running())
	assert.Equal(t, 2, pool.Free())
	assert.Equal(t, 0, pool.Waiting())
	assert.False(t, pool.IsClosed())

	// 提交任务后测试状态
	var wg sync.WaitGroup
	wg.Add(1)
	err = pool.Submit(func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)
	})
	require.NoError(t, err)

	assert.Equal(t, 2, pool.Cap())
	assert.Equal(t, 1, pool.Running())
	assert.Equal(t, 1, pool.Free())
	assert.Equal(t, 0, pool.Waiting())
	assert.False(t, pool.IsClosed())

	wg.Wait()

	// 关闭后测试状态
	cleanup()
	assert.True(t, pool.IsClosed())
}

// TestSubmit 测试默认池的任务提交。
func TestSubmit(t *testing.T) {
	tests := []struct {
		name    string
		task    func()
		wantErr bool
	}{
		{
			name: "提交正常任务到默认池",
			task: func() {
				time.Sleep(10 * time.Millisecond)
			},
			wantErr: false,
		},
		{
			name: "提交 panic 任务到默认池",
			task: func() {
				panic("test panic")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Submit(tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGoroutinePool_Concurrent 测试协程池的并发操作。
func TestGoroutinePool_Concurrent(t *testing.T) {
	pool, cleanup, err := NewGoroutinePool(WithSize(5))
	require.NoError(t, err)
	defer cleanup()

	var submitWg sync.WaitGroup
	var taskWg sync.WaitGroup
	count := 100 // 增加并发数
	submitWg.Add(count)
	taskWg.Add(count)

	// 并发提交任务
	for i := 0; i < count; i++ {
		go func() {
			defer submitWg.Done()
			err := pool.Submit(func() {
				defer taskWg.Done()
				time.Sleep(time.Duration(1+i%10) * time.Millisecond) // 随机延迟
			})
			assert.NoError(t, err)
		}()
	}

	// 等待所有任务提交完成
	submitWg.Wait()
	// 等待所有任务执行完成
	taskWg.Wait()

	// 只检查确定的状态
	assert.GreaterOrEqual(t, pool.Cap(), 5, "池容量应该大于等于初始大小")
	assert.Equal(t, 0, pool.Waiting(), "所有任务都应该执行完成")
}

// TestGoroutinePool_ConcurrentTune 测试并发调整池大小。
func TestGoroutinePool_ConcurrentTune(t *testing.T) {
	pool, cleanup, err := NewGoroutinePool(WithSize(5))
	require.NoError(t, err)
	defer cleanup()

	var wg sync.WaitGroup
	wg.Add(2)

	// 并发调整池大小
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			pool.Tune(5 + i%5)
			time.Sleep(time.Millisecond)
		}
	}()

	// 并发提交任务
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			err := pool.Submit(func() {
				time.Sleep(time.Millisecond)
			})
			assert.NoError(t, err)
		}
	}()

	wg.Wait()
	assert.GreaterOrEqual(t, pool.Cap(), 5, "池容量应该大于等于初始大小")
}

// TestGoroutinePool_Cleanup 测试清理函数。
func TestGoroutinePool_Cleanup(t *testing.T) {
	pool, cleanup, err := NewGoroutinePool()
	require.NoError(t, err)

	// 提交一些任务
	for i := 0; i < 5; i++ {
		err := pool.Submit(func() {
			time.Sleep(50 * time.Millisecond)
		})
		require.NoError(t, err)
	}

	// 立即调用清理函数
	cleanup()

	// 验证池已关闭
	assert.True(t, pool.IsClosed())
	err = pool.Submit(func() {})
	assert.Error(t, err, "向已清理的池提交任务应该返回错误")
}
