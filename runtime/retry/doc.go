// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package retry 提供基于 Backoff 的重试等待计算和重试循环。
//
// 本包既支持直接重试 func() error，也支持由 context.Context 控制取消和超时的重试。
// Backoff 使用最小值、最大值、增长因子和可选抖动计算每次等待时间，适合需要指数退避的失败重试场景。
//
// Retry 和 RetryWithContext 在被调函数返回普通错误时不会直接结束；它们会继续等待下一次重试，
// 直到函数成功返回 nil，或上下文被取消后返回 ctx.Err()。如果调用方需要限制总时长、总次数
// 或区分可重试与不可重试错误，需要自行通过 context 或包装逻辑实现。
package retry
