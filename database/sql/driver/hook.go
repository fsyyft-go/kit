// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package driver

import (
	"context"
	"database/sql/driver"
	"sync"
	"time"
)

// OpType 表示数据库操作类型。
type OpType int

const (
	// OpConnect 表示连接操作。
	OpConnect OpType = iota
	// OpBegin 表示开始事务操作。
	OpBegin
	// OpCommit 表示提交事务操作。
	OpCommit
	// OpRollback 表示回滚事务操作。
	OpRollback
	// OpPrepare 表示预处理语句操作。
	OpPrepare
	// OpStmtExec 表示执行预处理语句操作。
	OpStmtExec
	// OpStmtQuery 表示查询预处理语句操作。
	OpStmtQuery
	// OpStmtClose 表示关闭预处理语句操作。
	OpStmtClose
	// OpExec 表示执行操作。
	OpExec
	// OpQuery 表示查询操作。
	OpQuery
	// OpPing 表示 ping 操作。
	OpPing
)

// String 返回操作类型的字符串表示。
//
// 返回值：
//   - string：返回操作类型对应的字符串描述。
func (o OpType) String() string {
	switch o {
	case OpConnect:
		return "Connect"
	case OpBegin:
		return "Begin"
	case OpCommit:
		return "Commit"
	case OpRollback:
		return "Rollback"
	case OpPrepare:
		return "Prepare"
	case OpStmtExec:
		return "StmtExec"
	case OpStmtQuery:
		return "StmtQuery"
	case OpStmtClose:
		return "StmtClose"
	case OpExec:
		return "Exec"
	case OpQuery:
		return "Query"
	case OpPing:
		return "Ping"
	default:
		return "Unknown"
	}
}

// HookContext 包含数据库操作的上下文信息。
type HookContext struct {
	// 原始上下文对象。
	originContext context.Context
	// 操作类型。
	opType OpType
	// SQL 查询语句。
	query string
	// SQL 语句的参数列表。
	args []driver.NamedValue
	// 操作开始时间。
	startTime time.Time
	// 操作结束时间。
	endTime time.Time
	// 原始操作返回的错误。
	originError error
	// 原始操作的结果。
	originResult interface{}
	// 用于存储钩子相关的键值对数据。
	hookMap sync.Map
}

// NewHookContext 创建一个新的 HookContext 实例。
//
// 参数：
//   - ctx：原始上下文对象，用于控制操作的生命周期。
//   - opType：操作类型，表示当前执行的数据库操作。
//   - query：SQL 查询语句，可以为空。
//   - args：SQL 语句的参数列表，可以为 nil。
//
// 返回值：
//   - *HookContext：返回一个新创建的 HookContext 实例。
func NewHookContext(ctx context.Context, opType OpType, query string, args []driver.NamedValue) *HookContext {
	return &HookContext{
		originContext: ctx,
		opType:        opType,
		query:         query,
		args:          args,
		startTime:     time.Now(),
	}
}

// SetResult 设置操作结果和结束时间。
//
// 参数：
//   - result：操作的结果，可以是任意类型。
//   - err：操作过程中产生的错误，如果没有错误则为 nil。
func (h *HookContext) SetResult(result interface{}, err error) {
	h.originResult = result
	h.originError = err
	h.endTime = time.Now()
}

// StartTime 返回操作开始时间。
//
// 返回值：
//   - time.Time：返回操作的开始时间。
func (h *HookContext) StartTime() time.Time {
	return h.startTime
}

// EndTime 返回操作结束时间。
//
// 返回值：
//   - time.Time：返回操作的结束时间。
func (h *HookContext) EndTime() time.Time {
	return h.endTime
}

// OpType 返回操作类型。
//
// 返回值：
//   - OpType：返回当前操作的类型。
func (h *HookContext) OpType() OpType {
	return h.opType
}

// Query 返回操作语句。
//
// 返回值：
//   - string：返回 SQL 查询语句。
func (h *HookContext) Query() string {
	return h.query
}

// Args 返回操作参数。
//
// 返回值：
//   - []driver.NamedValue：返回 SQL 语句的参数列表。
func (h *HookContext) Args() []driver.NamedValue {
	return h.args
}

