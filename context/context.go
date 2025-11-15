// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package context

import (
	stdctx "context"
	"time"
)

// withoutCancelCtx 是一个包装的 context 实现，用于忽略父 context 的取消信号。
// 它继承父 context 的值，但永远不会被取消或超时。
type withoutCancelCtx struct {
	// parent 父 context，用于获取值。
	parent stdctx.Context
}

// Deadline 返回零时间和 false，表示该 context 没有截止时间。
// 无论父 context 是否有截止时间，该实现都返回无截止时间。
//
// 返回值：
//   - deadline：零时间，表示无截止时间。
//   - ok：false，表示没有截止时间。
func (c *withoutCancelCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done 返回 nil，表示该 context 永远不会被取消。
// 父 context 的 Done channel 被忽略。
//
// 返回值：
//   - <-chan struct{}：nil，表示永远不会关闭。
func (c *withoutCancelCtx) Done() <-chan struct{} {
	return nil
}

// Err 返回 nil，表示该 context 没有错误。
// 即使父 context 被取消，该实现也返回 nil。
//
// 返回值：
//   - error：nil，表示没有错误。
func (c *withoutCancelCtx) Err() error {
	return nil
}

// Value 返回父 context 中对应 key 的值。
// 该方法直接委托给父 context，以继承其值。
//
// 参数：
//   - key：要查找的键。
//
// 返回值：
//   - any：对应键的值，如果不存在则为 nil。
func (c *withoutCancelCtx) Value(key any) any {
	return c.parent.Value(key)
}

// WithoutCancel 返回一个新的 context，它继承父 context 的值，但忽略取消信号和超时。
// 这相当于 Go 1.21+ 中的 context.WithoutCancel，用于低版本 Go 的兼容性实现。
//
// 参数：
//   - parent：父 context，用于获取值和忽略其取消信号，不能为 nil。
//
// 返回值：
//   - stdctx.Context：新的 context，忽略取消和超时，但保留值。
func WithoutCancel(parent stdctx.Context) stdctx.Context {
	if parent == nil {
		panic("context: WithoutCancel with nil parent")
	}
	return &withoutCancelCtx{parent: parent}
}
