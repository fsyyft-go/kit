// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	kitlog "github.com/fsyyft-go/kit/log"
	kitgoroutine "github.com/fsyyft-go/kit/runtime/goroutine"
)

type (
	// logErrorHook 在请求返回原始错误后异步写入错误日志。
	logErrorHook struct {
		// logger 写入错误日志时使用的日志记录器。
		logger kitlog.Logger
	}
)

// Before 在请求发送前不做额外处理。
//
// 参数：
//   - ctx: 当前 HTTP Hook 上下文，本实现不会读取或修改它。
//
// 返回：
//   - error: 固定返回 nil。
func (h *logErrorHook) Before(ctx *HookContext) error {
	return nil
}

// After 在请求返回原始错误时异步写入错误日志。
//
// 参数：
//   - ctx: 当前 HTTP Hook 上下文，用于读取原始错误和请求 URL。
//
// 返回：
//   - error: 固定返回 nil；异步日志任务提交失败时错误会被忽略。
func (h *logErrorHook) After(ctx *HookContext) error {
	if nil != ctx.OriginError() {
		// 不等待日志任务执行完成，协程池提交失败也不改变原始请求结果。
		_ = kitgoroutine.Submit(func() {
			h.logger.
				WithField("error", ctx.OriginError()).
				WithField("url", ctx.Request().URL.String()).
				Error("")
		})
	}
	return nil
}

// NewLogErrorHook 创建一个在请求返回原始错误时异步记录错误日志的 Hook。
//
// 日志任务通过 runtime/goroutine 包级默认协程池提交；若协程池提交失败，错误会被忽略。
//
// 参数：
//   - logger: 写入错误日志时使用的日志记录器。
//
// 返回：
//   - *logErrorHook: 可注册到 HookManager 的错误日志 Hook。
func NewLogErrorHook(logger kitlog.Logger) *logErrorHook {
	h := &logErrorHook{
		logger: logger.WithField("hook", "log_error"),
	}
	return h
}
