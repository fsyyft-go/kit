// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package bytes 单元测试
//
// 测试文件设计思路：
// 1. 使用表格驱动测试方法，提高测试的可维护性和可读性。
// 2. 对GenerateNonce函数进行全面测试，包括正常场景、边界场景和异常场景。
// 3. 测试随机性特性，确保多次生成的随机字节不相同。
// 4. 使用stretchr/testify包进行断言，提高测试的清晰度和可读性。
//
// 使用方法：
// 1. 进入项目根目录。
// 2. 执行`go test ./bytes`命令运行测试。
// 3. 执行`go test -cover ./bytes`命令查看测试覆盖率。
package bytes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGenerateNonce 测试GenerateNonce函数的功能。
func TestGenerateNonce(t *testing.T) {
	// 定义测试用例表格。
	tests := []struct {
		name      string // 测试用例名称。
		length    int    // 输入的随机字节长度。
		wantLen   int    // 期望的随机字节长度。
		wantError bool   // 是否期望返回错误。
	}{
		{
			name:      "生成16字节随机数",
			length:    16,
			wantLen:   16,
			wantError: false,
		},
		{
			name:      "生成32字节随机数",
			length:    32,
			wantLen:   32,
			wantError: false,
		},
		{
			name:      "生成0字节随机数",
			length:    0,
			wantLen:   0,
			wantError: false,
		},
		{
			name:      "生成负数长度随机数",
			length:    -1,
			wantLen:   0,
			wantError: true,
		},
	}

	// 遍历测试用例。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 执行函数，获取结果。
			got, err := GenerateNonce(tt.length)

			// 验证错误情况。
			if tt.wantError {
				assert.Error(t, err, "当期望出错时，应当返回错误。")
			} else {
				assert.NoError(t, err, "当不期望出错时，不应当返回错误。")
				// 验证返回的字节切片长度。
				assert.Equal(t, tt.wantLen, len(got), "返回的随机字节长度应当与期望值相同。")
			}
		})
	}
}

// TestGenerateNonceRandomness 测试GenerateNonce函数的随机性。
func TestGenerateNonceRandomness(t *testing.T) {
	// 定义测试参数。
	length := 16
	iterations := 100

	// 存储生成的所有随机字节。
	allNonces := make([][]byte, iterations)

	// 生成多个随机字节切片。
	for i := 0; i < iterations; i++ {
		nonce, err := GenerateNonce(length)
		assert.NoError(t, err, "生成随机字节不应当返回错误。")
		assert.Equal(t, length, len(nonce), "返回的随机字节长度应当与输入长度相同。")
		allNonces[i] = nonce
	}

	// 验证随机性，确保不存在两个完全相同的随机字节切片。
	for i := 0; i < iterations; i++ {
		for j := i + 1; j < iterations; j++ {
			// 注意：理论上两个随机字节切片相同的可能性极小，但不为零。
			// 如果此测试偶尔失败，不一定表示函数存在问题。
			assert.NotEqualValues(t, allNonces[i], allNonces[j],
				"两个随机生成的字节切片不应当完全相同（除非极小概率事件发生）。")
		}
	}
}

// TestGenerateNoncePerformance 测试GenerateNonce函数的性能。
func BenchmarkGenerateNonce(b *testing.B) {
	// 测试16字节长度的随机数生成性能。
	b.Run("16字节", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = GenerateNonce(16)
		}
	})

	// 测试32字节长度的随机数生成性能。
	b.Run("32字节", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = GenerateNonce(32)
		}
	})

	// 测试64字节长度的随机数生成性能。
	b.Run("64字节", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = GenerateNonce(64)
		}
	})
}
