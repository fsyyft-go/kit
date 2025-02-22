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
	goid         int64 // Here it is!
}
