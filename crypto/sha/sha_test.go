// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information。

// Package sha_test 提供 sha 包字符串哈希函数的外部行为测试。
//
// 测试用例使用标准 SHA1/SHA256 摘要向量，覆盖空字符串、ASCII、数字、中文、特殊字符、
// 多行 Unicode 文本、长文本和大体量输入，确保公开函数返回稳定且兼容标准库的十六进制结果。
package sha_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kitsha "github.com/fsyyft-go/kit/crypto/sha"
)

// hashStringFixture 描述字符串哈希函数共享的标准输入和预期摘要。
type hashStringFixture struct {
	name        string
	description string
	giveSource  string
	wantSHA256  string
	wantSHA1    string
}

// standardHashStringFixtures 返回 SHA1 和 SHA256 字符串哈希测试共享的标准向量。
//
// 该辅助函数集中维护由独立工具计算得到的固定摘要值，避免各测试重复构造输入并确保
// SHA1HashString、SHA1HashStringWithoutError、SHA256HashString 和
// SHA256HashStringWithoutError 覆盖同一组行为语义。
//
// 返回：
//   - []hashStringFixture: 可直接用于字符串哈希函数表驱动测试的输入与预期摘要。
func standardHashStringFixtures() []hashStringFixture {
	return []hashStringFixture{
		{
			name:        "boundary/empty-string",
			description: "验证空字符串输入按照标准哈希定义生成空消息摘要。",
			giveSource:  "",
			wantSHA256:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantSHA1:    "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		},
		{
			name:        "success/ascii",
			description: "验证普通 ASCII 文本按照 UTF-8 字节序列计算标准摘要。",
			giveSource:  "hello world",
			wantSHA256:  "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
			wantSHA1:    "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
		},
		{
			name:        "success/digits",
			description: "验证数字字符串作为文本输入时不会被数值化处理。",
			giveSource:  "12345",
			wantSHA256:  "5994471abb01112afcc18159f6cc74b4f511b99806da59b3caf5a9c173cacfc5",
			wantSHA1:    "8cb2237d0679ca88db6464eac60da96345513964",
		},
		{
			name:        "success/chinese",
			description: "验证中文标点和汉字按照原始 UTF-8 字节序列计算摘要。",
			giveSource:  "你好，世界",
			wantSHA256:  "46932f1e6ea5216e77f58b1908d72ec9322ed129318c6d4bd4450b5eaab9d7e7",
			wantSHA1:    "3becb03b015ed48050611c8d7afe4b88f70d5a20",
		},
		{
			name:        "success/special-characters",
			description: "验证常见 ASCII 符号保持原样参与摘要计算。",
			giveSource:  "!@#$%^&*()_+",
			wantSHA256:  "36d3e1bc65f8b67935ae60f542abef3e55c5bbbd547854966400cc4f022566cb",
			wantSHA1:    "d0b9abafaf5a393954f53e47715c833f0c18075d",
		},
		{
			name:        "success/multiline-special-unicode",
			description: "验证换行、制表符、符号和 Unicode 字符混合时按完整字符串计算摘要。",
			giveSource:  "第一行\nSecond line\tTabbed\nSymbols: !@#$%^&*()_+\nEmoji: 🚀",
			wantSHA256:  "e5c85bfd7cc43c889e5b88ce300224b2fdd323b42c0798e179d4d09b1cf69b67",
			wantSHA1:    "bcb118dc8339622306be607540f29dccdde46f20",
		},
		{
			name:        "success/long-chinese-text",
			description: "验证较长中文文本输入返回固定长度且标准一致的十六进制摘要。",
			giveSource:  "这是一段较长的文本，用于测试SHA256哈希函数对长文本的处理能力。SHA256会生成固定长度的哈希值，无论输入多长。",
			wantSHA256:  "d9a75fbb24d37240199f1d719f497de5b5028fe611bdbff0fc50a997d0f2b48e",
			wantSHA1:    "ddff78fe3dc4b7bbad08f1b6e3ee15b2a268c572",
		},
		{
			name:        "success/repeated-long-text",
			description: "验证由重复非 ASCII 片段组成的长文本不会因长度增加改变编码语义。",
			giveSource:  strings.Repeat("Go语言测试", 256),
			wantSHA256:  "f7b8d5cb0abf061b0b2917e0a6540d97b30bc5dd3edc9738d2101ab6bf9f3464",
			wantSHA1:    "25aa2313e13ce858b3e7716923e4ce704e836ddc",
		},
		{
			name:        "boundary/one-mebibyte-zero-bytes",
			description: "验证包含 1MiB 零字节的大体量字符串仍按完整字节序列计算摘要。",
			giveSource:  string(make([]byte, 1024*1024)),
			wantSHA256:  "30e14955ebf1352266dc2ff8067e68104607e750abb9d3b36582b8af909fcb58",
			wantSHA1:    "3b71f43ff30f4b15b5cd85dd9e95ebc7e84eb5a3",
		},
	}
}

// TestSHA256HashString_StandardVectors 验证 SHA256HashString 返回标准 SHA256 十六进制摘要。
//
// 该测试通过表驱动用例覆盖空值、ASCII、Unicode、特殊字符、长文本和大体量输入，
// 并断言标准库 hash.Write 在这些真实字符串输入下不返回错误。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestSHA256HashString_StandardVectors(t *testing.T) {
	tests := standardHashStringFixtures()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := kitsha.SHA256HashString(tt.giveSource)

			require.NoError(t, err)
			assert.Equal(t, tt.wantSHA256, got)
		})
	}
}

