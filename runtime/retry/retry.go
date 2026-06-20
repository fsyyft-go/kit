// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package retry

import (
	"context"
	"time"
)

type (
	// RetryableFunc 表示不接收 context 的可重试函数。
	//
	// 返回 nil 时结束重试；返回非 nil error 时，[Retry] 会继续按退避策略等待下一次调用。
	RetryableFunc func() error

	// RetryableFuncWithContext 表示接收 context 的可重试函数。
	//
	// 实现应尊重 ctx.Done() 并在不再继续工作时尽快返回。返回 nil 时结束重试；
	// 返回非 nil error 时，[RetryWithContext] 会继续按退避策略等待下一次调用。
	RetryableFuncWithContext func(ctx context.Context) error
)

// Retry 使用 [context.Background] 按退避策略反复执行 fn，直到 fn 返回 nil。
//
// 当 fn 返回普通错误时，Retry 不会提前返回该错误，而是继续等待下一次重试。
// 由于本函数内部固定使用 [context.Background]，调用方不能直接取消本次重试循环；
// 如果 fn 始终失败且没有其他外部中断条件，本函数可能永不返回。
//
// 参数：
//   - fn: 待执行的可重试函数，不能为空。
//   - opts: Backoff 配置选项，按传入顺序应用。
//
// 返回：
//   - error: fn 成功时返回 nil；普通业务错误不会被直接返回。
func Retry(fn RetryableFunc, opts ...BackoffOption) error {
	return RetryWithContext(context.Background(), func(_ context.Context) error {
		return fn()
	}, opts...)
}

// RetryWithContext 按退避策略反复执行 fn，直到 fn 返回 nil 或 ctx 被取消。
//
// 普通业务错误会触发下一次重试，不会被直接返回。ctx.Done() 在调用前或等待期间触发时，
// RetryWithContext 返回 ctx.Err()。ctx 和 fn 都不能为空。
//
// 参数：
//   - ctx: 用于取消或超时控制的上下文。
//   - fn: 待执行的可重试函数。fn 应尊重 ctx.Done() 并尽快返回。
//   - opts: Backoff 配置选项，按传入顺序应用。
//
// 返回：
//   - error: fn 成功时返回 nil；ctx 取消或超时时返回 ctx.Err()。
func RetryWithContext(ctx context.Context, fn RetryableFuncWithContext, opts ...BackoffOption) error {
	var err error

	b := NewBackoff(opts...)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = fn(ctx)
			if err == nil {
				// 执行成功，返回 nil，退出重试。
				return nil
			}

			// 执行失败，等待下一次重试。
			delay := b.Duration()
			select {
			case <-ctx.Done():
				// 上下文已取消，返回错误。
				return ctx.Err()
			case <-time.After(delay):
				// 等待下一次重试。
				continue
			}
		}
	}
}
