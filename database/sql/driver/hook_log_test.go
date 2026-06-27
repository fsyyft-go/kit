// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package driver

import (
	"context"
	"database/sql/driver"
	"errors"
	"sync"
	"testing"
	"time"

	kitlog "github.com/fsyyft-go/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHookLogError_Behavior 验证错误日志 Hook 的构造、早退和日志字段行为。
//
// 该测试通过表驱动用例覆盖无错误不记录、错误记录完整字段和空可选字段不写入，确保错误日志 Hook 不影响数据库操作错误传播。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestHookLogError_Behavior(t *testing.T) {
	originErr := errors.New("exec failed")

	tests := []struct {
		name              string
		description       string
		giveNamespace     string
		giveQuery         string
		giveArgs          []driver.NamedValue
		giveErr           error
		wantLog           bool
		wantNamespace     string
		wantQuery         string
		wantArgs          string
		wantAbsentKeys    []string
		wantMessage       interface{}
		wantBeforeNoError bool
	}{
		{
			name:              "success/no-origin-error-skips-log",
			description:       "验证原始操作没有错误时 After 不记录错误日志。",
			giveNamespace:     "accounting",
			giveQuery:         "UPDATE users SET name=? WHERE id=?",
			giveArgs:          []driver.NamedValue{{Ordinal: 1, Value: "alice"}, {Ordinal: 2, Value: int64(7)}},
			wantBeforeNoError: true,
		},
		{
			name:              "success/origin-error-with-all-fields",
			description:       "验证原始操作错误时记录命名空间、SQL、参数、操作类型和耗时字段。",
			giveNamespace:     "accounting",
			giveQuery:         "UPDATE users SET name=? WHERE id=?",
			giveArgs:          []driver.NamedValue{{Ordinal: 1, Value: "alice"}, {Ordinal: 2, Value: int64(7)}},
			giveErr:           originErr,
			wantLog:           true,
			wantNamespace:     "accounting",
			wantQuery:         "UPDATE users SET name=? WHERE id=?",
			wantArgs:          "alice, 7",
			wantMessage:       originErr,
			wantBeforeNoError: true,
		},
		{
			name:              "boundary/origin-error-omits-empty-optional-fields",
			description:       "验证命名空间、SQL 和参数为空时错误日志仅保留核心字段。",
			giveErr:           originErr,
			wantLog:           true,
			wantAbsentKeys:    []string{"namespace", "query", "args"},
			wantMessage:       originErr,
			wantBeforeNoError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			logger := newCaptureLogger()
			hook := NewHookLogError(tt.giveNamespace, logger)
			ctx := NewHookContext(context.Background(), OpExec, tt.giveQuery, tt.giveArgs)
			ctx.SetResult(driver.RowsAffected(1), tt.giveErr)

			if tt.wantBeforeNoError {
				assert.NoError(t, hook.Before(ctx))
			}
			err := hook.After(ctx)

			require.NoError(t, err)
			if !tt.wantLog {
				assert.Empty(t, logger.snapshotEntries())
				return
			}

			entry := logger.requireEntry(t)
			assert.Equal(t, "error", entry.level)
			assert.Equal(t, tt.wantMessage, entry.message)
			assert.Equal(t, OpExec, entry.fields["operation"])
			assert.IsType(t, time.Duration(0), entry.fields["duration"])
			assert.GreaterOrEqual(t, entry.fields["duration"].(time.Duration), time.Duration(0))
			if tt.wantNamespace != "" {
				assert.Equal(t, tt.wantNamespace, entry.fields["namespace"])
			}
			if tt.wantQuery != "" {
				assert.Equal(t, tt.wantQuery, entry.fields["query"])
			}
			if tt.wantArgs != "" {
				assert.Equal(t, tt.wantArgs, entry.fields["args"])
			}
			for _, key := range tt.wantAbsentKeys {
				assert.NotContains(t, entry.fields, key)
			}
		})
	}
}

