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

// GetGoIDSlow 获取当前协程的 ID，当无法从 GetGoID 获取协程 ID 时使用此方法。
// 该方法通过获取协程的堆栈信息，然后解析堆栈信息来提取协程 ID。
func GetGoIDSlow() int64 {
	return getGoIDSlow()
}
