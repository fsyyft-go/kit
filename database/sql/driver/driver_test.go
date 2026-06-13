// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package driver

import (
	"context"
	"database/sql/driver"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKitDriver_Open 验证 KitDriver.Open 对原始驱动和 Hook 的编排行为。
//
// 该测试通过表驱动用例覆盖连接成功、前置 Hook 错误、原始驱动错误和后置 Hook 错误，确保连接操作的错误传播与 Hook 上下文稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestKitDriver_Open(t *testing.T) {
	beforeErr := errors.New("before open failed")
	driverErr := errors.New("driver open failed")
	afterErr := errors.New("after open failed")

	tests := []struct {
		name          string
		description   string
		giveBeforeErr error
		giveDriverErr error
		giveAfterErr  error
		wantErrIs     error
		wantOpenCalls int
		wantAfter     bool
		wantWrapped   bool
	}{
		{name: "success/wraps-opened-connection", description: "验证连接成功时返回包装连接并完整执行 Before 与 After。", wantOpenCalls: 1, wantAfter: true, wantWrapped: true},
		{name: "error/before-stops-open", description: "验证前置 Hook 返回错误时不会调用原始驱动 Open。", giveBeforeErr: beforeErr, wantErrIs: beforeErr},
		{name: "error/original-driver-error", description: "验证原始驱动 Open 错误会记录到 HookContext 并向调用方返回。", giveDriverErr: driverErr, wantErrIs: driverErr, wantOpenCalls: 1, wantAfter: true},
		{name: "error/after-overrides-open-result", description: "验证后置 Hook 返回错误时该错误优先返回给调用方。", giveAfterErr: afterErr, wantErrIs: afterErr, wantOpenCalls: 1, wantAfter: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			openedConn := &testFullConn{}
			fakeDriver := &testDriver{conn: openedConn, err: tt.giveDriverErr}
			var afterObserved bool
			hook := &recordingHook{
				name:      "open",
				beforeErr: tt.giveBeforeErr,
				afterErr:  tt.giveAfterErr,
				beforeFn: func(ctx *HookContext) {
					assert.Equal(t, OpConnect, ctx.OpType())
					assert.Empty(t, ctx.Query())
					assert.Empty(t, ctx.Args())
				},
				afterFn: func(ctx *HookContext) {
					afterObserved = true
					assert.Equal(t, OpConnect, ctx.OpType())
					assert.Same(t, openedConn, ctx.OriginResult())
					if tt.giveDriverErr != nil {
						assert.ErrorIs(t, ctx.OriginError(), tt.giveDriverErr)
					} else {
						assert.NoError(t, ctx.OriginError())
					}
				},
			}
			kitDriver := NewKitDriver(fakeDriver, hook)

			gotConn, err := kitDriver.Open("dsn-name")

			if tt.wantErrIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErrIs)
				assert.Nil(t, gotConn)
			} else {
				require.NoError(t, err)
				require.NotNil(t, gotConn)
			}
			if tt.wantOpenCalls == 0 {
				assert.Empty(t, fakeDriver.openNames)
			} else {
				assert.Equal(t, []string{"dsn-name"}[:tt.wantOpenCalls], fakeDriver.openNames)
			}
			assert.Equal(t, tt.wantAfter, afterObserved)
			if tt.wantWrapped {
				wrapped, ok := gotConn.(*kitConn)
				require.True(t, ok)
				assert.Same(t, openedConn, wrapped.Conn)
				assert.Same(t, hook, wrapped.hook)
			}
		})
	}
}

