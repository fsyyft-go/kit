// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package driver

import (
	"context"
	"database/sql/driver"
	"errors"
)

// 声明接口实现检查，确保类型实现了所有必需的接口。
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

// KitDriver 是一个数据库驱动包装器，用于添加钩子功能。
type KitDriver struct {
	// 原始数据库驱动实例。
	driver driver.Driver
	// 用于执行钩子操作的接口实例。
	hook Hook
}

// NewKitDriver 创建一个新的 KitDriver 实例。
//
// 参数：
//   - d：原始数据库驱动实例，用于执行实际的数据库操作。
//   - h：钩子接口实例，用于在数据库操作前后执行自定义逻辑。
//
// 返回值：
//   - *KitDriver：返回一个新创建的 KitDriver 实例。
func NewKitDriver(d driver.Driver, h Hook) *KitDriver {
	return &KitDriver{
		driver: d,
		hook:   h,
	}
}

// Open 实现 driver.Driver 接口，用于创建数据库连接。
//
// 参数：
//   - name：数据源名称（DSN），包含数据库连接所需的配置信息。
//
// 返回值：
//   - driver.Conn：数据库连接接口。
//   - error：如果连接过程中发生错误，返回相应的错误信息。
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

// kitConn 是一个数据库连接包装器，实现了多个数据库连接相关的接口。
type kitConn struct {
	// 原始数据库连接实例。
	driver.Conn
	// 用于执行钩子操作的接口实例。
	hook Hook
}

// PrepareContext 实现 driver.ConnPrepareContext 接口，用于创建预处理语句。
//
// 参数：
//   - ctx：上下文对象，用于控制操作的生命周期。
//   - query：预处理的 SQL 查询语句。
//
// 返回值：
//   - driver.Stmt：预处理语句接口。
//   - error：如果预处理过程中发生错误，返回相应的错误信息。
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

// ExecContext 实现 driver.ExecerContext 接口，用于执行 SQL 语句。
//
// 参数：
//   - ctx：上下文对象，用于控制操作的生命周期。
//   - query：要执行的 SQL 语句。
//   - args：SQL 语句的参数列表，使用命名参数形式。
//
// 返回值：
//   - driver.Result：执行结果接口，包含影响的行数等信息。
//   - error：如果执行过程中发生错误，返回相应的错误信息。
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

// QueryContext 实现 driver.QueryerContext 接口，用于执行查询操作。
//
// 参数：
//   - ctx：上下文对象，用于控制操作的生命周期。
//   - query：要执行的查询 SQL 语句。
//   - args：SQL 语句的参数列表，使用命名参数形式。
//
// 返回值：
//   - driver.Rows：查询结果集接口。
//   - error：如果查询过程中发生错误，返回相应的错误信息。
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

// Ping 实现 driver.Pinger 接口，用于检测数据库连接是否可用。
//
// 参数：
//   - ctx：上下文对象，用于控制操作的生命周期。
//
// 返回值：
//   - error：如果连接不可用或操作过程中发生错误，返回相应的错误信息。
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

// BeginTx 实现 driver.ConnBeginTx 接口，用于开始一个新的事务。
//
// 参数：
//   - ctx：上下文对象，用于控制操作的生命周期。
//   - opts：事务选项，包含隔离级别等配置信息。
//
// 返回值：
//   - driver.Tx：事务接口。
//   - error：如果开始事务过程中发生错误，返回相应的错误信息。
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

// CheckNamedValue 实现 driver.NamedValueChecker 接口，用于检查命名参数值。
//
// 参数：
//   - nv：命名参数值对象，包含参数名称和值。
//
// 返回值：
//   - error：如果参数值不合法，返回相应的错误信息；如果不支持检查，返回 driver.ErrSkip。
func (c *kitConn) CheckNamedValue(nv *driver.NamedValue) error {
	// 检查原始连接是否支持 CheckNamedValue。
	if checker, ok := c.Conn.(driver.NamedValueChecker); ok {
		return checker.CheckNamedValue(nv)
	}
	return driver.ErrSkip
}

// ResetSession 实现 driver.SessionResetter 接口，用于重置会话状态。
//
// 参数：
//   - ctx：上下文对象，用于控制操作的生命周期。
//
// 返回值：
//   - error：如果重置会话过程中发生错误，返回相应的错误信息。
func (c *kitConn) ResetSession(ctx context.Context) error {
	// 检查原始连接是否支持 ResetSession。
	if resetter, ok := c.Conn.(driver.SessionResetter); ok {
		return resetter.ResetSession(ctx)
	}
	return nil
}

// kitStmt 是一个预处理语句包装器，用于在执行预处理语句时添加钩子功能。
type kitStmt struct {
	// 原始预处理语句实例。
	driver.Stmt
	// 用于执行钩子操作的接口实例。
	hook Hook
	// SQL 查询语句。
	query string
}

// ExecContext 实现预处理语句的执行操作。
//
// 参数：
//   - ctx：上下文对象，用于控制操作的生命周期。
//   - args：SQL 语句的参数列表，使用命名参数形式。
//
// 返回值：
//   - driver.Result：执行结果接口，包含影响的行数等信息。
//   - error：如果执行过程中发生错误，返回相应的错误信息。
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

// QueryContext 实现预处理语句的查询操作。
//
// 参数：
//   - ctx：上下文对象，用于控制操作的生命周期。
//   - args：SQL 语句的参数列表，使用命名参数形式。
//
// 返回值：
//   - driver.Rows：查询结果集接口。
//   - error：如果查询过程中发生错误，返回相应的错误信息。
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

// Close 实现预处理语句的关闭操作。
//
// 返回值：
//   - error：如果关闭过程中发生错误，返回相应的错误信息。
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

// kitTx 是一个事务包装器，用于在事务操作时添加钩子功能。
type kitTx struct {
	// 原始事务实例。
	driver.Tx
	// 用于执行钩子操作的接口实例。
	hook Hook
}

// Commit 实现事务的提交操作。
//
// 返回值：
//   - error：如果提交过程中发生错误，返回相应的错误信息。
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

// Rollback 实现事务的回滚操作。
//
// 返回值：
//   - error：如果回滚过程中发生错误，返回相应的错误信息。
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
