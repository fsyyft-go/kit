// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !windows && arm64

package goroutine

// getg 返回当前 goroutine 对应的 runtime.g 指针。
//
// SAFETY: 此声明与 goid_arm64.s 配套，依赖当前平台通过 TLS 保存 g 指针的
// runtime 内部实现。升级 Go 版本或调整目标架构后需要重新验证汇编与 g 结构定义。
//
// 参数：无。
//
// 返回：
//   - *g：当前 goroutine 对应的 runtime.g 指针。
func getg() *g

// GetGoID 返回当前 goroutine 的 ID。
//
// 该快速路径直接读取 runtime.g.goid，依赖本包维护的 runtime 内部结构定义
// 与 goid_arm64.s 汇编实现保持一致。
//
// 参数：无。
//
// 返回：
//   - int64：当前 goroutine 的 ID。
func GetGoID() int64 {
	return getg().goid
}
