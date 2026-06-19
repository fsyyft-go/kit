// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package goroutine 提供 goroutine ID 读取工具和基于 ants 的协程池封装。
//
// GetGoID 会按当前架构和 Go 版本选择快速路径；在未提供快速路径的平台上会退回到基于
// runtime.Stack 的慢速解析实现，调用方也可显式使用 GetGoIDSlow。amd64 构建下，
// Offset 返回快速路径使用的 runtime.g.goid 字段偏移。NewGoroutinePool 用于创建独立
// 协程池实例，返回的 cleanup 负责停止指标采集协程并释放底层 ants.Pool 资源；包级
// Submit 会惰性创建并复用默认池，在任务 panic 时 recover 并记录日志，不会把 panic
// 继续向调用方传播。
//
// 本包的快速路径依赖 runtime 内部结构、汇编实现和按 Go 版本维护的偏移信息；升级
// Go 版本或切换目标架构后需要重新验证对应实现。
package goroutine
