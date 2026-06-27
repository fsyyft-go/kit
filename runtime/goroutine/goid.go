// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !arm64 && !amd64

package goroutine

// GetGoID 返回当前 goroutine 的 ID。
//
// 该实现仅在未提供 amd64 或 arm64 快速路径的平台上编译，并退回到基于
// runtime.Stack 的慢速解析路径。
//
// 参数：无。
//
// 返回：
//   - int64：当前 goroutine 的 ID。
func GetGoID() int64 {
	return getGoIDSlow()
}
