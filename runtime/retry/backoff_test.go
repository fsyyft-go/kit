// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package retry

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewBackoff_DefaultsAndOptions 验证 NewBackoff 的默认配置和选项覆盖行为。
//
// 该测试通过表驱动用例覆盖默认值、自定义选项和重复选项覆盖语义，确保 Backoff 构造契约稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewBackoff_DefaultsAndOptions(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveOptions []BackoffOption
		wantMin     time.Duration
		wantMax     time.Duration
		wantFactor  float64
		wantJitter  bool
	}{
		{
			name:        "success/defaults",
			description: "验证 NewBackoff 在未传入选项时使用稳定的默认退避参数。",
			wantMin:     100 * time.Millisecond,
			wantMax:     10 * time.Second,
			wantFactor:  2,
			wantJitter:  false,
		},
		{
			name:        "success/custom-options",
			description: "验证 WithMin、WithMax、WithFactor 和 WithJitter 能完整覆盖默认退避参数。",
			giveOptions: []BackoffOption{
				WithMin(25 * time.Millisecond),
				WithMax(750 * time.Millisecond),
				WithFactor(3.5),
				WithJitter(true),
			},
			wantMin:    25 * time.Millisecond,
			wantMax:    750 * time.Millisecond,
			wantFactor: 3.5,
			wantJitter: true,
		},
		{
			name:        "success/last-option-wins",
			description: "验证同一字段被多个选项设置时，后传入的选项具有最终生效语义。",
			giveOptions: []BackoffOption{
				WithMin(10 * time.Millisecond),
				WithMin(20 * time.Millisecond),
				WithMax(100 * time.Millisecond),
				WithMax(200 * time.Millisecond),
				WithFactor(1.25),
				WithFactor(1.5),
				WithJitter(false),
				WithJitter(true),
			},
			wantMin:    20 * time.Millisecond,
			wantMax:    200 * time.Millisecond,
			wantFactor: 1.5,
			wantJitter: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := NewBackoff(tt.giveOptions...)

			require.NotNil(t, got)
			assert.Equal(t, tt.wantMin, got.min)
			assert.Equal(t, tt.wantMax, got.max)
			assert.Equal(t, tt.wantFactor, got.factor)
			assert.Equal(t, tt.wantJitter, got.jitter)
		})
	}
}

