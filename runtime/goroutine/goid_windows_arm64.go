// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build windows && arm64

package goroutine

// GetGoID 获取当前协程的 ID。
// 此函数在 Windows 的 arm64 架构下暂时使用慢速方法获取协程 ID。
//
// 已废弃：请考虑使用其他替代方法获取协程 ID。
//
// 返回值：
//   - int64：返回当前协程的 ID。
func GetGoID() int64 {
	// TODO 汇编的方法未实现，先使用开销较大的方法。
	return getGoIDSlow()
}
