// Copyright (c) 2025 fsyyft-go
//
// Licensed under the MIT License (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://github.com/fsyyft-go/kit/blob/main/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goroutine

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	kitTesting "github.com/fsyyft-go/kit/testing"
)

// TestGetGoIDSlow 使用 assert 包来验证 getGoIDSlow 函数的行为。
func TestGetGoIDSlow(t *testing.T) {
	// 创建一个 assert 对象用于断言。
	a := assert.New(t)

	t.Run("测试获取 getGoIDSlow 是非零整数", func(t *testing.T) {
		// 调用被测函数。
		gid := getGoIDSlow()

		// 验证返回值是否是非零整数。
		a.NotEqual(int64(0), gid, "getGoIDSlow() 得到的是一个非零的整数。")
	})

	t.Run("测试获取 getGoIDSlow 内部 ID 比外部大", func(t *testing.T) {
		var wg sync.WaitGroup
		var idOuter, idInternal int64
		wg.Add(1)
		idOuter = getGoIDSlow()
		go func() {
			idInternal = getGoIDSlow()
			wg.Done()
		}()
		wg.Wait()
		// 值每次都不一样，有需要的情况可以打印出来查看。
		a.NotEqual(idOuter, idInternal)
		// 在没有复用的情况下，里的一般会比外的大。
		a.LessOrEqual(idOuter, idInternal)
		// fmt.Println(idInternal, idOuter)
		kitTesting.Println(idOuter, idInternal)
	})
}