// TestKitConn_ContextOperations 验证 kitConn 对上下文数据库操作的 Hook 包装行为。
//
// 该测试覆盖 PrepareContext、ExecContext、QueryContext、Ping 和 BeginTx 的成功路径，确保原始结果被返回且 HookContext 包含操作类型、SQL 和参数。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestKitConn_ContextOperations(t *testing.T) {
	execArgs := []driver.NamedValue{{Ordinal: 1, Value: "alice"}}
	queryArgs := []driver.NamedValue{{Ordinal: 1, Value: int64(7)}}
	execResult := driver.RowsAffected(2)
	queryRows := &testRows{columns: []string{"id"}}
	preparedStmt := &testFullStmt{}
	begunTx := &testTx{}

	tests := []struct {
		name        string
		description string
		act         func(t *testing.T, conn *kitConn) (interface{}, error)
		wantOp      OpType
		wantQuery   string
		wantArgs    []driver.NamedValue
		wantResult  interface{}
	}{
		{name: "success/prepare-context", description: "验证 PrepareContext 创建预处理语句并返回携带相同 Hook 的包装语句。", act: func(t *testing.T, conn *kitConn) (interface{}, error) {
			return conn.PrepareContext(context.Background(), "SELECT * FROM users WHERE id=?")
		}, wantOp: OpPrepare, wantQuery: "SELECT * FROM users WHERE id=?", wantResult: preparedStmt},
		{name: "success/exec-context", description: "验证 ExecContext 传递 SQL 和命名参数并返回原始执行结果。", act: func(t *testing.T, conn *kitConn) (interface{}, error) {
			return conn.ExecContext(context.Background(), "UPDATE users SET name=?", execArgs)
		}, wantOp: OpExec, wantQuery: "UPDATE users SET name=?", wantArgs: execArgs, wantResult: execResult},
		{name: "success/query-context", description: "验证 QueryContext 传递 SQL 和命名参数并返回原始查询结果集。", act: func(t *testing.T, conn *kitConn) (interface{}, error) {
			return conn.QueryContext(context.Background(), "SELECT id FROM users WHERE age>?", queryArgs)
		}, wantOp: OpQuery, wantQuery: "SELECT id FROM users WHERE age>?", wantArgs: queryArgs, wantResult: queryRows},
		{name: "success/ping", description: "验证 Ping 成功时记录 Ping 操作且结果为空。", act: func(t *testing.T, conn *kitConn) (interface{}, error) { return nil, conn.Ping(context.Background()) }, wantOp: OpPing},
		{name: "success/begin-tx", description: "验证 BeginTx 成功时返回携带相同 Hook 的包装事务。", act: func(t *testing.T, conn *kitConn) (interface{}, error) {
			return conn.BeginTx(context.Background(), driver.TxOptions{Isolation: driver.IsolationLevel(1), ReadOnly: true})
		}, wantOp: OpBegin, wantResult: begunTx},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			var beforeObserved bool
			var afterObserved bool
			hook := &recordingHook{
				name: "conn",
				beforeFn: func(ctx *HookContext) {
					beforeObserved = true
					assert.Equal(t, tt.wantOp, ctx.OpType())
					assert.Equal(t, tt.wantQuery, ctx.Query())
					assert.Equal(t, tt.wantArgs, ctx.Args())
				},
				afterFn: func(ctx *HookContext) {
					afterObserved = true
					assert.Equal(t, tt.wantOp, ctx.OpType())
					assert.Equal(t, tt.wantQuery, ctx.Query())
					assert.Equal(t, tt.wantArgs, ctx.Args())
					assert.NoError(t, ctx.OriginError())
					assert.Equal(t, tt.wantResult, ctx.OriginResult())
				},
			}
			baseConn := &testFullConn{
				prepareContextFn: func(ctx context.Context, query string) (driver.Stmt, error) {
					assert.Equal(t, "SELECT * FROM users WHERE id=?", query)
					return preparedStmt, nil
				},
				execContextFn: func(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
					assert.Equal(t, "UPDATE users SET name=?", query)
					assert.Equal(t, execArgs, args)
					return execResult, nil
				},
				queryContextFn: func(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
					assert.Equal(t, "SELECT id FROM users WHERE age>?", query)
					assert.Equal(t, queryArgs, args)
					return queryRows, nil
				},
				pingFn: func(ctx context.Context) error { return nil },
				beginTxFn: func(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
					assert.True(t, opts.ReadOnly)
					assert.Equal(t, driver.IsolationLevel(1), opts.Isolation)
					return begunTx, nil
				},
			}
			conn := &kitConn{Conn: baseConn, hook: hook}

			got, err := tt.act(t, conn)

			require.NoError(t, err)
			assert.True(t, beforeObserved)
			assert.True(t, afterObserved)
			switch tt.wantOp {
			case OpPrepare:
				stmt, ok := got.(*kitStmt)
				require.True(t, ok)
				assert.Same(t, preparedStmt, stmt.Stmt)
				assert.Same(t, hook, stmt.hook)
				assert.Equal(t, tt.wantQuery, stmt.query)
			case OpBegin:
				tx, ok := got.(*kitTx)
				require.True(t, ok)
				assert.Same(t, begunTx, tx.Tx)
				assert.Same(t, hook, tx.hook)
			default:
				assert.Equal(t, tt.wantResult, got)
			}
		})
	}
}

