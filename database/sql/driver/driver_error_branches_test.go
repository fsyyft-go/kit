// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package driver

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKitConn_PrepareContextErrorBranches 验证连接预处理操作的错误分支。
//
// 该测试通过表驱动用例覆盖前置 Hook 短路、原始 PrepareContext 错误和后置 Hook 覆盖错误，确保预处理包装层错误优先级稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestKitConn_PrepareContextErrorBranches(t *testing.T) {
	beforeErr := errors.New("prepare before failed")
	originErr := errors.New("prepare origin failed")
	afterErr := errors.New("prepare after failed")
	preparedStmt := &testFullStmt{}

	tests := []struct {
		name             string
		description      string
		giveBeforeErr    error
		giveOriginErr    error
		giveAfterErr     error
		wantErrIs        error
		wantOriginCalled bool
		wantAfterCalled  bool
	}{
		{
			name:          "error/before-stops-prepare",
			description:   "验证预处理前置 Hook 错误会阻止底层 PrepareContext 调用。",
			giveBeforeErr: beforeErr,
			wantErrIs:     beforeErr,
		},
		{
			name:             "error/origin-prepare-error-returned",
			description:      "验证底层 PrepareContext 错误会进入 HookContext 并返回给调用方。",
			giveOriginErr:    originErr,
			wantErrIs:        originErr,
			wantOriginCalled: true,
			wantAfterCalled:  true,
		},
		{
			name:             "error/after-overrides-prepare-success",
			description:      "验证预处理后置 Hook 错误会覆盖底层 PrepareContext 的成功结果。",
			giveAfterErr:     afterErr,
			wantErrIs:        afterErr,
			wantOriginCalled: true,
			wantAfterCalled:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			var originCalled bool
			var afterCalled bool
			hook := &recordingHook{
				name:      "prepare",
				beforeErr: tt.giveBeforeErr,
				afterErr:  tt.giveAfterErr,
				afterFn: func(ctx *HookContext) {
					afterCalled = true
					assert.Equal(t, OpPrepare, ctx.OpType())
					assert.Equal(t, "SELECT * FROM users WHERE id=?", ctx.Query())
					assert.Same(t, preparedStmt, ctx.OriginResult())
					if tt.giveOriginErr != nil {
						assert.ErrorIs(t, ctx.OriginError(), tt.giveOriginErr)
					} else {
						assert.NoError(t, ctx.OriginError())
					}
				},
			}
			baseConn := &testFullConn{prepareContextFn: func(ctx context.Context, query string) (driver.Stmt, error) {
				originCalled = true
				assert.Equal(t, "SELECT * FROM users WHERE id=?", query)
				return preparedStmt, tt.giveOriginErr
			}}
			conn := &kitConn{Conn: baseConn, hook: hook}

			got, err := conn.PrepareContext(context.Background(), "SELECT * FROM users WHERE id=?")

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErrIs)
			assert.Nil(t, got)
			assert.Equal(t, tt.wantOriginCalled, originCalled)
			assert.Equal(t, tt.wantAfterCalled, afterCalled)
		})
	}
}

