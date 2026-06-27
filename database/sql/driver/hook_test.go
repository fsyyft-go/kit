// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package driver

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpType_String 验证数据库操作类型到稳定字符串的映射关系。
//
// 该测试通过表驱动用例覆盖全部已定义操作类型和未知类型，确保日志、诊断和 Hook 上下文中的操作名称保持稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestOpType_String(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveOpType  OpType
		want        string
	}{
		{name: "success/connect", description: "验证连接操作类型返回稳定的 Connect 名称。", giveOpType: OpConnect, want: "Connect"},
		{name: "success/begin", description: "验证开启事务操作类型返回稳定的 Begin 名称。", giveOpType: OpBegin, want: "Begin"},
		{name: "success/commit", description: "验证提交事务操作类型返回稳定的 Commit 名称。", giveOpType: OpCommit, want: "Commit"},
		{name: "success/rollback", description: "验证回滚事务操作类型返回稳定的 Rollback 名称。", giveOpType: OpRollback, want: "Rollback"},
		{name: "success/prepare", description: "验证预处理语句操作类型返回稳定的 Prepare 名称。", giveOpType: OpPrepare, want: "Prepare"},
		{name: "success/stmt-exec", description: "验证预处理执行操作类型返回稳定的 StmtExec 名称。", giveOpType: OpStmtExec, want: "StmtExec"},
		{name: "success/stmt-query", description: "验证预处理查询操作类型返回稳定的 StmtQuery 名称。", giveOpType: OpStmtQuery, want: "StmtQuery"},
		{name: "success/stmt-close", description: "验证预处理关闭操作类型返回稳定的 StmtClose 名称。", giveOpType: OpStmtClose, want: "StmtClose"},
		{name: "success/exec", description: "验证直接执行操作类型返回稳定的 Exec 名称。", giveOpType: OpExec, want: "Exec"},
		{name: "success/query", description: "验证直接查询操作类型返回稳定的 Query 名称。", giveOpType: OpQuery, want: "Query"},
		{name: "success/ping", description: "验证 Ping 操作类型返回稳定的 Ping 名称。", giveOpType: OpPing, want: "Ping"},
		{name: "boundary/unknown", description: "验证未知操作类型返回 Unknown，避免诊断信息出现空字符串。", giveOpType: OpType(999), want: "Unknown"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := tt.giveOpType.String()

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestHookContext_AccessorsAndContextDelegation 验证 HookContext 保存操作信息并委托 context.Context 行为。
//
// 该测试覆盖查询、参数、结果、错误、耗时、hook 私有值和原始上下文委托，确保 Hook 实现可依赖稳定的上下文契约。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestHookContext_AccessorsAndContextDelegation(t *testing.T) {
	type contextKey string

	deadlineTime := time.Now().Add(time.Hour)
	originCtx, cancel := context.WithDeadline(context.WithValue(context.Background(), contextKey("trace"), "trace-001"), deadlineTime)
	t.Cleanup(cancel)
	args := []driver.NamedValue{
		{Name: "name", Ordinal: 1, Value: "alice"},
		{Ordinal: 2, Value: int64(42)},
	}
	ctx := NewHookContext(originCtx, OpQuery, "SELECT * FROM users WHERE name=? AND age=?", args)

	assert.False(t, ctx.StartTime().IsZero(), "开始时间应在构造 HookContext 时记录")
	assert.True(t, ctx.EndTime().IsZero(), "未设置结果前结束时间应为零值")
	assert.LessOrEqual(t, ctx.Duration(), time.Duration(0), "未设置结果前 Duration 不应表现为正耗时")
	assert.Equal(t, OpQuery, ctx.OpType())
	assert.Equal(t, "SELECT * FROM users WHERE name=? AND age=?", ctx.Query())
	assert.Equal(t, args, ctx.Args())
	assert.Nil(t, ctx.OriginResult())
	assert.NoError(t, ctx.OriginError())

	_, ok := ctx.GetHookValue("missing")
	assert.False(t, ok)
	ctx.SetHookValue("request-id", "req-001")
	gotValue, ok := ctx.GetHookValue("request-id")
	assert.True(t, ok)
	assert.Equal(t, "req-001", gotValue)

	result := driver.RowsAffected(3)
	originErr := errors.New("query failed")
	ctx.SetResult(result, originErr)

	assert.Equal(t, result, ctx.OriginResult())
	assert.ErrorIs(t, ctx.OriginError(), originErr)
	assert.False(t, ctx.EndTime().IsZero(), "设置结果后应记录结束时间")
	assert.GreaterOrEqual(t, ctx.Duration(), time.Duration(0))

	deadline, ok := ctx.Deadline()
	require.True(t, ok)
	assert.Equal(t, deadlineTime, deadline)
	assert.Equal(t, originCtx.Done(), ctx.Done())
	assert.NoError(t, ctx.Err())
	assert.Equal(t, "trace-001", ctx.Value(contextKey("trace")))
}

// TestHookContext_ContextCancellation 验证 HookContext 透传原始上下文取消状态。
//
// 该测试确保原始上下文取消后 HookContext 的 Done 与 Err 可观测到相同状态，便于 Hook 实现响应取消信号。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestHookContext_ContextCancellation(t *testing.T) {
	originCtx, cancel := context.WithCancel(context.Background())
	ctx := NewHookContext(originCtx, OpPing, "", nil)

	cancel()

	select {
	case <-ctx.Done():
		assert.ErrorIs(t, ctx.Err(), context.Canceled)
	case <-time.After(time.Second):
		require.Fail(t, "HookContext 未在原始上下文取消后关闭 Done channel")
	}
}

