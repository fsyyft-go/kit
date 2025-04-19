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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool.Tune(tt.size)
			assert.Equal(t, tt.size, pool.Cap())
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
	count := 10
	submitWg.Add(count)
	taskWg.Add(count)

	// 并发提交任务
	for i := 0; i < count; i++ {
		go func() {
			defer submitWg.Done()
			err := pool.Submit(func() {
				defer taskWg.Done()
				time.Sleep(10 * time.Millisecond)
			})
			assert.NoError(t, err)
		}()
	}

	// 等待所有任务提交完成
	submitWg.Wait()
	// 等待所有任务执行完成
	taskWg.Wait()

	// 只检查确定的状态
	assert.Equal(t, 5, pool.Cap(), "池容量应该保持不变")
	assert.Equal(t, 0, pool.Waiting(), "所有任务都应该执行完成")
}