// TestKitConn_QueryPingBeginErrorBranches 验证连接查询、Ping 和开启事务的错误分支。
//
// 该测试通过表驱动用例覆盖 QueryContext、Ping 和 BeginTx 的原始错误与 Hook 覆盖错误，确保连接级可选操作的错误传播稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestKitConn_QueryPingBeginErrorBranches(t *testing.T) {
	beforeErr := errors.New("before failed")
	originErr := errors.New("origin failed")
	afterErr := errors.New("after failed")
	queryRows := &testRows{columns: []string{"id"}}
	originTx := &testTx{}

	tests := []struct {
		name             string
		description      string
		giveBeforeErr    error
		giveOriginErr    error
		giveAfterErr     error
		act              func(conn *kitConn) (interface{}, error)
		wantOp           OpType
		wantErrIs        error
		wantResult       interface{}
		wantOriginCalled *bool
		wantAfterCalled  bool
	}{
		{
			name:          "error/query-before-stops-origin",
			description:   "验证查询前置 Hook 错误会阻止底层 QueryContext 调用。",
			giveBeforeErr: beforeErr,
			act: func(conn *kitConn) (interface{}, error) {
				return conn.QueryContext(context.Background(), "SELECT id FROM users", nil)
			},
			wantOp:    OpQuery,
			wantErrIs: beforeErr,
		},
		{
			name:          "error/query-origin-error-returned-with-rows",
			description:   "验证底层 QueryContext 同时返回结果集和错误时包装层按原样返回。",
			giveOriginErr: originErr,
			act: func(conn *kitConn) (interface{}, error) {
				return conn.QueryContext(context.Background(), "SELECT id FROM users", nil)
			},
			wantOp:          OpQuery,
			wantErrIs:       originErr,
			wantResult:      queryRows,
			wantAfterCalled: true,
		},
		{
			name:         "error/query-after-overrides-success",
			description:  "验证查询后置 Hook 错误会覆盖底层 QueryContext 的成功结果。",
			giveAfterErr: afterErr,
			act: func(conn *kitConn) (interface{}, error) {
				return conn.QueryContext(context.Background(), "SELECT id FROM users", nil)
			},
			wantOp:          OpQuery,
			wantErrIs:       afterErr,
			wantAfterCalled: true,
		},
		{
			name:          "error/ping-before-stops-origin",
			description:   "验证 Ping 前置 Hook 错误会阻止底层 Ping 调用。",
			giveBeforeErr: beforeErr,
			act: func(conn *kitConn) (interface{}, error) {
				return nil, conn.Ping(context.Background())
			},
			wantOp:    OpPing,
			wantErrIs: beforeErr,
		},
		{
			name:          "error/ping-origin-error-returned",
			description:   "验证底层 Ping 错误会进入 HookContext 并返回给调用方。",
			giveOriginErr: originErr,
			act: func(conn *kitConn) (interface{}, error) {
				return nil, conn.Ping(context.Background())
			},
			wantOp:          OpPing,
			wantErrIs:       originErr,
			wantAfterCalled: true,
		},
		{
			name:         "error/ping-after-overrides-success",
			description:  "验证 Ping 后置 Hook 错误会覆盖底层 Ping 的成功结果。",
			giveAfterErr: afterErr,
			act: func(conn *kitConn) (interface{}, error) {
				return nil, conn.Ping(context.Background())
			},
			wantOp:          OpPing,
			wantErrIs:       afterErr,
			wantAfterCalled: true,
		},
		{
			name:          "error/begin-before-stops-origin",
			description:   "验证开启事务前置 Hook 错误会阻止底层 BeginTx 调用。",
			giveBeforeErr: beforeErr,
			act: func(conn *kitConn) (interface{}, error) {
				return conn.BeginTx(context.Background(), driver.TxOptions{})
			},
			wantOp:    OpBegin,
			wantErrIs: beforeErr,
		},
		{
			name:          "error/begin-origin-error-returned",
			description:   "验证底层 BeginTx 错误会进入 HookContext 并返回给调用方。",
			giveOriginErr: originErr,
			act: func(conn *kitConn) (interface{}, error) {
				return conn.BeginTx(context.Background(), driver.TxOptions{})
			},
			wantOp:          OpBegin,
			wantErrIs:       originErr,
			wantAfterCalled: true,
		},
		{
			name:         "error/begin-after-overrides-success",
			description:  "验证开启事务后置 Hook 错误会覆盖底层 BeginTx 的成功结果。",
			giveAfterErr: afterErr,
			act: func(conn *kitConn) (interface{}, error) {
				return conn.BeginTx(context.Background(), driver.TxOptions{})
			},
			wantOp:          OpBegin,
			wantErrIs:       afterErr,
			wantAfterCalled: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			var queryCalled bool
			var pingCalled bool
			var beginCalled bool
			var afterCalled bool
			hook := &recordingHook{
				name:      "conn-errors",
				beforeErr: tt.giveBeforeErr,
				afterErr:  tt.giveAfterErr,
				beforeFn: func(ctx *HookContext) {
					assert.Equal(t, tt.wantOp, ctx.OpType())
				},
				afterFn: func(ctx *HookContext) {
					afterCalled = true
					assert.Equal(t, tt.wantOp, ctx.OpType())
					if tt.giveOriginErr != nil {
						assert.ErrorIs(t, ctx.OriginError(), tt.giveOriginErr)
					} else {
						assert.NoError(t, ctx.OriginError())
					}
				},
			}
			baseConn := &testFullConn{
				queryContextFn: func(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
					queryCalled = true
					return queryRows, tt.giveOriginErr
				},
				pingFn: func(ctx context.Context) error {
					pingCalled = true
					return tt.giveOriginErr
				},
				beginTxFn: func(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
					beginCalled = true
					return originTx, tt.giveOriginErr
				},
			}
			conn := &kitConn{Conn: baseConn, hook: hook}

			got, err := tt.act(conn)

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErrIs)
			assert.Equal(t, tt.wantResult, got)
			assert.Equal(t, tt.wantAfterCalled, afterCalled)
			switch tt.wantOp {
			case OpQuery:
				assert.Equal(t, tt.giveBeforeErr == nil, queryCalled)
			case OpPing:
				assert.Equal(t, tt.giveBeforeErr == nil, pingCalled)
			case OpBegin:
				assert.Equal(t, tt.giveBeforeErr == nil, beginCalled)
			}
		})
	}
}

