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

// BackoffOption 类型用于配置 Backoff 实例的参数。
// 每个选项函数会修改 Backoff 的一个或多个字段。
type BackoffOption func(*Backoff)

// WithMin 设置 Backoff 的最小等待时间。
// 参数：
//   - min time.Duration：最小等待时间。
//
// 返回值：
//   - BackoffOption：用于设置 min 字段的选项函数。
func WithMin(min time.Duration) BackoffOption {
	return func(b *Backoff) {
		b.min = min
	}
}

// WithMax 设置 Backoff 的最大等待时间。
// 参数：
//   - max time.Duration：最大等待时间。
//
// 返回值：
//   - BackoffOption：用于设置 max 字段的选项函数。
func WithMax(max time.Duration) BackoffOption {
	return func(b *Backoff) {
		b.max = max
	}
}

// WithFactor 设置 Backoff 的增长因子。
// 参数：
//   - factor float64：每次递增时的乘数因子。
//
// 返回值：
//   - BackoffOption：用于设置 factor 字段的选项函数。
func WithFactor(factor float64) BackoffOption {
	return func(b *Backoff) {
		b.factor = factor
	}
}

// WithJitter 设置 Backoff 是否启用抖动机制。
// 参数：
//   - jitter bool：是否启用抖动。
//
// 返回值：
//   - BackoffOption：用于设置 jitter 字段的选项函数。
func WithJitter(jitter bool) BackoffOption {
	return func(b *Backoff) {
		b.jitter = jitter
	}
}
