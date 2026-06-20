// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !arm64 && !amd64

package goroutine

// GetGoID 返回当前 goroutine 的 ID。
//
// 在未提供架构专用快速路径的平台上，本实现退回到基于 runtime.Stack 的慢速解析路径。
//
// 返回值：
//   - int64：当前 goroutine 的 ID。
func GetGoID() int64 {
	return getGoIDSlow()
}
