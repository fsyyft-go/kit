// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// 本包提供了通用的重试机制，支持带上下文和不带上下文的函数重试。
package retry

import (
	"context"
	"time"
)

type (
	// RetryableFunc 定义了可重试的函数类型。
	//
	// 签名：
	//   - func() error
	//
	// 参数：
	//   - 无参数。
	//
	// 返回值：
	//   - error：执行过程中发生的错误。
	RetryableFunc func() error

	// RetryableFuncWithContext 定义了带上下文的可重试函数类型。
	//
	// 签名：
	//   - func(ctx context.Context) error
	//
	// 参数：
	//   - ctx context.Context：上下文对象，用于控制取消、超时等。
	//
	// 返回值：
	//   - error：执行过程中发生的错误。
	RetryableFuncWithContext func(ctx context.Context) error
)

// Retry 对传入的 RetryableFunc 类型函数进行重试。
//
// 参数：
//   - fn RetryableFunc：需要重试的函数，签名为 func() error。
//
// 返回值：
//   - error：如果所有重试均失败，则返回最后一次的错误；否则返回 nil。
//
// 当前实现仅为占位，实际重试逻辑需后续补充。
func Retry(fn RetryableFunc, opts ...BackoffOption) error {
	return RetryWithContext(context.Background(), func(_ context.Context) error {
		return fn()
	}, opts...)
}

// RetryWithContext 对传入的带上下文的 RetryableFuncWithContext 类型函数进行重试。
// 支持通过 context.Context 控制重试过程，如取消或超时。
//
// 参数：
//   - ctx context.Context：上下文对象，用于控制重试过程的取消与超时。
//   - fn RetryableFuncWithContext：需要重试的函数，签名为 func(ctx context.Context) error。
//
// 返回值：
//   - error：如果所有重试均失败，则返回最后一次的错误；否则返回 nil。
//
// 当前实现仅为占位，实际重试逻辑需后续补充。
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
