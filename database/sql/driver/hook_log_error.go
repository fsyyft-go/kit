// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package driver

import (
	"fmt"
	"strings"

	"github.com/fsyyft-go/kit/log"
)

type (
	// HookLogError 用于记录数据库操作中的错误信息。
	HookLogError struct {
		// namespace 是日志记录的命名空间。
		namespace string
		// logger 是用于记录错误信息的日志记录器。
		logger log.Logger
	}
)

// NewHookLogError 创建一个新的 HookLogError 实例。
//
// 参数：
//   - namespace：日志记录的命名空间。
//   - logger：用于记录错误信息的日志记录器。
//
// 返回值：
//   - *HookLogError：返回一个新创建的 HookLogError 实例。
func NewHookLogError(namespace string, logger log.Logger) *HookLogError {
	return &HookLogError{
		namespace: namespace,
		logger:    logger,
	}
}

// Before 实现 Hook 接口的 Before 方法。
//
// 参数：
//   - ctx：钩子上下文，包含操作的相关信息。
//
// 返回值：
//   - error：始终返回 nil，因为这个钩子不会中断操作。
func (h *HookLogError) Before(ctx *HookContext) error {
	return nil
}

// After 实现 Hook 接口的 After 方法，用于记录操作中的错误信息。
//
// 参数：
//   - ctx：钩子上下文，包含操作的相关信息和结果。
//
// 返回值：
//   - error：始终返回 nil，因为这个钩子不会中断操作。
func (h *HookLogError) After(ctx *HookContext) error {
	// 只有在出现错误时才记录日志。
	if nil == ctx.OriginError() {
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
		"duration":  ctx.Duration(),
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

	// 记录错误日志。
	go h.logger.WithFields(m).Error(ctx.OriginError())

	return nil
}
