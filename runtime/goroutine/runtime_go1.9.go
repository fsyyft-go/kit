// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build gc && go1.9 && !go1.23 && arm64
// +build gc,go1.9,!go1.23,arm64

package goroutine

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

type g struct {
	stack       stack
	stackguard0 uintptr
	stackguard1 uintptr

	_panic       uintptr
	_defer       uintptr
	m            uintptr
	sched        gobuf
	syscallsp    uintptr
	syscallpc    uintptr
	stktopsp     uintptr
	param        uintptr
	atomicstatus uint32
	stackLock    uint32
	goid         int64 // Here it is!
}
