// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package gorm

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormlogger "gorm.io/gorm/logger"

	kitlog "github.com/fsyyft-go/kit/log"
)

type logEvent struct {
	level string
	args  []interface{}
}

type recordingLogger struct {
	kitlog.Logger

	level  kitlog.Level
	events chan logEvent
}

// TestGetGormLevel_Mapping 验证 kit 日志级别到 gorm 日志级别的映射契约。
//
// 该测试通过表驱动用例覆盖已知级别与未知级别，确保适配器初始化时使用稳定的 gorm 日志阈值。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGetGormLevel_Mapping(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveLevel   kitlog.Level
		wantLevel   gormlogger.LogLevel
	}{
		{
			name:        "success/debug-to-info",
			description: "验证 Debug 级别映射为 gorm Info，以保留 SQL 明细日志。",
			giveLevel:   kitlog.DebugLevel,
			wantLevel:   gormlogger.Info,
		},
		{
			name:        "success/info-to-info",
			description: "验证 Info 级别映射为 gorm Info，保持信息日志输出。",
			giveLevel:   kitlog.InfoLevel,
			wantLevel:   gormlogger.Info,
		},
		{
			name:        "success/warn-to-warn",
			description: "验证 Warn 级别映射为 gorm Warn，仅输出警告及以上日志。",
			giveLevel:   kitlog.WarnLevel,
			wantLevel:   gormlogger.Warn,
		},
		{
			name:        "success/error-to-error",
			description: "验证 Error 级别映射为 gorm Error，仅输出错误日志。",
			giveLevel:   kitlog.ErrorLevel,
			wantLevel:   gormlogger.Error,
		},
		{
			name:        "success/fatal-to-silent",
			description: "验证 Fatal 级别映射为 gorm Silent，避免继续输出 SQL 日志。",
			giveLevel:   kitlog.FatalLevel,
			wantLevel:   gormlogger.Silent,
		},
		{
			name:        "boundary/unknown-to-info",
			description: "验证未知 kit 日志级别使用 gorm Info 作为兼容默认值。",
			giveLevel:   kitlog.Level(999),
			wantLevel:   gormlogger.Info,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			gotLevel := getGormLevel(tt.giveLevel)

			assert.Equal(t, tt.wantLevel, gotLevel)
		})
	}
}

// TestNewLogger_InitialLevel 验证 NewLogger 根据底层 kit logger 级别初始化 gorm logger。
//
// 该测试通过表驱动用例覆盖典型初始化级别，确保公开构造函数返回可用的 gorm logger 适配器。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewLogger_InitialLevel(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveLevel   kitlog.Level
		wantLevel   gormlogger.LogLevel
	}{
		{
			name:        "success/info-logger",
			description: "验证 Info 级别 kit logger 构造出的适配器启用 gorm Info。",
			giveLevel:   kitlog.InfoLevel,
			wantLevel:   gormlogger.Info,
		},
		{
			name:        "success/error-logger",
			description: "验证 Error 级别 kit logger 构造出的适配器启用 gorm Error。",
			giveLevel:   kitlog.ErrorLevel,
			wantLevel:   gormlogger.Error,
		},
		{
			name:        "boundary/fatal-logger",
			description: "验证 Fatal 级别 kit logger 构造出的适配器进入 gorm Silent。",
			giveLevel:   kitlog.FatalLevel,
			wantLevel:   gormlogger.Silent,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			giveLogger := newRecordingLogger(tt.giveLevel)

			gotLogger := NewLogger(giveLogger)

			gotGormLogger, ok := gotLogger.(*gormLogger)
			require.True(t, ok)
			assert.Equal(t, tt.wantLevel, gotGormLogger.level)
			assert.True(t, gotGormLogger.logger == giveLogger, "适配器应保留传入的底层 logger")
		})
	}
}

// TestGormLogger_LogMode 验证 LogMode 返回独立副本并保留底层 logger。
//
// 该测试覆盖 gorm logger.Interface 的级别切换语义，确保调用 LogMode 不会修改原适配器状态。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGormLogger_LogMode(t *testing.T) {
	// 验证 LogMode 只影响返回的新实例，原实例仍保持 Info 级别。
	giveLogger := newRecordingLogger(kitlog.InfoLevel)
	giveGormLogger := &gormLogger{
		logger: giveLogger,
		level:  gormlogger.Info,
	}

	gotLogger := giveGormLogger.LogMode(gormlogger.Error)

	gotGormLogger, ok := gotLogger.(*gormLogger)
	require.True(t, ok)
	assert.NotSame(t, giveGormLogger, gotGormLogger)
	assert.Equal(t, gormlogger.Info, giveGormLogger.level)
	assert.Equal(t, gormlogger.Error, gotGormLogger.level)
	assert.True(t, gotGormLogger.logger == giveLogger, "LogMode 返回的新实例应复用底层 logger")
}

