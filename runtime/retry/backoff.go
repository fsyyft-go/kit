// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package retry

import (
	"math"
	"math/rand"
	"sync/atomic"
	"time"
)

type (
	// Backoff 结构体用于实现带有指数退避和可选抖动机制的重试等待时间生成器。
	// 支持设置最小、最大等待时间、增长因子等参数。
	// 注意：Backoff 结构体本身不是并发安全的，但 ForAttempt 方法是并发安全的。
	Backoff struct {
		// attempt 用于记录当前的重试次数。
		attempt uint64

		// factor 为每次递增时的乘数因子。
		// 默认为 2。
		factor float64

		// jitter 表示是否启用抖动机制，用于在多并发场景下减少竞争。
		// 默认为 false。
		jitter bool

		// min 表示等待时间的最小值。
		// 默认为 100 毫秒。
		min time.Duration

		// max 表示等待时间的最大值。
		// 默认为 10 秒。
		max time.Duration
	}
)

const (
	// maxInt64 常量用于防止 float64 溢出 int64，留有一定安全余量。
	maxInt64 = float64(math.MaxInt64 - 512)
)

// Copy 返回一个与当前 Backoff 实例参数相同的新实例。
// 新实例不会复制尝试次数，只复制参数配置。
//
// 返回值：
//   - *Backoff：新建的 Backoff 实例，参数与当前实例一致。
func (b *Backoff) Copy() *Backoff {
	return &Backoff{
		factor: b.factor,
		jitter: b.jitter,
		min:    b.min,
		max:    b.max,
	}
}

// Reset 将当前尝试次数重置为零。
//
// 无参数，无返回值。
func (b *Backoff) Reset() {
	atomic.StoreUint64(&b.attempt, 0)
}

// Duration 返回当前尝试次数对应的等待时间，并将尝试次数加一。
// 本方法不是并发安全的，若需并发安全请使用 ForAttempt 方法。
//
// 返回值：
//   - time.Duration：当前尝试次数对应的等待时间。
func (b *Backoff) Duration() time.Duration {
	// 先自增 attempt 计数器，再计算对应的等待时间。
	d := b.ForAttempt(float64(atomic.AddUint64(&b.attempt, 1) - 1))
	return d
}

// ForAttempt 根据指定的尝试次数计算对应的等待时间。
// 该方法是并发安全的，适用于多个独立 Backoff 实例共享参数的场景。
//
// 参数：
//   - attempt float64：尝试次数，从 0 开始，表示第 0 次尝试。
//
// 返回值：
//   - time.Duration：指定尝试次数对应的等待时间。
func (b *Backoff) ForAttempt(attempt float64) time.Duration {
	// 若参数为零值，则使用默认值。
	min := b.min
	if min <= 0 {
		min = 100 * time.Millisecond
	}
	max := b.max
	if max <= 0 {
		max = 10 * time.Second
	}
	// 若最小值大于等于最大值，直接返回最大值。
	if min >= max {
		return max
	}
	factor := b.factor
	if factor <= 0 {
		factor = 2
	}
	// 计算当前尝试次数对应的等待时间。
	minf := float64(min)
	durf := minf * math.Pow(factor, attempt)
	// 若启用抖动机制，则在 [min, durf] 区间内随机取值。
	if b.jitter {
		durf = rand.Float64()*(durf-minf) + minf
	}
	// 防止 float64 溢出 int64。
	if durf > maxInt64 {
		return max
	}
	dur := time.Duration(durf)
	// 保证返回值在 [min, max] 区间内。
	if dur < min {
		return min
	}
	if dur > max {
		return max
	}
	return dur
}

// Attempt 返回当前的尝试次数。
// 返回值为 float64 类型，便于与 ForAttempt 方法配合使用。
//
// 返回值：
//   - float64：当前的尝试次数。
func (b *Backoff) Attempt() float64 {
	return float64(atomic.LoadUint64(&b.attempt))
}

// NewBackoff 创建一个新的 Backoff 实例，并应用所有给定的选项。
// 参数：
//   - opts ...BackoffOption：可选参数，用于配置 Backoff。
//
// 返回值：
//   - *Backoff：新建的 Backoff 实例。
func NewBackoff(opts ...BackoffOption) *Backoff {
	b := &Backoff{
		factor: factorDefault,
		jitter: jitterDefault,
		min:    minDefault,
		max:    maxDefault,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}
