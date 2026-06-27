// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build go1.5 && !go1.6 && arm64
// +build go1.5,!go1.6,arm64

package goroutine

// Just enough of the structs from runtime/runtime2.go to get the offset to goid.
// See https://github.com/golang/go/blob/release-branch.go1.5/src/runtime/runtime2.go

type stack struct {
	lo uintptr
	hi uintptr
}

type gobuf struct {
	sp   uintptr
	pc   uintptr
	g    uintptr
	ctxt uintptr
	ret  uintptr
	lr   uintptr
	bp   uintptr
}

// g 镜像 Go 1.5 arm64 的 runtime.g 最小前缀布局。
// SAFETY: 本结构仅用于让 goid 字段偏移与目标运行时保持一致，字段顺序不得随意改动。
type g struct {
	stack       stack
	stackguard0 uintptr
	stackguard1 uintptr

	_panic       uintptr
	_defer       uintptr
	m            uintptr
	stackAlloc   uintptr
	sched        gobuf
	syscallsp    uintptr
	syscallpc    uintptr
	stkbar       []uintptr
	stkbarPos    uintptr
	param        uintptr
	atomicstatus uint32
	stackLock    uint32
	goid         int64 // goroutine 的唯一标识符
}
