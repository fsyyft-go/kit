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
	initialBufferSize = 128 // 初始缓冲区大小。
)

// getGoIDSlow 获取当前协程的 ID，当无法从快速通道获取协程 ID 时使用此方法。
// 该方法通过获取协程的堆栈信息，然后解析堆栈信息来提取协程 ID。
func getGoIDSlow() int64 {
	var buf [initialBufferSize]byte
	stackBytes := buf[:]

	// 使用 runtime.Stack 填充缓冲区，获取当前协程的堆栈信息。
	// 参数 false 表示不希望获取完整堆栈信息，只需要足够信息来提取协程 ID。
	// 此处是关键操作，解释为何要这么做：因为直接获取完整堆栈会非常昂贵， 而协程 ID 通常位于堆栈信息的起始部分，通过限制获取的堆栈信息量，可以更高效地提取到协程 ID。
	stackBytes = stackBytes[:runtime.Stack(stackBytes, false)]

	return extractGID(stackBytes)
}

// extractGID 从一个字节切片中提取并解析 goroutine 的 ID。
func extractGID(s []byte) int64 {
	// 移除前缀"goroutine "，以便于后续处理。
	s = s[len("goroutine "):]
	// 找到第一个空格的位置，以确定 ID 的结束位置。
	s = s[:bytes.IndexByte(s, ' ')]
	// 将 ID 部分解析为int64类型，这里忽略了解析错误的情况。
	gid, _ := strconv.ParseInt(string(s), 10, 64)
	return gid
}

// GetGoIDSlow 获取当前协程的 ID，当无法从 GetGoID 获取协程 ID 时使用此方法。
// 该方法通过获取协程的堆栈信息，然后解析堆栈信息来提取协程 ID。
func GetGoIDSlow() int64 {
	return getGoIDSlow()
}
