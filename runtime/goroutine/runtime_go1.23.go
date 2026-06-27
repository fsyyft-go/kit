// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build gc && go1.23 && !go1.25 && arm64

package goroutine

// stack 镜像 Go 1.23 到 Go 1.24 arm64 的 runtime.stack 最小前缀布局。
type stack struct { // nolint:unused
	// lo 栈的低地址边界。
	lo uintptr
	// hi 栈的高地址边界。
	hi uintptr
}

// gobuf 镜像 Go 1.23 到 Go 1.24 arm64 的 runtime.gobuf 最小前缀布局。
type gobuf struct { // nolint:unused
	// sp 栈指针。
	sp uintptr
	// pc 程序计数器。
	pc uintptr
	// g 关联的 g 结构体指针。
	g uintptr
	// ctxt 上下文信息。
	ctxt uintptr
	// ret 返回值。
	ret uintptr
	// lr 链接寄存器。
	lr uintptr
	// bp 基址指针。
	bp uintptr
}

// g 镜像 Go 1.23 到 Go 1.24 arm64 的 runtime.g 最小前缀布局。
// SAFETY: 本结构仅用于让 goid 字段偏移与目标运行时保持一致，字段顺序不得随意改动。
type g struct {
	stack       stack   // nolint:unused // 协程的栈
	stackguard0 uintptr // nolint:unused // 栈溢出检测，快速路径
	stackguard1 uintptr // nolint:unused // 栈溢出检测，慢速路径

	_panic       uintptr // nolint:unused // 内部 panic 记录
	_defer       uintptr // nolint:unused // 内部 defer 记录
	m            uintptr // nolint:unused // 当前关联的 M
	sched        gobuf   // nolint:unused // 调度信息
	syscallsp    uintptr // nolint:unused // 系统调用时的栈指针
	syscallpc    uintptr // nolint:unused // 系统调用时的程序计数器
	syscallbp    uintptr // nolint:unused // 系统调用时的基址指针
	stktopsp     uintptr // nolint:unused // 预留的栈顶指针
	param        uintptr // nolint:unused // 唤醒参数
	atomicstatus uint32  // nolint:unused // goroutine 状态
	stackLock    uint32  // nolint:unused // 栈锁
	goid         int64   // goroutine 的唯一标识符
}