// TestBackoff_ForAttemptBoundaries 验证 ForAttempt 在增长、默认值和边界输入下的退避计算。
//
// 该测试覆盖指数增长、最大值截断、min >= max、非正因子、负 attempt 和溢出保护等公共行为。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestBackoff_ForAttemptBoundaries(t *testing.T) {
	tests := []struct {
		name         string
		description  string
		giveBackoff  *Backoff
		giveAttempt  float64
		wantDuration time.Duration
	}{
		{
			name:         "success/exponential-growth",
			description:  "验证 ForAttempt 按 min * factor^attempt 计算未触顶的指数退避时间。",
			giveBackoff:  NewBackoff(WithMin(10*time.Millisecond), WithMax(time.Second), WithFactor(2)),
			giveAttempt:  3,
			wantDuration: 80 * time.Millisecond,
		},
		{
			name:         "boundary/caps-at-max",
			description:  "验证指数退避结果超过 max 时返回 max 以保持上界约束。",
			giveBackoff:  NewBackoff(WithMin(10*time.Millisecond), WithMax(25*time.Millisecond), WithFactor(2)),
			giveAttempt:  2,
			wantDuration: 25 * time.Millisecond,
		},
		{
			name:         "boundary/min-greater-than-max",
			description:  "验证 min 大于 max 时 ForAttempt 直接返回 max，避免产生无效区间。",
			giveBackoff:  NewBackoff(WithMin(5*time.Second), WithMax(time.Second), WithFactor(2)),
			giveAttempt:  0,
			wantDuration: time.Second,
		},
		{
			name:         "boundary/min-equals-max",
			description:  "验证 min 等于 max 时退避时间保持为 max，不受 attempt 和 factor 影响。",
			giveBackoff:  NewBackoff(WithMin(17*time.Millisecond), WithMax(17*time.Millisecond), WithFactor(8)),
			giveAttempt:  4,
			wantDuration: 17 * time.Millisecond,
		},
		{
			name:         "boundary/zero-value-uses-defaults",
			description:  "验证 Backoff 零值通过 ForAttempt 使用默认 min、max 和 factor 参数。",
			giveBackoff:  &Backoff{},
			giveAttempt:  2,
			wantDuration: 400 * time.Millisecond,
		},
		{
			name:         "boundary/negative-factor-uses-default",
			description:  "验证 factor 小于零时使用默认增长因子并保留有效的自定义边界。",
			giveBackoff:  NewBackoff(WithMin(5*time.Millisecond), WithMax(100*time.Millisecond), WithFactor(-3)),
			giveAttempt:  2,
			wantDuration: 20 * time.Millisecond,
		},
		{
			name:         "boundary/zero-factor-uses-default",
			description:  "验证 factor 等于零时使用默认增长因子并保留有效的自定义边界。",
			giveBackoff:  NewBackoff(WithMin(5*time.Millisecond), WithMax(100*time.Millisecond), WithFactor(0)),
			giveAttempt:  2,
			wantDuration: 20 * time.Millisecond,
		},
		{
			name:         "boundary/factor-below-one-clamps-to-min",
			description:  "验证正因子小于一导致计算值低于 min 时，ForAttempt 返回 min。",
			giveBackoff:  NewBackoff(WithMin(100*time.Millisecond), WithMax(time.Second), WithFactor(0.5)),
			giveAttempt:  3,
			wantDuration: 100 * time.Millisecond,
		},
		{
			name:         "boundary/negative-attempt-clamps-to-min",
			description:  "验证负 attempt 导致计算值低于 min 时，ForAttempt 返回 min。",
			giveBackoff:  NewBackoff(WithMin(100*time.Millisecond), WithMax(time.Second), WithFactor(2)),
			giveAttempt:  -1,
			wantDuration: 100 * time.Millisecond,
		},
		{
			name:         "boundary/overflow-returns-max",
			description:  "验证指数计算超过可安全转换的整数范围时，ForAttempt 返回 max 作为溢出保护。",
			giveBackoff:  NewBackoff(WithMin(time.Nanosecond), WithMax(2*time.Nanosecond), WithFactor(1e18)),
			giveAttempt:  1000,
			wantDuration: 2 * time.Nanosecond,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := tt.giveBackoff.ForAttempt(tt.giveAttempt)

			assert.Equal(t, tt.wantDuration, got)
		})
	}
}

// TestBackoff_DurationAttemptAndReset 验证 Duration、Attempt 和 Reset 的状态协作语义。
//
// 该测试覆盖连续 Duration 调用时的 attempt 自增、退避序列计算、最大值截断和 Reset 后重新开始计数的行为。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestBackoff_DurationAttemptAndReset(t *testing.T) {
	tests := []struct {
		name           string
		description    string
		setup          func() *Backoff
		wantDurations  []time.Duration
		wantAfterReset time.Duration
	}{
		{
			name:        "success/growing-sequence",
			description: "验证 Duration 按当前 attempt 返回指数增长序列，并在每次调用后递增计数。",
			setup: func() *Backoff {
				return NewBackoff(WithMin(10*time.Millisecond), WithMax(time.Second), WithFactor(2))
			},
			wantDurations:  []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 40 * time.Millisecond},
			wantAfterReset: 10 * time.Millisecond,
		},
		{
			name:        "boundary/capped-sequence",
			description: "验证 Duration 序列触达 max 后保持在上界，同时 attempt 仍继续递增。",
			setup: func() *Backoff {
				return NewBackoff(WithMin(10*time.Millisecond), WithMax(25*time.Millisecond), WithFactor(2))
			},
			wantDurations:  []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 25 * time.Millisecond, 25 * time.Millisecond},
			wantAfterReset: 10 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			b := tt.setup()
			require.NotNil(t, b)
			assert.Equal(t, float64(0), b.Attempt())

			for i, want := range tt.wantDurations {
				assert.Equal(t, float64(i), b.Attempt())
				got := b.Duration()
				assert.Equal(t, want, got)
				assert.Equal(t, float64(i+1), b.Attempt())
			}

			b.Reset()
			assert.Equal(t, float64(0), b.Attempt())
			assert.Equal(t, tt.wantAfterReset, b.Duration())
			assert.Equal(t, float64(1), b.Attempt())
		})
	}
}

