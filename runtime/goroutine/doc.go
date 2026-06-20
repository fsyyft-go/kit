// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package goroutine 提供 goroutine ID 读取工具和基于 ants 的协程池封装。
//
// GetGoID 优先使用与架构、Go 版本绑定的快速路径；在未提供快速路径的平台上，
// GetGoID 会退回到基于 runtime.Stack 的慢速解析实现，调用方也可显式使用 GetGoIDSlow。
// Offset 返回 amd64 快速路径使用的 runtime.g.goid 偏移。包级 Submit 会惰性创建并复用默认协程池，
// 并在任务 panic 时 recover 后记录日志，不会把 panic 继续向调用方传播。
// NewGoroutinePool 用于创建独立池实例，并返回 cleanup 用于停止指标采集协程并释放底层 ants.Pool 资源；
// 未调用 cleanup 的池会继续保留后台状态与资源。快速路径依赖 runtime 内部布局、汇编实现和版本偏移表，
// 升级 Go 版本或目标架构后需要重新验证。
package goroutine
