// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package driver

import (
	"context"
	"database/sql/driver"
	"errors"
)

// 以下断言确保包装类型满足 database/sql/driver 相关接口。
var (
	_ driver.Driver             = (*KitDriver)(nil)
	_ driver.Conn               = (*kitConn)(nil)
	_ driver.ConnPrepareContext = (*kitConn)(nil)
	_ driver.ExecerContext      = (*kitConn)(nil)
	_ driver.QueryerContext     = (*kitConn)(nil)
	_ driver.Pinger             = (*kitConn)(nil)
	_ driver.ConnBeginTx        = (*kitConn)(nil)
	_ driver.NamedValueChecker  = (*kitConn)(nil)
	_ driver.SessionResetter    = (*kitConn)(nil)
)

// KitDriver 包装底层 driver.Driver，并在连接及其派生对象的操作前后执行 Hook。
//
// KitDriver 只负责把 Open 创建出的连接，以及后续 PrepareContext 和 BeginTx 创建出的语句、事务再包装为带 Hook 的实现，
// 不改变底层驱动对 DSN、结果类型或错误值的基础语义。调用方应为 d 和 h 提供
// 可用的非 nil 实现。
type KitDriver struct {
	// 原始数据库驱动实例。
	driver driver.Driver
	// 用于执行钩子操作的接口实例。
	hook Hook
}

// NewKitDriver 创建一个带 Hook 的 driver.Driver 包装器。
//
// 参数：
//   - d: 实际执行数据库协议的底层 driver。
//   - h: 在 Open 以及后续连接、语句和事务操作前后执行的 Hook；调用方应传入非 nil 实现。
//
// 返回：
//   - *KitDriver: 对 d 的包装实例。
func NewKitDriver(d driver.Driver, h Hook) *KitDriver {
	return &KitDriver{
		driver: d,
		hook:   h,
	}
}

// Open 打开一个新的底层数据库连接，并为该连接安装 Hook 包装。
//
// Open 会先以 OpConnect 调用 Hook.Before，再调用底层 driver 的 Open，随后
// 将连接或错误写入 HookContext 并调用 Hook.After。由于 driver.Driver.Open
// 不接收上下文，HookContext 使用 context.Background 作为原始上下文。
//
// 参数：
//   - name: 原样传递给底层 driver 的 DSN。
//
// 返回：
//   - driver.Conn: 成功时返回带 Hook 包装的连接。
//   - error: Hook.Before 返回错误、底层 driver.Open 失败或 Hook.After 返回错误时返回错误。
func (d *KitDriver) Open(name string) (driver.Conn, error) {
	// 创建一个新的钩子上下文，用于连接操作。
	ctx := NewHookContext(context.Background(), OpConnect, "", nil)

	// 执行前置钩子。
	if err := d.hook.Before(ctx); err != nil {
		return nil, err
	}

	// 调用原始驱动的 Open 方法。
	conn, err := d.driver.Open(name)
	// 设置操作结果。
	ctx.SetResult(conn, err)

	// 执行后置钩子。
	if err := d.hook.After(ctx); err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	// 返回包装后的连接实例。
	return &kitConn{
		Conn: conn,
		hook: d.hook,
	}, nil
}

// kitConn 包装底层 driver.Conn，并为连接级操作执行 Hook。
//
// kitConn 在 PrepareContext、ExecContext、QueryContext、Ping 和 BeginTx 中按统一流程执行 Hook；
// 对 CheckNamedValue 和 ResetSession 则直接转发到底层可选接口。
type kitConn struct {
	// 原始数据库连接实例。
	driver.Conn
	// 用于执行钩子操作的接口实例。
	hook Hook
}

