// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package testing

import (
	"fmt"
)

const (
	// logHeader 定义了日志输出的统一前缀，用于在测试输出中快速识别来自测试包的日志信息。
	// 前缀格式为 "=-=       "，包含特殊标识符和空格，使输出更加醒目。
	logHeader = "=-=       "
)

// Println 将 a 按 fmt.Println 语义写入标准输出，并在内容前添加统一日志前缀。
//
// 每次调用先写入 logHeader，再由 fmt.Println 写入 a 和换行。该函数不提供额外同步；
// 并发调用时前缀和正文可能与其他输出交错。
//
// 参数：
//   - a: 要输出的值列表；未传入时仅输出前缀和换行，多个值按 fmt.Println 规则以空格分隔。
func Println(a ...interface{}) {
	fmt.Print(logHeader)
	fmt.Println(a...)
}

// Printf 将 format 和 a 按 fmt.Printf 语义写入标准输出，并在内容前添加统一日志前缀。
//
// 每次调用先写入 logHeader，再由 fmt.Printf 写入格式化结果。该函数不提供额外同步；
// 并发调用时前缀和正文可能与其他输出交错，换行完全由 format 决定。
//
// 参数：
//   - format: 格式字符串，支持 fmt.Printf 的格式化指令；空字符串时仅输出前缀。
//   - a: 用于 format 的值列表；数量或类型不匹配时保留 fmt.Printf 的诊断输出。
func Printf(format string, a ...interface{}) {
	fmt.Print(logHeader)
	fmt.Printf(format, a...)
}