// TestGormLogger_MessageLogging 验证 Info、Warn 和 Error 的格式化输出与级别门控行为。
//
// 该测试通过表驱动用例覆盖普通消息日志的成功输出和被阈值抑制的边界行为。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGormLogger_MessageLogging(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name        string
		description string
		giveLevel   gormlogger.LogLevel
		act         func(logger *gormLogger)
		wantLog     bool
		wantLevel   string
		wantMessage string
	}{
		{
			name:        "success/info-formats-message",
			description: "验证 Info 级别允许信息日志输出，并按 fmt 语义格式化消息。",
			giveLevel:   gormlogger.Info,
			act: func(logger *gormLogger) {
				logger.Info(ctx, "hello %s", "gorm")
			},
			wantLog:     true,
			wantLevel:   "info",
			wantMessage: "hello gorm",
		},
		{
			name:        "boundary/info-suppressed-by-warn-level",
			description: "验证 Warn 阈值会抑制 Info 消息日志。",
			giveLevel:   gormlogger.Warn,
			act: func(logger *gormLogger) {
				logger.Info(ctx, "hidden %s", "info")
			},
			wantLog: false,
		},
		{
			name:        "success/warn-formats-message",
			description: "验证 Warn 阈值允许警告日志输出，并按 fmt 语义格式化消息。",
			giveLevel:   gormlogger.Warn,
			act: func(logger *gormLogger) {
				logger.Warn(ctx, "slow %s", "query")
			},
			wantLog:     true,
			wantLevel:   "warn",
			wantMessage: "slow query",
		},
		{
			name:        "boundary/warn-suppressed-by-error-level",
			description: "验证 Error 阈值会抑制 Warn 消息日志。",
			giveLevel:   gormlogger.Error,
			act: func(logger *gormLogger) {
				logger.Warn(ctx, "hidden %s", "warn")
			},
			wantLog: false,
		},
		{
			name:        "success/error-formats-message",
			description: "验证 Error 阈值允许错误日志输出，并按 fmt 语义格式化消息。",
			giveLevel:   gormlogger.Error,
			act: func(logger *gormLogger) {
				logger.Error(ctx, "failed %s", "query")
			},
			wantLog:     true,
			wantLevel:   "error",
			wantMessage: "failed query",
		},
		{
			name:        "boundary/error-suppressed-by-silent-level",
			description: "验证 Silent 阈值会抑制 Error 消息日志。",
			giveLevel:   gormlogger.Silent,
			act: func(logger *gormLogger) {
				logger.Error(ctx, "hidden %s", "error")
			},
			wantLog: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			giveLogger := newRecordingLogger(kitlog.InfoLevel)
			giveGormLogger := &gormLogger{
				logger: giveLogger,
				level:  tt.giveLevel,
			}

			tt.act(giveGormLogger)

			if !tt.wantLog {
				assertNoLogEvent(t, giveLogger)
				return
			}

			gotEvent := requireLogEvent(t, giveLogger, tt.wantLevel)
			require.Len(t, gotEvent.args, 1)
			assert.Equal(t, tt.wantMessage, gotEvent.args[0])
			assertNoLogEvent(t, giveLogger)
		})
	}
}