// OriginError 返回原始操作错误。
//
// 返回值：
//   - error：返回操作过程中产生的原始错误。
func (h *HookContext) OriginError() error {
	return h.originError
}

// OriginResult 返回原始操作结果。
//
// 返回值：
//   - interface{}：返回操作的原始结果。
func (h *HookContext) OriginResult() interface{} {
	return h.originResult
}

// GetHookValue 获取 hook 中的值。
//
// 参数：
//   - key：要获取的键名。
//
// 返回值：
//   - interface{}：返回与键关联的值。
//   - bool：如果键存在返回 true，否则返回 false。
func (h *HookContext) GetHookValue(key string) (interface{}, bool) {
	return h.hookMap.Load(key)
}

// SetHookValue 设置 hook 中的值。
//
// 参数：
//   - key：要设置的键名。
//   - value：要存储的值。
func (h *HookContext) SetHookValue(key string, value interface{}) {
	h.hookMap.Store(key, value)
}

// Deadline 实现 context.Context 接口。
//
// 返回值：
//   - deadline：返回上下文的截止时间。
//   - ok：如果设置了截止时间返回 true，否则返回 false。
func (h *HookContext) Deadline() (deadline time.Time, ok bool) {
	return h.originContext.Deadline()
}

// Done 实现 context.Context 接口。
//
// 返回值：
//   - <-chan struct{}：返回一个 channel，当上下文被取消时会被关闭。
func (h *HookContext) Done() <-chan struct{} {
	return h.originContext.Done()
}

// Err 实现 context.Context 接口。
//
// 返回值：
//   - error：如果上下文被取消，返回取消的原因。
func (h *HookContext) Err() error {
	return h.originContext.Err()
}

// Value 实现 context.Context 接口。
//
// 参数：
//   - key：要获取的值的键。
//
// 返回值：
//   - interface{}：返回与键关联的值。
func (h *HookContext) Value(key interface{}) interface{} {
	return h.originContext.Value(key)
}

// Hook 定义数据库操作的钩子接口。
type Hook interface {
	// Before 在操作执行前调用。
	//
	// 参数：
	//   - ctx：钩子上下文，包含操作的相关信息。
	//
	// 返回值：
	//   - error：如果钩子执行出错，返回相应的错误信息。
	Before(ctx *HookContext) error

	// After 在操作执行后调用。
	//
	// 参数：
	//   - ctx：钩子上下文，包含操作的相关信息和结果。
	//
	// 返回值：
	//   - error：如果钩子执行出错，返回相应的错误信息。
	After(ctx *HookContext) error
}

// HookManager 管理多个 Hook 的执行。
type HookManager struct {
	// hooks 存储注册的所有钩子。
	hooks []Hook
}

// NewHookManager 创建一个新的 HookManager 实例。
//
// 返回值：
//   - *HookManager：返回一个新创建的 HookManager 实例。
func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make([]Hook, 0),
	}
}

// AddHook 添加一个 Hook。
//
// 参数：
//   - hook：要添加的钩子实例。
func (m *HookManager) AddHook(hook Hook) {
	m.hooks = append(m.hooks, hook)
}

// Before 实现 Hook 接口，按顺序执行所有 Hook 的 Before 方法。
//
// 参数：
//   - ctx：钩子上下文，包含操作的相关信息。
//
// 返回值：
//   - error：如果任何钩子执行出错，返回第一个错误信息。
func (m *HookManager) Before(ctx *HookContext) error {
	for _, hook := range m.hooks {
		if err := hook.Before(ctx); err != nil {
			return err
		}
	}
	return nil
}

// After 实现 Hook 接口，按逆序执行所有 Hook 的 After 方法。
//
// 参数：
//   - ctx：钩子上下文，包含操作的相关信息和结果。
//
// 返回值：
//   - error：如果任何钩子执行出错，返回第一个错误信息。
func (m *HookManager) After(ctx *HookContext) error {
	for i := len(m.hooks) - 1; i >= 0; i-- {
		if err := m.hooks[i].After(ctx); err != nil {
			return err
		}
	}
	return nil
}
