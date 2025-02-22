// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

// GetGoID 获取当前 goroutine 的 ID。
func GetGoID() uint64 {
	var buf [64]byte
	// 获取当前 goroutine 的调用栈。
	n := runtime.Stack(buf[:], false)
	// 调用栈的第一行包含 goroutine ID。
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	var id uint64
	fmt.Sscanf(idField, "%d", &id)
	return id
}

func main() {
	// 创建一个等待组，用于等待所有 goroutine 完成。
	var wg sync.WaitGroup

	// 打印主 goroutine 的 ID。
	fmt.Printf("主 goroutine ID: %d\n", GetGoID())

	// 启动 3 个新的 goroutine。
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			id := GetGoID()
			fmt.Printf("goroutine %d 的 ID: %d\n", index+1, id)
		}(i)
	}

	// 等待所有 goroutine 完成。
	wg.Wait()
}
