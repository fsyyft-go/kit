// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package goroutine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetGoIDSlow_Behavior 验证慢速 goroutine ID 获取逻辑的核心行为。
//
// 该测试覆盖当前 goroutine 的非零返回值，以及父子 goroutine ID 隔离语义，确保基于
// runtime.Stack 的慢速实现能够稳定区分不同 goroutine。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGetGoIDSlow_Behavior(t *testing.T) {
	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "success/current-goroutine",
			description: "验证 getGoIDSlow 和 GetGoIDSlow 在当前 goroutine 中返回一致的非零 ID。",
			assert: func(t *testing.T) {
				t.Helper()

				gotInternal := getGoIDSlow()
				gotExported := GetGoIDSlow()

				require.NotZero(t, gotInternal)
				assert.Equal(t, gotInternal, gotExported)
			},
		},
		{
			name:        "concurrency/child-goroutine",
			description: "验证慢速实现可以稳定区分父 goroutine 与子 goroutine。",
			assert: func(t *testing.T) {
				t.Helper()

				parentID := getGoIDSlow()
				require.NotZero(t, parentID)
				childID := make(chan int64, 1)

				go func() {
					childID <- getGoIDSlow()
				}()

				gotChildID := receiveWithin(t, childID, "child slow goroutine ID collection")
				require.NotZero(t, gotChildID)
				assert.NotEqual(t, parentID, gotChildID)
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

// TestExtractGID_ParseStackHeader 验证 extractGID 对 runtime.Stack 头部格式的解析契约。
//
// 该测试通过表驱动用例覆盖正常运行态、包含前导零的 ID，以及非法 ID 回退为零的场景，
// 确保慢速 goroutine ID 解析逻辑对关键输入格式保持稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestExtractGID_ParseStackHeader(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveStack   []byte
		wantGID     int64
	}{
		{
			name:        "success/running-state",
			description: "验证标准 runtime.Stack 头部中的十进制 goroutine ID 可以被正确解析。",
			giveStack:   []byte("goroutine 123 [running]:\nexample.stack()"),
			wantGID:     123,
		},
		{
			name:        "success/leading-zero",
			description: "验证包含前导零的 goroutine ID 会按十进制数值解析。",
			giveStack:   []byte("goroutine 00042 [chan receive]:\nexample.stack()"),
			wantGID:     42,
		},
		{
			name:        "error/non-numeric-id",
			description: "验证 ID 字段不是数字时，解析失败会按照当前实现返回零值。",
			giveStack:   []byte("goroutine abc [running]:\nexample.stack()"),
			wantGID:     0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			gotGID := extractGID(tt.giveStack)

			assert.Equal(t, tt.wantGID, gotGID)
		})
	}
}
