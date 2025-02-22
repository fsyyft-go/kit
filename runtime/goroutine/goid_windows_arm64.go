// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build windows && arm64

package goroutine

// Deprecated: GetGoID  获取 goroutine ID。
func GetGoID() int64 {
	// TODO 汇编的方法未实现，先使用开销较大的。
	return getGoIDSlow()
}
