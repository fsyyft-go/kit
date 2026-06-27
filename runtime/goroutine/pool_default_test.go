// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package goroutine

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSubmit_DefaultPoolCreationError 验证默认协程池创建失败时 Submit 返回错误且不触发 panic。
//
// 该测试通过隔离并临时篡改默认池配置，覆盖默认池初始化错误分支，确保调用方能够收到
// ants.NewPool 返回的配置错误，同时默认池不会被部分初始化。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestSubmit_DefaultPoolCreationError(t *testing.T) {
	isolateDefaultPoolForTest(t)

	// 验证默认池配置无效时，Submit 应该将创建错误返回给调用方，而不是因 nil cleanup panic。
	sizeDefault = 0
	preAllocDefault = true
	metricsDefault = false

	err := Submit(func() {})

	require.Error(t, err)
	assert.ErrorContains(t, err, "PreAlloc")
	assert.Nil(t, poolDefault)
}

// TestSubmit_DefaultPoolBehavior 验证包级 Submit 的默认池初始化、任务执行和 panic 隔离行为。
//
// 该测试通过表驱动用例覆盖普通任务和 panic 任务，确保默认池懒加载成功、普通任务能够异步执行，
// 且任务 panic 会被 Submit 包装层恢复而不影响后续任务提交。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestSubmit_DefaultPoolBehavior(t *testing.T) {
	tests := []struct {
		name        string
		description string
		setup       func(t *testing.T) chan struct{}
		assert      func(t *testing.T, executed chan struct{})
	}{
		{
			name:        "success/executes-task",
			description: "验证 Submit 会懒加载默认池并执行提交的普通任务。",
			setup: func(t *testing.T) chan struct{} {
				t.Helper()
				executed := make(chan struct{}, 1)
				err := Submit(func() {
					executed <- struct{}{}
				})
				require.NoError(t, err)
				return executed
			},
			assert: func(t *testing.T, executed chan struct{}) {
				t.Helper()

				receiveWithin(t, executed, "default pool task execution")
				assert.NotNil(t, poolDefault)
				assert.False(t, poolDefault.IsClosed())
			},
		},
		{
			name:        "panic/recovers-and-keeps-pool-usable",
			description: "验证 Submit 包装层会恢复任务 panic，并且默认池仍可执行后续任务。",
			setup: func(t *testing.T) chan struct{} {
				t.Helper()
				executed := make(chan struct{}, 1)
				err := Submit(func() {
					panic("default pool task panic")
				})
				require.NoError(t, err)
				return executed
			},
			assert: func(t *testing.T, executed chan struct{}) {
				t.Helper()

				err := Submit(func() {
					executed <- struct{}{}
				})
				require.NoError(t, err)
				receiveWithin(t, executed, "default pool task execution after recovered panic")
				assert.NotNil(t, poolDefault)
				assert.False(t, poolDefault.IsClosed())
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			isolateDefaultPoolForTest(t)
			metricsDefault = false

			executed := tt.setup(t)
			tt.assert(t, executed)
		})
	}
}

// TestSubmit_DefaultPoolConcurrentInitialization 验证并发调用 Submit 时默认池只被安全初始化一次。
//
// 该测试通过多个 goroutine 同时提交任务，覆盖并发 Submit 下默认池安全懒加载路径，确保所有任务都会执行，
// 且并发初始化不会导致多个可见默认池或任务丢失。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestSubmit_DefaultPoolConcurrentInitialization(t *testing.T) {
	isolateDefaultPoolForTest(t)
	metricsDefault = false

	const giveTaskCount = 16
	var executed atomic.Int32
	errs := make(chan error, giveTaskCount)
	done := make(chan struct{}, giveTaskCount)

	// 验证默认池在并发首次提交下能安全完成懒加载，并执行所有提交的任务。
	for range giveTaskCount {
		go func() {
			errs <- Submit(func() {
				executed.Add(1)
				done <- struct{}{}
			})
		}()
	}

	for range giveTaskCount {
		assert.NoError(t, receiveWithin(t, errs, "concurrent Submit call"))
		receiveWithin(t, done, "concurrent default pool task execution")
	}

	assert.Equal(t, int32(giveTaskCount), executed.Load())
	require.NotNil(t, poolDefault)
	assert.False(t, poolDefault.IsClosed())
}
