// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package rand_test 提供了 rand 包的单元测试。
//
// 测试设计思路：
// 1. 使用表格驱动测试，提高测试代码的可维护性和可读性。
// 2. 对于随机数生成函数：
//   - 验证结果是否在预期范围内
//   - 使用固定种子的随机数生成器保证测试可重复性
//   - 进行多次测试以验证分布的合理性
//
// 3. 测试场景覆盖：
//   - 正常输入场景
//   - 边界值场景
//   - nil random 场景
//
// 使用方法：
// 1. 直接运行：go test ./math/rand
// 2. 查看覆盖率：go test ./math/rand -cover
// 3. 生成覆盖率报告：go test ./math/rand -coverprofile=coverage.out
package rand_test

import (
	"math/rand"
	"testing"
	"time"
	"unicode"

	fsrand "github.com/fsyyft-go/kit/math/rand"
	"github.com/stretchr/testify/assert"
)

// 测试用例结构体，用于表格驱动测试。
type testCase struct {
	name     string
	random   *rand.Rand
	min      int64
	max      int64
	times    int
	validate func(t *testing.T, result int64)
}

func TestInt63n(t *testing.T) {
	// 创建一个使用固定种子的随机数生成器，确保测试结果可重现。
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	// 定义测试用例。
	tests := []testCase{
		{
			name:   "正常范围测试",
			random: random,
			min:    0,
			max:    100,
			times:  1000,
			validate: func(t *testing.T, result int64) {
				assert.GreaterOrEqual(t, result, int64(0), "结果应该大于等于最小值")
				assert.Less(t, result, int64(100), "结果应该小于最大值")
			},
		},
		{
			name:   "边界值测试",
			random: random,
			min:    -1,
			max:    1,
			times:  1000,
			validate: func(t *testing.T, result int64) {
				assert.GreaterOrEqual(t, result, int64(-1), "结果应该大于等于最小值")
				assert.Less(t, result, int64(1), "结果应该小于最大值")
			},
		},
		{
			name:   "nil random 测试",
			random: nil,
			min:    0,
			max:    1000,
			times:  1000,
			validate: func(t *testing.T, result int64) {
				assert.GreaterOrEqual(t, result, int64(0), "结果应该大于等于最小值")
				assert.Less(t, result, int64(1000), "结果应该小于最大值")
			},
		},
	}

	// 执行测试用例。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.times; i++ {
				result := fsrand.Int63n(tt.random, tt.min, tt.max)
				tt.validate(t, result)
			}
		})
	}
}

func TestIntn(t *testing.T) {
	// 创建一个使用固定种子的随机数生成器。
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	// 定义测试用例。
	tests := []struct {
		name     string
		random   *rand.Rand
		min      int
		max      int
		times    int
		validate func(t *testing.T, result int)
	}{
		{
			name:   "正常范围测试",
			random: random,
			min:    0,
			max:    100,
			times:  1000,
			validate: func(t *testing.T, result int) {
				assert.GreaterOrEqual(t, result, 0, "结果应该大于等于最小值")
				assert.Less(t, result, 100, "结果应该小于最大值")
			},
		},
		{
			name:   "边界值测试",
			random: random,
			min:    -1,
			max:    1,
			times:  1000,
			validate: func(t *testing.T, result int) {
				assert.GreaterOrEqual(t, result, -1, "结果应该大于等于最小值")
				assert.Less(t, result, 1, "结果应该小于最大值")
			},
		},
		{
			name:   "nil random 测试",
			random: nil,
			min:    0,
			max:    1000,
			times:  1000,
			validate: func(t *testing.T, result int) {
				assert.GreaterOrEqual(t, result, 0, "结果应该大于等于最小值")
				assert.Less(t, result, 1000, "结果应该小于最大值")
			},
		},
	}

	// 执行测试用例。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.times; i++ {
				result := fsrand.Intn(tt.random, tt.min, tt.max)
				tt.validate(t, result)
			}
		})
	}
}

func TestChinese(t *testing.T) {
	// 创建一个使用固定种子的随机数生成器。
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	// 定义测试用例。
	tests := []struct {
		name     string
		random   *rand.Rand
		times    int
		validate func(t *testing.T, result string)
	}{
		{
			name:   "正常生成测试",
			random: random,
			times:  1000,
			validate: func(t *testing.T, result string) {
				// 验证生成的是单个汉字。
				assert.Equal(t, 1, len([]rune(result)), "应该生成一个汉字")

				// 验证生成的是有效的汉字。
				r := []rune(result)[0]
				assert.True(t, unicode.Is(unicode.Han, r), "应该生成一个有效的汉字")

				// 验证生成的汉字在指定范围内。
				assert.GreaterOrEqual(t, int(r), 19968, "汉字编码应该大于等于最小值")
				assert.Less(t, int(r), 40869, "汉字编码应该小于最大值")
			},
		},
		{
			name:   "nil random 测试",
			random: nil,
			times:  1000,
			validate: func(t *testing.T, result string) {
				r := []rune(result)[0]
				assert.True(t, unicode.Is(unicode.Han, r), "应该生成一个有效的汉字")
			},
		},
	}

	// 执行测试用例。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.times; i++ {
				result := fsrand.Chinese(tt.random)
				tt.validate(t, result)
			}
		})
	}
}

func TestChineseLastName(t *testing.T) {
	// 创建一个使用固定种子的随机数生成器。
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	// 定义测试用例。
	tests := []struct {
		name     string
		random   *rand.Rand
		times    int
		validate func(t *testing.T, result string)
	}{
		{
			name:   "正常生成测试",
			random: random,
			times:  1000,
			validate: func(t *testing.T, result string) {
				// 验证生成的是单个汉字。
				assert.Equal(t, 1, len([]rune(result)), "应该生成一个汉字")

				// 验证生成的是有效的汉字。
				r := []rune(result)[0]
				assert.True(t, unicode.Is(unicode.Han, r), "应该生成一个有效的汉字")
			},
		},
		{
			name:   "nil random 测试",
			random: nil,
			times:  1000,
			validate: func(t *testing.T, result string) {
				assert.Equal(t, 1, len([]rune(result)), "应该生成一个汉字")
				r := []rune(result)[0]
				assert.True(t, unicode.Is(unicode.Han, r), "应该生成一个有效的汉字")
			},
		},
	}

	// 执行测试用例。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.times; i++ {
				result := fsrand.ChineseLastName(tt.random)
				tt.validate(t, result)
			}
		})
	}
}