// TestGormLogger_Trace 验证 Trace 对 SQL 执行日志的分支选择和结构化字段。
//
// 该测试通过表驱动用例覆盖静默、错误、慢查询、普通查询与阈值抑制场景，确保 Trace 行为稳定且无需真实数据库。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGormLogger_Trace(t *testing.T) {
	ctx := context.Background()
	giveErr := errors.New("query failed")
	tests := []struct {
		name           string
		description    string
		giveLevel      gormlogger.LogLevel
		giveBegin      func() time.Time
		giveSQL        string
		giveRows       int64
		giveErr        error
		wantCallback   bool
		wantLog        bool
		wantLevel      string
		assertLogEvent func(t *testing.T, event logEvent)
	}{
		{
			name:         "boundary/silent-skips-callback",
			description:  "验证 Silent 阈值直接返回，不执行 SQL 回调且不输出日志。",
			giveLevel:    gormlogger.Silent,
			giveBegin:    time.Now,
			giveSQL:      "SELECT skipped",
			giveRows:     0,
			wantCallback: false,
			wantLog:      false,
		},
		{
			name:         "success/error-includes-sql-rows-and-error",
			description:  "验证 SQL 执行失败时输出错误日志，并携带 SQL、行数和错误字段。",
			giveLevel:    gormlogger.Error,
			giveBegin:    time.Now,
			giveSQL:      "SELECT * FROM users WHERE id = 1",
			giveRows:     0,
			giveErr:      giveErr,
			wantCallback: true,
			wantLog:      true,
			wantLevel:    "error",
			assertLogEvent: func(t *testing.T, event logEvent) {
				require.Len(t, event.args, 3)
				assertTraceMessage(t, event.args[0], "SELECT * FROM users WHERE id = 1", 0)
				assert.Equal(t, "err", event.args[1])
				gotErr, ok := event.args[2].(error)
				require.True(t, ok)
				assert.ErrorIs(t, gotErr, giveErr)
			},
		},
		{
			name:        "success/slow-query-warns-with-elapsed",
			description: "验证超过一秒的成功 SQL 在 Warn 阈值下输出慢查询警告，并携带耗时字段。",
			giveLevel:   gormlogger.Warn,
			giveBegin: func() time.Time {
				return time.Now().Add(-1100 * time.Millisecond)
			},
			giveSQL:      "UPDATE users SET name = 'alice' WHERE id = 1",
			giveRows:     1,
			wantCallback: true,
			wantLog:      true,
			wantLevel:    "warn",
			assertLogEvent: func(t *testing.T, event logEvent) {
				require.Len(t, event.args, 3)
				assertTraceMessage(t, event.args[0], "UPDATE users SET name = 'alice' WHERE id = 1", 1)
				assert.Equal(t, "elapsed", event.args[1])
				gotElapsed, ok := event.args[2].(time.Duration)
				require.True(t, ok)
				assert.Greater(t, gotElapsed, time.Second)
			},
		},
		{
			name:         "success/info-query-includes-sql-and-rows",
			description:  "验证普通成功 SQL 在 Info 阈值下输出包含 SQL 和影响行数的跟踪日志。",
			giveLevel:    gormlogger.Info,
			giveBegin:    time.Now,
			giveSQL:      "SELECT count(*) FROM users",
			giveRows:     7,
			wantCallback: true,
			wantLog:      true,
			wantLevel:    "info",
			assertLogEvent: func(t *testing.T, event logEvent) {
				require.Len(t, event.args, 1)
				assertTraceMessage(t, event.args[0], "SELECT count(*) FROM users", 7)
			},
		},
		{
			name:         "boundary/warn-skips-fast-success",
			description:  "验证 Warn 阈值下未达到慢查询标准的成功 SQL 不输出日志，但仍执行 SQL 回调。",
			giveLevel:    gormlogger.Warn,
			giveBegin:    time.Now,
			giveSQL:      "SELECT fast",
			giveRows:     1,
			wantCallback: true,
			wantLog:      false,
		},
		{
			name:        "boundary/error-skips-slow-success",
			description: "验证 Error 阈值下成功慢查询不会降级输出 Warn 日志，但仍执行 SQL 回调。",
			giveLevel:   gormlogger.Error,
			giveBegin: func() time.Time {
				return time.Now().Add(-1100 * time.Millisecond)
			},
			giveSQL:      "SELECT slow_without_error",
			giveRows:     1,
			wantCallback: true,
			wantLog:      false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			giveLogger := newRecordingLogger(kitlog.InfoLevel)
			giveGormLogger := &gormLogger{
				logger: giveLogger,
				level:  tt.giveLevel,
			}
			callbackCalled := false

			giveGormLogger.Trace(ctx, tt.giveBegin(), func() (string, int64) {
				callbackCalled = true
				return tt.giveSQL, tt.giveRows
			}, tt.giveErr)

			assert.Equal(t, tt.wantCallback, callbackCalled)
			if !tt.wantLog {
				assertNoLogEvent(t, giveLogger)
				return
			}

			gotEvent := requireLogEvent(t, giveLogger, tt.wantLevel)
			require.NotNil(t, tt.assertLogEvent)
			tt.assertLogEvent(t, gotEvent)
			assertNoLogEvent(t, giveLogger)
		})
	}
}

