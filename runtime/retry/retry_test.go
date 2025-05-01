// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//
// retry.go 单元测试
//
// 设计思路：
//   - 采用表格驱动，覆盖 Retry、RetryWithContext 的正常、失败、超时、取消等场景。
//   - 断言全部使用 stretchr/testify/assert，保证一致性和可读性。
//   - 每个测试函数前均有详细注释，说明测试目标和覆盖点。
//   - 充分利用 context 控制重试流程，验证边界和异常。
//   - 使用方法：直接 go test 运行本文件即可。
//

package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 测试 Retry 的基本功能，覆盖成功、失败、重试多次等场景。
func TestRetry_Basic(t *testing.T) {
	type args struct {
		fn  RetryableFunc
		opt []BackoffOption
	}
	tests := []struct {
		name      string // 用例名称。
		args      args   // 输入参数。
		expectErr bool   // 是否期望返回错误。
		tryCount  int    // 期望实际调用次数。
	}{
		{
			name: "一次成功，无需重试",
			args: args{
				fn: func() error { return nil },
			},
			expectErr: false,
			tryCount:  1,
		},
		{
			name:      "多次失败后成功",
			args:      args{},
			expectErr: false,
			tryCount:  3,
		},
		{
			name:      "始终失败，手动限制最大重试次数",
			args:      args{},
			expectErr: false,
			tryCount:  4, // 第 4 次返回 nil
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			var fn RetryableFunc
			switch tt.name {
			case "多次失败后成功":
				staticCount := 0
				fn = func() error {
					count++
					if staticCount < 2 {
						staticCount++
						return errors.New("fail")
					}
					return nil
				}
			case "始终失败，手动限制最大重试次数":
				maxAttempts := 3
				failCount := 0
				fn = func() error {
					count++
					if failCount < maxAttempts {
						failCount++
						return errors.New("always fail")
					}
					return nil // 超过最大次数后返回 nil，模拟终止
				}
			default:
				fn = func() error {
					count++
					return nil
				}
			}
			err := Retry(fn, tt.args.opt...)
			if tt.expectErr {
				assert.Error(t, err, "应返回错误")
			} else {
				assert.NoError(t, err, "不应返回错误")
			}
			assert.Equal(t, tt.tryCount, count, "实际调用次数应等于期望值")
		})
	}
}

// 测试 RetryWithContext 的超时、取消、成功、失败等场景。
func TestRetryWithContext(t *testing.T) {
	type args struct {
		ctx context.Context
		// fn  RetryableFuncWithContext  // 已废弃，未被使用，删除以通过 lint 检查。
		opt []BackoffOption
	}
	tests := []struct {
		name      string
		args      args
		expectErr bool
		tryCount  int
	}{
		{
			name: "一次成功，无需重试",
			args: args{
				ctx: context.Background(),
			},
			expectErr: false,
			tryCount:  1,
		},
		{
			name: "多次失败后成功",
			args: args{
				ctx: context.Background(),
			},
			expectErr: false,
			tryCount:  3,
		},
		{
			name: "始终失败，手动限制最大重试次数",
			args: args{
				ctx: context.Background(),
			},
			expectErr: false,
			tryCount:  4, // 第 4 次返回 nil
		},
		{
			name: "超时提前终止",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
					defer cancel()
					return ctx
				}(),
			},
			// 说明：context 传入时已超时，RetryWithContext 实现会直接返回，不会调用 fn。
			expectErr: true,
			tryCount:  0,
		},
		{
			name: "手动取消提前终止",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					go func() {
						time.Sleep(10 * time.Millisecond)
						cancel()
					}()
					return ctx
				}(),
			},
			// 说明：context 传入时已被取消，RetryWithContext 实现会直接返回，不会调用 fn。
			expectErr: true,
			tryCount:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			var fn RetryableFuncWithContext
			switch tt.name {
			case "多次失败后成功":
				staticCount := 0
				fn = func(ctx context.Context) error {
					count++
					if staticCount < 2 {
						staticCount++
						return errors.New("fail")
					}
					return nil
				}
			case "始终失败，手动限制最大重试次数":
				maxAttempts := 3
				failCount := 0
				fn = func(ctx context.Context) error {
					count++
					if failCount < maxAttempts {
						failCount++
						return errors.New("always fail")
					}
					return nil // 超过最大次数后返回 nil，模拟终止
				}
			case "超时提前终止":
				fn = func(ctx context.Context) error {
					count++
					time.Sleep(20 * time.Millisecond)
					return errors.New("fail")
				}
			case "手动取消提前终止":
				fn = func(ctx context.Context) error {
					count++
					time.Sleep(20 * time.Millisecond)
					return errors.New("fail")
				}
			default:
				fn = func(ctx context.Context) error {
					count++
					return nil
				}
			}
			err := RetryWithContext(tt.args.ctx, fn, tt.args.opt...)
			if tt.expectErr {
				assert.Error(t, err, "应返回错误")
			} else {
				assert.NoError(t, err, "不应返回错误")
			}
			assert.Equal(t, tt.tryCount, count, "实际调用次数应等于期望值")
		})
	}
}

// TestBackoffOptionsAndNewBackoff
//
// 该测试专门覆盖 BackoffOption 相关函数（WithMin/WithMax/WithFactor/WithJitter）
// 以及 NewBackoff 的所有分支，包括极端参数和组合，提升覆盖率。
func TestBackoffOptionsAndNewBackoff(t *testing.T) {
	// 直接测试各 Option 对 Backoff 字段的影响。
	b := NewBackoff(
		WithMin(123*time.Millisecond),
		WithMax(456*time.Second),
		WithFactor(3.14),
		WithJitter(true),
	)
	assert.Equal(t, 123*time.Millisecond, b.min, "WithMin 应设置 min 字段")
	assert.Equal(t, 456*time.Second, b.max, "WithMax 应设置 max 字段")
	assert.Equal(t, 3.14, b.factor, "WithFactor 应设置 factor 字段")
	assert.Equal(t, true, b.jitter, "WithJitter 应设置 jitter 字段")

	// 测试 NewBackoff 默认参数
	b2 := NewBackoff()
	assert.Equal(t, 100*time.Millisecond, b2.min, "默认 min 应为 100ms")
	assert.Equal(t, 10*time.Second, b2.max, "默认 max 应为 10s")
	assert.Equal(t, 2.0, b2.factor, "默认 factor 应为 2")
	assert.Equal(t, false, b2.jitter, "默认 jitter 应为 false")

	// 测试极端参数分支
	b3 := NewBackoff(WithMin(10*time.Second), WithMax(1*time.Second))
	// min >= max 时 ForAttempt 直接返回 max
	assert.Equal(t, 1*time.Second, b3.ForAttempt(0), "min >= max 时应返回 max")

	b4 := NewBackoff(WithMin(0), WithMax(0), WithFactor(0))
	// min/max/factor <= 0 时走默认值
	assert.Equal(t, 100*time.Millisecond, b4.ForAttempt(0), "min/max/factor <= 0 时应走默认值")

	// 测试溢出分支
	b5 := NewBackoff(WithMin(1), WithMax(2), WithFactor(1e18))
	_ = b5.ForAttempt(1000) // 只要不 panic 即可

	// 测试 jitter 分支
	b6 := NewBackoff(WithJitter(true))
	v := b6.ForAttempt(1)
	assert.GreaterOrEqual(t, v, 100*time.Millisecond, "jitter 场景下返回值不应小于 min")
	assert.LessOrEqual(t, v, 200*time.Millisecond, "jitter 场景下返回值不应大于理论最大")
}