// PrepareContext 创建带 Hook 包装的预处理语句。
//
// 参数：
//   - ctx: 控制预处理操作生命周期的上下文。
//   - query: 要预处理的 SQL 语句文本。
//
// 返回：
//   - driver.Stmt: 成功时返回带 Hook 包装的预处理语句。
//   - error: 底层连接不支持 driver.ConnPrepareContext、Hook.Before 返回错误、底层预处理失败或 Hook.After 返回错误时返回错误。
func (c *kitConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	// 检查原始连接是否支持 PrepareContext。
	if preparerCtx, ok := c.Conn.(driver.ConnPrepareContext); ok {
		// 创建预处理语句的钩子上下文。
		hookCtx := NewHookContext(ctx, OpPrepare, query, nil)

		// 执行前置钩子。
		if err := c.hook.Before(hookCtx); err != nil {
			return nil, err
		}

		// 调用原始连接的 PrepareContext 方法。
		stmt, err := preparerCtx.PrepareContext(ctx, query)
		// 设置操作结果。
		hookCtx.SetResult(stmt, err)

		// 执行后置钩子。
		if err := c.hook.After(hookCtx); err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}

		// 返回包装后的预处理语句实例。
		return &kitStmt{
			Stmt:  stmt,
			hook:  c.hook,
			query: query,
		}, nil
	}
	return nil, errors.New("driver does not support prepare context")
}

// ExecContext 执行 SQL 语句并在操作前后执行 Hook。
//
// 参数：
//   - ctx: 控制执行操作生命周期的上下文。
//   - query: 要执行的 SQL 语句。
//   - args: SQL 语句的命名参数列表；没有参数时可为 nil。
//
// 返回：
//   - driver.Result: 底层驱动返回的执行结果。
//   - error: 底层连接不支持 driver.ExecerContext、Hook.Before 返回错误、底层执行失败或 Hook.After 返回错误时返回错误。
func (c *kitConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	// 检查原始连接是否支持 ExecContext。
	if execer, ok := c.Conn.(driver.ExecerContext); ok {
		// 创建执行操作的钩子上下文。
		hookCtx := NewHookContext(ctx, OpExec, query, args)

		// 执行前置钩子。
		if err := c.hook.Before(hookCtx); err != nil {
			return nil, err
		}

		// 调用原始连接的 ExecContext 方法。
		result, err := execer.ExecContext(ctx, query, args)
		// 设置操作结果。
		hookCtx.SetResult(result, err)

		// 执行后置钩子。
		if err := c.hook.After(hookCtx); err != nil {
			return nil, err
		}

		return result, err
	}
	return nil, errors.New("driver does not support exec context")
}

// QueryContext 执行查询语句并在操作前后执行 Hook。
//
// 参数：
//   - ctx: 控制查询操作生命周期的上下文。
//   - query: 要执行的查询 SQL 语句。
//   - args: SQL 语句的命名参数列表；没有参数时可为 nil。
//
// 返回：
//   - driver.Rows: 底层驱动返回的查询结果集。
//   - error: 底层连接不支持 driver.QueryerContext、Hook.Before 返回错误、底层查询失败或 Hook.After 返回错误时返回错误。
func (c *kitConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	// 检查原始连接是否支持 QueryContext。
	if queryer, ok := c.Conn.(driver.QueryerContext); ok {
		// 创建查询操作的钩子上下文。
		hookCtx := NewHookContext(ctx, OpQuery, query, args)

		// 执行前置钩子。
		if err := c.hook.Before(hookCtx); err != nil {
			return nil, err
		}

		// 调用原始连接的 QueryContext 方法。
		rows, err := queryer.QueryContext(ctx, query, args)
		// 设置操作结果。
		hookCtx.SetResult(rows, err)

		// 执行后置钩子。
		if err := c.hook.After(hookCtx); err != nil {
			return nil, err
		}

		return rows, err
	}
	return nil, errors.New("driver does not support query context")
}

// Ping 检测数据库连接是否可用并在操作前后执行 Hook。
//
// 参数：
//   - ctx: 控制 Ping 操作生命周期的上下文。
//
// 返回：
//   - error: 底层连接不支持 driver.Pinger、Hook.Before 返回错误、底层 Ping 失败或 Hook.After 返回错误时返回错误。
func (c *kitConn) Ping(ctx context.Context) error {
	// 检查原始连接是否支持 Ping。
	if pinger, ok := c.Conn.(driver.Pinger); ok {
		// 创建 ping 操作的钩子上下文。
		hookCtx := NewHookContext(ctx, OpPing, "", nil)

		// 执行前置钩子。
		if err := c.hook.Before(hookCtx); err != nil {
			return err
		}

		// 调用原始连接的 Ping 方法。
		err := pinger.Ping(ctx)
		// 设置操作结果。
		hookCtx.SetResult(nil, err)

		// 执行后置钩子。
		if err := c.hook.After(hookCtx); err != nil {
			return err
		}

		return err
	}
	return errors.New("driver does not support ping")
}

