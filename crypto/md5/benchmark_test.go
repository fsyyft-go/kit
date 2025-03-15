// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// md5_test 基准测试包，用于评估 md5 包中函数的性能。
package md5_test

import (
	"testing"

	"github.com/fsyyft-go/kit/crypto/md5"
)

// 基准测试数据集，用于测试不同大小和类型的输入数据。
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

// BenchmarkHashStringVariousData 对 HashString 函数进行多种数据的基准测试。
func BenchmarkHashStringVariousData(b *testing.B) {
	for _, bm := range benchmarkData {
		// 为每个数据集创建子基准测试。
		b.Run(bm.name, func(b *testing.B) {
			input := ""
			// 根据 repeat 值重复输入字符串，构建实际测试数据。
			for i := 0; i < bm.repeat; i++ {
				input += bm.input
			}

			// 重置定时器，避免初始化代码影响基准测试结果。
			b.ResetTimer()

			// 执行 b.N 次测试，b.N 由测试框架自动确定。
			for i := 0; i < b.N; i++ {
				_, _ = md5.HashString(input)
			}
		})
	}
}

// BenchmarkHashStringWithoutErrorVariousData 对 HashStringWithoutError 函数进行多种数据的基准测试。
func BenchmarkHashStringWithoutErrorVariousData(b *testing.B) {
	for _, bm := range benchmarkData {
		// 为每个数据集创建子基准测试。
		b.Run(bm.name, func(b *testing.B) {
			input := ""
			// 根据 repeat 值重复输入字符串，构建实际测试数据。
			for i := 0; i < bm.repeat; i++ {
				input += bm.input
			}

			// 重置定时器，避免初始化代码影响基准测试结果。
			b.ResetTimer()

			// 执行 b.N 次测试，b.N 由测试框架自动确定。
			for i := 0; i < b.N; i++ {
				_ = md5.HashStringWithoutError(input)
			}
		})
	}
}

// BenchmarkHashStringParallel 对 HashString 函数进行并行基准测试。
func BenchmarkHashStringParallel(b *testing.B) {
	// 对较大数据进行并行基准测试。
	input := string(make([]byte, 100*1024)) // 100KB

	// 重置定时器，避免初始化代码影响基准测试结果。
	b.ResetTimer()

	// 使用 RunParallel 进行并行基准测试。
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = md5.HashString(input)
		}
	})
}

// BenchmarkHashStringWithoutErrorParallel 对 HashStringWithoutError 函数进行并行基准测试。
func BenchmarkHashStringWithoutErrorParallel(b *testing.B) {
	// 对较大数据进行并行基准测试。
	input := string(make([]byte, 100*1024)) // 100KB

	// 重置定时器，避免初始化代码影响基准测试结果。
	b.ResetTimer()

	// 使用 RunParallel 进行并行基准测试。
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = md5.HashStringWithoutError(input)
		}
	})
}
