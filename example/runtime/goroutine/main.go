// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// goroutine 示例比较当前 goroutine 在快速路径和慢速路径下的 ID 获取结果。
//
// 该命令会打印主 goroutine 和新建 goroutine 的标识，并并排展示 GetGoID 与 GetGoIDSlow 的返回值。
package main

import (
	"fmt"
	"sync"

	kitgoroutine "github.com/fsyyft-go/kit/runtime/goroutine"
)

func main() {
	// 创建一个等待组，用于等待所有 goroutine 完成。
	var wg sync.WaitGroup

	// 打印主 goroutine 的 ID。
	fmt.Printf("主 goroutine ID: %d\n", kitgoroutine.GetGoID()) // nolint:staticcheck

	// 启动 3 个新的 goroutine。
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			id := kitgoroutine.GetGoID() // nolint:staticcheck
			idSlow := kitgoroutine.GetGoIDSlow()
			fmt.Printf("goroutine %d 的 ID: %d %d\n", index+1, id, idSlow)
		}(i)
	}

	// 等待所有 goroutine 完成。
	wg.Wait()
}
