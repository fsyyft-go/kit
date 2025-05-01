// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information。

// Package sha_test 提供了对 sha 包功能的测试。
//
// 测试设计思路：
// 1. 采用表格驱动测试，覆盖正常情况、边界情况和特殊情况；
// 2. 对 SHA256HashString 和 SHA256HashStringWithoutError 两个函数分别进行测试；
// 3. 使用 stretchr/testify 包进行断言验证；
// 4. 保证测试覆盖率尽可能高。
//
// 使用方法：
// 1. 在项目根目录执行 `go test github.com/fsyyft-go/kit/crypto/sha` 运行测试；
// 2. 添加 `-v` 参数可查看详细测试输出；
// 3. 添加 `-cover` 参数可查看测试覆盖率。
package sha_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	kitsha "github.com/fsyyft-go/kit/crypto/sha"
)

// TestSHA256HashString 测试 SHA256HashString 函数的各种情况。
func TestSHA256HashString(t *testing.T) {
	// 定义测试用例表格，包含输入值和期望的输出值。
	tests := []struct {
		name        string // 测试用例名称
		input       string // 输入字符串
		expected    string // 期望的 SHA256 哈希值
		expectError bool   // 是否期望出现错误
	}{
		{
			name:        "空字符串测试",
			input:       "",
			expected:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expectError: false,
		},
		{
			name:        "简单ASCII字符串测试",
			input:       "hello world",
			expected:    "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			expectError: false,
		},
		{
			name:        "数字字符串测试",
			input:       "12345",
			expected:    "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
			expectError: false,
		},
		{
			name:        "中文字符串测试",
			input:       "你好，世界",
			expected:    "46932f1e6ea5216e77f58b1908d72ec9322ed129318c6d4bd4450b5eaab9d7e7",
			expectError: false,
		},
		{
			name:        "特殊字符测试",
			input:       "!@#$%^&*()_+",
			expected:    "36d3e1bc65f8b67935ae60f542abef3e55c5bbbd547854966400cc4f022566cb",
			expectError: false,
		},
		{
			name:        "长文本测试",
			input:       "这是一段较长的文本，用于测试SHA256哈希函数对长文本的处理能力。SHA256会生成固定长度的哈希值，无论输入多长。",
			expected:    "d9a75fbb24d37240199f1d719f497de5b5028fe611bdbff0fc50a997d0f2b48e",
			expectError: false,
		},
		{
			name:        "大数据量测试",
			input:       string(make([]byte, 1024*1024)), // 1MB 的数据
			expected:    "30e14955ebf1352266dc2ff8067e68104607e750abb9d3b36582b8af909fcb58",
			expectError: false,
		},
	}

	// 遍历测试用例表格，执行测试。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测试函数。
			result, err := kitsha.SHA256HashString(tt.input)

			// 断言测试结果。
			if tt.expectError {
				assert.Error(t, err, "期望返回错误，但实际无错误发生。")
			} else {
				assert.NoError(t, err, "不期望错误，但实际发生了错误: %v", err)
				assert.Equal(t, tt.expected, result, "SHA256哈希值不匹配，期望: %s, 实际: %s", tt.expected, result)
			}
		})
	}
}

// TestSHA256HashStringWithoutError 测试 SHA256HashStringWithoutError 函数的各种情况。
func TestSHA256HashStringWithoutError(t *testing.T) {
	// 定义测试用例表格，包含输入值和期望的输出值。
	tests := []struct {
		name     string // 测试用例名称
		input    string // 输入字符串
		expected string // 期望的 SHA256 哈希值
	}{
		{
			name:     "空字符串测试",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "简单ASCII字符串测试",
			input:    "hello world",
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:     "数字字符串测试",
			input:    "12345",
			expected: "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
		},
		{
			name:     "中文字符串测试",
			input:    "你好，世界",
			expected: "46932f1e6ea5216e77f58b1908d72ec9322ed129318c6d4bd4450b5eaab9d7e7",
		},
		{
			name:     "特殊字符测试",
			input:    "!@#$%^&*()_+",
			expected: "36d3e1bc65f8b67935ae60f542abef3e55c5bbbd547854966400cc4f022566cb",
		},
		{
			name:     "长文本测试",
			input:    "这是一段较长的文本，用于测试SHA256哈希函数对长文本的处理能力。SHA256会生成固定长度的哈希值，无论输入多长。",
			expected: "d9a75fbb24d37240199f1d719f497de5b5028fe611bdbff0fc50a997d0f2b48e",
		},
		{
			name:     "大数据量测试",
			input:    string(make([]byte, 1024*1024)), // 1MB 的数据
			expected: "30e14955ebf1352266dc2ff8067e68104607e750abb9d3b36582b8af909fcb58",
		},
	}

	// 遍历测试用例表格，执行测试。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测试函数。
			result := kitsha.SHA256HashStringWithoutError(tt.input)

			// 断言测试结果。
			assert.Equal(t, tt.expected, result, "SHA256哈希值不匹配，期望: %s, 实际: %s", tt.expected, result)
		})
	}
}

