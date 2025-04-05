// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !arm64 && !amd64

package goroutine

// GetGoID 获取当前协程的 ID。
// 此函数在非 arm64 和非 amd64 架构下使用较慢的方法获取协程 ID。
//
// 已废弃：请考虑使用特定平台的实现或其他替代方法。
//
// 返回值：
//   - int64：返回当前协程的 ID。
func GetGoID() int64 {
	return getGoIDSlow()
}