// TestKitConn_ContextOperationErrors 验证 kitConn 上下文操作的错误传播优先级。
//
// 该测试通过表驱动用例覆盖前置 Hook 短路、原始操作错误和后置 Hook 覆盖错误，确保错误语义不被包装层吞掉。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestKitConn_ContextOperationErrors(t *testing.T) {
	beforeErr := errors.New("before failed")
	originErr := errors.New("origin failed")
	afterErr := errors.New("after failed")

	tests := []struct {
		name             string
		description      string
		giveBeforeErr    error
		giveOriginErr    error
		giveAfterErr     error
		wantErrIs        error
		wantResult       driver.Result
		wantOriginCalled bool
		wantAfterCalled  bool
	}{
		{name: "error/before-stops-exec", description: "验证前置 Hook 错误会阻止 ExecContext 调用原始连接。", giveBeforeErr: beforeErr, wantErrIs: beforeErr},
		{name: "error/origin-error-returned", description: "验证原始 ExecContext 错误会进入 HookContext 并返回给调用方。", giveOriginErr: originErr, wantErrIs: originErr, wantResult: driver.RowsAffected(1), wantOriginCalled: true, wantAfterCalled: true},
		{name: "error/after-overrides-origin-success", description: "验证后置 Hook 错误会覆盖原始 ExecContext 的成功结果。", giveAfterErr: afterErr, wantErrIs: afterErr, wantOriginCalled: true, wantAfterCalled: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			var originCalled bool
			var afterCalled bool
			hook := &recordingHook{
				name:      "exec",
				beforeErr: tt.giveBeforeErr,
				afterErr:  tt.giveAfterErr,
				afterFn: func(ctx *HookContext) {
					afterCalled = true
					assert.Equal(t, OpExec, ctx.OpType())
					if tt.giveOriginErr != nil {
						assert.ErrorIs(t, ctx.OriginError(), tt.giveOriginErr)
					} else {
						assert.NoError(t, ctx.OriginError())
					}
				},
			}
			baseConn := &testFullConn{execContextFn: func(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
				originCalled = true
				return driver.RowsAffected(1), tt.giveOriginErr
			}}
			conn := &kitConn{Conn: baseConn, hook: hook}

			got, err := conn.ExecContext(context.Background(), "UPDATE users SET active=1", nil)

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErrIs)
			assert.Equal(t, tt.wantResult, got)
			assert.Equal(t, tt.wantOriginCalled, originCalled)
			assert.Equal(t, tt.wantAfterCalled, afterCalled)
		})
	}
}

