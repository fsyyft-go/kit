// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build gc && go1.23 && arm64

package goroutine

type stack struct { // nolint:unused
	lo uintptr
	hi uintptr
}

type gobuf struct { // nolint:unused
	sp   uintptr
	pc   uintptr
	g    uintptr
	ctxt uintptr
	ret  uintptr
	lr   uintptr
	bp   uintptr
}

type g struct {
	stack       stack   // nolint:unused
	stackguard0 uintptr // nolint:unused
	stackguard1 uintptr // nolint:unused

	_panic       uintptr // nolint:unused
	_defer       uintptr // nolint:unused
	m            uintptr // nolint:unused
	sched        gobuf   // nolint:unused
	syscallsp    uintptr // nolint:unused
	syscallpc    uintptr // nolint:unused
	syscallbp    uintptr // nolint:unused
	stktopsp     uintptr // nolint:unused
	param        uintptr // nolint:unused
	atomicstatus uint32  // nolint:unused
	stackLock    uint32  // nolint:unused
	goid         int64   // Here it is!
}
