// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build amd64

package goroutine

import (
	"runtime"
	"strings"
)

var (
	// offsetDict 记录各 Go 版本下 runtime.g 中 goid 字段的字节偏移。
	// 该表与本包维护的 runtime 内部结构定义和 amd64 汇编快速路径配套使用。
	offsetDict = map[string]int64{
		"go1.4":  128,
		"go1.5":  184,
		"go1.6":  192,
		"go1.7":  192,
		"go1.8":  192,
		"go1.9":  152,
		"go1.10": 152,
		"go1.11": 152,
		"go1.12": 152,
		"go1.13": 152,
		"go1.14": 152,
		"go1.15": 152,
		"go1.16": 152,
		"go1.17": 152,
		"go1.18": 152,
		"go1.19": 152,
		"go1.20": 152,
		"go1.21": 152,
		"go1.22": 152,
		"go1.23": 160, // 多了 syscallbp 8 个字节。
		"go1.24": 160,
		"go1.25": 152, // 少了 gobuf.ret 8 个字节。
		"go1.26": 152,
	}

	// offset 缓存当前 Go 运行时版本对应的 goid 偏移量。
	// 若 runtime.Version() 未命中 offsetDict，结果会是 0；升级 Go 版本后必须先补齐偏移表并重新验证。
	offset = func() int64 {
		ver := strings.Join(strings.Split(runtime.Version(), ".")[:2], ".")
		return offsetDict[ver]
	}()
)

// Offset 返回 amd64 快速路径当前使用的 runtime.g.goid 字段偏移量。
//
// 返回 0 通常表示当前 Go 版本尚未写入 offsetDict。
//
// 返回值：
//   - int64：当前 Go 版本对应的 runtime.g.goid 字段偏移量；未命中 offsetDict 时返回 0。
func Offset() int64 {
	return offset
}
