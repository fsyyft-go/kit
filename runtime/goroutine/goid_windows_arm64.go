// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build windows && arm64

package goroutine

// GetGoID 返回当前 goroutine 的 ID。
//
// windows/arm64 目前未提供汇编快速路径，因此该实现固定退回到基于
// runtime.Stack 的慢速解析路径。
//
// 参数：无。
//
// 返回：
//   - int64：当前 goroutine 的 ID。
func GetGoID() int64 {
	// TODO(fsyyft): 为 windows/arm64 补齐汇编快速路径，并与对应的 runtime.g 布局一起验证。
	return getGoIDSlow()
}
