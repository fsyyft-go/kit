// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package goroutine

import (
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	kittesting "github.com/fsyyft-go/kit/testing"
)

func TestGetGoID(t *testing.T) {
	t.Run("测试获取 GoroutineID", func(t *testing.T) {
		if isDarwinArm64() {
			kittesting.Println("M CPU 架构的 Mac 未能实现此方法。")
		} else {
			assertion := assert.New(t)

			var wg sync.WaitGroup
			var idOuter, idInternal int64
			wg.Add(1)
			idOuter = GetGoID()
			go func() {
				idInternal = GetGoID()
				wg.Done()
			}()
			wg.Wait()
			// 值每次都不一样，有需要的情况可以打印出来查看。
			assertion.NotEqual(idOuter, idInternal)
			// 在没有复用的情况下，里的一般会比外的大。
			assertion.LessOrEqual(idOuter, idInternal)
			// fmt.Println(idInternal, idOuter)
			kittesting.Println(idOuter, idInternal)
		}
	})
}

func TestGetGoID_Equal(t *testing.T) {
	t.Run("GetGoID GetGoIDSlow 需要返回相同的值", func(t *testing.T) {
		a := assert.New(t)

		// 获取快速版本的 goroutine ID。
		id := GetGoID()
		// 获取慢速版本的 goroutine ID。
		idSlow := GetGoIDSlow()

		a.Equal(id, idSlow, "GetGoID GetGoIDSlow 需要返回相同的值")
	})
}

func BenchmarkGetGoID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go func() { GetGoID() }()
	}
}

func BenchmarkGetGoIDSlow(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go func() { GetGoIDSlow() }()
	}
}

func isDarwinArm64() bool {
	return runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"
}
