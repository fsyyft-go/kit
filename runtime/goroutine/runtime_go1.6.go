// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build gc && go1.6 && !go1.9 && arm64
// +build gc,go1.6,!go1.9,arm64

package goroutine

// 这里包含了来自 runtime/runtime2.go 的部分结构体定义，用于获取 goid 的偏移量。
// 参考：https://github.com/golang/go/blob/release-branch.go1.6/src/runtime/runtime2.go

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

// g 结构体包含了 goroutine 的运行时信息。
type g struct {
	stack       stack   // goroutine 的栈空间
	stackguard0 uintptr // 栈溢出检查的边界值
	stackguard1 uintptr // 栈溢出检查的第二边界值

	_panic       uintptr   // 内部 panic 记录
	_defer       uintptr   // 内部 defer 记录
	m            uintptr   // 当前关联的 M（系统线程）
	stackAlloc   uintptr   // 栈的分配大小
	sched        gobuf     // goroutine 调度信息
	syscallsp    uintptr   // 系统调用时的栈指针
	syscallpc    uintptr   // 系统调用时的程序计数器
	stkbar       []uintptr // 栈障碍记录
	stkbarPos    uintptr   // 栈障碍位置
	stktopsp     uintptr   // 期望的栈指针位置
	param        uintptr   // wakeup 参数
	atomicstatus uint32    // goroutine 的状态
	stackLock    uint32    // 栈的锁状态
	goid         int64     // goroutine 的唯一标识符
}
