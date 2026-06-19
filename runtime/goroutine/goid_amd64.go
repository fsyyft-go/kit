// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build amd64

package goroutine

// GetGoID 返回当前 goroutine 的 ID。
//
// 该快速路径通过 goid_amd64.s 直接读取当前 runtime.g 的 goid 字段。
// SAFETY: 此实现依赖 Offset 返回的偏移值与目标 Go 版本的 runtime.g 布局保持一致；
// 升级 Go 版本后需要同步校验偏移表、汇编代码和配套结构定义。
//
// 参数：无。
//
// 返回：
//   - int64：当前 goroutine 的 ID。
func GetGoID() int64