// TestKitConn_UnsupportedAndDelegatedOptionalInterfaces 验证 kitConn 对可选接口的降级与代理行为。
//
// 该测试覆盖底层连接不支持上下文扩展时的稳定错误，以及 CheckNamedValue 和 ResetSession 对底层实现的代理语义。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestKitConn_UnsupportedAndDelegatedOptionalInterfaces(t *testing.T) {
	t.Run("error/unsupported-context-operations", func(t *testing.T) {
		t.Log("验证底层连接不支持上下文扩展接口时返回可诊断错误。")

		conn := &kitConn{Conn: &testBasicConn{}, hook: &recordingHook{name: "unsupported"}}

		stmt, err := conn.PrepareContext(context.Background(), "SELECT 1")
		assert.Nil(t, stmt)
		assert.EqualError(t, err, "driver does not support prepare context")

		result, err := conn.ExecContext(context.Background(), "UPDATE users SET active=1", nil)
		assert.Nil(t, result)
		assert.EqualError(t, err, "driver does not support exec context")

		rows, err := conn.QueryContext(context.Background(), "SELECT 1", nil)
		assert.Nil(t, rows)
		assert.EqualError(t, err, "driver does not support query context")

		assert.EqualError(t, conn.Ping(context.Background()), "driver does not support ping")

		tx, err := conn.BeginTx(context.Background(), driver.TxOptions{})
		assert.Nil(t, tx)
		assert.EqualError(t, err, "driver does not support begin tx")
	})

	delegatedErr := errors.New("invalid named value")
	tests := []struct {
		name        string
		description string
		conn        *kitConn
		act         func(conn *kitConn) error
		wantErrIs   error
	}{
		{name: "success/check-named-value-delegated", description: "验证底层连接实现 NamedValueChecker 时 CheckNamedValue 代理到底层实现。", wantErrIs: delegatedErr},
		{name: "success/reset-session-delegated", description: "验证底层连接实现 SessionResetter 时 ResetSession 代理到底层实现。"},
		{name: "boundary/check-named-value-unsupported", description: "验证底层连接未实现 NamedValueChecker 时返回 driver.ErrSkip。", conn: &kitConn{Conn: &testBasicConn{}, hook: &recordingHook{name: "checker"}}, act: func(conn *kitConn) error { return conn.CheckNamedValue(&driver.NamedValue{Ordinal: 1, Value: "value"}) }, wantErrIs: driver.ErrSkip},
		{name: "boundary/reset-session-unsupported", description: "验证底层连接未实现 SessionResetter 时 ResetSession 保持空操作成功。", conn: &kitConn{Conn: &testBasicConn{}, hook: &recordingHook{name: "reset"}}, act: func(conn *kitConn) error { return conn.ResetSession(context.Background()) }},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			called := false
			conn := tt.conn
			act := tt.act
			if conn == nil {
				baseConn := &testFullConn{}
				switch tt.name {
				case "success/check-named-value-delegated":
					baseConn.checkNamedValueFn = func(nv *driver.NamedValue) error {
						called = true
						assert.Equal(t, "value", nv.Value)
						return delegatedErr
					}
					act = func(conn *kitConn) error { return conn.CheckNamedValue(&driver.NamedValue{Ordinal: 1, Value: "value"}) }
				case "success/reset-session-delegated":
					baseConn.resetSessionFn = func(ctx context.Context) error {
						called = true
						return nil
					}
					act = func(conn *kitConn) error { return conn.ResetSession(context.Background()) }
				}
				conn = &kitConn{Conn: baseConn, hook: &recordingHook{name: "delegate"}}
			}

			err := act(conn)

			if tt.wantErrIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErrIs)
			} else {
				require.NoError(t, err)
			}
			if tt.conn == nil {
				assert.True(t, called)
			}
		})
	}
}

// TestKitStmt_ContextOperations 验证 kitStmt 对预处理语句上下文操作和关闭操作的 Hook 包装行为。
//
// 该测试覆盖 ExecContext、QueryContext 与 Close 成功路径，确保预处理 SQL、参数、结果和 Hook 调用语义稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestKitStmt_ContextOperations(t *testing.T) {
	stmtArgs := []driver.NamedValue{{Ordinal: 1, Value: "alice"}}
	execResult := driver.RowsAffected(4)
	queryRows := &testRows{columns: []string{"name"}}

	tests := []struct {
		name        string
		description string
		act         func(stmt *kitStmt) (interface{}, error)
		wantOp      OpType
		wantArgs    []driver.NamedValue
		wantResult  interface{}
	}{
		{name: "success/stmt-exec-context", description: "验证预处理语句 ExecContext 传递参数并返回原始执行结果。", act: func(stmt *kitStmt) (interface{}, error) { return stmt.ExecContext(context.Background(), stmtArgs) }, wantOp: OpStmtExec, wantArgs: stmtArgs, wantResult: execResult},
		{name: "success/stmt-query-context", description: "验证预处理语句 QueryContext 传递参数并返回原始查询结果集。", act: func(stmt *kitStmt) (interface{}, error) { return stmt.QueryContext(context.Background(), stmtArgs) }, wantOp: OpStmtQuery, wantArgs: stmtArgs, wantResult: queryRows},
		{name: "success/stmt-close", description: "验证预处理语句 Close 通过 Hook 记录关闭操作并返回原始关闭结果。", act: func(stmt *kitStmt) (interface{}, error) { return nil, stmt.Close() }, wantOp: OpStmtClose},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			var beforeObserved bool
			var afterObserved bool
			hook := &recordingHook{
				name: "stmt",
				beforeFn: func(ctx *HookContext) {
					beforeObserved = true
					assert.Equal(t, tt.wantOp, ctx.OpType())
					assert.Equal(t, "SELECT name FROM users WHERE name=?", ctx.Query())
					assert.Equal(t, tt.wantArgs, ctx.Args())
				},
				afterFn: func(ctx *HookContext) {
					afterObserved = true
					assert.Equal(t, tt.wantOp, ctx.OpType())
					assert.NoError(t, ctx.OriginError())
					assert.Equal(t, tt.wantResult, ctx.OriginResult())
				},
			}
			baseStmt := &testFullStmt{
				execContextFn: func(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
					assert.Equal(t, stmtArgs, args)
					return execResult, nil
				},
				queryContextFn: func(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
					assert.Equal(t, stmtArgs, args)
					return queryRows, nil
				},
				closeFn: func() error { return nil },
			}
			stmt := &kitStmt{Stmt: baseStmt, hook: hook, query: "SELECT name FROM users WHERE name=?"}

			got, err := tt.act(stmt)

			require.NoError(t, err)
			assert.True(t, beforeObserved)
			assert.True(t, afterObserved)
			assert.Equal(t, tt.wantResult, got)
		})
	}
}

