// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

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

// HookContext 描述一次数据库驱动操作在 Hook 链中的共享上下文。
//
// HookContext 记录操作类型、SQL、参数、开始与结束时间、底层操作的原始结果
// 和原始错误，并实现 context.Context 以透传取消信号、截止时间和上下文值。
// NewHookContext 创建后会立即记录开始时间；调用 SetResult 后，Duration 才表示
// 本次操作的实际耗时。Hook 之间还可以通过 SetHookValue 和 GetHookValue 在当前
// 操作内共享数据。
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

// NewHookContext 创建一次数据库操作对应的 HookContext。
//
// 参数：
//   - ctx：底层操作使用的原始上下文；调用方应传入非 nil 上下文。
//   - opType：当前数据库操作的类型。
//   - query：当前操作关联的 SQL；对连接、事务或 Ping 等无 SQL 的操作通常为空字符串。
//   - args：当前操作的命名参数列表；没有参数时可为 nil。
//
// 返回值：
//   - *HookContext：已记录开始时间的 HookContext。
func NewHookContext(ctx context.Context, opType OpType, query string, args []driver.NamedValue) *HookContext {
	return &HookContext{
		originContext: ctx,
		opType:        opType,
		query:         query,
		args:          args,
		startTime:     time.Now(),
	}
}

// SetResult 记录底层操作的原始结果、原始错误和结束时间。
//
// SetResult 通常在底层 driver 调用返回后由包装器调用一次；调用后 Duration
// 才表示本次操作的实际耗时。
//
// 参数：
//   - result：底层操作返回的原始结果；类型随操作而变化。
//   - err：底层操作返回的原始错误；成功时为 nil。
func (h *HookContext) SetResult(result interface{}, err error) {
	h.originResult = result
	h.originError = err
	h.endTime = time.Now()
}

// StartTime 返回当前操作开始执行的时间戳。
//
// 返回值：
//   - time.Time：记录 HookContext 创建时的开始时间。
func (h *HookContext) StartTime() time.Time {
	return h.startTime
}

// EndTime 返回当前操作记录结束结果的时间戳。
//
// 在 SetResult 调用前，EndTime 返回零值。
//
// 返回值：
//   - time.Time：记录底层操作结果写入 HookContext 的时间。
func (h *HookContext) EndTime() time.Time {
	return h.endTime
}

// Duration 返回当前操作从开始到记录结果之间的耗时。
//
// 在 SetResult 调用前，Duration 的返回值不表示有效耗时。
//
// 返回值：
//   - time.Duration：当前操作的记录耗时。
func (h *HookContext) Duration() time.Duration {
	return h.endTime.Sub(h.startTime)
}

// OpType 返回当前操作的类型。
//
// 返回值：
//   - OpType：当前 HookContext 记录的数据库操作类型。
func (h *HookContext) OpType() OpType {
	return h.opType
}

// Query 返回当前操作关联的 SQL 语句。
//
// 对连接、Ping、事务提交和回滚等无 SQL 的操作，Query 可能为空字符串。
//
// 返回值：
//   - string：当前操作关联的 SQL 文本。
func (h *HookContext) Query() string {
	return h.query
}

// Args 返回当前操作携带的命名参数列表。
//
// 返回的切片应视为只读，供 Hook 观察参数内容使用。
//
// 返回值：
//   - []driver.NamedValue：当前操作的参数列表。
func (h *HookContext) Args() []driver.NamedValue {
	return h.args
}

// OriginError 返回底层操作产生的原始错误。
//
// 如果底层操作成功，OriginError 返回 nil。
//
// 返回值：
//   - error：底层操作返回的原始错误。
func (h *HookContext) OriginError() error {
	return h.originError
}

// OriginResult 返回底层操作产生的原始结果。
//
// 结果类型取决于当前操作，例如 driver.Conn、driver.Stmt、driver.Result、
// driver.Rows、driver.Tx 或 nil。
//
// 返回值：
//   - interface{}：底层操作返回的原始结果。
func (h *HookContext) OriginResult() interface{} {
	return h.originResult
}

// GetHookValue 读取当前操作中由 Hook 保存的共享数据。
//
// 参数：
//   - key：要读取的共享数据键名。
//
// 返回值：
//   - interface{}：与 key 关联的值。
//   - bool：key 存在时返回 true，否则返回 false。
func (h *HookContext) GetHookValue(key string) (interface{}, bool) {
	return h.hookMap.Load(key)
}

