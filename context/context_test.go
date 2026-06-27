// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package context

import (
	stdctx "context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 编译期确保 withoutCancelCtx 满足标准库 context.Context 接口。
var _ stdctx.Context = (*withoutCancelCtx)(nil)

// TestWithoutCancel_NilParentPanics 验证 WithoutCancel 在 nil 父 context 下的 panic 契约。
//
// 该测试覆盖 nil parent 输入的错误分支，确保 panic 文本与现有实现保持兼容。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestWithoutCancel_NilParentPanics(t *testing.T) {
	// 验证 nil parent 会触发明确的 panic 文本，便于调用方定位错误用法。
	require.PanicsWithValue(t, "context: WithoutCancel with nil parent", func() {
		_ = WithoutCancel(nil) //nolint:staticcheck // 需要传入 nil 以验证 WithoutCancel(nil) panic 契约。
	})
}

// TestWithoutCancel_DeadlineIgnored 验证 WithoutCancel 返回的 context 不继承父 context 的截止时间。
//
// 该测试覆盖父 context 存在 deadline 的场景，确保返回 context 始终报告无截止时间。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestWithoutCancel_DeadlineIgnored(t *testing.T) {
	// 验证父 context 具有有效 deadline，确保用例前置条件成立。
	parent, cancel := stdctx.WithDeadline(stdctx.Background(), time.Now().Add(time.Hour))
	t.Cleanup(cancel)
	parentDeadline, parentOK := parent.Deadline()
	require.True(t, parentOK)
	require.False(t, parentDeadline.IsZero())

	// 验证 WithoutCancel 返回的 context 忽略父 context 的 deadline。
	got := WithoutCancel(parent)
	gotDeadline, gotOK := got.Deadline()
	assert.False(t, gotOK)
	assert.True(t, gotDeadline.IsZero())
}

// TestWithoutCancel_DoneIgnored 验证 WithoutCancel 返回的 context 不继承父 context 的 Done 信号。
//
// 该测试通过表驱动用例覆盖父 context 可取消和已取消场景，确保返回 context 的 Done 始终为 nil。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestWithoutCancel_DoneIgnored(t *testing.T) {
	tests := []struct {
		name        string
		description string
		setup       func(t *testing.T) stdctx.Context
	}{
		{
			name:        "success/cancelable-parent",
			description: "验证父 context 可取消但尚未取消时，返回 context 的 Done 仍为 nil。",
			setup: func(t *testing.T) stdctx.Context {
				t.Helper()

				parent, cancel := stdctx.WithCancel(stdctx.Background())
				t.Cleanup(cancel)
				require.NotNil(t, parent.Done())
				return parent
			},
		},
		{
			name:        "success/canceled-parent",
			description: "验证父 context 已取消时，返回 context 的 Done 仍为 nil。",
			setup: func(t *testing.T) stdctx.Context {
				t.Helper()

				parent, cancel := stdctx.WithCancel(stdctx.Background())
				cancel()
				t.Cleanup(cancel)

				select {
				case <-parent.Done():
				default:
					require.FailNow(t, "父 context 应已完成取消")
				}

				return parent
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			parent := tt.setup(t)
			got := WithoutCancel(parent)

			assert.Nil(t, got.Done())
		})
	}
}

// TestWithoutCancel_ErrIgnored 验证 WithoutCancel 返回的 context 不继承父 context 的错误状态。
//
// 该测试通过表驱动用例覆盖父 context 已取消和已超时场景，确保返回 context 的 Err 始终为 nil。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestWithoutCancel_ErrIgnored(t *testing.T) {
	tests := []struct {
		name          string
		description   string
		setup         func(t *testing.T) stdctx.Context
		wantParentErr error
	}{
		{
			name:          "success/canceled-parent",
			description:   "验证父 context 已取消时，返回 context 的 Err 仍为 nil。",
			wantParentErr: stdctx.Canceled,
			setup: func(t *testing.T) stdctx.Context {
				t.Helper()

				parent, cancel := stdctx.WithCancel(stdctx.Background())
				cancel()
				t.Cleanup(cancel)
				return parent
			},
		},
		{
			name:          "success/deadline-exceeded-parent",
			description:   "验证父 context 已超时时，返回 context 的 Err 仍为 nil。",
			wantParentErr: stdctx.DeadlineExceeded,
			setup: func(t *testing.T) stdctx.Context {
				t.Helper()

				parent, cancel := stdctx.WithDeadline(stdctx.Background(), time.Now().Add(-time.Nanosecond))
				t.Cleanup(cancel)
				return parent
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			parent := tt.setup(t)
			got := WithoutCancel(parent)

			require.ErrorIs(t, parent.Err(), tt.wantParentErr)
			assert.NoError(t, got.Err())
		})
	}
}

// TestWithoutCancel_ValueDelegatesToParent 验证 WithoutCancel 返回的 context 从父 context 读取值。
//
// 该测试通过表驱动用例覆盖直接父节点值、父子链值和缺失 key，确保值读取契约稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestWithoutCancel_ValueDelegatesToParent(t *testing.T) {
	type contextValueKey string

	rootKey := contextValueKey("root")
	childKey := contextValueKey("child")
	missingKey := contextValueKey("missing")

	root := stdctx.WithValue(stdctx.Background(), rootKey, "root-value")
	parent := stdctx.WithValue(root, childKey, "child-value")
	got := WithoutCancel(parent)

	tests := []struct {
		name        string
		description string
		giveKey     any
		wantValue   any
	}{
		{
			name:        "success/direct-parent-value",
			description: "验证返回 context 能读取直接父 context 上保存的值。",
			giveKey:     childKey,
			wantValue:   "child-value",
		},
		{
			name:        "success/ancestor-parent-value",
			description: "验证返回 context 能读取父子链上祖先 context 保存的值。",
			giveKey:     rootKey,
			wantValue:   "root-value",
		},
		{
			name:        "success/missing-value",
			description: "验证父 context 链上不存在指定 key 时返回 nil。",
			giveKey:     missingKey,
			wantValue:   nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			assert.Equal(t, tt.wantValue, got.Value(tt.giveKey))
		})
	}
}