// BeginTx 开始一个新事务并返回带 Hook 包装的事务。
//
// 参数：
//   - ctx: 控制开启事务操作生命周期的上下文。
//   - opts: 事务选项，包含隔离级别和只读标志。
//
// 返回：
//   - driver.Tx: 成功时返回带 Hook 包装的事务。
//   - error: 底层连接不支持 driver.ConnBeginTx、Hook.Before 返回错误、底层开启事务失败或 Hook.After 返回错误时返回错误。
func (c *kitConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	// 检查原始连接是否支持 BeginTx。
	if beginner, ok := c.Conn.(driver.ConnBeginTx); ok {
		// 创建开始事务的钩子上下文。
		hookCtx := NewHookContext(ctx, OpBegin, "", nil)

		// 执行前置钩子。
		if err := c.hook.Before(hookCtx); err != nil {
			return nil, err
		}

		// 调用原始连接的 BeginTx 方法。
		tx, err := beginner.BeginTx(ctx, opts)
		// 设置操作结果。
		hookCtx.SetResult(tx, err)

		// 执行后置钩子。
		if err := c.hook.After(hookCtx); err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}

		// 返回包装后的事务实例。
		return &kitTx{
			Tx:   tx,
			hook: c.hook,
		}, nil
	}
	return nil, errors.New("driver does not support begin tx")
}

// CheckNamedValue 检查命名参数值是否可被底层驱动接受。
//
// 参数：
//   - nv: 待检查的命名参数值。
//
// 返回：
//   - error: 底层检查返回的错误；底层连接不支持 driver.NamedValueChecker 时返回 driver.ErrSkip。
func (c *kitConn) CheckNamedValue(nv *driver.NamedValue) error {
	// 检查原始连接是否支持 CheckNamedValue。
	if checker, ok := c.Conn.(driver.NamedValueChecker); ok {
		return checker.CheckNamedValue(nv)
	}
	return driver.ErrSkip
}

// ResetSession 重置底层连接的会话状态。
//
// 参数：
//   - ctx: 控制重置会话操作生命周期的上下文。
//
// 返回：
//   - error: 底层重置失败时返回错误；底层连接不支持 driver.SessionResetter 时返回 nil。
func (c *kitConn) ResetSession(ctx context.Context) error {
	// 检查原始连接是否支持 ResetSession。
	if resetter, ok := c.Conn.(driver.SessionResetter); ok {
		return resetter.ResetSession(ctx)
	}
	return nil
}

// kitStmt 包装底层 driver.Stmt，并为预处理语句操作执行 Hook。
//
// kitStmt 保留预处理 SQL 文本，用于 ExecContext、QueryContext 和 Close 的 HookContext。
type kitStmt struct {
	// 原始预处理语句实例。
	driver.Stmt
	// 用于执行钩子操作的接口实例。
	hook Hook
	// query 是预处理 SQL 语句文本。
	query string
}

// ExecContext 执行预处理语句并在操作前后执行 Hook。
//
// 参数：
//   - ctx: 控制执行操作生命周期的上下文。
//   - args: SQL 语句的命名参数列表；没有参数时可为 nil。
//
// 返回：
//   - driver.Result: 底层语句返回的执行结果。
//   - error: 底层语句不支持 driver.StmtExecContext、Hook.Before 返回错误、底层执行失败或 Hook.After 返回错误时返回错误。
func (s *kitStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	// 检查原始语句是否支持 ExecContext。
	if execer, ok := s.Stmt.(driver.StmtExecContext); ok {
		// 创建执行预处理语句的钩子上下文。
		hookCtx := NewHookContext(ctx, OpStmtExec, s.query, args)

		// 执行前置钩子。
		if err := s.hook.Before(hookCtx); err != nil {
			return nil, err
		}

		// 调用原始语句的 ExecContext 方法。
		result, err := execer.ExecContext(ctx, args)
		// 设置操作结果。
		hookCtx.SetResult(result, err)

		// 执行后置钩子。
		if err := s.hook.After(hookCtx); err != nil {
			return nil, err
		}

		return result, err
	}
	return nil, errors.New("stmt does not support exec context")
}