// TestSHA256HashStringWithoutError_StandardVectors 验证 SHA256HashStringWithoutError 返回标准 SHA256 摘要。
//
// 该测试通过与错误返回版本一致的标准向量，确保无错误返回包装函数不会改变字符串哈希结果。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestSHA256HashStringWithoutError_StandardVectors(t *testing.T) {
	tests := standardHashStringFixtures()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := kitsha.SHA256HashStringWithoutError(tt.giveSource)

			assert.Equal(t, tt.wantSHA256, got)
		})
	}
}

// TestSHA1HashString_StandardVectors 验证 SHA1HashString 返回标准 SHA1 十六进制摘要。
//
// 该测试通过表驱动用例覆盖空值、ASCII、Unicode、特殊字符、长文本和大体量输入，
// 并断言标准库 hash.Write 在这些真实字符串输入下不返回错误。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestSHA1HashString_StandardVectors(t *testing.T) {
	tests := standardHashStringFixtures()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := kitsha.SHA1HashString(tt.giveSource)

			require.NoError(t, err)
			assert.Equal(t, tt.wantSHA1, got)
		})
	}
}

// TestSHA1HashStringWithoutError_StandardVectors 验证 SHA1HashStringWithoutError 返回标准 SHA1 摘要。
//
// 该测试通过与错误返回版本一致的标准向量，确保无错误返回包装函数不会改变字符串哈希结果。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestSHA1HashStringWithoutError_StandardVectors(t *testing.T) {
	tests := standardHashStringFixtures()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := kitsha.SHA1HashStringWithoutError(tt.giveSource)

			assert.Equal(t, tt.wantSHA1, got)
		})
	}
}

// benchmarkData 定义字符串哈希函数基准测试使用的输入规模。
var benchmarkData = []struct {
	name       string
	giveSource string
}{
	{name: "empty-string", giveSource: ""},
	{name: "short-ascii", giveSource: "hello world"},
	{name: "short-ascii-repeated", giveSource: strings.Repeat("hello world", 100)},
	{name: "chinese", giveSource: "你好，世界"},
	{name: "chinese-repeated", giveSource: strings.Repeat("你好，世界", 100)},
	{name: "digits", giveSource: "12345"},
	{name: "special-characters", giveSource: "!@#$%^&*()_+"},
	{name: "one-kib-zero-bytes", giveSource: string(make([]byte, 1024))},
	{name: "ten-kib-zero-bytes", giveSource: string(make([]byte, 10*1024))},
	{name: "one-hundred-kib-zero-bytes", giveSource: string(make([]byte, 100*1024))},
}

// BenchmarkSHA256HashStringVariousData 度量 SHA256HashString 在不同输入规模下的性能。
//
// 参数：
//   - b: 基准测试上下文，用于运行子基准和记录迭代性能。
func BenchmarkSHA256HashStringVariousData(b *testing.B) {
	for _, bm := range benchmarkData {
		bm := bm
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = kitsha.SHA256HashString(bm.giveSource)
			}
		})
	}
}

// BenchmarkSHA256HashStringWithoutErrorVariousData 度量 SHA256HashStringWithoutError 在不同输入规模下的性能。
//
// 参数：
//   - b: 基准测试上下文，用于运行子基准和记录迭代性能。
func BenchmarkSHA256HashStringWithoutErrorVariousData(b *testing.B) {
	for _, bm := range benchmarkData {
		bm := bm
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = kitsha.SHA256HashStringWithoutError(bm.giveSource)
			}
		})
	}
}

// BenchmarkSHA256HashStringParallel 度量 SHA256HashString 在并行调用场景下的性能。
//
// 参数：
//   - b: 基准测试上下文，用于运行并行基准和记录迭代性能。
func BenchmarkSHA256HashStringParallel(b *testing.B) {
	input := string(make([]byte, 100*1024))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = kitsha.SHA256HashString(input)
		}
	})
}

// BenchmarkSHA256HashStringWithoutErrorParallel 度量 SHA256HashStringWithoutError 在并行调用场景下的性能。
//
// 参数：
//   - b: 基准测试上下文，用于运行并行基准和记录迭代性能。
func BenchmarkSHA256HashStringWithoutErrorParallel(b *testing.B) {
	input := string(make([]byte, 100*1024))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = kitsha.SHA256HashStringWithoutError(input)
		}
	})
}
