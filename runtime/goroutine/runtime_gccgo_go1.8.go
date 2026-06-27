// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !gc && gccgo && go1.8 && arm64
// +build !gc,gccgo,go1.8,arm64

package goroutine

// https://github.com/gcc-mirror/gcc/blob/releases/gcc-7/libgo/go/runtime/runtime2.go#L329-L354

// g 镜像 gccgo 1.8 arm64 的 runtime.g 最小前缀布局。
// SAFETY: 本结构仅用于让 goid 字段偏移与目标运行时保持一致，字段顺序不得随意改动。
type g struct {
	_panic       uintptr
	_defer       uintptr
	m            uintptr
	syscallsp    uintptr
	syscallpc    uintptr
	param        uintptr
	atomicstatus uint32
	goid         int64 // goroutine 的唯一标识符
}