// TestKitStmt_ErrorAndUnsupportedOperations 验证 kitStmt 的错误传播和不支持接口分支。
//
// 该测试覆盖前置 Hook 短路、原始语句错误、后置 Hook 覆盖错误，以及底层语句不支持上下文执行或查询时的错误信息。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestKitStmt_ErrorAndUnsupportedOperations(t *testing.T) {
	beforeErr := errors.New("before stmt failed")
	originErr := errors.New("origin stmt failed")
	afterErr := errors.New("after stmt failed")

	tests := []struct {
		name             string
		description      string
		giveBeforeErr    error
		giveOriginErr    error
		giveAfterErr     error
		wantErrIs        error
		wantResult       driver.Result
		wantOriginCalled bool
		wantAfterCalled  bool
	}{
		{name: "error/before-stops-stmt-exec", description: "验证预处理语句前置 Hook 错误会阻止底层 ExecContext。", giveBeforeErr: beforeErr, wantErrIs: beforeErr},
		{name: "error/origin-stmt-exec-error-returned", description: "验证底层预处理 ExecContext 错误会进入 HookContext 并返回给调用方。", giveOriginErr: originErr, wantErrIs: originErr, wantResult: driver.RowsAffected(1), wantOriginCalled: true, wantAfterCalled: true},
		{name: "error/after-overrides-stmt-exec-result", description: "验证预处理语句后置 Hook 错误会覆盖底层 ExecContext 的成功结果。", giveAfterErr: afterErr, wantErrIs: afterErr, wantOriginCalled: true, wantAfterCalled: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			var originCalled bool
			var afterCalled bool
			hook := &recordingHook{
				name:      "stmt-exec",
				beforeErr: tt.giveBeforeErr,
				afterErr:  tt.giveAfterErr,
				afterFn: func(ctx *HookContext) {
					afterCalled = true
					assert.Equal(t, OpStmtExec, ctx.OpType())
					if tt.giveOriginErr != nil {
						assert.ErrorIs(t, ctx.OriginError(), tt.giveOriginErr)
					} else {
						assert.NoError(t, ctx.OriginError())
					}
				},
			}
			baseStmt := &testFullStmt{execContextFn: func(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
				originCalled = true
				return driver.RowsAffected(1), tt.giveOriginErr
			}}
			stmt := &kitStmt{Stmt: baseStmt, hook: hook, query: "UPDATE users SET active=1"}

			got, err := stmt.ExecContext(context.Background(), nil)

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErrIs)
			assert.Equal(t, tt.wantResult, got)
			assert.Equal(t, tt.wantOriginCalled, originCalled)
			assert.Equal(t, tt.wantAfterCalled, afterCalled)
		})
	}

	t.Run("error/unsupported-stmt-context-operations", func(t *testing.T) {
		t.Log("验证底层语句不支持上下文执行或查询时返回可诊断错误。")

		stmt := &kitStmt{Stmt: &testBasicStmt{}, hook: &recordingHook{name: "stmt-unsupported"}, query: "SELECT 1"}

		result, err := stmt.ExecContext(context.Background(), nil)
		assert.Nil(t, result)
		assert.EqualError(t, err, "stmt does not support exec context")

		rows, err := stmt.QueryContext(context.Background(), nil)
		assert.Nil(t, rows)
		assert.EqualError(t, err, "stmt does not support query context")
	})
}

