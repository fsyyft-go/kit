// Copyright (c) 2025 fsyyft-go
//
// Licensed under the MIT License (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://github.com/fsyyft-go/kit/blob/main/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
