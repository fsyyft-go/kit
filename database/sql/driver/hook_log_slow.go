// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package driver

import (
	"fmt"
	"strings"
	"time"

	kitlog "github.com/fsyyft-go/kit/log"
	kitgoroutine "github.com/fsyyft-go/kit/runtime/goroutine"
)

type (
	// HookLogSlow 是一个记录慢操作的 Hook。
	//
	// HookLogSlow 会在 After 阶段比较 HookContext.Duration 与 threshold；当耗时
	// 大于等于阈值时，它会异步提交一条包含操作类型、耗时以及可选 namespace、
	// SQL 和参数摘要的警告日志。
	HookLogSlow struct {
		// namespace 是日志记录的命名空间。
		namespace string
		// logger 是用于记录慢查询信息的日志记录器。
		logger kitlog.Logger
		// threshold 是慢查询的时间阈值。
		threshold time.Duration
	}
)

// NewHookLogSlow 创建一个慢操作日志 Hook。
//
// 参数：
//   - namespace：写入日志字段的命名空间；为空时省略该字段。
//   - logger：用于输出慢操作日志的记录器。
//   - threshold：慢操作阈值；当 Duration 大于等于该值时记录日志。
//
// 返回值：
//   - *HookLogSlow：在数据库操作达到慢阈值时异步写日志的 Hook。
func NewHookLogSlow(namespace string, logger kitlog.Logger, threshold time.Duration) *HookLogSlow {
	return &HookLogSlow{
		namespace: namespace,
		logger:    logger,
		threshold: threshold,
	}
}

// Before 在执行数据库操作前不做任何处理。
//
// 参数：
//   - ctx：当前操作的 HookContext。
//
// 返回值：
//   - error：始终返回 nil，不会阻止底层操作执行。
func (h *HookLogSlow) Before(ctx *HookContext) error {
	return nil
}

// After 在操作耗时达到阈值时异步记录慢操作日志。
//
// After 仅在 HookContext.Duration 大于等于 threshold 时写日志。日志字段包含
// operation、duration，以及存在时的 namespace、query 和 args。
//
// 参数：
//   - ctx：当前操作的 HookContext。
//
// 返回值：
//   - error：始终返回 nil，不会覆盖原始操作结果。
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
	_ = kitgoroutine.Submit(func() {
		h.logger.WithFields(m).Warn("")
	})

	return nil
}