// TestKitTx_CommitAndRollback 验证 kitTx 对事务提交与回滚的 Hook 包装行为。
//
// 该测试通过表驱动用例覆盖 Commit 和 Rollback 的成功与错误传播，确保事务操作的 Hook 顺序和错误优先级稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestKitTx_CommitAndRollback(t *testing.T) {
	originErr := errors.New("origin tx failed")
	afterErr := errors.New("after tx failed")

	tests := []struct {
		name             string
		description      string
		giveOriginErr    error
		giveAfterErr     error
		act              func(tx *kitTx) error
		wantOp           OpType
		wantErrIs        error
		wantOriginCalled func(tx *testTx) bool
	}{
		{name: "success/commit", description: "验证 Commit 成功时执行事务提交并记录提交 Hook。", act: func(tx *kitTx) error { return tx.Commit() }, wantOp: OpCommit, wantOriginCalled: func(tx *testTx) bool { return tx.commitCalled }},
		{name: "success/rollback", description: "验证 Rollback 成功时执行事务回滚并记录回滚 Hook。", act: func(tx *kitTx) error { return tx.Rollback() }, wantOp: OpRollback, wantOriginCalled: func(tx *testTx) bool { return tx.rollbackCalled }},
		{name: "error/commit-origin-error", description: "验证 Commit 底层错误会进入 HookContext 并返回给调用方。", giveOriginErr: originErr, act: func(tx *kitTx) error { return tx.Commit() }, wantOp: OpCommit, wantErrIs: originErr, wantOriginCalled: func(tx *testTx) bool { return tx.commitCalled }},
		{name: "error/rollback-after-overrides-success", description: "验证 Rollback 后置 Hook 错误会覆盖底层成功结果。", giveAfterErr: afterErr, act: func(tx *kitTx) error { return tx.Rollback() }, wantOp: OpRollback, wantErrIs: afterErr, wantOriginCalled: func(tx *testTx) bool { return tx.rollbackCalled }},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			var afterObserved bool
			baseTx := &testTx{commitErr: tt.giveOriginErr, rollbackErr: tt.giveOriginErr}
			hook := &recordingHook{
				name:     "tx",
				afterErr: tt.giveAfterErr,
				beforeFn: func(ctx *HookContext) { assert.Equal(t, tt.wantOp, ctx.OpType()) },
				afterFn: func(ctx *HookContext) {
					afterObserved = true
					assert.Equal(t, tt.wantOp, ctx.OpType())
					if tt.giveOriginErr != nil {
						assert.ErrorIs(t, ctx.OriginError(), tt.giveOriginErr)
					} else {
						assert.NoError(t, ctx.OriginError())
					}
				},
			}
			tx := &kitTx{Tx: baseTx, hook: hook}

			err := tt.act(tx)

			if tt.wantErrIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErrIs)
			} else {
				require.NoError(t, err)
			}
			assert.True(t, afterObserved)
			assert.True(t, tt.wantOriginCalled(baseTx))
		})
	}
}

// testDriver 构造可观测 Open 调用的内存数据库驱动。
//
// 该辅助类型避免测试依赖真实数据库服务，并允许用例注入连接结果或错误。
type testDriver struct {
	conn      driver.Conn
	err       error
	openNames []string
}

// Open 记录数据源名称并返回预设连接或错误。
//
// 参数：
//   - name: 数据源名称，用于断言包装驱动是否按原样传递。
//
// 返回：
//   - driver.Conn: 预设的内存连接。
//   - error: 预设的打开错误。
func (d *testDriver) Open(name string) (driver.Conn, error) {
	d.openNames = append(d.openNames, name)
	return d.conn, d.err
}

// testBasicConn 提供仅满足 driver.Conn 的最小内存连接。
//
// 该辅助类型用于验证 kitConn 在底层连接缺少上下文扩展接口时的降级行为。
type testBasicConn struct{}

// Prepare 返回一个最小预处理语句以满足 driver.Conn 接口。
//
// 参数：
//   - query: SQL 语句，本辅助实现不解释该参数。
//
// 返回：
//   - driver.Stmt: 最小内存语句。
//   - error: 始终为 nil。
func (c *testBasicConn) Prepare(query string) (driver.Stmt, error) { return &testBasicStmt{}, nil }

