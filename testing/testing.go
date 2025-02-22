// Copyright 2024 fsyyft-go Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package testing 提供了一组用于测试时输出日志的辅助函数。
// 这个包封装了标准库 fmt 包的功能，并在输出内容前添加统一的日志前缀，
// 使测试输出更加清晰和易于识别。
package testing

import (
	"fmt"
)

const (
	// logHeader 定义了日志输出的统一前缀。
	// 这个前缀用于在测试输出中快速识别来自测试包的日志信息。
	// 前缀格式为 "=-=       "，其中包含了特殊的标识符和空格，使输出更加醒目。
	logHeader = "=-=       "
)

// Println 函数用于输出带有统一前缀的日志信息，并在末尾自动添加换行符。
// 该函数会在实际内容前添加 logHeader 前缀，并使用空格分隔多个参数。
//
// 参数：
//   - a ...interface{}：要输出的任意类型参数列表，支持多个参数。
//
// 使用示例：
//
//	testing.Println("测试信息")
//	testing.Println("值：", 100, "状态：", "成功")
func Println(a ...interface{}) {
	fmt.Print(logHeader)
	// fmt.Print(a...) 所有参数连在一起。
	// fmt.Println(a...) 参数之间空格分割。
	fmt.Println(a...)
}

// Printf 函数用于输出带有统一前缀的格式化日志信息。
// 该函数会在实际内容前添加 logHeader 前缀，并根据提供的格式字符串格式化输出内容。
//
// 参数：
//   - format string：格式化字符串，支持所有 fmt.Printf 的格式化指令。
//   - a ...interface{}：要格式化输出的参数列表。
//
// 使用示例：
//
//	testing.Printf("当前进度：%d%%\n", 50)
//	testing.Printf("用户：%s，年龄：%d\n", "张三", 25)
func Printf(format string, a ...interface{}) {
	fmt.Print(logHeader)
	fmt.Printf(format, a...)
}
