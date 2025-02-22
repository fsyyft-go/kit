// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !gc && gccgo && go1.8 && arm64
// +build !gc,gccgo,go1.8,arm64

package goroutine

// https://github.com/gcc-mirror/gcc/blob/releases/gcc-7/libgo/go/runtime/runtime2.go#L329-L354

type g struct {
	_panic       uintptr
	_defer       uintptr
	m            uintptr
	syscallsp    uintptr
	syscallpc    uintptr
	param        uintptr
	atomicstatus uint32
	goid         int64 // Here it is!
}