// TestBackoff_Copy 验证 Copy 复制参数但不复制运行时 attempt 状态。
//
// 该测试确认 Copy 返回独立实例，复制 min、max、factor 和 jitter 配置，并从零 attempt 开始独立计数。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestBackoff_Copy(t *testing.T) {
	// 先推进原始实例的 attempt，以验证 Copy 不复制运行时计数器。
	original := NewBackoff(
		WithMin(25*time.Millisecond),
		WithMax(500*time.Millisecond),
		WithFactor(3),
		WithJitter(true),
	)
	assert.Equal(t, 25*time.Millisecond, original.Duration())
	assert.Equal(t, float64(1), original.Attempt())

	copied := original.Copy()

	require.NotNil(t, copied)
	assert.NotSame(t, original, copied)
	assert.Equal(t, original.min, copied.min)
	assert.Equal(t, original.max, copied.max)
	assert.Equal(t, original.factor, copied.factor)
	assert.Equal(t, original.jitter, copied.jitter)
	assert.Equal(t, float64(1), original.Attempt())
	assert.Equal(t, float64(0), copied.Attempt())

	// 分别推进原始实例和副本，验证两个实例的 attempt 彼此独立。
	_ = original.Duration()
	assert.Equal(t, float64(2), original.Attempt())
	assert.Equal(t, float64(0), copied.Attempt())
	assert.Equal(t, 25*time.Millisecond, copied.Duration())
	assert.Equal(t, float64(1), copied.Attempt())
	assert.Equal(t, float64(2), original.Attempt())
}

// TestBackoff_JitterRange 验证启用 jitter 后返回值始终落在稳定边界内。
//
// 该测试不依赖随机值的具体分布，只断言 jitter 结果满足 min、理论退避值和 max 共同形成的范围约束。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestBackoff_JitterRange(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveBackoff *Backoff
		giveAttempt float64
		wantMin     time.Duration
		wantMax     time.Duration
		repeat      int
	}{
		{
			name:        "boundary/first-attempt-is-min",
			description: "验证 attempt 为零时 jitter 区间退化为 min，返回值稳定等于 min。",
			giveBackoff: NewBackoff(WithMin(10*time.Millisecond), WithMax(time.Second), WithFactor(4), WithJitter(true)),
			giveAttempt: 0,
			wantMin:     10 * time.Millisecond,
			wantMax:     10 * time.Millisecond,
			repeat:      5,
		},
		{
			name:        "success/within-theoretical-range",
			description: "验证未触达 max 的 jitter 结果位于 min 与理论指数退避值之间。",
			giveBackoff: NewBackoff(WithMin(10*time.Millisecond), WithMax(time.Second), WithFactor(4), WithJitter(true)),
			giveAttempt: 1,
			wantMin:     10 * time.Millisecond,
			wantMax:     40 * time.Millisecond,
			repeat:      20,
		},
		{
			name:        "boundary/clamped-by-max",
			description: "验证 jitter 结果超过 max 时会被上界截断，最终返回值不超过 max。",
			giveBackoff: NewBackoff(WithMin(10*time.Millisecond), WithMax(25*time.Millisecond), WithFactor(4), WithJitter(true)),
			giveAttempt: 2,
			wantMin:     10 * time.Millisecond,
			wantMax:     25 * time.Millisecond,
			repeat:      20,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			for i := 0; i < tt.repeat; i++ {
				got := tt.giveBackoff.ForAttempt(tt.giveAttempt)
				assert.GreaterOrEqual(t, got, tt.wantMin)
				assert.LessOrEqual(t, got, tt.wantMax)
			}
		})
	}
}

// TestBackoff_ForAttemptConcurrent 验证 ForAttempt 可在并发调用场景下稳定计算退避值。
//
// 该测试使用多个 goroutine 并发读取同一 Backoff 参数，并通过超时保护避免并发回归导致测试挂起。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestBackoff_ForAttemptConcurrent(t *testing.T) {
	const workers = 16

	b := NewBackoff(WithMin(time.Millisecond), WithMax(512*time.Millisecond), WithFactor(2))
	results := make([]time.Duration, workers)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		attempt := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[attempt] = b.ForAttempt(float64(attempt))
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		require.FailNow(t, "ForAttempt 并发调用未在预期时间内完成")
	}

	for attempt, got := range results {
		want := time.Duration(float64(time.Millisecond) * math.Pow(2, float64(attempt)))
		if want > 512*time.Millisecond {
			want = 512 * time.Millisecond
		}
		assert.Equal(t, want, got)
	}
}
