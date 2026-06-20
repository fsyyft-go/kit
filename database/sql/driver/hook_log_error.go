// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package driver

import (
	"fmt"
	"strings"

	kitlog "github.com/fsyyft-go/kit/log"
	kitgoroutine "github.com/fsyyft-go/kit/runtime/goroutine"
)

type (
	// HookLogError 是一个只记录失败操作的 Hook。
	//
	// HookLogError 会在 After 阶段检查 HookContext.OriginError；当底层操作返回
	// 错误时，它会异步提交一条包含操作类型、耗时以及可选 namespace、SQL 和
	// 参数摘要的错误日志。
	HookLogError struct {
		// namespace 是日志记录的命名空间。
		namespace string
		// logger 是用于记录错误信息的日志记录器。
		logger kitlog.Logger
	}
)

// NewHookLogError 创建一个错误日志 Hook。
//
// 参数：
//   - namespace：写入日志字段的命名空间；为空时省略该字段。
//   - logger：用于输出错误日志的记录器。
//
// 返回值：
//   - *HookLogError：在数据库操作失败时异步写日志的 Hook。
func NewHookLogError(namespace string, logger kitlog.Logger) *HookLogError {
	return &HookLogError{
		namespace: namespace,
		logger:    logger,
	}
}

// Before 在执行数据库操作前不做任何处理。
//
// 参数：
//   - ctx：当前操作的 HookContext。
//
// 返回值：
//   - error：始终返回 nil，不会阻止底层操作执行。
func (h *HookLogError) Before(ctx *HookContext) error {
	return nil
}

// After 在底层操作返回错误时异步记录错误日志。
//
// After 仅在 HookContext.OriginError 非 nil 时写日志。日志字段包含 operation、
// duration，以及存在时的 namespace、query 和 args。
//
// 参数：
//   - ctx：当前操作的 HookContext。
//
// 返回值：
//   - error：始终返回 nil，不会覆盖原始操作结果。
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
	_ = kitgoroutine.Submit(func() {
		h.logger.WithFields(m).Error(ctx.OriginError())
	})

	return nil
}
