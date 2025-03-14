// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build amd64

package goroutine

// GetGoID 获取当前协程的 ID。
// 此函数在 amd64 架构下使用汇编实现，以获取更高效的性能。
//
// 已废弃：请考虑使用其他替代方法获取协程 ID。
//
// 返回值：
//   - int64：返回当前协程的 ID。
func GetGoID() int64
