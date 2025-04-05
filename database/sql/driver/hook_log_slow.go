// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package driver

import (
	"fmt"
	"strings"
	"time"

	"github.com/fsyyft-go/kit/log"
)

type (
	// HookLogSlow 用于记录数据库慢查询操作。
	HookLogSlow struct {
		// namespace 是日志记录的命名空间。
		namespace string
		// logger 是用于记录慢查询信息的日志记录器。
		logger log.Logger
		// threshold 是慢查询的时间阈值。
		threshold time.Duration
	}
)

// NewHookLogSlow 创建一个新的 HookLogSlow 实例。
//
// 参数：
//   - namespace：日志记录的命名空间。
//   - logger：用于记录慢查询信息的日志记录器。
//   - threshold：慢查询的时间阈值，超过这个时间的查询会被记录。
//
// 返回值：
//   - *HookLogSlow：返回一个新创建的 HookLogSlow 实例。
func NewHookLogSlow(namespace string, logger log.Logger, threshold time.Duration) *HookLogSlow {
	return &HookLogSlow{
		namespace: namespace,
		logger:    logger,
		threshold: threshold,
	}
}

// Before 实现 Hook 接口的 Before 方法。
//
// 参数：
//   - ctx：钩子上下文，包含操作的相关信息。
//
// 返回值：
//   - error：始终返回 nil，因为这个钩子不会中断操作。
func (h *HookLogSlow) Before(ctx *HookContext) error {
	return nil
}

// After 实现 Hook 接口的 After 方法，用于记录慢查询信息。
//
// 参数：
//   - ctx：钩子上下文，包含操作的相关信息和结果。
//
// 返回值：
//   - error：始终返回 nil，因为这个钩子不会中断操作。
func (h *HookLogSlow) After(ctx *HookContext) error {
	// 只记录超过阈值的查询。
	duration := ctx.Duration()
	if duration < h.threshold {
		return nil
	}

	// 构建参数字符串。
	var args []string
	for _, arg := range ctx.Args() {
		args = append(args, fmt.Sprintf("%v", arg.Value))
	}
	argsStr := strings.Join(args, ", ")

	m := map[string]interface{}{
		"operation": ctx.OpType(),
		"duration":  duration,
	}
	if h.namespace != "" {
		m["namespace"] = h.namespace
	}
	if ctx.Query() != "" {
		m["query"] = ctx.Query()
	}
	if argsStr != "" {
		m["args"] = argsStr
	}

	// 记录慢查询日志。
	go h.logger.WithFields(m).Warn("")

	return nil
}
