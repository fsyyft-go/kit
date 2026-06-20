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
	slowHook struct {
		logger    kitlog.Logger
		threshold time.Duration
	}
)

func (h *slowHook) Before(ctx *HookContext) error {
	return nil
}

func (h *slowHook) After(ctx *HookContext) error {
	if ctx.Duration() > h.threshold {
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
func NewSlowHook(logger kitlog.Logger, threshold time.Duration) *slowHook {
	h := &slowHook{
		logger:    logger.WithField("hook", "log_slow"),
		threshold: threshold}
	return h
}