// Close 满足 driver.Conn 接口并保持空操作成功。
//
// 返回：
//   - error: 始终为 nil。
func (c *testBasicConn) Close() error { return nil }

// Begin 返回一个最小事务以满足 driver.Conn 接口。
//
// 返回：
//   - driver.Tx: 最小内存事务。
//   - error: 始终为 nil。
func (c *testBasicConn) Begin() (driver.Tx, error) { return &testTx{}, nil }

// testFullConn 提供支持上下文扩展接口的内存连接。
//
// 该辅助类型通过函数字段注入行为，便于各用例精确断言包装层调用语义。
type testFullConn struct {
	testBasicConn
	prepareContextFn  func(context.Context, string) (driver.Stmt, error)
	execContextFn     func(context.Context, string, []driver.NamedValue) (driver.Result, error)
	queryContextFn    func(context.Context, string, []driver.NamedValue) (driver.Rows, error)
	pingFn            func(context.Context) error
	beginTxFn         func(context.Context, driver.TxOptions) (driver.Tx, error)
	checkNamedValueFn func(*driver.NamedValue) error
	resetSessionFn    func(context.Context) error
}

// PrepareContext 调用预设函数或返回默认内存语句。
//
// 参数：
//   - ctx: 操作上下文。
//   - query: SQL 语句。
//
// 返回：
//   - driver.Stmt: 预设或默认语句。
//   - error: 预设错误；未设置时为 nil。
func (c *testFullConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if c.prepareContextFn != nil {
		return c.prepareContextFn(ctx, query)
	}
	return &testFullStmt{}, nil
}

// ExecContext 调用预设函数或返回默认影响行数。
//
// 参数：
//   - ctx: 操作上下文。
//   - query: SQL 语句。
//   - args: 命名参数列表。
//
// 返回：
//   - driver.Result: 预设或默认执行结果。
//   - error: 预设错误；未设置时为 nil。
func (c *testFullConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if c.execContextFn != nil {
		return c.execContextFn(ctx, query, args)
	}
	return driver.RowsAffected(1), nil
}

// QueryContext 调用预设函数或返回默认结果集。
//
// 参数：
//   - ctx: 操作上下文。
//   - query: SQL 语句。
//   - args: 命名参数列表。
//
// 返回：
//   - driver.Rows: 预设或默认结果集。
//   - error: 预设错误；未设置时为 nil。
func (c *testFullConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if c.queryContextFn != nil {
		return c.queryContextFn(ctx, query, args)
	}
	return &testRows{}, nil
}

// Ping 调用预设函数或保持空操作成功。
//
// 参数：
//   - ctx: 操作上下文。
//
// 返回：
//   - error: 预设错误；未设置时为 nil。
func (c *testFullConn) Ping(ctx context.Context) error {
	if c.pingFn != nil {
		return c.pingFn(ctx)
	}
	return nil
}

// BeginTx 调用预设函数或返回默认内存事务。
//
// 参数：
//   - ctx: 操作上下文。
//   - opts: 事务选项。
//
// 返回：
//   - driver.Tx: 预设或默认事务。
//   - error: 预设错误；未设置时为 nil。
func (c *testFullConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if c.beginTxFn != nil {
		return c.beginTxFn(ctx, opts)
	}
	return &testTx{}, nil
}

// CheckNamedValue 调用预设函数或保持参数检查成功。
//
// 参数：
//   - nv: 待检查的命名参数。
//
// 返回：
//   - error: 预设错误；未设置时为 nil。
func (c *testFullConn) CheckNamedValue(nv *driver.NamedValue) error {
	if c.checkNamedValueFn != nil {
		return c.checkNamedValueFn(nv)
	}
	return nil
}

// ResetSession 调用预设函数或保持空操作成功。
//
// 参数：
//   - ctx: 操作上下文。
//
// 返回：
//   - error: 预设错误；未设置时为 nil。
func (c *testFullConn) ResetSession(ctx context.Context) error {
	if c.resetSessionFn != nil {
		return c.resetSessionFn(ctx)
	}
	return nil
}

// testBasicStmt 提供仅满足 driver.Stmt 的最小内存语句。
//
// 该辅助类型用于验证 kitStmt 在底层语句缺少上下文扩展接口时的降级行为。
type testBasicStmt struct{}