// newRecordingLogger 构造记录日志调用的 kit logger 测试替身。
//
// 该辅助函数使用带缓冲 channel 捕获日志事件，便于断言异步日志调用的级别和参数。
//
// 参数：
//   - giveLevel: 测试替身暴露给适配器的 kit 日志级别。
//
// 返回：
//   - *recordingLogger: 可作为 kitlog.Logger 使用的日志记录测试替身。
func newRecordingLogger(giveLevel kitlog.Level) *recordingLogger {
	return &recordingLogger{
		level:  giveLevel,
		events: make(chan logEvent, 8),
	}
}

// GetLevel 返回测试替身配置的 kit 日志级别。
//
// 该辅助方法用于验证 NewLogger 根据底层 logger 级别初始化 gorm 日志阈值。
//
// 返回：
//   - kitlog.Level: 当前测试替身配置的日志级别。
func (l *recordingLogger) GetLevel() kitlog.Level {
	return l.level
}

// Info 记录一次信息级别日志调用。
//
// 该辅助方法实现 kitlog.Logger 的信息日志行为，将调用参数写入事件 channel 供测试断言。
//
// 参数：
//   - args: 日志调用传入的消息和结构化字段。
func (l *recordingLogger) Info(args ...interface{}) {
	l.events <- logEvent{level: "info", args: cloneArgs(args)}
}

// Warn 记录一次警告级别日志调用。
//
// 该辅助方法实现 kitlog.Logger 的警告日志行为，将调用参数写入事件 channel 供测试断言。
//
// 参数：
//   - args: 日志调用传入的消息和结构化字段。
func (l *recordingLogger) Warn(args ...interface{}) {
	l.events <- logEvent{level: "warn", args: cloneArgs(args)}
}

// Error 记录一次错误级别日志调用。
//
// 该辅助方法实现 kitlog.Logger 的错误日志行为，将调用参数写入事件 channel 供测试断言。
//
// 参数：
//   - args: 日志调用传入的消息和结构化字段。
func (l *recordingLogger) Error(args ...interface{}) {
	l.events <- logEvent{level: "error", args: cloneArgs(args)}
}

// cloneArgs 复制日志调用参数，避免后续修改影响断言结果。
//
// 参数：
//   - giveArgs: 原始日志调用参数。
//
// 返回：
//   - []interface{}: 可安全用于断言的参数副本。
func cloneArgs(giveArgs []interface{}) []interface{} {
	gotArgs := make([]interface{}, len(giveArgs))
	copy(gotArgs, giveArgs)
	return gotArgs
}

// requireLogEvent 等待并返回一次期望级别的日志事件。
//
// 该辅助函数为异步日志调用提供超时保护，并在捕获到事件后校验日志级别。
//
// 参数：
//   - t: 测试上下文，用于报告等待或断言失败。
//   - logger: 记录日志调用的测试替身。
//   - wantLevel: 期望捕获到的日志级别名称。
//
// 返回：
//   - logEvent: 捕获到的日志事件。
func requireLogEvent(t *testing.T, logger *recordingLogger, wantLevel string) logEvent {
	t.Helper()

	select {
	case gotEvent := <-logger.events:
		require.Equal(t, wantLevel, gotEvent.level)
		return gotEvent
	case <-time.After(time.Second):
		require.Fail(t, "等待日志事件超时", "want level %s", wantLevel)
		return logEvent{}
	}
}

// assertNoLogEvent 断言测试替身没有捕获到日志事件。
//
// 该辅助函数用于验证日志阈值抑制行为，并通过短超时给异步日志 goroutine 调度机会。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
//   - logger: 记录日志调用的测试替身。
func assertNoLogEvent(t *testing.T, logger *recordingLogger) {
	t.Helper()

	select {
	case gotEvent := <-logger.events:
		assert.Failf(t, "不应捕获日志事件", "got level %s args %v", gotEvent.level, gotEvent.args)
	case <-time.After(20 * time.Millisecond):
	}
}

// assertTraceMessage 断言 Trace 生成的日志消息包含 SQL 和影响行数。
//
// 该辅助函数只校验稳定的消息契约，避免对动态耗时文本做精确匹配导致测试脆弱。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
//   - giveMessage: Trace 写入底层 logger 的消息参数。
//   - wantSQL: 期望消息包含的 SQL 文本。
//   - wantRows: 期望消息包含的影响行数。
func assertTraceMessage(t *testing.T, giveMessage interface{}, wantSQL string, wantRows int64) {
	t.Helper()

	gotMessage := fmt.Sprint(giveMessage)
	assert.Contains(t, gotMessage, "ms]", "Trace 消息应包含毫秒耗时片段")
	assert.Contains(t, gotMessage, fmt.Sprintf("[rows:%d]", wantRows))
	assert.Contains(t, gotMessage, wantSQL)
}