// TestHookLogSlow_Behavior 验证慢查询日志 Hook 的构造、阈值和日志字段行为。
//
// 该测试通过表驱动用例覆盖低于阈值不记录、达到阈值记录完整字段和空可选字段省略，确保慢查询 Hook 的边界语义稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestHookLogSlow_Behavior(t *testing.T) {
	tests := []struct {
		name              string
		description       string
		giveNamespace     string
		giveThreshold     time.Duration
		giveQuery         string
		giveArgs          []driver.NamedValue
		wantLog           bool
		wantNamespace     string
		wantQuery         string
		wantArgs          string
		wantAbsentKeys    []string
		wantBeforeNoError bool
	}{
		{
			name:              "success/below-threshold-skips-log",
			description:       "验证操作耗时低于阈值时 After 不记录慢查询日志。",
			giveNamespace:     "billing",
			giveThreshold:     time.Hour,
			giveQuery:         "SELECT id FROM users WHERE name=?",
			giveArgs:          []driver.NamedValue{{Ordinal: 1, Value: "alice"}},
			wantBeforeNoError: true,
		},
		{
			name:              "success/at-or-above-threshold-logs-fields",
			description:       "验证操作耗时达到阈值时记录命名空间、SQL、参数、操作类型和耗时字段。",
			giveNamespace:     "billing",
			giveThreshold:     -time.Nanosecond,
			giveQuery:         "SELECT id FROM users WHERE name=?",
			giveArgs:          []driver.NamedValue{{Ordinal: 1, Value: "alice"}},
			wantLog:           true,
			wantNamespace:     "billing",
			wantQuery:         "SELECT id FROM users WHERE name=?",
			wantArgs:          "alice",
			wantBeforeNoError: true,
		},
		{
			name:              "boundary/log-omits-empty-optional-fields",
			description:       "验证慢查询日志在命名空间、SQL 和参数为空时仅保留核心字段。",
			giveThreshold:     -time.Nanosecond,
			wantLog:           true,
			wantAbsentKeys:    []string{"namespace", "query", "args"},
			wantBeforeNoError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			logger := newCaptureLogger()
			hook := NewHookLogSlow(tt.giveNamespace, logger, tt.giveThreshold)
			ctx := NewHookContext(context.Background(), OpQuery, tt.giveQuery, tt.giveArgs)
			ctx.SetResult(&testRows{}, nil)

			if tt.wantBeforeNoError {
				assert.NoError(t, hook.Before(ctx))
			}
			err := hook.After(ctx)

			require.NoError(t, err)
			if !tt.wantLog {
				assert.Empty(t, logger.snapshotEntries())
				return
			}

			entry := logger.requireEntry(t)
			assert.Equal(t, "warn", entry.level)
			assert.Equal(t, "", entry.message)
			assert.Equal(t, OpQuery, entry.fields["operation"])
			assert.IsType(t, time.Duration(0), entry.fields["duration"])
			assert.GreaterOrEqual(t, entry.fields["duration"].(time.Duration), time.Duration(0))
			if tt.wantNamespace != "" {
				assert.Equal(t, tt.wantNamespace, entry.fields["namespace"])
			}
			if tt.wantQuery != "" {
				assert.Equal(t, tt.wantQuery, entry.fields["query"])
			}
			if tt.wantArgs != "" {
				assert.Equal(t, tt.wantArgs, entry.fields["args"])
			}
			for _, key := range tt.wantAbsentKeys {
				assert.NotContains(t, entry.fields, key)
			}
		})
	}
}

// captureLogEntry 保存一次测试日志调用的级别、字段和消息。
//
// 该辅助结构用于断言 HookLogError 和 HookLogSlow 传递给 logger 的结构化字段。
type captureLogEntry struct {
	level   string
	fields  map[string]interface{}
	message interface{}
}

// captureLogger 是实现 kitlog.Logger 的内存日志记录器。
//
// 该辅助类型通过嵌入 kitlog.Logger 保持接口兼容，并覆盖 Hook 使用的 WithFields、Error 和 Warn 方法。
type captureLogger struct {
	kitlog.Logger
	mu      sync.Mutex
	fields  map[string]interface{}
	entries []captureLogEntry
	done    chan struct{}
	root    *captureLogger
}

// newCaptureLogger 构造用于日志 Hook 测试的内存日志记录器。
//
// 该辅助函数集中初始化同步状态，确保异步日志写入可被测试稳定观测。
//
// 返回：
//   - *captureLogger: 可传入日志 Hook 的内存 logger。
func newCaptureLogger() *captureLogger {
	logger := &captureLogger{done: make(chan struct{}, 8)}
	logger.root = logger
	return logger
}

// WithFields 保存结构化字段并返回派生日志记录器。
//
// 参数：
//   - fields: Hook 传入的结构化日志字段。
//
// 返回：
//   - kitlog.Logger: 携带字段副本的派生日志记录器。
func (l *captureLogger) WithFields(fields map[string]interface{}) kitlog.Logger {
	copied := make(map[string]interface{}, len(fields))
	for key, value := range fields {
		copied[key] = value
	}
	return &captureLogger{Logger: l.Logger, fields: copied, done: l.done, root: l.root}
}

// Error 记录一次 error 级别日志调用。
//
// 参数：
//   - args: Hook 传入的日志消息参数。
func (l *captureLogger) Error(args ...interface{}) {
	l.record("error", firstLogArg(args))
}

// Warn 记录一次 warn 级别日志调用。
//
// 参数：
//   - args: Hook 传入的日志消息参数。
func (l *captureLogger) Warn(args ...interface{}) {
	l.record("warn", firstLogArg(args))
}

// record 保存日志调用并通知等待中的测试。
//
// 参数：
//   - level: 日志级别。
//   - message: 日志消息。
func (l *captureLogger) record(level string, message interface{}) {
	root := l.root
	root.mu.Lock()
	root.entries = append(root.entries, captureLogEntry{level: level, fields: l.fields, message: message})
	root.mu.Unlock()
	root.done <- struct{}{}
}

// entries 返回已记录日志的快照。
//
// 返回：
//   - []captureLogEntry: 当前已记录的日志条目副本。
func (l *captureLogger) snapshotEntries() []captureLogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	copied := make([]captureLogEntry, len(l.entries))
	copy(copied, l.entries)
	return copied
}

// requireEntry 等待并返回唯一一条日志记录。
//
// 参数：
//   - t: 测试上下文，用于报告异步日志未完成或日志数量异常。
//
// 返回：
//   - captureLogEntry: 捕获到的日志条目。
func (l *captureLogger) requireEntry(t *testing.T) captureLogEntry {
	t.Helper()
	select {
	case <-l.done:
	case <-time.After(time.Second):
		require.Fail(t, "未在超时时间内捕获异步日志")
	}
	entries := l.snapshotEntries()
	require.Len(t, entries, 1)
	return entries[0]
}

// firstLogArg 返回日志参数中的第一个消息值。
//
// 参数：
//   - args: logger 接收到的变长参数。
//
// 返回：
//   - interface{}: 第一个参数；没有参数时返回 nil。
func firstLogArg(args []interface{}) interface{} {
	if len(args) == 0 {
		return nil
	}
	return args[0]
}
