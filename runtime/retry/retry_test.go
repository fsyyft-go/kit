// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package retry

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRetry_Behavior 验证 Retry 在成功和失败后重试场景下的行为。
//
// 该测试通过短退避配置避免慢测试，并断言 Retry 委托到无取消上下文后的调用次数与成功返回语义。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRetry_Behavior(t *testing.T) {
	transientErr := errors.New("transient failure")
	tests := []struct {
		name        string
		description string
		setup       func(t *testing.T) (RetryableFunc, []BackoffOption, *int)
		wantCalls   int
	}{
		{
			name:        "success/first-attempt",
			description: "验证 Retry 在函数首次成功时立即返回 nil 且不执行额外重试。",
			setup: func(t *testing.T) (RetryableFunc, []BackoffOption, *int) {
				t.Helper()
				calls := 0
				return func() error {
					calls++
					return nil
				}, shortBackoffOptions(), &calls
			},
			wantCalls: 1,
		},
		{
			name:        "success/after-retries",
			description: "验证 Retry 在前两次返回错误后继续重试，并在后续成功时返回 nil。",
			setup: func(t *testing.T) (RetryableFunc, []BackoffOption, *int) {
				t.Helper()
				calls := 0
				return func() error {
					calls++
					if calls < 3 {
						return transientErr
					}
					return nil
				}, shortBackoffOptions(), &calls
			},
			wantCalls: 3,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			fn, opts, calls := tt.setup(t)
			err := Retry(fn, opts...)

			require.NoError(t, err)
			assert.Equal(t, tt.wantCalls, *calls)
		})
	}
}

// TestRetryWithContext_Behavior 验证 RetryWithContext 的重试、取消、超时和错误返回行为。
//
// 该测试使用表驱动用例覆盖成功、失败后重试、调用前取消、重试等待期间取消和函数内部取消等关键语义。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRetryWithContext_Behavior(t *testing.T) {
	transientErr := errors.New("transient failure")
	tests := []struct {
		name        string
		description string
		setup       func(t *testing.T) (context.Context, RetryableFuncWithContext, []BackoffOption, *int)
		wantErr     bool
		wantErrIs   error
		wantCalls   int
	}{
		{
			name:        "success/first-attempt",
			description: "验证 RetryWithContext 在首次调用成功时返回 nil，且只调用函数一次。",
			setup: func(t *testing.T) (context.Context, RetryableFuncWithContext, []BackoffOption, *int) {
				t.Helper()
				calls := 0
				return context.Background(), func(ctx context.Context) error {
					calls++
					assert.NoError(t, ctx.Err())
					return nil
				}, shortBackoffOptions(), &calls
			},
			wantCalls: 1,
		},
		{
			name:        "success/after-retries",
			description: "验证 RetryWithContext 在暂时性错误后按退避策略重试，并在成功后返回 nil。",
			setup: func(t *testing.T) (context.Context, RetryableFuncWithContext, []BackoffOption, *int) {
				t.Helper()
				calls := 0
				return context.Background(), func(ctx context.Context) error {
					calls++
					assert.NoError(t, ctx.Err())
					if calls < 3 {
						return transientErr
					}
					return nil
				}, shortBackoffOptions(), &calls
			},
			wantCalls: 3,
		},
		{
			name:        "error/canceled-before-first-attempt",
			description: "验证传入已取消 context 时 RetryWithContext 直接返回 context.Canceled 且不调用函数。",
			setup: func(t *testing.T) (context.Context, RetryableFuncWithContext, []BackoffOption, *int) {
				t.Helper()
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				calls := 0
				return ctx, func(ctx context.Context) error {
					calls++
					return transientErr
				}, shortBackoffOptions(), &calls
			},
			wantErr:   true,
			wantErrIs: context.Canceled,
			wantCalls: 0,
		},
		{
			name:        "error/deadline-before-first-attempt",
			description: "验证传入已超时 context 时 RetryWithContext 直接返回 context.DeadlineExceeded 且不调用函数。",
			setup: func(t *testing.T) (context.Context, RetryableFuncWithContext, []BackoffOption, *int) {
				t.Helper()
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Nanosecond))
				t.Cleanup(cancel)
				calls := 0
				return ctx, func(ctx context.Context) error {
					calls++
					return transientErr
				}, shortBackoffOptions(), &calls
			},
			wantErr:   true,
			wantErrIs: context.DeadlineExceeded,
			wantCalls: 0,
		},
		{
			name:        "error/canceled-during-backoff",
			description: "验证函数失败后处于退避等待时 context 被取消，RetryWithContext 返回 context.Canceled 并停止重试。",
			setup: func(t *testing.T) (context.Context, RetryableFuncWithContext, []BackoffOption, *int) {
				t.Helper()
				ctx, cancel := context.WithCancel(context.Background())
				t.Cleanup(cancel)
				calls := 0
				return ctx, func(ctx context.Context) error {
					calls++
					cancel()
					return transientErr
				}, []BackoffOption{WithMin(time.Hour), WithMax(time.Hour), WithFactor(1)}, &calls
			},
			wantErr:   true,
			wantErrIs: context.Canceled,
			wantCalls: 1,
		},
		{
			name:        "error/deadline-during-backoff",
			description: "验证函数失败后处于退避等待时 context 超时，RetryWithContext 返回 context.DeadlineExceeded 并停止重试。",
			setup: func(t *testing.T) (context.Context, RetryableFuncWithContext, []BackoffOption, *int) {
				t.Helper()
				ctx := newControllableContext()
				calls := 0
				return ctx, func(giveCtx context.Context) error {
					calls++
					require.NoError(t, giveCtx.Err())
					ctx.complete(context.DeadlineExceeded)
					return transientErr
				}, []BackoffOption{WithMin(time.Hour), WithMax(time.Hour), WithFactor(1)}, &calls
			},
			wantErr:   true,
			wantErrIs: context.DeadlineExceeded,
			wantCalls: 1,
		},
		{
			name:        "error/function-observes-cancellation",
			description: "验证函数内部因 context 取消返回错误后，RetryWithContext 优先返回 context.Canceled 语义。",
			setup: func(t *testing.T) (context.Context, RetryableFuncWithContext, []BackoffOption, *int) {
				t.Helper()
				ctx, cancel := context.WithCancel(context.Background())
				t.Cleanup(cancel)
				calls := 0
				return ctx, func(ctx context.Context) error {
					calls++
					cancel()
					return ctx.Err()
				}, shortBackoffOptions(), &calls
			},
			wantErr:   true,
			wantErrIs: context.Canceled,
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			ctx, fn, opts, calls := tt.setup(t)
			err := RetryWithContext(ctx, fn, opts...)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					assert.ErrorIs(t, err, tt.wantErrIs)
				}
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantCalls, *calls)
		})
	}
}

