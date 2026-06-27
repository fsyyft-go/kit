// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package goroutine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetGoID_Consistency 验证快速 goroutine ID 获取逻辑的核心行为。
//
// 该测试通过表驱动用例覆盖当前 goroutine 与子 goroutine 场景，确保 GetGoID 返回非零 ID，
// 且与 GetGoIDSlow 在同一 goroutine 内保持一致。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGetGoID_Consistency(t *testing.T) {
	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "success/current-goroutine",
			description: "验证当前 goroutine 中 GetGoID 返回非零值，并与 GetGoIDSlow 保持一致。",
			assert: func(t *testing.T) {
				t.Helper()

				gotFast := GetGoID()
				gotSlow := GetGoIDSlow()

				require.NotZero(t, gotFast)
				assert.Equal(t, gotFast, gotSlow)
			},
		},
		{
			name:        "concurrency/child-goroutine",
			description: "验证父子 goroutine 的 ID 不同，且子 goroutine 内快速与慢速获取结果一致。",
			assert: func(t *testing.T) {
				t.Helper()

				parentID := GetGoID()
				require.NotZero(t, parentID)
				childIDs := make(chan struct {
					fast int64
					slow int64
				}, 1)

				go func() {
					childIDs <- struct {
						fast int64
						slow int64
					}{
						fast: GetGoID(),
						slow: GetGoIDSlow(),
					}
				}()

				gotChild := receiveWithin(t, childIDs, "child goroutine ID collection")
				require.NotZero(t, gotChild.fast)
				assert.Equal(t, gotChild.fast, gotChild.slow)
				assert.NotEqual(t, parentID, gotChild.fast)
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

// BenchmarkGetGoID 衡量快速 goroutine ID 获取方法的调用开销。
//
// 参数：
//   - b: 基准测试上下文，用于控制迭代次数并报告性能数据。
func BenchmarkGetGoID(b *testing.B) {
	for b.Loop() {
		GetGoID()
	}
}

// BenchmarkGetGoIDSlow 衡量慢速 goroutine ID 获取方法的调用开销。
//
// 参数：
//   - b: 基准测试上下文，用于控制迭代次数并报告性能数据。
func BenchmarkGetGoIDSlow(b *testing.B) {
	for b.Loop() {
		GetGoIDSlow()
	}
}
