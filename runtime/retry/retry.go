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
	// 返回 nil 时结束重试；返回非 nil error 时，[Retry] 会把该错误视为一次可重试失败并继续等待下一次调用。
	//
	// 参数：无。
	//
	// 返回：
	//   - error: 返回 nil 时停止重试；返回非 nil 时由 [Retry] 继续按退避策略重试，不会被直接返回给调用方。
	RetryableFunc func() error

	// RetryableFuncWithContext 表示接收 context 的可重试函数。
	//
	// 实现应尊重 ctx.Done() 并在不再继续工作时尽快返回。返回 nil 时结束重试；返回非 nil error 时，
	// [RetryWithContext] 会把该错误视为一次可重试失败并继续等待下一次调用。
	//
	// 参数：
	//   - ctx: 当前重试调用使用的上下文；实现应监听其取消信号并尽快终止不再需要的工作。
	//
	// 返回：
	//   - error: 返回 nil 时停止重试；返回非 nil 时由 [RetryWithContext] 继续按退避策略重试，不会被直接返回给调用方。
	RetryableFuncWithContext func(ctx context.Context) error
)

// Retry 使用 [context.Background] 按退避策略反复执行 fn，直到 fn 返回 nil。
//
// 当 fn 返回普通错误时，Retry 不会提前返回该错误，而是继续等待下一次重试。
// 由于本函数内部固定使用 [context.Background]，调用方不能直接取消本次重试循环；
// 如果 fn 始终失败且没有其他外部中断条件，本函数可能永不返回。传入 nil fn 会在首次调用时 panic。
//
// 参数：
//   - fn: 待执行的可重试函数；返回非 nil error 时会继续重试。
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
// RetryWithContext 返回 ctx.Err()。传入 nil ctx 或 nil fn 会在首次使用时 panic。
//
// 参数：
//   - ctx: 用于取消或超时控制的上下文；取消或超时后会终止后续重试等待。
//   - fn: 待执行的可重试函数。fn 应尊重 ctx.Done() 并尽快返回；返回非 nil error 时会继续重试。
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
			// 每轮执行前先检查 ctx 是否已取消，避免继续调用不再需要的重试逻辑。
			return ctx.Err()
		default:
			// 调用可重试函数；成功时立即结束整个重试循环。
			err = fn(ctx)
			if err == nil {
				return nil
			}

			// 本轮失败后按退避策略计算下一次重试前的等待时间。
			delay := b.Duration()
			select {
			case <-ctx.Done():
				// 等待期间仍要监听 ctx 取消信号，避免在退避窗口内继续阻塞。
				return ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
	}
}
