// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package retry

import (
	"time"
)

// 以下为 Backoff 的默认参数配置。
// 可通过 BackoffOption 机制覆盖。
var (
	// minDefault 为 Backoff 的最小等待时间。
	minDefault = 100 * time.Millisecond
	// maxDefault 为 Backoff 的最大等待时间。
	maxDefault = 10 * time.Second
	// factorDefault 为 Backoff 的增长因子。
	factorDefault = float64(2)
	// jitterDefault 为 Backoff 是否启用抖动。
	jitterDefault = false
)

// BackoffOption 配置 [Backoff] 的等待参数。
//
// 多个选项按传入顺序依次应用；同一字段以后传入的值为准。
type BackoffOption func(*Backoff)

// WithMin 设置 [Backoff] 的最小等待时间。
//
// min 小于等于 0 时不会在构造阶段报错；实际计算等待时间时会回退到默认最小值。
//
// 参数：
//   - min: 期望设置的最小等待时间。
//
// 返回：
//   - BackoffOption: 写入最小等待时间配置的选项函数。
func WithMin(min time.Duration) BackoffOption {
	return func(b *Backoff) {
		b.min = min
	}
}

// WithMax 设置 [Backoff] 的最大等待时间。
//
// max 小于等于 0 时不会在构造阶段报错；实际计算等待时间时会回退到默认最大值。
//
// 参数：
//   - max: 期望设置的最大等待时间。
//
// 返回：
//   - BackoffOption: 写入最大等待时间配置的选项函数。
func WithMax(max time.Duration) BackoffOption {
	return func(b *Backoff) {
		b.max = max
	}
}

// WithFactor 设置 [Backoff] 的增长因子。
//
// factor 小于等于 0 时不会在构造阶段报错；实际计算等待时间时会回退到默认增长因子。
//
// 参数：
//   - factor: 期望设置的退避增长因子。
//
// 返回：
//   - BackoffOption: 写入增长因子配置的选项函数。
func WithFactor(factor float64) BackoffOption {
	return func(b *Backoff) {
		b.factor = factor
	}
}

// WithJitter 设置 [Backoff] 是否在退避结果中加入抖动。
//
// 启用后，[Backoff.ForAttempt] 会在最小值和理论退避值之间取随机值，再按最大值上限截断。
//
// 参数：
//   - jitter: 是否启用抖动。
//
// 返回：
//   - BackoffOption: 写入抖动配置的选项函数。
func WithJitter(jitter bool) BackoffOption {
	return func(b *Backoff) {
		b.jitter = jitter
	}
}
