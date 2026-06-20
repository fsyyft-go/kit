// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package retry 提供基于 Backoff 的重试循环，支持不带上下文和带 context.Context 的函数重试。
//
// Retry 和 RetryWithContext 都会在函数返回 nil 前持续重试；普通错误不会直接返回，
// 只有成功或 context 取消时才结束。
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

// Retry 使用 context.Background() 按退避策略反复执行 fn，直到 fn 返回 nil。
//
// 当 fn 返回普通错误时，本函数不会提前返回该错误，而是继续等待下一次重试。
// 如果 fn 始终失败且调用方未通过外部机制取消流程，本函数可能永不返回。
//
// 参数：
//   - fn：需要执行的可重试函数。
//   - opts：Backoff 配置选项。
//
// 返回值：
//   - error：fn 成功时返回 nil；若 fn 持续返回普通错误，本函数会持续重试而不返回。
func Retry(fn RetryableFunc, opts ...BackoffOption) error {
	return RetryWithContext(context.Background(), func(_ context.Context) error {
		return fn()
	}, opts...)
}

// RetryWithContext 按退避策略反复执行 fn，直到 fn 返回 nil 或 ctx 被取消。
//
// fn 成功时返回 nil；fn 返回非 nil error 时会继续重试。
// ctx.Done() 触发后返回 ctx.Err()，不会返回最近一次业务错误。
//
// 参数：
//   - ctx：上下文对象，用于控制取消或超时。
//   - fn：需要执行的可重试函数。
//   - opts：Backoff 配置选项。
//
// 返回值：
//   - error：fn 成功时返回 nil；ctx 取消或超时时返回 ctx.Err()；普通业务错误不会被直接返回。
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
