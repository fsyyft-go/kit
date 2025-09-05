// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build gc && go1.25 && arm64

package goroutine

// stack 表示协程栈的结构。
// 该结构体与 Go 1.25 版本的运行时实现相匹配。
type stack struct { // nolint:unused
	// lo 栈的低地址边界。
	lo uintptr
	// hi 栈的高地址边界。
	hi uintptr
}

// gobuf 表示协程的执行上下文。
// 该结构体与 Go 1.25 版本的运行时实现相匹配。
type gobuf struct { // nolint:unused
	// sp 栈指针。
	sp uintptr
	// pc 程序计数器。
	pc uintptr
	// g 关联的 g 结构体指针。
	g uintptr
	// ctxt 上下文信息。
	ctxt uintptr
	// lr 链接寄存器。
	lr uintptr
	// bp 基址指针。
	bp uintptr
}

// g 表示一个 Go 协程的运行时结构。
// 该结构体与 Go 1.23 版本的运行时实现相匹配，
// 只关注 goid 字段的位置，其他字段仅用于确保正确的内存偏移。
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
	goid         int64   // 协程的唯一标识符
}
