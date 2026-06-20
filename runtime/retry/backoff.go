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
	// Backoff 根据尝试次数计算退避等待时间。
	//
	// 零值 Backoff 在计算时会回退到默认最小值、最大值和增长因子。
	// Backoff 持有内部 attempt 计数器；[Backoff.Duration] 会推进该计数器，
	// 而 [Backoff.ForAttempt] 只按显式给定的尝试次数计算等待时间。
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

// Copy 返回一个参数配置与当前实例一致的新 [Backoff]。
//
// 新实例不会复制当前的 attempt 计数，返回后的第一次 [Backoff.Duration] 调用会从第 0 次尝试重新开始。
//
// 参数：无。
//
// 返回：
//   - *Backoff: 与当前实例配置一致、但 attempt 计数重置为 0 的新实例。
func (b *Backoff) Copy() *Backoff {
	return &Backoff{
		factor: b.factor,
		jitter: b.jitter,
		min:    b.min,
		max:    b.max,
	}
}

// Reset 将内部 attempt 计数重置为 0。
//
// 调用后，下一次 [Backoff.Duration] 会重新按第 0 次尝试计算等待时间。
//
// 参数：无。
func (b *Backoff) Reset() {
	atomic.StoreUint64(&b.attempt, 0)
}

// Duration 返回当前 attempt 对应的等待时间，并将 attempt 加 1。
//
// 多个 goroutine 并发调用 Duration 时会共享同一 attempt 序列，返回顺序取决于竞争结果；
// 需要按指定尝试次数独立计算等待时间时，请使用 [Backoff.ForAttempt]。
//
// 参数：无。
//
// 返回：
//   - time.Duration: 当前 attempt 对应的退避等待时间。
func (b *Backoff) Duration() time.Duration {
	// 先自增 attempt 计数器，再计算对应的等待时间。
	d := b.ForAttempt(float64(atomic.AddUint64(&b.attempt, 1) - 1))
	return d
}

// ForAttempt 根据指定的尝试次数计算等待时间。
//
// 当 min、max 或 factor 为非正值时，本方法会回退到默认配置；当 min 大于等于 max 时，
// 直接返回 max。启用 jitter 后，结果会在最小值和理论退避值之间随机取值，再按 max 截断。
//
// 参数：
//   - attempt: 目标尝试次数，从 0 开始；第 0 次尝试的理论等待时间为 min。
//
// 返回：
//   - time.Duration: 指定尝试次数对应的退避等待时间。
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

// Attempt 返回当前内部 attempt 计数。
//
// 返回值使用 float64，以便直接作为 [Backoff.ForAttempt] 的参数复用。
//
// 参数：无。
//
// 返回：
//   - float64: 当前内部 attempt 计数。
func (b *Backoff) Attempt() float64 {
	return float64(atomic.LoadUint64(&b.attempt))
}

// NewBackoff 创建一个新的 [Backoff] 实例。
//
// 默认配置为：min 100ms、max 10s、factor 2、jitter false。多个选项按传入顺序依次应用，
// 同一字段以后传入的值为准。
//
// 参数：
//   - opts: 可选的 Backoff 配置项。
//
// 返回：
//   - *Backoff: 应用全部选项后的退避配置实例。
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