// QueryContext 执行预处理查询并在操作前后执行 Hook。
//
// 参数：
//   - ctx: 控制查询操作生命周期的上下文。
//   - args: SQL 语句的命名参数列表；没有参数时可为 nil。
//
// 返回：
//   - driver.Rows: 底层语句返回的查询结果集。
//   - error: 底层语句不支持 driver.StmtQueryContext、Hook.Before 返回错误、底层查询失败或 Hook.After 返回错误时返回错误。
func (s *kitStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	// 检查原始语句是否支持 QueryContext。
	if queryer, ok := s.Stmt.(driver.StmtQueryContext); ok {
		// 创建查询预处理语句的钩子上下文。
		hookCtx := NewHookContext(ctx, OpStmtQuery, s.query, args)

		// 执行前置钩子。
		if err := s.hook.Before(hookCtx); err != nil {
			return nil, err
		}

		// 调用原始语句的 QueryContext 方法。
		rows, err := queryer.QueryContext(ctx, args)
		// 设置操作结果。
		hookCtx.SetResult(rows, err)

		// 执行后置钩子。
		if err := s.hook.After(hookCtx); err != nil {
			return nil, err
		}

		return rows, err
	}
	return nil, errors.New("stmt does not support query context")
}

// Close 关闭预处理语句并在操作前后执行 Hook。
//
// 参数：无。
//
// 返回：
//   - error: Hook.Before 返回错误、底层语句关闭失败或 Hook.After 返回错误时返回错误。
func (s *kitStmt) Close() error {
	// 创建关闭预处理语句的钩子上下文。
	hookCtx := NewHookContext(context.Background(), OpStmtClose, s.query, nil)

	// 执行前置钩子。
	if err := s.hook.Before(hookCtx); err != nil {
		return err
	}

	// 调用原始语句的 Close 方法。
	err := s.Stmt.Close()
	// 设置操作结果。
	hookCtx.SetResult(nil, err)

	// 执行后置钩子。
	if err := s.hook.After(hookCtx); err != nil {
		return err
	}

	return err
}

// kitTx 包装底层 driver.Tx，并为提交和回滚操作执行 Hook。
type kitTx struct {
	// 原始事务实例。
	driver.Tx
	// 用于执行钩子操作的接口实例。
	hook Hook
}

// Commit 提交事务并在操作前后执行 Hook。
//
// 参数：无。
//
// 返回：
//   - error: Hook.Before 返回错误、底层事务提交失败或 Hook.After 返回错误时返回错误。
func (t *kitTx) Commit() error {
	// 创建提交事务的钩子上下文。
	hookCtx := NewHookContext(context.Background(), OpCommit, "", nil)

	// 执行前置钩子。
	if err := t.hook.Before(hookCtx); err != nil {
		return err
	}

	// 调用原始事务的 Commit 方法。
	err := t.Tx.Commit()
	// 设置操作结果。
	hookCtx.SetResult(nil, err)

	// 执行后置钩子。
	if err := t.hook.After(hookCtx); err != nil {
		return err
	}

	return err
}

// Rollback 回滚事务并在操作前后执行 Hook。
//
// 参数：无。
//
// 返回：
//   - error: Hook.Before 返回错误、底层事务回滚失败或 Hook.After 返回错误时返回错误。
func (t *kitTx) Rollback() error {
	// 创建回滚事务的钩子上下文。
	hookCtx := NewHookContext(context.Background(), OpRollback, "", nil)

	// 执行前置钩子。
	if err := t.hook.Before(hookCtx); err != nil {
		return err
	}

	// 调用原始事务的 Rollback 方法。
	err := t.Tx.Rollback()
	// 设置操作结果。
	hookCtx.SetResult(nil, err)

	// 执行后置钩子。
	if err := t.hook.After(hookCtx); err != nil {
		return err
	}

	return err
}
