// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build amd64

package goroutine

// GetGoID 返回当前 goroutine 的 ID。
//
// SAFETY: 该快速路径依赖 goid_amd64.s 汇编实现以及 Offset 返回的 runtime.g.goid 字段偏移。
// 升级 Go 版本后需要同步验证偏移表与配套结构定义。
//
// 返回值：
//   - int64：当前 goroutine 的 ID。
func GetGoID() int64
