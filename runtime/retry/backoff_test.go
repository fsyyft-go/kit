// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//
// Backoff 单元测试
//
// 设计思路：
//   - 所有测试用例均采用表格驱动，便于批量验证多种参数组合和边界场景。
//   - 断言全部使用 stretchr/testify/assert，提升可读性和一致性。
//   - 每个测试函数前均有详细注释，说明测试目标和覆盖点。
//   - 涵盖 Backoff 的所有公开方法，包括 Copy、Reset、Duration、ForAttempt、Attempt。
//   - 针对极端参数、默认值、抖动、并发等场景均有覆盖。
//   - 使用方法：直接 go test 运行本文件即可。
//

package retry

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 测试 Backoff.Duration、Reset、factor、min/max 逻辑。
func TestBackoff_DurationAndReset(t *testing.T) {
	// 定义测试用例，覆盖不同 factor、min、max 组合及重置场景。
	tests := []struct {
		name        string          // 用例名称。
		b           Backoff         // 待测试的 Backoff 实例。
		calls       []time.Duration // 期望每次 Duration 的返回值。
		reset       int             // 第几次调用后重置。
		resetExpect time.Duration   // 重置后第一次期望值。
	}{
		{
			name:        "factor=2, min=100ms, max=10s",
			b:           Backoff{min: 100 * time.Millisecond, max: 10 * time.Second, factor: 2},
			calls:       []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond},
			reset:       3,
			resetExpect: 100 * time.Millisecond,
		},
		{
			name:        "factor=1.5, min=100ms, max=10s",
			b:           Backoff{min: 100 * time.Millisecond, max: 10 * time.Second, factor: 1.5},
			calls:       []time.Duration{100 * time.Millisecond, 150 * time.Millisecond, 225 * time.Millisecond},
			reset:       3,
			resetExpect: 100 * time.Millisecond,
		},
		{
			name:        "factor=1.75, min=100ns, max=10s",
			b:           Backoff{min: 100 * time.Nanosecond, max: 10 * time.Second, factor: 1.75},
			calls:       []time.Duration{100 * time.Nanosecond, 175 * time.Nanosecond, 306 * time.Nanosecond},
			reset:       3,
			resetExpect: 100 * time.Nanosecond,
		},
		{
			name:        "min>=max, 直接返回max",
			b:           Backoff{min: 500 * time.Second, max: 100 * time.Second, factor: 1},
			calls:       []time.Duration{100 * time.Second},
			reset:       1,
			resetExpect: 100 * time.Second,
		},
	}
	for _, tt := range tests {
		// 使用 t.Run 子测试，便于并行和单独调试。
		t.Run(tt.name, func(t *testing.T) {
			b := tt.b
			for i, expect := range tt.calls {
				// 依次调用 Duration 并断言返回值。
				actual := b.Duration()
				assert.Equal(t, expect, actual, "第 %d 次 Duration 期望值不符", i+1)
			}
			// 调用 Reset 后再次断言。
			b.Reset()
			actual := b.Duration()
			assert.Equal(t, tt.resetExpect, actual, "Reset 后第一次 Duration 期望值不符")
		})
	}
}

// 测试 ForAttempt 方法，覆盖不同 attempt 值和参数边界。
func TestBackoff_ForAttempt(t *testing.T) {
	// 定义测试用例，覆盖不同 factor、min、max 组合及边界。
	tests := []struct {
		name    string          // 用例名称。
		b       Backoff         // 待测试的 Backoff 实例。
		inputs  []float64       // 输入的 attempt 值。
		expects []time.Duration // 期望的返回值。
	}{
		{
			name:    "factor=2, min=100ms, max=10s",
			b:       Backoff{min: 100 * time.Millisecond, max: 10 * time.Second, factor: 2},
			inputs:  []float64{0, 1, 2},
			expects: []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond},
		},
		{
			name:    "factor=1.5, min=100ms, max=10s",
			b:       Backoff{min: 100 * time.Millisecond, max: 10 * time.Second, factor: 1.5},
			inputs:  []float64{0, 1, 2},
			expects: []time.Duration{100 * time.Millisecond, 150 * time.Millisecond, 225 * time.Millisecond},
		},
		{
			name:    "min>=max, 直接返回max",
			b:       Backoff{min: 500 * time.Second, max: 100 * time.Second, factor: 1},
			inputs:  []float64{0, 1},
			expects: []time.Duration{100 * time.Second, 100 * time.Second},
		},
		{
			name:    "factor<=0, min/max<=0, 走默认值",
			b:       Backoff{factor: 0, min: 0, max: 0},
			inputs:  []float64{0, 1, 2},
			expects: []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond},
		},
	}
	for _, tt := range tests {
		// 使用 t.Run 子测试，便于并行和单独调试。
		t.Run(tt.name, func(t *testing.T) {
			for i, in := range tt.inputs {
				// 调用 ForAttempt 并断言返回值。
				actual := tt.b.ForAttempt(in)
				assert.Equal(t, tt.expects[i], actual, "ForAttempt(%v) 期望值不符", in)
			}
		})
	}
}