// controllableContext 是测试专用的可控 context 实现。
//
// 该类型允许测试在函数首次调用后同步完成 context，避免使用极短 timeout 造成时序脆弱性。
type controllableContext struct {
	// done 是 context 完成通知通道，complete 会关闭该通道。
	done chan struct{}
	// once 保证 complete 多次调用时只关闭一次 done。
	once sync.Once
	// err 保存 context 完成后 Err 返回的错误语义。
	err error
}

// newControllableContext 构造未完成的可控测试 context。
//
// 该辅助函数为 RetryWithContext 的等待阶段错误路径提供可同步触发的 context.Done 信号。
//
// 返回：
//   - *controllableContext: 初始未完成、可通过 complete 触发 Done 的测试 context。
func newControllableContext() *controllableContext {
	return &controllableContext{done: make(chan struct{})}
}

// Deadline 返回该测试 context 不包含真实截止时间的语义。
//
// 参数：无。
//
// 返回：
//   - time.Time: 零值时间。
//   - bool: 固定为 false，表示没有真实 deadline。
func (c *controllableContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

// Done 返回该测试 context 的完成通知通道。
//
// 参数：无。
//
// 返回：
//   - <-chan struct{}: complete 被调用后关闭的完成通道。
func (c *controllableContext) Done() <-chan struct{} {
	return c.done
}

// Err 返回该测试 context 完成后的错误语义。
//
// 参数：无。
//
// 返回：
//   - error: complete 设置的错误；未完成时返回 nil。
func (c *controllableContext) Err() error {
	select {
	case <-c.done:
		return c.err
	default:
		return nil
	}
}

// Value 返回该测试 context 不携带任何键值的语义。
//
// 参数：
//   - key: 调用方查询的 context key，本实现始终忽略。
//
// 返回：
//   - any: 固定为 nil，表示不存在对应值。
func (c *controllableContext) Value(key any) any {
	return nil
}

// complete 同步完成该测试 context 并设置 Err 返回值。
//
// 该辅助方法只在首次调用时关闭 Done 通道，保证重复调用不会 panic。
//
// 参数：
//   - err: context 完成后 Err 返回的错误语义。
func (c *controllableContext) complete(err error) {
	c.once.Do(func() {
		c.err = err
		close(c.done)
	})
}

// shortBackoffOptions 返回适合单元测试的短退避配置。
//
// 该辅助函数集中提供纳秒级退避参数，避免 Retry 成功路径测试因默认退避时间变慢。
//
// 返回：
//   - []BackoffOption: 可传入 Retry 或 RetryWithContext 的短退避选项。
func shortBackoffOptions() []BackoffOption {
	return []BackoffOption{
		WithMin(time.Nanosecond),
		WithMax(time.Nanosecond),
		WithFactor(1),
	}
}