// TestKitStmt_QueryCloseAndTxAdditionalErrorBranches 验证语句查询、语句关闭和事务操作的补充错误分支。
//
// 该测试覆盖 Stmt QueryContext、Stmt Close、Commit 和 Rollback 的前置 Hook、原始错误与后置 Hook 分支，确保剩余包装层错误行为稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestKitStmt_QueryCloseAndTxAdditionalErrorBranches(t *testing.T) {
	beforeErr := errors.New("before failed")
	originErr := errors.New("origin failed")
	afterErr := errors.New("after failed")
	queryRows := &testRows{columns: []string{"name"}}

	tests := []struct {
		name             string
		description      string
		giveBeforeErr    error
		giveOriginErr    error
		giveAfterErr     error
		act              func(hook Hook) (interface{}, *testTx, *bool, error)
		wantOp           OpType
		wantErrIs        error
		wantResult       interface{}
		wantAfterCalled  bool
		wantOriginCalled bool
	}{
		{
			name:          "error/stmt-query-before-stops-origin",
			description:   "验证预处理查询前置 Hook 错误会阻止底层 QueryContext 调用。",
			giveBeforeErr: beforeErr,
			act: func(hook Hook) (interface{}, *testTx, *bool, error) {
				originCalled := false
				stmt := &kitStmt{Stmt: &testFullStmt{queryContextFn: func(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
					originCalled = true
					return queryRows, nil
				}}, hook: hook, query: "SELECT name FROM users"}
				got, err := stmt.QueryContext(context.Background(), nil)
				return got, nil, &originCalled, err
			},
			wantOp:    OpStmtQuery,
			wantErrIs: beforeErr,
		},
		{
			name:          "error/stmt-query-origin-error-returned-with-rows",
			description:   "验证预处理查询同时返回结果集和错误时包装层按原样返回。",
			giveOriginErr: originErr,
			act: func(hook Hook) (interface{}, *testTx, *bool, error) {
				originCalled := false
				stmt := &kitStmt{Stmt: &testFullStmt{queryContextFn: func(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
					originCalled = true
					return queryRows, originErr
				}}, hook: hook, query: "SELECT name FROM users"}
				got, err := stmt.QueryContext(context.Background(), nil)
				return got, nil, &originCalled, err
			},
			wantOp:           OpStmtQuery,
			wantErrIs:        originErr,
			wantResult:       queryRows,
			wantAfterCalled:  true,
			wantOriginCalled: true,
		},
		{
			name:         "error/stmt-query-after-overrides-success",
			description:  "验证预处理查询后置 Hook 错误会覆盖底层 QueryContext 成功结果。",
			giveAfterErr: afterErr,
			act: func(hook Hook) (interface{}, *testTx, *bool, error) {
				originCalled := false
				stmt := &kitStmt{Stmt: &testFullStmt{queryContextFn: func(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
					originCalled = true
					return queryRows, nil
				}}, hook: hook, query: "SELECT name FROM users"}
				got, err := stmt.QueryContext(context.Background(), nil)
				return got, nil, &originCalled, err
			},
			wantOp:           OpStmtQuery,
			wantErrIs:        afterErr,
			wantAfterCalled:  true,
			wantOriginCalled: true,
		},
		{
			name:          "error/stmt-close-before-stops-origin",
			description:   "验证语句关闭前置 Hook 错误会阻止底层 Close 调用。",
			giveBeforeErr: beforeErr,
			act: func(hook Hook) (interface{}, *testTx, *bool, error) {
				originCalled := false
				stmt := &kitStmt{Stmt: &testFullStmt{closeFn: func() error { originCalled = true; return nil }}, hook: hook, query: "SELECT 1"}
				return nil, nil, &originCalled, stmt.Close()
			},
			wantOp:    OpStmtClose,
			wantErrIs: beforeErr,
		},
		{
			name:          "error/stmt-close-origin-error-returned",
			description:   "验证底层 Close 错误会进入 HookContext 并返回给调用方。",
			giveOriginErr: originErr,
			act: func(hook Hook) (interface{}, *testTx, *bool, error) {
				originCalled := false
				stmt := &kitStmt{Stmt: &testFullStmt{closeFn: func() error { originCalled = true; return originErr }}, hook: hook, query: "SELECT 1"}
				return nil, nil, &originCalled, stmt.Close()
			},
			wantOp:           OpStmtClose,
			wantErrIs:        originErr,
			wantAfterCalled:  true,
			wantOriginCalled: true,
		},
		{
			name:         "error/stmt-close-after-overrides-success",
			description:  "验证语句关闭后置 Hook 错误会覆盖底层 Close 成功结果。",
			giveAfterErr: afterErr,
			act: func(hook Hook) (interface{}, *testTx, *bool, error) {
				originCalled := false
				stmt := &kitStmt{Stmt: &testFullStmt{closeFn: func() error { originCalled = true; return nil }}, hook: hook, query: "SELECT 1"}
				return nil, nil, &originCalled, stmt.Close()
			},
			wantOp:           OpStmtClose,
			wantErrIs:        afterErr,
			wantAfterCalled:  true,
			wantOriginCalled: true,
		},
		{
			name:          "error/commit-before-stops-origin",
			description:   "验证 Commit 前置 Hook 错误会阻止底层事务提交。",
			giveBeforeErr: beforeErr,
			act: func(hook Hook) (interface{}, *testTx, *bool, error) {
				tx := &testTx{}
				return nil, tx, nil, (&kitTx{Tx: tx, hook: hook}).Commit()
			},
			wantOp:    OpCommit,
			wantErrIs: beforeErr,
		},
		{
			name:         "error/commit-after-overrides-success",
			description:  "验证 Commit 后置 Hook 错误会覆盖底层提交成功结果。",
			giveAfterErr: afterErr,
			act: func(hook Hook) (interface{}, *testTx, *bool, error) {
				tx := &testTx{}
				return nil, tx, nil, (&kitTx{Tx: tx, hook: hook}).Commit()
			},
			wantOp:           OpCommit,
			wantErrIs:        afterErr,
			wantAfterCalled:  true,
			wantOriginCalled: true,
		},
		{
			name:          "error/rollback-before-stops-origin",
			description:   "验证 Rollback 前置 Hook 错误会阻止底层事务回滚。",
			giveBeforeErr: beforeErr,
			act: func(hook Hook) (interface{}, *testTx, *bool, error) {
				tx := &testTx{}
				return nil, tx, nil, (&kitTx{Tx: tx, hook: hook}).Rollback()
			},
			wantOp:    OpRollback,
			wantErrIs: beforeErr,
		},
		{
			name:          "error/rollback-origin-error-returned",
			description:   "验证 Rollback 底层错误会进入 HookContext 并返回给调用方。",
			giveOriginErr: originErr,
			act: func(hook Hook) (interface{}, *testTx, *bool, error) {
				tx := &testTx{rollbackErr: originErr}
				return nil, tx, nil, (&kitTx{Tx: tx, hook: hook}).Rollback()
			},
			wantOp:           OpRollback,
			wantErrIs:        originErr,
			wantAfterCalled:  true,
			wantOriginCalled: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			var afterCalled bool
			hook := &recordingHook{
				name:      "stmt-tx-errors",
				beforeErr: tt.giveBeforeErr,
				afterErr:  tt.giveAfterErr,
				beforeFn: func(ctx *HookContext) {
					assert.Equal(t, tt.wantOp, ctx.OpType())
				},
				afterFn: func(ctx *HookContext) {
					afterCalled = true
					assert.Equal(t, tt.wantOp, ctx.OpType())
					if tt.giveOriginErr != nil {
						assert.ErrorIs(t, ctx.OriginError(), tt.giveOriginErr)
					} else {
						assert.NoError(t, ctx.OriginError())
					}
				},
			}

			got, tx, originCalled, err := tt.act(hook)

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErrIs)
			assert.Equal(t, tt.wantResult, got)
			assert.Equal(t, tt.wantAfterCalled, afterCalled)
			if originCalled != nil {
				assert.Equal(t, tt.wantOriginCalled, *originCalled)
			}
			if tx != nil {
				switch tt.wantOp {
				case OpCommit:
					assert.Equal(t, tt.wantOriginCalled, tx.commitCalled)
				case OpRollback:
					assert.Equal(t, tt.wantOriginCalled, tx.rollbackCalled)
				}
			}
		})
	}
}