// TestSHA256HashString_ErrorCase 记录错误分支的测试思路。
func TestSHA256HashString_ErrorCase(t *testing.T) {
	// 注释：由于标准库 sha256.New().Write 不会对字符串输入返回错误，
	// 这里仅记录如何处理理论上的错误分支。
	t.Run("错误处理示例", func(t *testing.T) {
		t.Log("sha256.New().Write 理论上不会返回错误，若发生错误应返回空字符串和该错误")
	})
}

// TestSHA256HashStringWithoutError_ErrorHandling 测试 SHA256HashStringWithoutError 对错误的处理。
func TestSHA256HashStringWithoutError_ErrorHandling(t *testing.T) {
	// 注释：SHA256HashStringWithoutError 会忽略错误，若底层发生错误应返回空字符串。
	t.Run("错误处理测试", func(t *testing.T) {
		result := kitsha.SHA256HashStringWithoutError("test")
		assert.Equal(t, "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", result, "测试字符串的SHA256哈希值不匹配")
		t.Log("SHA256HashStringWithoutError 应该在任何情况下都不返回错误，即使计算过程中出现问题")
	})
}

// 性能测试数据集，用于测试不同大小和类型的输入数据。
var benchmarkData = []struct {
	name   string
	input  string
	repeat int
}{
	{"空字符串", "", 1},
	{"短ASCII", "hello world", 1},
	{"短ASCII重复", "hello world", 100},
	{"中文字符串", "你好，世界", 1},
	{"中文字符串重复", "你好，世界", 100},
	{"数字", "12345", 1},
	{"特殊字符", "!@#$%^&*()_+", 1},
	{"1KB数据", string(make([]byte, 1024)), 1},
	{"10KB数据", string(make([]byte, 10*1024)), 1},
	{"100KB数据", string(make([]byte, 100*1024)), 1},
}

// BenchmarkSHA256HashStringVariousData 对 SHA256HashString 函数进行多种数据的基准测试。
func BenchmarkSHA256HashStringVariousData(b *testing.B) {
	for _, bm := range benchmarkData {
		b.Run(bm.name, func(b *testing.B) {
			input := ""
			for i := 0; i < bm.repeat; i++ {
				input += bm.input
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = kitsha.SHA256HashString(input)
			}
		})
	}
}

// BenchmarkSHA256HashStringWithoutErrorVariousData 对 SHA256HashStringWithoutError 函数进行多种数据的基准测试。
func BenchmarkSHA256HashStringWithoutErrorVariousData(b *testing.B) {
	for _, bm := range benchmarkData {
		b.Run(bm.name, func(b *testing.B) {
			input := ""
			for i := 0; i < bm.repeat; i++ {
				input += bm.input
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = kitsha.SHA256HashStringWithoutError(input)
			}
		})
	}
}

// BenchmarkSHA256HashStringParallel 对 SHA256HashString 函数进行并行基准测试。
func BenchmarkSHA256HashStringParallel(b *testing.B) {
	input := string(make([]byte, 100*1024)) // 100KB
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = kitsha.SHA256HashString(input)
		}
	})
}

// BenchmarkSHA256HashStringWithoutErrorParallel 对 SHA256HashStringWithoutError 函数进行并行基准测试。
func BenchmarkSHA256HashStringWithoutErrorParallel(b *testing.B) {
	input := string(make([]byte, 100*1024)) // 100KB
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = kitsha.SHA256HashStringWithoutError(input)
		}
	})
}

