// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package http

import (
	kitlog "github.com/fsyyft-go/kit/log"
	kitgoroutine "github.com/fsyyft-go/kit/runtime/goroutine"
)

type (
	logErrorHook struct {
		logger kitlog.Logger
	}
)

func (h *logErrorHook) Before(ctx *HookContext) error {
	return nil
}

func (h *logErrorHook) After(ctx *HookContext) error {
	if nil != ctx.OriginError() {
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
func NewLogErrorHook(logger kitlog.Logger) *logErrorHook {
	h := &logErrorHook{
		logger: logger.WithField("hook", "log_error"),
	}
	return h
}