// Close 满足 driver.Stmt 接口并保持空操作成功。
//
// 返回：
//   - error: 始终为 nil。
func (s *testBasicStmt) Close() error { return nil }

// NumInput 返回固定参数数量以满足 driver.Stmt 接口。
//
// 返回：
//   - int: 返回 -1 表示参数数量不固定。
func (s *testBasicStmt) NumInput() int { return -1 }

// Exec 返回默认影响行数以满足 driver.Stmt 接口。
//
// 参数：
//   - args: 旧式参数列表，本辅助实现不解释该参数。
//
// 返回：
//   - driver.Result: 默认执行结果。
//   - error: 始终为 nil。
func (s *testBasicStmt) Exec(args []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

// Query 返回默认结果集以满足 driver.Stmt 接口。
//
// 参数：
//   - args: 旧式参数列表，本辅助实现不解释该参数。
//
// 返回：
//   - driver.Rows: 默认结果集。
//   - error: 始终为 nil。
func (s *testBasicStmt) Query(args []driver.Value) (driver.Rows, error) { return &testRows{}, nil }

// testFullStmt 提供支持上下文执行与查询的内存语句。
//
// 该辅助类型通过函数字段注入行为，便于各用例精确断言预处理语句包装层调用语义。
type testFullStmt struct {
	testBasicStmt
	execContextFn  func(context.Context, []driver.NamedValue) (driver.Result, error)
	queryContextFn func(context.Context, []driver.NamedValue) (driver.Rows, error)
	closeFn        func() error
}

// ExecContext 调用预设函数或返回默认影响行数。
//
// 参数：
//   - ctx: 操作上下文。
//   - args: 命名参数列表。
//
// 返回：
//   - driver.Result: 预设或默认执行结果。
//   - error: 预设错误；未设置时为 nil。
func (s *testFullStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	if s.execContextFn != nil {
		return s.execContextFn(ctx, args)
	}
	return driver.RowsAffected(1), nil
}

// QueryContext 调用预设函数或返回默认结果集。
//
// 参数：
//   - ctx: 操作上下文。
//   - args: 命名参数列表。
//
// 返回：
//   - driver.Rows: 预设或默认结果集。
//   - error: 预设错误；未设置时为 nil。
func (s *testFullStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if s.queryContextFn != nil {
		return s.queryContextFn(ctx, args)
	}
	return &testRows{}, nil
}

// Close 调用预设函数或保持空操作成功。
//
// 返回：
//   - error: 预设错误；未设置时为 nil。
func (s *testFullStmt) Close() error {
	if s.closeFn != nil {
		return s.closeFn()
	}
	return nil
}

// testTx 提供可观测提交与回滚调用的内存事务。
//
// 该辅助类型用于验证 kitTx 是否调用底层事务并正确传播错误。
type testTx struct {
	commitCalled   bool
	rollbackCalled bool
	commitErr      error
	rollbackErr    error
}

// Commit 记录提交调用并返回预设错误。
//
// 返回：
//   - error: 预设提交错误；未设置时为 nil。
func (tx *testTx) Commit() error {
	tx.commitCalled = true
	return tx.commitErr
}

// Rollback 记录回滚调用并返回预设错误。
//
// 返回：
//   - error: 预设回滚错误；未设置时为 nil。
func (tx *testTx) Rollback() error {
	tx.rollbackCalled = true
	return tx.rollbackErr
}

// testRows 提供最小只读结果集实现。
//
// 该辅助类型用于避免查询相关测试依赖真实数据库服务。
type testRows struct {
	columns []string
}

// Columns 返回预设列名。
//
// 返回：
//   - []string: 预设列名；未设置时返回默认单列。
func (r *testRows) Columns() []string {
	if len(r.columns) > 0 {
		return r.columns
	}
	return []string{"value"}
}

// Close 满足 driver.Rows 接口并保持空操作成功。
//
// 返回：
//   - error: 始终为 nil。
func (r *testRows) Close() error { return nil }

// Next 表示结果集已经耗尽。
//
// 参数：
//   - dest: 目标值切片，本辅助实现不写入任何数据。
//
// 返回：
//   - error: 始终返回 io.EOF 表示没有更多数据。
func (r *testRows) Next(dest []driver.Value) error { return io.EOF }