// TestSHA1HashString 测试 SHA1HashString 函数的各种情况。
func TestSHA1HashString(t *testing.T) {
	// 定义测试用例表格，包含输入值和期望的输出值。
	tests := []struct {
		name        string // 测试用例名称
		input       string // 输入字符串
		expected    string // 期望的 SHA1 哈希值
		expectError bool   // 是否期望出现错误
	}{
		{
			name:        "空字符串测试",
			input:       "",
			expected:    "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			expectError: false,
		},
		{
			name:        "简单ASCII字符串测试",
			input:       "hello world",
			expected:    "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
			expectError: false,
		},
		{
			name:        "数字字符串测试",
			input:       "12345",
			expected:    "8cb2237d0679ca88db6464eac60da96345513964",
			expectError: false,
		},
		{
			name:        "中文字符串测试",
			input:       "你好，世界",
			expected:    "3becb03b015ed48050611c8d7afe4b88f70d5a20",
			expectError: false,
		},
		{
			name:        "特殊字符测试",
			input:       "!@#$%^&*()_+",
			expected:    "d0b9abafaf5a393954f53e47715c833f0c18075d",
			expectError: false,
		},
		{
			name:        "长文本测试",
			input:       "这是一段较长的文本，用于测试SHA256哈希函数对长文本的处理能力。SHA256会生成固定长度的哈希值，无论输入多长。",
			expected:    "ddff78fe3dc4b7bbad08f1b6e3ee15b2a268c572",
			expectError: false,
		},
		{
			name:        "大数据量测试",
			input:       string(make([]byte, 1024*1024)), // 1MB 的数据
			expected:    "3b71f43ff30f4b15b5cd85dd9e95ebc7e84eb5a3",
			expectError: false,
		},
	}

	// 遍历测试用例表格，执行测试。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测试函数。
			result, err := kitsha.SHA1HashString(tt.input)

			// 断言测试结果。
			if tt.expectError {
				assert.Error(t, err, "期望返回错误，但实际无错误发生。")
			} else {
				assert.NoError(t, err, "不期望错误，但实际发生了错误: %v", err)
				assert.Equal(t, tt.expected, result, "SHA1哈希值不匹配，期望: %s, 实际: %s", tt.expected, result)
			}
		})
	}
}

// TestSHA1HashStringWithoutError 测试 SHA1HashStringWithoutError 函数的各种情况。
func TestSHA1HashStringWithoutError(t *testing.T) {
	// 定义测试用例表格，包含输入值和期望的输出值。
	tests := []struct {
		name     string // 测试用例名称
		input    string // 输入字符串
		expected string // 期望的 SHA1 哈希值
	}{
		{
			name:     "空字符串测试",
			input:    "",
			expected: "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		},
		{
			name:     "简单ASCII字符串测试",
			input:    "hello world",
			expected: "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
		},
		{
			name:     "数字字符串测试",
			input:    "12345",
			expected: "8cb2237d0679ca88db6464eac60da96345513964",
		},
		{
			name:     "中文字符串测试",
			input:    "你好，世界",
			expected: "3becb03b015ed48050611c8d7afe4b88f70d5a20",
		},
		{
			name:     "特殊字符测试",
			input:    "!@#$%^&*()_+",
			expected: "d0b9abafaf5a393954f53e47715c833f0c18075d",
		},
		{
			name:     "长文本测试",
			input:    "这是一段较长的文本，用于测试SHA256哈希函数对长文本的处理能力。SHA256会生成固定长度的哈希值，无论输入多长。",
			expected: "ddff78fe3dc4b7bbad08f1b6e3ee15b2a268c572",
		},
		{
			name:     "大数据量测试",
			input:    string(make([]byte, 1024*1024)), // 1MB 的数据
			expected: "3b71f43ff30f4b15b5cd85dd9e95ebc7e84eb5a3",
		},
	}

	// 遍历测试用例表格，执行测试。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测试函数。
			result := kitsha.SHA1HashStringWithoutError(tt.input)

			// 断言测试结果。
			assert.Equal(t, tt.expected, result, "SHA1哈希值不匹配，期望: %s, 实际: %s", tt.expected, result)
		})
	}
}