// SetHookValue 在当前操作的 Hook 链中保存共享数据。
//
// 参数：
//   - key：要写入的共享数据键名。
//   - value：与 key 关联的值。
func (h *HookContext) SetHookValue(key string, value interface{}) {
	h.hookMap.Store(key, value)
}

// Deadline 返回原始上下文的截止时间。
//
// 返回值：
//   - deadline：原始上下文的截止时间。
//   - ok：原始上下文设置了截止时间时返回 true，否则返回 false。
func (h *HookContext) Deadline() (deadline time.Time, ok bool) {
	return h.originContext.Deadline()
}

// Done 返回原始上下文的 Done channel。
//
// 返回值：
//   - <-chan struct{}：当原始上下文被取消时关闭的 channel。
func (h *HookContext) Done() <-chan struct{} {
	return h.originContext.Done()
}

// Err 返回原始上下文的取消错误。
//
// 返回值：
//   - error：原始上下文被取消或超时时返回对应错误，否则返回 nil。
func (h *HookContext) Err() error {
	return h.originContext.Err()
}

// Value 返回原始上下文中与 key 关联的值。
//
// 参数：
//   - key：要读取的上下文键。
//
// 返回值：
//   - interface{}：原始上下文中与 key 关联的值。
func (h *HookContext) Value(key interface{}) interface{} {
	return h.originContext.Value(key)
}

// Hook 定义数据库驱动操作前后的扩展点。
//
// Before 在底层操作执行前调用；返回错误会阻止底层操作继续执行。
// After 在底层操作返回并写入 HookContext 后调用；返回错误会覆盖底层操作
// 原本准备返回给调用方的错误。
type Hook interface {
	// Before 在底层数据库操作开始前执行。
	//
	// 参数：
	//   - ctx：当前操作的 HookContext；Before 阶段尚未写入 OriginResult、OriginError 和 EndTime。
	//
	// 返回值：
	//   - error：返回非 nil 错误会中止底层操作，并将该错误直接返回给调用方。
	Before(ctx *HookContext) error

	// After 在底层数据库操作完成后执行。
	//
	// 参数：
	//   - ctx：当前操作的 HookContext；After 阶段可以读取底层操作写入的结果、错误和耗时。
	//
	// 返回值：
	//   - error：返回非 nil 错误会覆盖底层操作原本要返回给调用方的错误。
	After(ctx *HookContext) error
}

// HookManager 按顺序编排多个 Hook。
//
// HookManager 会按 AddHook 的注册顺序调用 Before，并按相反顺序调用 After，
// 以便成对组织前置和后置逻辑。HookManager 不做并发保护，通常应在初始化阶段
// 完成 AddHook，再作为只读 Hook 链共享使用。
type HookManager struct {
	// hooks 存储注册的所有钩子。
	hooks []Hook
}

// NewHookManager 创建一个空的 HookManager。
//
// 返回值：
//   - *HookManager：可继续通过 AddHook 注册 Hook 的管理器。
func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make([]Hook, 0),
	}
}

// AddHook 按注册顺序向 HookManager 追加一个 Hook。
//
// 后续 Before 会按照追加顺序执行该 Hook，After 会按逆序执行。
//
// 参数：
//   - hook：要注册到管理器末尾的 Hook。
func (m *HookManager) AddHook(hook Hook) {
	m.hooks = append(m.hooks, hook)
}

// Before 按 Hook 的注册顺序执行所有前置逻辑。
//
// 参数：
//   - ctx：当前操作的 HookContext。
//
// 返回值：
//   - error：第一个返回的 Hook 错误；发生错误后不会继续执行后续 Hook。
func (m *HookManager) Before(ctx *HookContext) error {
	for _, hook := range m.hooks {
		if err := hook.Before(ctx); err != nil {
			return err
		}
	}
	return nil
}

// After 按 Hook 的注册逆序执行所有后置逻辑。
//
// 参数：
//   - ctx：当前操作的 HookContext。
//
// 返回值：
//   - error：逆序执行过程中遇到的第一个 Hook 错误；发生错误后不会继续执行剩余 Hook。
func (m *HookManager) After(ctx *HookContext) error {
	for i := len(m.hooks) - 1; i >= 0; i-- {
		if err := m.hooks[i].After(ctx); err != nil {
			return err
		}
	}
	return nil
}
