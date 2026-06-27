// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package context

import (
	stdctx "context"
	"time"
)

// withoutCancelCtx 封装父 context，并只继承其键值查询能力。
//
// withoutCancelCtx 的 Deadline、Done 和 Err 均返回零值，因此不会因父 context
// 取消或超时而完成。parent 必须非 nil；该前提由 WithoutCancel 负责校验。
type withoutCancelCtx struct {
	// parent 保存被委托读取 Value 的父 context，必须非 nil。
	parent stdctx.Context
}

// Deadline 返回零时间和 false，表示该 context 没有截止时间。
//
// 参数：无。
//
// 返回：
//   - deadline: 零时间，调用方不应将其视为有效截止时间。
//   - ok: false，表示该 context 不提供截止时间。
func (c *withoutCancelCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

// Done 返回 nil，表示该 context 永远不会因取消或超时完成。
//
// 参数：无。
//
// 返回：
//   - <-chan struct{}: nil；调用方不能通过该 channel 监听父 context 的取消信号。
func (c *withoutCancelCtx) Done() <-chan struct{} {
	return nil
}

// Err 返回 nil，表示该 context 不记录取消或超时错误。
//
// 参数：无。
//
// 返回：
//   - error: nil；即使父 context 已取消或超时也不会返回 stdctx.Canceled 或
//     stdctx.DeadlineExceeded。
func (c *withoutCancelCtx) Err() error {
	return nil
}

// Value 返回父 context 中与 key 关联的值。
//
// 参数：
//   - key: 待查找的 context 键，语义和可比较性要求与标准库 context.Context.Value 一致。
//
// 返回：
//   - any: 父 context 链上对应键的值；不存在时返回 nil。
func (c *withoutCancelCtx) Value(key any) any {
	return c.parent.Value(key)
}

// WithoutCancel 返回继承父 context 值且不受父取消或超时影响的 context。
//
// 返回 context 的 Deadline 返回零值，Done 返回 nil，Err 返回 nil；Value 委托给 parent。
// 当 parent 为 nil 时，WithoutCancel 会 panic，行为与标准库 context.WithoutCancel 保持一致。
//
// 参数：
//   - parent: 提供 Value 查询来源的父 context，必须非 nil。
//
// 返回：
//   - stdctx.Context: 保留 parent 值查找能力、但不继承取消信号、截止时间和错误状态的 context。
func WithoutCancel(parent stdctx.Context) stdctx.Context {
	if parent == nil {
		panic("context: WithoutCancel with nil parent")
	}
	return &withoutCancelCtx{parent: parent}
}
