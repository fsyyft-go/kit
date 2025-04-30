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

func NewSlowHook(logger kitlog.Logger, threshold time.Duration) *slowHook {
	h := &slowHook{
		logger:    logger.WithField("hook", "log_slow"),
		threshold: threshold}
	return h
}
