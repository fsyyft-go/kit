// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package goroutine

import (
	"bytes"
	"runtime"
	"strconv"
)

const (
	// initialBufferSize 定义慢速 goroutine ID 解析使用的初始栈缓冲区大小。
	initialBufferSize = 128
)

// getGoIDSlow 通过解析当前 goroutine 的栈头信息提取 goroutine ID。
//
// 该慢速路径依赖 runtime.Stack 输出以 "goroutine <id> " 开头；仅在未提供
// 架构专用快速路径的平台上，或调用方显式使用 GetGoIDSlow 时使用。
//
// 参数：无。
//
// 返回：
//   - int64：当前 goroutine 的 ID。
func getGoIDSlow() int64 {
	var buf [initialBufferSize]byte
	stackBytes := buf[:]

	// 仅抓取当前 goroutine 的栈头信息，足以解析 ID，同时避免构造完整堆栈的额外开销。
	stackBytes = stackBytes[:runtime.Stack(stackBytes, false)]

	return extractGID(stackBytes)
}

// extractGID 从 runtime.Stack 返回的栈头字节串中提取 goroutine ID。
//
// 参数：
//   - s：来源于 runtime.Stack 首行输出、并以 "goroutine <id> " 开头的字节切片。
//
// 返回：
//   - int64：解析出的 goroutine ID。
func extractGID(s []byte) int64 {
	s = s[len("goroutine "):]
	s = s[:bytes.IndexByte(s, ' ')]
	gid, _ := strconv.ParseInt(string(s), 10, 64)
	return gid
}

// GetGoIDSlow 通过慢速解析路径返回当前 goroutine 的 ID。
//
// 该函数始终使用 runtime.Stack 解析首行文本提取 goroutine ID，不依赖架构专用
// 汇编快速路径，适合在需要显式回退行为时使用。
//
// 参数：无。
//
// 返回：
//   - int64：当前 goroutine 的 ID。
func GetGoIDSlow() int64 {
	return getGoIDSlow()
}
