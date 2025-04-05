// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !windows && arm64

package goroutine

// getg 获取当前 G 结构体的指针。
// 此函数通过汇编实现，直接访问 TLS 获取 G 指针。
func getg() *g

// GetGoID 获取当前协程的 ID。
// 此函数在非 Windows 的 arm64 架构下通过 G 结构体获取协程 ID。
//
// 已废弃：请考虑使用其他替代方法获取协程 ID。
//
// 返回值：
//   - int64：返回当前协程的 ID。
func GetGoID() int64 {
	return getg().goid
}
