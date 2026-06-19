// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	"time"

	kitlog "github.com/fsyyft-go/kit/log"
	kitgoroutine "github.com/fsyyft-go/kit/runtime/goroutine"
)

type (
	// slowHook 在请求耗时超过阈值后异步写入慢请求日志。
	slowHook struct {
		// logger 写入慢请求日志时使用的日志记录器。
		logger kitlog.Logger

		// threshold 触发慢请求日志的耗时阈值。
		threshold time.Duration
	}
)

// Before 在请求发送前不做额外处理。
//
// 参数：
//   - ctx: 当前 HTTP Hook 上下文，本实现不会读取或修改它。
//
// 返回：
//   - error: 固定返回 nil。
func (h *slowHook) Before(ctx *HookContext) error {
	return nil
}

// After 在请求耗时超过阈值时异步写入慢请求日志。
//
// 参数：
//   - ctx: 当前 HTTP Hook 上下文，用于读取请求耗时和 URL。
//
// 返回：
//   - error: 固定返回 nil；异步日志任务提交失败时错误会被忽略。
func (h *slowHook) After(ctx *HookContext) error {
	if ctx.Duration() > h.threshold {
		// 不等待日志任务执行完成，协程池提交失败也不改变原始请求结果。
		_ = kitgoroutine.Submit(func() {
			h.logger.
				WithField("duration", ctx.Duration()).
				WithField("url", ctx.Request().URL.String()).
				Warn("")
		})
	}
	return nil
}

// NewSlowHook 创建一个在请求耗时超过 threshold 时异步记录慢请求日志的 Hook。
//
// threshold 不会在构造时校验；当 threshold 小于等于 0 时，绝大多数已完成请求都会满足记录条件。
// 日志任务通过 runtime/goroutine 包级默认协程池提交；若协程池提交失败，错误会被忽略。
//
// 参数：
//   - logger: 写入慢请求日志时使用的日志记录器。
//   - threshold: 触发慢请求日志的耗时阈值。
//
// 返回：
//   - *slowHook: 可注册到 HookManager 的慢请求日志 Hook。
func NewSlowHook(logger kitlog.Logger, threshold time.Duration) *slowHook {
	h := &slowHook{
		logger:    logger.WithField("hook", "log_slow"),
		threshold: threshold}
	return h
}