// TestHookManager_OrderAndErrorPropagation 验证 HookManager 的执行顺序和错误短路语义。
//
// 该测试通过表驱动用例覆盖 Before 正序、After 逆序、空管理器和错误短路，确保多个 Hook 组合时行为可预测。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestHookManager_OrderAndErrorPropagation(t *testing.T) {
	beforeErr := errors.New("before hook failed")
	afterErr := errors.New("after hook failed")

	tests := []struct {
		name        string
		description string
		setup       func(t *testing.T, calls *[]string) (*HookManager, *HookContext)
		act         func(manager *HookManager, ctx *HookContext) error
		wantCalls   []string
		wantErrIs   error
	}{
		{
			name:        "success/before-in-add-order",
			description: "验证 Before 按 Hook 添加顺序依次执行。",
			setup: func(t *testing.T, calls *[]string) (*HookManager, *HookContext) {
				manager := NewHookManager()
				manager.AddHook(&recordingHook{name: "first", calls: calls})
				manager.AddHook(&recordingHook{name: "second", calls: calls})
				return manager, NewHookContext(context.Background(), OpExec, "UPDATE users SET name=?", nil)
			},
			act:       func(manager *HookManager, ctx *HookContext) error { return manager.Before(ctx) },
			wantCalls: []string{"before:first:Exec", "before:second:Exec"},
		},
		{
			name:        "success/after-in-reverse-order",
			description: "验证 After 按 Hook 添加顺序的反向依次执行。",
			setup: func(t *testing.T, calls *[]string) (*HookManager, *HookContext) {
				manager := NewHookManager()
				manager.AddHook(&recordingHook{name: "first", calls: calls})
				manager.AddHook(&recordingHook{name: "second", calls: calls})
				return manager, NewHookContext(context.Background(), OpQuery, "SELECT 1", nil)
			},
			act:       func(manager *HookManager, ctx *HookContext) error { return manager.After(ctx) },
			wantCalls: []string{"after:second:Query", "after:first:Query"},
		},
		{
			name:        "success/empty-manager",
			description: "验证空 HookManager 执行 Before 和 After 均不返回错误。",
			setup: func(t *testing.T, calls *[]string) (*HookManager, *HookContext) {
				return NewHookManager(), NewHookContext(context.Background(), OpPing, "", nil)
			},
			act: func(manager *HookManager, ctx *HookContext) error {
				require.NoError(t, manager.Before(ctx))
				return manager.After(ctx)
			},
			wantCalls: []string{},
		},
		{
			name:        "error/before-stops-at-first-error",
			description: "验证 Before 遇到错误后立即返回且不会继续执行后续 Hook。",
			setup: func(t *testing.T, calls *[]string) (*HookManager, *HookContext) {
				manager := NewHookManager()
				manager.AddHook(&recordingHook{name: "first", calls: calls, beforeErr: beforeErr})
				manager.AddHook(&recordingHook{name: "second", calls: calls})
				return manager, NewHookContext(context.Background(), OpBegin, "", nil)
			},
			act:       func(manager *HookManager, ctx *HookContext) error { return manager.Before(ctx) },
			wantCalls: []string{"before:first:Begin"},
			wantErrIs: beforeErr,
		},
		{
			name:        "error/after-stops-at-first-reverse-error",
			description: "验证 After 逆序执行遇到错误后立即返回且不会继续执行剩余 Hook。",
			setup: func(t *testing.T, calls *[]string) (*HookManager, *HookContext) {
				manager := NewHookManager()
				manager.AddHook(&recordingHook{name: "first", calls: calls})
				manager.AddHook(&recordingHook{name: "second", calls: calls, afterErr: afterErr})
				return manager, NewHookContext(context.Background(), OpCommit, "", nil)
			},
			act:       func(manager *HookManager, ctx *HookContext) error { return manager.After(ctx) },
			wantCalls: []string{"after:second:Commit"},
			wantErrIs: afterErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			calls := make([]string, 0)
			manager, ctx := tt.setup(t, &calls)

			err := tt.act(manager, ctx)

			if tt.wantErrIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErrIs)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantCalls, calls)
		})
	}
}

// recordingHook 是用于记录 Hook 调用顺序和注入错误的测试辅助 Hook。
//
// 该辅助类型集中表达 HookManager 和驱动包装用例需要的调用观测能力，避免依赖外部 mock 框架。
type recordingHook struct {
	name      string
	calls     *[]string
	beforeErr error
	afterErr  error
	beforeFn  func(ctx *HookContext)
	afterFn   func(ctx *HookContext)
}

// Before 记录前置 Hook 调用并按需返回预设错误。
//
// 参数：
//   - ctx: Hook 上下文，用于读取当前操作类型并传递给自定义断言函数。
//
// 返回：
//   - error: 预设的前置 Hook 错误；未设置时返回 nil。
func (h *recordingHook) Before(ctx *HookContext) error {
	if h.calls != nil {
		*h.calls = append(*h.calls, "before:"+h.name+":"+ctx.OpType().String())
	}
	if h.beforeFn != nil {
		h.beforeFn(ctx)
	}
	return h.beforeErr
}

// After 记录后置 Hook 调用并按需返回预设错误。
//
// 参数：
//   - ctx: Hook 上下文，用于读取当前操作类型并传递给自定义断言函数。
//
// 返回：
//   - error: 预设的后置 Hook 错误；未设置时返回 nil。
func (h *recordingHook) After(ctx *HookContext) error {
	if h.calls != nil {
		*h.calls = append(*h.calls, "after:"+h.name+":"+ctx.OpType().String())
	}
	if h.afterFn != nil {
		h.afterFn(ctx)
	}
	return h.afterErr
}