// 测试 Attempt 方法和 Duration 计数器自增。
func TestBackoff_Attempt(t *testing.T) {
	// 构造 Backoff 并逐步调用 Duration，检查 Attempt 是否同步递增。
	b := &Backoff{min: 100 * time.Millisecond, max: 10 * time.Second, factor: 2}
	// 检查初始 Attempt。
	assert.Equal(t, float64(0), b.Attempt(), "初始 Attempt 应为 0")
	b.Duration()
	// 检查第一次 Duration 后 Attempt。
	assert.Equal(t, float64(1), b.Attempt(), "第一次 Duration 后 Attempt 应为 1")
	b.Duration()
	// 检查第二次 Duration 后 Attempt。
	assert.Equal(t, float64(2), b.Attempt(), "第二次 Duration 后 Attempt 应为 2")
	b.Duration()
	// 检查第三次 Duration 后 Attempt。
	assert.Equal(t, float64(3), b.Attempt(), "第三次 Duration 后 Attempt 应为 3")
	b.Reset()
	// 检查 Reset 后 Attempt。
	assert.Equal(t, float64(0), b.Attempt(), "Reset 后 Attempt 应为 0")
}

// 测试 Copy 方法，确保参数复制但计数器不共享。
func TestBackoff_Copy(t *testing.T) {
	b := &Backoff{min: 100 * time.Millisecond, max: 10 * time.Second, factor: 2, jitter: true}
	b2 := b.Copy()
	// 检查 factor 是否一致。
	assert.Equal(t, b.factor, b2.factor, "factor 应一致")
	// 检查 jitter 是否一致。
	assert.Equal(t, b.jitter, b2.jitter, "jitter 应一致")
	// 检查 min 是否一致。
	assert.Equal(t, b.min, b2.min, "min 应一致")
	// 检查 max 是否一致。
	assert.Equal(t, b.max, b2.max, "max 应一致")
	// 检查实例地址是否不同。
	assert.NotSame(t, b, b2, "Copy 返回的新实例地址应不同")
	b.Duration()
	// 检查原实例计数器递增。
	assert.Equal(t, float64(1), b.Attempt(), "原实例计数器应递增")
	// 检查新实例计数器为 0。
	assert.Equal(t, float64(0), b2.Attempt(), "新实例计数器应为 0")
}

// 测试 jitter 场景，确保返回值在合理区间。
func TestBackoff_Jitter(t *testing.T) {
	b := &Backoff{min: 100 * time.Millisecond, max: 10 * time.Second, factor: 2, jitter: true}
	// 第一次必定为 min。
	assert.Equal(t, 100*time.Millisecond, b.Duration(), "第一次 Duration 应为 min")
	// 后续带抖动，区间断言。
	v := b.Duration()
	assert.GreaterOrEqual(t, v, 100*time.Millisecond, "带 jitter 的 Duration 不应小于 min")
	assert.LessOrEqual(t, v, 200*time.Millisecond, "带 jitter 的 Duration 不应大于理论最大")
	v2 := b.Duration()
	assert.GreaterOrEqual(t, v2, 100*time.Millisecond, "带 jitter 的 Duration 不应小于 min")
	assert.LessOrEqual(t, v2, 400*time.Millisecond, "带 jitter 的 Duration 不应大于理论最大")
	b.Reset()
	// Reset 后第一次必定为 min。
	assert.Equal(t, 100*time.Millisecond, b.Duration(), "Reset 后第一次 Duration 应为 min")
}

// 并发场景测试，确保 ForAttempt 并发安全。
func TestBackoff_Concurrent(t *testing.T) {
	b := &Backoff{min: 100 * time.Millisecond, max: 10 * time.Second, factor: 2}
	wg := &sync.WaitGroup{}
	results := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		// 启动 10 个 goroutine 并发调用 ForAttempt。
		go func(idx int) {
			defer wg.Done()
			results[idx] = b.ForAttempt(float64(idx))
		}(i)
	}
	wg.Wait()
	for i := 0; i < 10; i++ {
		// 期望值为 min * factor^i。
		expect := time.Duration(float64(100*time.Millisecond) * math.Pow(2, float64(i)))
		if expect > 10*time.Second {
			expect = 10 * time.Second
		}
		assert.Equal(t, expect, results[i], "并发 ForAttempt(%d) 期望值不符", i)
	}
}
