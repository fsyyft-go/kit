// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	stdLoggerFatalHelperEnv = "KIT_LOG_TEST_STD_FATAL_HELPER"
	stdLoggerFatalOutputEnv = "KIT_LOG_TEST_STD_FATAL_OUTPUT"
)

type (
	// fakeLoggerCall 记录 fakeLogger 接收到的一次方法调用。
	//
	// 该结构体用于验证全局日志代理函数是否把参数完整转发到已配置的 Logger。
	fakeLoggerCall struct {
		method string
		level  Level
		format string
		args   []interface{}
		fields map[string]interface{}
	}

	// fakeLogger 是用于全局代理测试的内存 Logger 实现。
	//
	// 该实现不会写 stdout，也不会在 Fatal/Fatalf 调用时退出进程，便于稳定验证代理行为。
	fakeLogger struct {
		level  Level
		fields map[string]interface{}
		calls  []fakeLoggerCall
	}

	// logrusExitPanic 表示测试中由 logrus ExitFunc 触发的退出信号。
	//
	// 该结构体用于携带退出状态码，避免 Fatal/Fatalf 测试真实终止进程。
	logrusExitPanic struct {
		code int
	}
)

// TestLevelString_ReturnsStableNames 验证 Level.String 返回稳定的日志级别名称。
//
// 该测试覆盖所有已定义级别和未知级别，确保日志级别的字符串契约稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestLevelString_ReturnsStableNames(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveLevel   Level
		want        string
	}{
		{
			name:        "success/debug",
			description: "验证 DebugLevel 的字符串表示为 debug。",
			giveLevel:   DebugLevel,
			want:        "debug",
		},
		{
			name:        "success/info",
			description: "验证 InfoLevel 的字符串表示为 info。",
			giveLevel:   InfoLevel,
			want:        "info",
		},
		{
			name:        "success/warn",
			description: "验证 WarnLevel 的字符串表示为 warn。",
			giveLevel:   WarnLevel,
			want:        "warn",
		},
		{
			name:        "success/error",
			description: "验证 ErrorLevel 的字符串表示为 error。",
			giveLevel:   ErrorLevel,
			want:        "error",
		},
		{
			name:        "success/fatal",
			description: "验证 FatalLevel 的字符串表示为 fatal。",
			giveLevel:   FatalLevel,
			want:        "fatal",
		},
		{
			name:        "boundary/unknown",
			description: "验证未知日志级别返回 unknown，避免泄露未定义数值。",
			giveLevel:   Level(99),
			want:        "unknown",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := tt.giveLevel.String()

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParseLevel_ParsesKnownLevelsAndRejectsUnknown 验证 ParseLevel 的成功解析和错误回退行为。
//
// 该测试覆盖所有支持的级别字符串以及未知字符串，确保解析结果和错误信息符合契约。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestParseLevel_ParsesKnownLevelsAndRejectsUnknown(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		giveLevelString string
		wantLevel       Level
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "success/debug",
			description:     "验证 debug 字符串解析为 DebugLevel。",
			giveLevelString: "debug",
			wantLevel:       DebugLevel,
		},
		{
			name:            "success/info",
			description:     "验证 info 字符串解析为 InfoLevel。",
			giveLevelString: "info",
			wantLevel:       InfoLevel,
		},
		{
			name:            "success/warn",
			description:     "验证 warn 字符串解析为 WarnLevel。",
			giveLevelString: "warn",
			wantLevel:       WarnLevel,
		},
		{
			name:            "success/error",
			description:     "验证 error 字符串解析为 ErrorLevel。",
			giveLevelString: "error",
			wantLevel:       ErrorLevel,
		},
		{
			name:            "success/fatal",
			description:     "验证 fatal 字符串解析为 FatalLevel。",
			giveLevelString: "fatal",
			wantLevel:       FatalLevel,
		},
		{
			name:            "error/unknown",
			description:     "验证未知级别字符串返回 InfoLevel 并携带可诊断错误。",
			giveLevelString: "trace",
			wantLevel:       InfoLevel,
			wantErr:         true,
			wantErrContains: "unknown level: trace",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			gotLevel, err := ParseLevel(tt.giveLevelString)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				assert.Equal(t, tt.wantLevel, gotLevel)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantLevel, gotLevel)
		})
	}
}

// TestLoggerOptions_ApplyOptions 验证 Option 对 LoggerOptions 的字段修改行为。
//
// 该测试逐项覆盖所有公开 Logger Option，确保函数式选项只修改对应配置字段。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestLoggerOptions_ApplyOptions(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveOption  Option
		assert      func(t *testing.T, got LoggerOptions)
	}{
		{
			name:        "success/log-type",
			description: "验证 WithLogType 设置日志实现类型。",
			giveOption:  WithLogType(LogTypeLogrus),
			assert: func(t *testing.T, got LoggerOptions) {
				assert.Equal(t, LogTypeLogrus, got.Type)
			},
		},
		{
			name:        "success/format-type",
			description: "验证 WithFormatType 设置日志格式类型。",
			giveOption:  WithFormatType(TextFormat),
			assert: func(t *testing.T, got LoggerOptions) {
				assert.Equal(t, TextFormat, got.FormatType)
			},
		},
		{
			name:        "success/level",
			description: "验证 WithLevel 设置日志过滤级别。",
			giveOption:  WithLevel(ErrorLevel),
			assert: func(t *testing.T, got LoggerOptions) {
				assert.Equal(t, ErrorLevel, got.Level)
			},
		},
		{
			name:        "success/output",
			description: "验证 WithOutput 设置日志输出路径。",
			giveOption:  WithOutput("/var/log/app.log"),
			assert: func(t *testing.T, got LoggerOptions) {
				assert.Equal(t, "/var/log/app.log", got.Output)
			},
		},
		{
			name:        "success/enable-rotate",
			description: "验证 WithEnableRotate 设置日志滚动开关。",
			giveOption:  WithEnableRotate(false),
			assert: func(t *testing.T, got LoggerOptions) {
				assert.False(t, got.EnableRotate)
			},
		},
		{
			name:        "success/rotate-time",
			description: "验证 WithRotateTime 设置日志滚动周期。",
			giveOption:  WithRotateTime(2 * time.Hour),
			assert: func(t *testing.T, got LoggerOptions) {
				assert.Equal(t, 2*time.Hour, got.RotateTime)
			},
		},
		{
			name:        "success/max-age",
			description: "验证 WithMaxAge 设置日志文件保留时间。",
			giveOption:  WithMaxAge(48 * time.Hour),
			assert: func(t *testing.T, got LoggerOptions) {
				assert.Equal(t, 48*time.Hour, got.MaxAge)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := LoggerOptions{
				Type:         LogTypeStd,
				Level:        InfoLevel,
				Output:       "",
				EnableRotate: true,
				RotateTime:   time.Hour,
				MaxAge:       24 * time.Hour,
				FormatType:   JSONFormat,
			}
			tt.giveOption(&got)

			tt.assert(t, got)
		})
	}
}

// TestNewLogger_ConstructsSupportedTypesAndFormats 验证 NewLogger 对支持类型和格式的创建行为。
//
// 该测试覆盖默认、console、std、logrus JSON、logrus text 和未知格式回退场景，确保构造结果可用且配置正确。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewLogger_ConstructsSupportedTypesAndFormats(t *testing.T) {
	tests := []struct {
		name        string
		description string
		setup       func(t *testing.T) ([]Option, string)
		assert      func(t *testing.T, got Logger, giveOutput string)
	}{
		{
			name:        "success/default-std",
			description: "验证无选项创建默认标准库 Logger，且默认级别为 InfoLevel。",
			setup: func(t *testing.T) ([]Option, string) {
				return nil, ""
			},
			assert: func(t *testing.T, got Logger, giveOutput string) {
				stdLogger, ok := got.(*StdLogger)
				require.True(t, ok)
				assert.Equal(t, InfoLevel, stdLogger.GetLevel())
				assert.Empty(t, stdLogger.fields)
			},
		},
		{
			name:        "success/console-ignores-output-path",
			description: "验证 console 类型创建标准库 Logger，并忽略文件输出路径。",
			setup: func(t *testing.T) ([]Option, string) {
				outputPath := filepath.Join(t.TempDir(), "ignored.log")
				t.Cleanup(func() {
					assert.NoFileExists(t, outputPath)
				})
				return []Option{
					WithLogType(LogTypeConsole),
					WithOutput(outputPath),
					WithLevel(WarnLevel),
				}, outputPath
			},
			assert: func(t *testing.T, got Logger, giveOutput string) {
				_, ok := got.(*StdLogger)
				require.True(t, ok)
				assert.Equal(t, WarnLevel, got.GetLevel())
			},
		},
		{
			name:        "success/std-file-output",
			description: "验证 std 类型按文件路径输出日志并应用 DebugLevel。",
			setup: func(t *testing.T) ([]Option, string) {
				outputPath := filepath.Join(t.TempDir(), "std", "app.log")
				return []Option{
					WithLogType(LogTypeStd),
					WithOutput(outputPath),
					WithLevel(DebugLevel),
				}, outputPath
			},
			assert: func(t *testing.T, got Logger, giveOutput string) {
				stdLogger, ok := got.(*StdLogger)
				require.True(t, ok)
				stdLogger.logger.SetFlags(0)
				stdLogger.Debug("std-debug-event")

				content := readFileString(t, giveOutput)
				assert.Contains(t, content, "[DEBUG] std-debug-event")
			},
		},
		{
			name:        "success/logrus-json-format",
			description: "验证 logrus 类型在 JSON 格式下输出结构化字段。",
			setup: func(t *testing.T) ([]Option, string) {
				outputPath := filepath.Join(t.TempDir(), "logrus-json.log")
				return []Option{
					WithLogType(LogTypeLogrus),
					WithOutput(outputPath),
					WithEnableRotate(false),
					WithFormatType(JSONFormat),
					WithLevel(DebugLevel),
				}, outputPath
			},
			assert: func(t *testing.T, got Logger, giveOutput string) {
				logrusLogger, ok := got.(*LogrusLogger)
				require.True(t, ok)
				_, ok = logrusLogger.logger.Logger.Formatter.(*logrus.JSONFormatter)
				require.True(t, ok)

				logrusLogger.WithField("component", "new-logger").Info("json-event")

				entries := readJSONLogEntries(t, giveOutput)
				entry := findJSONLogEntry(t, entries, "json-event")
				assert.Equal(t, "info", entry["level"])
				assert.Equal(t, "new-logger", entry["component"])
			},
		},
		{
			name:        "success/logrus-text-format",
			description: "验证 logrus 类型在文本格式下输出消息和结构化字段。",
			setup: func(t *testing.T) ([]Option, string) {
				outputPath := filepath.Join(t.TempDir(), "logrus-text.log")
				return []Option{
					WithLogType(LogTypeLogrus),
					WithOutput(outputPath),
					WithEnableRotate(false),
					WithFormatType(TextFormat),
					WithLevel(InfoLevel),
				}, outputPath
			},
			assert: func(t *testing.T, got Logger, giveOutput string) {
				logrusLogger, ok := got.(*LogrusLogger)
				require.True(t, ok)
				_, ok = logrusLogger.logger.Logger.Formatter.(*logrus.TextFormatter)
				require.True(t, ok)

				logrusLogger.WithField("component", "new-logger").Info("text-event")

				content := readFileString(t, giveOutput)
				assert.Contains(t, content, "level=info")
				assert.Contains(t, content, "msg=text-event")
				assert.Contains(t, content, "component=new-logger")
			},
		},
		{
			name:        "boundary/logrus-unknown-format-uses-default-json",
			description: "验证未知格式类型不会破坏 logrus 默认 JSON formatter。",
			setup: func(t *testing.T) ([]Option, string) {
				return []Option{
					WithLogType(LogTypeLogrus),
					WithFormatType(LoggerFormatType("yaml")),
				}, ""
			},
			assert: func(t *testing.T, got Logger, giveOutput string) {
				logrusLogger, ok := got.(*LogrusLogger)
				require.True(t, ok)
				_, ok = logrusLogger.logger.Logger.Formatter.(*logrus.JSONFormatter)
				assert.True(t, ok)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			giveOptions, giveOutput := tt.setup(t)
			got, err := NewLogger(giveOptions...)

			require.NoError(t, err)
			require.NotNil(t, got)
			cleanupLoggerOutput(t, got)
			tt.assert(t, got, giveOutput)
		})
	}
}

// TestNewLogger_ReturnsErrorsForUnsupportedOrInvalidOutput 验证 NewLogger 的错误返回行为。
//
// 该测试覆盖非法日志类型以及底层输出路径创建失败，确保错误可诊断且不会返回半初始化 Logger。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewLogger_ReturnsErrorsForUnsupportedOrInvalidOutput(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		setup           func(t *testing.T) []Option
		wantErrContains string
	}{
		{
			name:        "error/unsupported-log-type",
			description: "验证未知日志类型返回不支持错误。",
			setup: func(t *testing.T) []Option {
				return []Option{WithLogType(LogType("unsupported"))}
			},
			wantErrContains: "不支持的日志类型",
		},
		{
			name:        "error/std-output-is-directory",
			description: "验证 std 类型把目录作为日志文件路径时返回创建失败错误。",
			setup: func(t *testing.T) []Option {
				return []Option{
					WithLogType(LogTypeStd),
					WithOutput(t.TempDir()),
				}
			},
			wantErrContains: "创建日志实例失败",
		},
		{
			name:        "error/logrus-parent-is-file",
			description: "验证 logrus 类型在父路径为普通文件时返回创建失败错误。",
			setup: func(t *testing.T) []Option {
				blockedParent := filepath.Join(t.TempDir(), "blocked")
				require.NoError(t, os.WriteFile(blockedParent, []byte("not-a-dir"), 0600))
				return []Option{
					WithLogType(LogTypeLogrus),
					WithOutput(filepath.Join(blockedParent, "app.log")),
					WithEnableRotate(false),
				}
			},
			wantErrContains: "创建日志实例失败",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := NewLogger(tt.setup(t)...)

			require.Error(t, err)
			assert.Nil(t, got)
			assert.Contains(t, err.Error(), tt.wantErrContains)
		})
	}
}

// TestStdLogger_NewStdLoggerCreatesOutputsAndReportsErrors 验证 NewStdLogger 的输出创建和错误分支。
//
// 该测试覆盖 stdout 默认配置、文件输出、父路径异常和输出路径为目录等边界，确保标准库 Logger 初始化稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestStdLogger_NewStdLoggerCreatesOutputsAndReportsErrors(t *testing.T) {
	tests := []struct {
		name        string
		description string
		setup       func(t *testing.T) string
		wantErr     bool
		assert      func(t *testing.T, got Logger, giveOutput string)
	}{
		{
			name:        "success/stdout-default",
			description: "验证空输出路径创建默认 stdout 标准库 Logger。",
			setup: func(t *testing.T) string {
				return ""
			},
			assert: func(t *testing.T, got Logger, giveOutput string) {
				stdLogger, ok := got.(*StdLogger)
				require.True(t, ok)
				assert.Equal(t, InfoLevel, stdLogger.GetLevel())
				assert.Empty(t, stdLogger.fields)
			},
		},
		{
			name:        "success/file-output",
			description: "验证文件输出路径会自动创建目录和日志文件。",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nested", "std.log")
			},
			assert: func(t *testing.T, got Logger, giveOutput string) {
				stdLogger, ok := got.(*StdLogger)
				require.True(t, ok)
				stdLogger.logger.SetFlags(0)
				stdLogger.Info("file-output-event")

				content := readFileString(t, giveOutput)
				assert.Contains(t, content, "[INFO] file-output-event")
			},
		},
		{
			name:        "error/parent-is-file",
			description: "验证父路径为普通文件时目录创建失败。",
			setup: func(t *testing.T) string {
				blockedParent := filepath.Join(t.TempDir(), "blocked")
				require.NoError(t, os.WriteFile(blockedParent, []byte("not-a-dir"), 0600))
				return filepath.Join(blockedParent, "std.log")
			},
			wantErr: true,
		},
		{
			name:        "error/output-is-directory",
			description: "验证输出路径为目录时文件打开失败。",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			giveOutput := tt.setup(t)
			got, err := NewStdLogger(giveOutput)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			cleanupLoggerOutput(t, got)
			tt.assert(t, got, giveOutput)
		})
	}
}

// TestStdLogger_ShouldLogBoundaries 验证 StdLogger.shouldLog 的级别边界判断。
//
// 该测试覆盖低于、等于和高于当前级别的输入，确保过滤逻辑与 Level 顺序一致。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestStdLogger_ShouldLogBoundaries(t *testing.T) {
	logger, _ := newBufferedStdLogger(t, WarnLevel)
	tests := []struct {
		name        string
		description string
		giveLevel   Level
		want        bool
	}{
		{
			name:        "boundary/below-current-level",
			description: "验证低于 WarnLevel 的 InfoLevel 不应输出。",
			giveLevel:   InfoLevel,
			want:        false,
		},
		{
			name:        "boundary/equal-current-level",
			description: "验证等于 WarnLevel 的日志应输出。",
			giveLevel:   WarnLevel,
			want:        true,
		},
		{
			name:        "boundary/above-current-level",
			description: "验证高于 WarnLevel 的 ErrorLevel 应输出。",
			giveLevel:   ErrorLevel,
			want:        true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := logger.shouldLog(tt.giveLevel)

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestStdLogger_LevelFilteringAndFormatOutput 验证 StdLogger 的级别过滤和格式化输出。
//
// 该测试覆盖普通与格式化日志方法，确保低级别消息被过滤，符合级别的消息带有正确前缀和内容。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestStdLogger_LevelFilteringAndFormatOutput(t *testing.T) {
	logger, buffer := newBufferedStdLogger(t, WarnLevel)

	// 验证低于 WarnLevel 的日志方法不会产生输出。
	logger.Debug("debug-hidden")
	logger.Debugf("debugf-%s", "hidden")
	logger.Info("info-hidden")
	logger.Infof("infof-%s", "hidden")

	// 验证 WarnLevel 及以上的普通和格式化日志会输出稳定前缀与消息。
	logger.Warn("warn-visible")
	logger.Warnf("warnf-%s", "visible")
	logger.Error("error-visible")
	logger.Errorf("errorf-%s", "visible")

	content := buffer.String()
	assert.NotContains(t, content, "debug-hidden")
	assert.NotContains(t, content, "debugf-hidden")
	assert.NotContains(t, content, "info-hidden")
	assert.NotContains(t, content, "infof-hidden")
	assert.Contains(t, content, "[WARN] warn-visible")
	assert.Contains(t, content, "[WARN] warnf-visible")
	assert.Contains(t, content, "[ERROR] error-visible")
	assert.Contains(t, content, "[ERROR] errorf-visible")
}

// TestStdLogger_FieldsAreImmutableAndFormatted 验证 StdLogger 的字段不可变和字段格式输出。
//
// 该测试覆盖 WithField、WithFields 和原始 Logger 的隔离语义，避免字段上下文被后续派生 Logger 污染。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestStdLogger_FieldsAreImmutableAndFormatted(t *testing.T) {
	baseLogger, buffer := newBufferedStdLogger(t, InfoLevel)
	assert.Empty(t, baseLogger.formatFields())

	withRequest := baseLogger.WithField("request_id", "req-1")
	withRequestLogger, ok := withRequest.(*StdLogger)
	require.True(t, ok)
	assert.Equal(t, InfoLevel, withRequestLogger.GetLevel())

	fields := map[string]interface{}{
		"user_id": 42,
		"admin":   true,
	}
	withTrace := withRequest.WithField("trace_id", "trace-1")
	withUser := withRequest.WithFields(fields)
	fields["user_id"] = 100
	fields["ignored"] = "new-value"

	baseLogger.Info("base-event")
	withRequest.Info("request-event")
	withTrace.Info("trace-event")
	withUser.Infof("user-%s", "event")

	lines := outputLines(buffer.String())
	require.Len(t, lines, 4)

	assert.Contains(t, lines[0], "[INFO] base-event")
	assert.NotContains(t, lines[0], "request_id")
	assert.NotContains(t, lines[0], "user_id")

	assert.Contains(t, lines[1], "[INFO]")
	assert.Contains(t, lines[1], "request-event")
	assert.Contains(t, lines[1], "request_id=req-1")
	assert.NotContains(t, lines[1], "trace_id")
	assert.NotContains(t, lines[1], "user_id")

	assert.Contains(t, lines[2], "[INFO]")
	assert.Contains(t, lines[2], "trace-event")
	assert.Contains(t, lines[2], "request_id=req-1")
	assert.Contains(t, lines[2], "trace_id=trace-1")
	assert.NotContains(t, lines[2], "user_id")

	assert.Contains(t, lines[3], "[INFO]")
	assert.Contains(t, lines[3], "user-event")
	assert.Contains(t, lines[3], "request_id=req-1")
	assert.Contains(t, lines[3], "user_id=42")
	assert.Contains(t, lines[3], "admin=true")
	assert.NotContains(t, lines[3], "trace_id")
	assert.NotContains(t, lines[3], "user_id=100")
	assert.NotContains(t, lines[3], "ignored=new-value")
}

// TestStdLogger_FatalSubprocess 验证 StdLogger 的 Fatal/Fatalf 会输出日志并以状态码 1 退出。
//
// 该测试通过子进程隔离 os.Exit，确保不会破坏当前测试进程，同时验证致命日志内容写入文件。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestStdLogger_FatalSubprocess(t *testing.T) {
	if helperCase := os.Getenv(stdLoggerFatalHelperEnv); helperCase != "" {
		runStdLoggerFatalHelper(helperCase, os.Getenv(stdLoggerFatalOutputEnv))
		return
	}

	tests := []struct {
		name        string
		description string
		giveCase    string
		wantContent string
	}{
		{
			name:        "success/fatal",
			description: "验证 Fatal 输出致命日志并以状态码 1 退出。",
			giveCase:    "fatal",
			wantContent: "[FATAL] fatal-event",
		},
		{
			name:        "success/fatalf",
			description: "验证 Fatalf 输出格式化致命日志并以状态码 1 退出。",
			giveCase:    "fatalf",
			wantContent: "[FATAL] fatalf-event",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			outputPath := filepath.Join(t.TempDir(), "fatal.log")
			cmd := exec.Command(os.Args[0], "-test.run", "^TestStdLogger_FatalSubprocess$")
			cmd.Env = append(os.Environ(),
				stdLoggerFatalHelperEnv+"="+tt.giveCase,
				stdLoggerFatalOutputEnv+"="+outputPath,
			)

			combinedOutput, err := cmd.CombinedOutput()
			require.Error(t, err, string(combinedOutput))
			var exitErr *exec.ExitError
			require.ErrorAs(t, err, &exitErr)
			assert.Equal(t, 1, exitErr.ExitCode())

			content := readFileString(t, outputPath)
			assert.Contains(t, content, tt.wantContent)
		})
	}
}

// TestLogrusOptions_ApplyConfiguration 验证 LogrusOption 对 LogrusLoggerOptions 的配置作用。
//
// 该测试覆盖输出路径、formatter、级别、文件权限、目录权限和滚动参数，确保 Logrus 选项可预测地应用。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestLogrusOptions_ApplyConfiguration(t *testing.T) {
	customFormatter := &logrus.JSONFormatter{TimestampFormat: time.RFC3339}
	tests := []struct {
		name        string
		description string
		giveOption  LogrusOption
		assert      func(t *testing.T, got LogrusLoggerOptions)
	}{
		{
			name:        "success/output-path",
			description: "验证 WithOutputPath 设置 logrus 输出文件路径。",
			giveOption:  WithOutputPath("/tmp/app.log"),
			assert: func(t *testing.T, got LogrusLoggerOptions) {
				assert.Equal(t, "/tmp/app.log", got.OutputPath)
			},
		},
		{
			name:        "success/custom-formatter",
			description: "验证 WithFormatter 设置自定义 formatter 实例。",
			giveOption:  WithFormatter(customFormatter),
			assert: func(t *testing.T, got LogrusLoggerOptions) {
				assert.Same(t, customFormatter, got.Formatter)
			},
		},
		{
			name:        "success/json-formatter",
			description: "验证 WithJSONFormatter 设置时间格式和 PrettyPrint 配置。",
			giveOption:  WithJSONFormatter(time.RFC3339Nano, true),
			assert: func(t *testing.T, got LogrusLoggerOptions) {
				formatter, ok := got.Formatter.(*logrus.JSONFormatter)
				require.True(t, ok)
				assert.Equal(t, time.RFC3339Nano, formatter.TimestampFormat)
				assert.True(t, formatter.PrettyPrint)
			},
		},
		{
			name:        "success/text-formatter",
			description: "验证 WithTextFormatter 设置时间格式、完整时间戳和颜色开关。",
			giveOption:  WithTextFormatter(time.RFC3339, true, true),
			assert: func(t *testing.T, got LogrusLoggerOptions) {
				formatter, ok := got.Formatter.(*logrus.TextFormatter)
				require.True(t, ok)
				assert.Equal(t, time.RFC3339, formatter.TimestampFormat)
				assert.True(t, formatter.FullTimestamp)
				assert.True(t, formatter.DisableColors)
			},
		},
		{
			name:        "success/valid-level",
			description: "验证 WithLogrusLevel 将自定义级别映射为 logrus 级别。",
			giveOption:  WithLogrusLevel(ErrorLevel),
			assert: func(t *testing.T, got LogrusLoggerOptions) {
				assert.Equal(t, logrus.ErrorLevel, got.Level)
			},
		},
		{
			name:        "boundary/invalid-level-keeps-current",
			description: "验证未知自定义级别不会覆盖现有 logrus 级别。",
			giveOption:  WithLogrusLevel(Level(100)),
			assert: func(t *testing.T, got LogrusLoggerOptions) {
				assert.Equal(t, logrus.WarnLevel, got.Level)
			},
		},
		{
			name:        "success/file-mode",
			description: "验证 WithFileMode 设置日志文件权限模式。",
			giveOption:  WithFileMode(0600),
			assert: func(t *testing.T, got LogrusLoggerOptions) {
				assert.Equal(t, os.FileMode(0600), got.FileMode)
			},
		},
		{
			name:        "success/dir-mode",
			description: "验证 WithDirMode 设置日志目录权限模式。",
			giveOption:  WithDirMode(0700),
			assert: func(t *testing.T, got LogrusLoggerOptions) {
				assert.Equal(t, os.FileMode(0700), got.DirMode)
			},
		},
		{
			name:        "success/enable-rotate",
			description: "验证 WithLogrusEnableRotate 设置滚动开关。",
			giveOption:  WithLogrusEnableRotate(false),
			assert: func(t *testing.T, got LogrusLoggerOptions) {
				assert.False(t, got.EnableRotate)
			},
		},
		{
			name:        "success/rotate-time",
			description: "验证 WithLogrusRotateTime 设置滚动周期。",
			giveOption:  WithLogrusRotateTime(30 * time.Minute),
			assert: func(t *testing.T, got LogrusLoggerOptions) {
				assert.Equal(t, 30*time.Minute, got.RotateTime)
			},
		},
		{
			name:        "success/max-age",
			description: "验证 WithLogrusMaxAge 设置日志保留时间。",
			giveOption:  WithLogrusMaxAge(72 * time.Hour),
			assert: func(t *testing.T, got LogrusLoggerOptions) {
				assert.Equal(t, 72*time.Hour, got.MaxAge)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := LogrusLoggerOptions{
				Formatter:    &logrus.TextFormatter{},
				Level:        logrus.WarnLevel,
				FileMode:     0666,
				DirMode:      0755,
				EnableRotate: true,
				RotateTime:   time.Hour,
				MaxAge:       24 * time.Hour,
			}
			tt.giveOption(&got)

			tt.assert(t, got)
		})
	}
}

// TestLogrusLogger_NewLoggerOutputsAndErrors 验证 NewLogrusLogger 的输出配置和错误分支。
//
// 该测试覆盖默认配置、普通文件输出、滚动文件输出、目录创建失败和文件打开失败，确保初始化行为可诊断。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestLogrusLogger_NewLoggerOutputsAndErrors(t *testing.T) {
	tests := []struct {
		name        string
		description string
		setup       func(t *testing.T) ([]LogrusOption, string)
		wantErr     bool
		assert      func(t *testing.T, got Logger, giveOutput string)
	}{
		{
			name:        "success/default-options",
			description: "验证默认 LogrusLogger 使用 InfoLevel 和 JSON formatter。",
			setup: func(t *testing.T) ([]LogrusOption, string) {
				return nil, ""
			},
			assert: func(t *testing.T, got Logger, giveOutput string) {
				logrusLogger, ok := got.(*LogrusLogger)
				require.True(t, ok)
				assert.Equal(t, InfoLevel, logrusLogger.GetLevel())
				_, ok = logrusLogger.logger.Logger.Formatter.(*logrus.JSONFormatter)
				assert.True(t, ok)
			},
		},
		{
			name:        "success/no-rotate-file-output",
			description: "验证禁用滚动时 LogrusLogger 直接写入指定文件。",
			setup: func(t *testing.T) ([]LogrusOption, string) {
				outputPath := filepath.Join(t.TempDir(), "logrus", "app.log")
				return []LogrusOption{
					WithOutputPath(outputPath),
					WithLogrusEnableRotate(false),
					WithJSONFormatter(time.RFC3339, false),
					WithLogrusLevel(DebugLevel),
				}, outputPath
			},
			assert: func(t *testing.T, got Logger, giveOutput string) {
				logrusLogger, ok := got.(*LogrusLogger)
				require.True(t, ok)
				logrusLogger.Debug("debug-event")

				entries := readJSONLogEntries(t, giveOutput)
				entry := findJSONLogEntry(t, entries, "debug-event")
				assert.Equal(t, "debug", entry["level"])
			},
		},
		{
			name:        "success/rotate-file-output",
			description: "验证启用滚动时 LogrusLogger 写入真实轮转文件，且不依赖 link path 内容断言。",
			setup: func(t *testing.T) ([]LogrusOption, string) {
				requireSymlinkSupport(t)
				outputPath := filepath.Join(t.TempDir(), "rotated", "app.log")
				return []LogrusOption{
					WithOutputPath(outputPath),
					WithLogrusEnableRotate(true),
					WithLogrusRotateTime(time.Hour),
					WithLogrusMaxAge(24 * time.Hour),
					WithJSONFormatter(time.RFC3339, false),
				}, outputPath
			},
			assert: func(t *testing.T, got Logger, giveOutput string) {
				logrusLogger, ok := got.(*LogrusLogger)
				require.True(t, ok)
				logrusLogger.Info("rotate-event")

				entries := readRotatedJSONLogEntries(t, giveOutput)
				entry := findJSONLogEntry(t, entries, "rotate-event")
				assert.Equal(t, "info", entry["level"])
			},
		},
		{
			name:        "error/parent-is-file",
			description: "验证输出父路径为普通文件时返回目录创建错误。",
			setup: func(t *testing.T) ([]LogrusOption, string) {
				blockedParent := filepath.Join(t.TempDir(), "blocked")
				require.NoError(t, os.WriteFile(blockedParent, []byte("not-a-dir"), 0600))
				outputPath := filepath.Join(blockedParent, "app.log")
				return []LogrusOption{
					WithOutputPath(outputPath),
					WithLogrusEnableRotate(false),
				}, outputPath
			},
			wantErr: true,
		},
		{
			name:        "error/rotate-invalid-strftime-pattern",
			description: "验证启用滚动且输出路径包含非法 strftime 片段时返回 rotatelogs 构造错误。",
			setup: func(t *testing.T) ([]LogrusOption, string) {
				outputPath := filepath.Join(t.TempDir(), "bad%")
				return []LogrusOption{
					WithOutputPath(outputPath),
					WithLogrusEnableRotate(true),
				}, outputPath
			},
			wantErr: true,
		},
		{
			name:        "error/output-is-directory",
			description: "验证禁用滚动且输出路径为目录时返回文件打开错误。",
			setup: func(t *testing.T) ([]LogrusOption, string) {
				outputPath := t.TempDir()
				return []LogrusOption{
					WithOutputPath(outputPath),
					WithLogrusEnableRotate(false),
				}, outputPath
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			giveOptions, giveOutput := tt.setup(t)
			got, err := NewLogrusLogger(giveOptions...)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			cleanupLoggerOutput(t, got)
			tt.assert(t, got, giveOutput)
		})
	}
}

// TestLogrusLogger_LevelFilteringAndFormattedMethods 验证 LogrusLogger 的级别过滤和格式化方法。
//
// 该测试覆盖 Debug/Debugf/Info/Infof/Warn/Warnf/Error/Errorf，确保低级别日志被过滤且输出消息稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestLogrusLogger_LevelFilteringAndFormattedMethods(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "logrus-methods.log")
	loggerInterface, err := NewLogrusLogger(
		WithOutputPath(outputPath),
		WithLogrusEnableRotate(false),
		WithJSONFormatter(time.RFC3339, false),
		WithLogrusLevel(WarnLevel),
	)
	require.NoError(t, err)
	cleanupLoggerOutput(t, loggerInterface)
	logger, ok := loggerInterface.(*LogrusLogger)
	require.True(t, ok)

	// 验证低于 WarnLevel 的日志方法执行但不会写入文件。
	logger.Debug("debug-hidden")
	logger.Debugf("debugf-%s", "hidden")
	logger.Info("info-hidden")
	logger.Infof("infof-%s", "hidden")

	// 验证 WarnLevel 及以上的普通和格式化日志会写入 JSON 记录。
	logger.Warn("warn-visible")
	logger.Warnf("warnf-%s", "visible")
	logger.Error("error-visible")
	logger.Errorf("errorf-%s", "visible")

	entries := readJSONLogEntries(t, outputPath)
	messages := jsonLogMessages(entries)
	assert.NotContains(t, messages, "debug-hidden")
	assert.NotContains(t, messages, "debugf-hidden")
	assert.NotContains(t, messages, "info-hidden")
	assert.NotContains(t, messages, "infof-hidden")
	assert.Contains(t, messages, "warn-visible")
	assert.Contains(t, messages, "warnf-visible")
	assert.Contains(t, messages, "error-visible")
	assert.Contains(t, messages, "errorf-visible")
}

// TestLogrusLogger_FatalMethodsUseConfiguredExitFunc 验证 LogrusLogger 的致命日志方法会输出日志并触发退出函数。
//
// 该测试替换底层 logrus Logger 的 ExitFunc 为可捕获 panic，避免真实 os.Exit，同时验证 Fatal 和 Fatalf 的消息输出和退出码。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestLogrusLogger_FatalMethodsUseConfiguredExitFunc(t *testing.T) {
	tests := []struct {
		name        string
		description string
		act         func(logger *LogrusLogger)
		wantMessage string
	}{
		{
			name:        "success/fatal",
			description: "验证 Fatal 输出致命日志并触发状态码 1 的退出函数。",
			act: func(logger *LogrusLogger) {
				logger.Fatal("fatal-event")
			},
			wantMessage: "fatal-event",
		},
		{
			name:        "success/fatalf",
			description: "验证 Fatalf 输出格式化致命日志并触发状态码 1 的退出函数。",
			act: func(logger *LogrusLogger) {
				logger.Fatalf("fatalf-%s", "event")
			},
			wantMessage: "fatalf-event",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			buffer := &bytes.Buffer{}
			loggerInterface, err := NewLogrusLogger(
				WithFormatter(&logrus.JSONFormatter{}),
				WithLogrusLevel(DebugLevel),
			)
			require.NoError(t, err)
			logger, ok := loggerInterface.(*LogrusLogger)
			require.True(t, ok)
			logger.logger.Logger.SetOutput(buffer)
			logger.logger.Logger.ExitFunc = func(code int) {
				panic(logrusExitPanic{code: code})
			}

			panicValue := capturePanic(func() {
				tt.act(logger)
			})

			exit, ok := panicValue.(logrusExitPanic)
			require.True(t, ok)
			assert.Equal(t, 1, exit.code)

			entries := parseJSONLogEntries(t, buffer.String())
			entry := findJSONLogEntry(t, entries, tt.wantMessage)
			assert.Equal(t, "fatal", entry["level"])
		})
	}
}

// TestLogrusLogger_LevelMappingBoundaries 验证 LogrusLogger 的级别映射边界。
//
// 该测试覆盖有效级别设置、未知自定义级别忽略以及底层未知 logrus 级别回退到 InfoLevel 的行为。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestLogrusLogger_LevelMappingBoundaries(t *testing.T) {
	loggerInterface, err := NewLogrusLogger(WithLogrusLevel(InfoLevel))
	require.NoError(t, err)
	logger, ok := loggerInterface.(*LogrusLogger)
	require.True(t, ok)

	logger.SetLevel(DebugLevel)
	assert.Equal(t, DebugLevel, logger.GetLevel())

	logger.SetLevel(Level(200))
	assert.Equal(t, DebugLevel, logger.GetLevel())

	logger.logger.Logger.SetLevel(logrus.TraceLevel)
	assert.Equal(t, InfoLevel, logger.GetLevel())
}

// TestLogrusLogger_FieldsAreImmutableAndSerialized 验证 LogrusLogger 的字段不可变和 JSON 序列化行为。
//
// 该测试覆盖 WithField、WithFields、字段输入 map 后续变更以及原始 Logger 隔离语义，确保结构化上下文稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestLogrusLogger_FieldsAreImmutableAndSerialized(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "logrus-fields.log")
	loggerInterface, err := NewLogrusLogger(
		WithOutputPath(outputPath),
		WithLogrusEnableRotate(false),
		WithJSONFormatter(time.RFC3339, false),
	)
	require.NoError(t, err)
	cleanupLoggerOutput(t, loggerInterface)
	baseLogger, ok := loggerInterface.(*LogrusLogger)
	require.True(t, ok)

	withRequest := baseLogger.WithField("request_id", "req-1")
	fields := map[string]interface{}{
		"user_id": float64(42),
		"admin":   true,
	}
	withUser := withRequest.WithFields(fields)
	fields["user_id"] = float64(100)
	fields["ignored"] = "new-value"

	baseLogger.Info("base-event")
	withRequest.Info("request-event")
	withUser.Infof("user-%s", "event")

	entries := readJSONLogEntries(t, outputPath)
	baseEntry := findJSONLogEntry(t, entries, "base-event")
	requestEntry := findJSONLogEntry(t, entries, "request-event")
	userEntry := findJSONLogEntry(t, entries, "user-event")

	assert.NotContains(t, baseEntry, "request_id")
	assert.NotContains(t, baseEntry, "user_id")

	assert.Equal(t, "req-1", requestEntry["request_id"])
	assert.NotContains(t, requestEntry, "user_id")

	assert.Equal(t, "req-1", userEntry["request_id"])
	assert.Equal(t, float64(42), userEntry["user_id"])
	assert.Equal(t, true, userEntry["admin"])
	assert.NotContains(t, userEntry, "ignored")
}

// TestGlobalLogger_ProxyFunctionsUseConfiguredLogger 验证全局日志代理函数委托到已配置 Logger。
//
// 该测试使用 fake Logger 覆盖 SetLevel、GetLevel、普通日志、格式化日志、Fatal、Fatalf 和字段代理，避免 stdout 与 os.Exit 副作用。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGlobalLogger_ProxyFunctionsUseConfiguredLogger(t *testing.T) {
	preserveGlobalLogger(t)
	fake := newFakeLogger(InfoLevel)
	SetLogger(fake)

	assert.Same(t, fake, GetLogger())

	SetLevel(WarnLevel)
	assert.Equal(t, WarnLevel, fake.level)
	assert.Equal(t, WarnLevel, GetLevel())

	Debug("debug-message")
	Debugf("debug-%s", "formatted")
	Info("info-message")
	Infof("info-%s", "formatted")
	Warn("warn-message")
	Warnf("warn-%s", "formatted")
	Error("error-message")
	Errorf("error-%s", "formatted")
	Fatal("fatal-message")
	Fatalf("fatal-%s", "formatted")

	withFieldLogger := WithField("request_id", "req-1")
	withFieldFake, ok := withFieldLogger.(*fakeLogger)
	require.True(t, ok)
	assert.Equal(t, "req-1", withFieldFake.fields["request_id"])

	withFieldsLogger := WithFields(map[string]interface{}{"user_id": 42})
	withFieldsFake, ok := withFieldsLogger.(*fakeLogger)
	require.True(t, ok)
	assert.Equal(t, 42, withFieldsFake.fields["user_id"])

	wantMethods := []string{
		"SetLevel",
		"GetLevel",
		"Debug",
		"Debugf",
		"Info",
		"Infof",
		"Warn",
		"Warnf",
		"Error",
		"Errorf",
		"Fatal",
		"Fatalf",
		"WithField",
		"WithFields",
	}
	assert.Equal(t, wantMethods, fakeLoggerCallMethods(fake.calls))
	assert.Equal(t, []interface{}{"debug-message"}, fake.calls[2].args)
	assert.Equal(t, "debug-%s", fake.calls[3].format)
	assert.Equal(t, []interface{}{"formatted"}, fake.calls[3].args)
	assert.Equal(t, []interface{}{"fatal-message"}, fake.calls[10].args)
	assert.Equal(t, "fatal-%s", fake.calls[11].format)
}

// TestGlobalLogger_DefaultAndInitLogger 验证全局 Logger 的默认懒加载和 InitLogger 初始化行为。
//
// 该测试覆盖 nil 全局状态下的默认创建、成功初始化替换以及初始化失败不覆盖既有 Logger 的行为。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGlobalLogger_DefaultAndInitLogger(t *testing.T) {
	preserveGlobalLogger(t)

	globalLoggerLock.Lock()
	globalLogger = nil
	globalLoggerLock.Unlock()

	defaultLogger := GetLogger()
	require.NotNil(t, defaultLogger)
	_, ok := defaultLogger.(*StdLogger)
	require.True(t, ok)
	assert.Same(t, defaultLogger, GetLogger())

	outputPath := filepath.Join(t.TempDir(), "global", "app.log")
	require.NoError(t, InitLogger(
		WithLogType(LogTypeStd),
		WithOutput(outputPath),
		WithLevel(ErrorLevel),
	))
	cleanupLoggerOutput(t, GetLogger())
	assert.Equal(t, ErrorLevel, GetLevel())
	Error("global-error-event")
	assert.Contains(t, readFileString(t, outputPath), "global-error-event")

	sentinel := newFakeLogger(InfoLevel)
	SetLogger(sentinel)
	err := InitLogger(WithLogType(LogType("unsupported")))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "初始化日志实例失败")
	assert.Same(t, sentinel, GetLogger())
}

// newBufferedStdLogger 构造写入内存缓冲区的 StdLogger 测试实例。
//
// 该辅助函数避免测试依赖 stdout 或固定文件，同时保留 StdLogger 的级别过滤、字段和格式化逻辑。
//
// 参数：
//   - t: 测试上下文，用于标记辅助函数调用栈。
//   - level: StdLogger 初始日志级别。
//
// 返回：
//   - *StdLogger: 写入内存缓冲区的标准库 Logger。
//   - *bytes.Buffer: 用于读取日志输出的缓冲区。
func newBufferedStdLogger(t *testing.T, level Level) (*StdLogger, *bytes.Buffer) {
	t.Helper()

	buffer := &bytes.Buffer{}
	return &StdLogger{
		logger: stdlog.New(buffer, "", 0),
		fields: make(map[string]interface{}),
		level:  level,
	}, buffer
}

// cleanupLoggerOutput 注册 Logger 底层输出的关闭清理逻辑。
//
// 该辅助函数用于关闭文件型 StdLogger、LogrusLogger 以及 rotatelogs writer，避免文件描述符泄漏并确保临时目录可删除。
//
// 参数：
//   - t: 测试上下文，用于注册清理函数并报告关闭失败。
//   - logger: 可能持有可关闭输出的 Logger。
func cleanupLoggerOutput(t *testing.T, logger Logger) {
	t.Helper()

	switch typedLogger := logger.(type) {
	case *StdLogger:
		registerCloserCleanup(t, typedLogger.logger.Writer())
	case *LogrusLogger:
		registerCloserCleanup(t, typedLogger.logger.Logger.Out)
	}
}

// registerCloserCleanup 在测试结束时关闭实现 io.Closer 的输出对象。
//
// 该辅助函数只处理真实可关闭资源，stdout、buffer 等非 closer 输出会被安全忽略。
//
// 参数：
//   - t: 测试上下文，用于注册清理函数并报告关闭失败。
//   - output: 待检查和关闭的输出对象。
func registerCloserCleanup(t *testing.T, output interface{}) {
	t.Helper()

	closer, ok := output.(io.Closer)
	if !ok || closer == os.Stdout || closer == os.Stderr {
		return
	}

	t.Cleanup(func() {
		assert.NoError(t, closer.Close())
	})
}

// requireSymlinkSupport 验证当前环境支持创建符号链接。
//
// 该辅助函数用于 gate rotatelogs 场景；rotatelogs 初始化会创建 link path，不支持 symlink 的环境会跳过该用例。
//
// 参数：
//   - t: 测试上下文，用于跳过不具备符号链接能力的环境。
func requireSymlinkSupport(t *testing.T) {
	t.Helper()

	dir := t.TempDir()
	target := filepath.Join(dir, "target.log")
	link := filepath.Join(dir, "link.log")
	require.NoError(t, os.WriteFile(target, []byte("probe"), 0600))
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("当前环境不支持符号链接，跳过依赖 rotatelogs link path 创建能力的用例：%v", err)
	}
}

// readRotatedJSONLogEntries 读取 rotatelogs 写入的真实轮转 JSON 日志文件。
//
// 该辅助函数不依赖 link path 内容，而是在 link path 同目录下查找 rotatelogs 生成的实际文件，提升跨平台稳定性。
//
// 参数：
//   - t: 测试上下文，用于报告读取或解析失败并标记辅助函数调用栈。
//   - linkPath: rotatelogs 配置中的链接路径。
//
// 返回：
//   - []map[string]interface{}: 实际轮转日志文件中的 JSON 记录。
func readRotatedJSONLogEntries(t *testing.T, linkPath string) []map[string]interface{} {
	t.Helper()

	matches, err := filepath.Glob(rotatedLogGlobPattern(linkPath))
	require.NoError(t, err)
	require.NotEmpty(t, matches)
	assert.NotContains(t, matches, linkPath)

	var entries []map[string]interface{}
	for _, match := range matches {
		entries = append(entries, readJSONLogEntries(t, match)...)
	}
	return entries
}

// rotatedLogGlobPattern 构造 rotatelogs 实际输出文件的 glob 模式。
//
// 该辅助函数根据生产代码 base+"-%Y%m%d%H"+ext 的命名规则生成匹配模式，避免断言符号链接路径内容。
//
// 参数：
//   - linkPath: rotatelogs 配置中的链接路径。
//
// 返回：
//   - string: 可用于 filepath.Glob 的实际轮转文件匹配模式。
func rotatedLogGlobPattern(linkPath string) string {
	ext := filepath.Ext(linkPath)
	base := strings.TrimSuffix(linkPath, ext)
	return base + "-*" + ext
}

// runStdLoggerFatalHelper 在子进程中执行 StdLogger 的致命日志方法。
//
// 该辅助函数仅供 TestStdLogger_FatalSubprocess 通过环境变量触发，用于隔离 os.Exit 副作用。
//
// 参数：
//   - helperCase: 要执行的致命日志方法标识。
//   - outputPath: 子进程写入致命日志的文件路径。
func runStdLoggerFatalHelper(helperCase string, outputPath string) {
	loggerInterface, err := NewStdLogger(outputPath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	logger := loggerInterface.(*StdLogger)
	logger.logger.SetFlags(0)

	switch helperCase {
	case "fatal":
		logger.Fatal("fatal-event")
	case "fatalf":
		logger.Fatalf("fatalf-%s", "event")
	default:
		os.Exit(3)
	}
}

// capturePanic 执行函数并返回其触发的 panic 值。
//
// 该辅助函数用于验证会通过 panic 表达可捕获退出信号的测试路径，同时保持调用方断言逻辑清晰。
//
// 参数：
//   - fn: 需要执行并捕获 panic 的函数。
//
// 返回：
//   - interface{}: 捕获到的 panic 值；如果未发生 panic，则返回 nil。
func capturePanic(fn func()) (panicValue interface{}) {
	defer func() {
		panicValue = recover()
	}()

	fn()
	return nil
}

// readFileString 读取文件内容并以字符串返回。
//
// 该辅助函数集中处理测试文件读取失败，保证调用方可以专注于行为断言。
//
// 参数：
//   - t: 测试上下文，用于报告读取失败并标记辅助函数调用栈。
//   - path: 要读取的文件路径。
//
// 返回：
//   - string: 文件的完整文本内容。
func readFileString(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}

// outputLines 将日志输出按非空行拆分。
//
// 该辅助函数用于在不依赖尾随换行的情况下断言标准库日志输出。
//
// 参数：
//   - output: 原始日志输出文本。
//
// 返回：
//   - []string: 去除空白后的非空日志行列表。
func outputLines(output string) []string {
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

// readJSONLogEntries 读取并解析按行输出的 JSON 日志。
//
// 该辅助函数避免测试依赖字段顺序，适用于 logrus JSON formatter 生成的日志文件。
//
// 参数：
//   - t: 测试上下文，用于报告读取或解析失败并标记辅助函数调用栈。
//   - path: JSON 日志文件路径。
//
// 返回：
//   - []map[string]interface{}: 每行 JSON 日志解析得到的结构化记录。
func readJSONLogEntries(t *testing.T, path string) []map[string]interface{} {
	t.Helper()

	return parseJSONLogEntries(t, readFileString(t, path))
}

// parseJSONLogEntries 解析按行输出的 JSON 日志文本。
//
// 该辅助函数用于文件日志和内存日志的共同解析，避免调用方依赖 JSON 字段顺序。
//
// 参数：
//   - t: 测试上下文，用于报告解析失败并标记辅助函数调用栈。
//   - content: 按行输出的 JSON 日志文本。
//
// 返回：
//   - []map[string]interface{}: 每行 JSON 日志解析得到的结构化记录。
func parseJSONLogEntries(t *testing.T, content string) []map[string]interface{} {
	t.Helper()

	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	lines := strings.Split(content, "\n")
	entries := make([]map[string]interface{}, 0, len(lines))
	for _, line := range lines {
		entry := make(map[string]interface{})
		require.NoError(t, json.Unmarshal([]byte(line), &entry))
		entries = append(entries, entry)
	}
	return entries
}

// findJSONLogEntry 查找指定消息对应的 JSON 日志记录。
//
// 该辅助函数通过 msg 字段定位日志记录，避免测试依赖日志行顺序之外的格式细节。
//
// 参数：
//   - t: 测试上下文，用于在未找到记录时报告断言失败并标记辅助函数调用栈。
//   - entries: 待查找的 JSON 日志记录列表。
//   - message: 期望匹配的日志消息。
//
// 返回：
//   - map[string]interface{}: 匹配到的 JSON 日志记录。
func findJSONLogEntry(t *testing.T, entries []map[string]interface{}, message string) map[string]interface{} {
	t.Helper()

	for _, entry := range entries {
		if entry["msg"] == message {
			return entry
		}
	}
	require.Failf(t, "missing JSON log entry", "message %q not found in entries %#v", message, entries)
	return nil
}

// jsonLogMessages 提取 JSON 日志记录中的消息字段。
//
// 该辅助函数用于断言日志过滤结果，避免测试关注完整 JSON 结构。
//
// 参数：
//   - entries: JSON 日志记录列表。
//
// 返回：
//   - []string: 按日志行顺序提取的消息列表。
func jsonLogMessages(entries []map[string]interface{}) []string {
	messages := make([]string, 0, len(entries))
	for _, entry := range entries {
		message, _ := entry["msg"].(string)
		messages = append(messages, message)
	}
	return messages
}

// preserveGlobalLogger 保存并在测试结束后恢复全局 Logger。
//
// 该辅助函数用于隔离会修改 globalLogger 的测试，避免污染其他测试用例。
//
// 参数：
//   - t: 测试上下文，用于注册清理函数并标记辅助函数调用栈。
func preserveGlobalLogger(t *testing.T) {
	t.Helper()

	globalLoggerLock.Lock()
	original := globalLogger
	globalLoggerLock.Unlock()

	t.Cleanup(func() {
		globalLoggerLock.Lock()
		defer globalLoggerLock.Unlock()
		globalLogger = original
	})
}

// newFakeLogger 构造用于全局代理断言的 fake Logger。
//
// 该辅助函数集中初始化 fakeLogger 的级别和字段容器，保证测试记录状态可预测。
//
// 参数：
//   - level: fake Logger 的初始日志级别。
//
// 返回：
//   - *fakeLogger: 可记录所有 Logger 接口调用的测试替身。
func newFakeLogger(level Level) *fakeLogger {
	return &fakeLogger{
		level:  level,
		fields: make(map[string]interface{}),
	}
}

// fakeLoggerCallMethods 提取 fakeLogger 调用记录中的方法名。
//
// 该辅助函数用于稳定断言全局代理函数的调用顺序。
//
// 参数：
//   - calls: fakeLogger 记录的方法调用列表。
//
// 返回：
//   - []string: 调用记录中的方法名列表。
func fakeLoggerCallMethods(calls []fakeLoggerCall) []string {
	methods := make([]string, 0, len(calls))
	for _, call := range calls {
		methods = append(methods, call.method)
	}
	return methods
}

// SetLevel 记录 SetLevel 调用并更新 fakeLogger 的当前级别。
//
// 参数：
//   - level: 要设置的日志级别。
func (f *fakeLogger) SetLevel(level Level) {
	f.level = level
	f.calls = append(f.calls, fakeLoggerCall{method: "SetLevel", level: level})
}

// GetLevel 记录 GetLevel 调用并返回 fakeLogger 的当前级别。
//
// 返回：
//   - Level: 当前日志级别。
func (f *fakeLogger) GetLevel() Level {
	f.calls = append(f.calls, fakeLoggerCall{method: "GetLevel", level: f.level})
	return f.level
}

// Debug 记录 Debug 调用参数。
//
// 参数：
//   - args: 代理转发的日志参数。
func (f *fakeLogger) Debug(args ...interface{}) {
	f.recordArgs("Debug", args...)
}

// Debugf 记录 Debugf 调用格式和参数。
//
// 参数：
//   - format: 代理转发的格式字符串。
//   - args: 代理转发的格式化参数。
func (f *fakeLogger) Debugf(format string, args ...interface{}) {
	f.recordFormat("Debugf", format, args...)
}

// Info 记录 Info 调用参数。
//
// 参数：
//   - args: 代理转发的日志参数。
func (f *fakeLogger) Info(args ...interface{}) {
	f.recordArgs("Info", args...)
}

// Infof 记录 Infof 调用格式和参数。
//
// 参数：
//   - format: 代理转发的格式字符串。
//   - args: 代理转发的格式化参数。
func (f *fakeLogger) Infof(format string, args ...interface{}) {
	f.recordFormat("Infof", format, args...)
}

// Warn 记录 Warn 调用参数。
//
// 参数：
//   - args: 代理转发的日志参数。
func (f *fakeLogger) Warn(args ...interface{}) {
	f.recordArgs("Warn", args...)
}

// Warnf 记录 Warnf 调用格式和参数。
//
// 参数：
//   - format: 代理转发的格式字符串。
//   - args: 代理转发的格式化参数。
func (f *fakeLogger) Warnf(format string, args ...interface{}) {
	f.recordFormat("Warnf", format, args...)
}

// Error 记录 Error 调用参数。
//
// 参数：
//   - args: 代理转发的日志参数。
func (f *fakeLogger) Error(args ...interface{}) {
	f.recordArgs("Error", args...)
}

// Errorf 记录 Errorf 调用格式和参数。
//
// 参数：
//   - format: 代理转发的格式字符串。
//   - args: 代理转发的格式化参数。
func (f *fakeLogger) Errorf(format string, args ...interface{}) {
	f.recordFormat("Errorf", format, args...)
}

// Fatal 记录 Fatal 调用参数且不会退出进程。
//
// 参数：
//   - args: 代理转发的日志参数。
func (f *fakeLogger) Fatal(args ...interface{}) {
	f.recordArgs("Fatal", args...)
}

// Fatalf 记录 Fatalf 调用格式和参数且不会退出进程。
//
// 参数：
//   - format: 代理转发的格式字符串。
//   - args: 代理转发的格式化参数。
func (f *fakeLogger) Fatalf(format string, args ...interface{}) {
	f.recordFormat("Fatalf", format, args...)
}

// WithField 记录 WithField 调用并返回包含新增字段的 fakeLogger 副本。
//
// 参数：
//   - key: 字段名。
//   - value: 字段值。
//
// 返回：
//   - Logger: 包含新增字段的 fakeLogger 副本。
func (f *fakeLogger) WithField(key string, value interface{}) Logger {
	f.calls = append(f.calls, fakeLoggerCall{
		method: "WithField",
		fields: map[string]interface{}{key: value},
	})
	child := f.clone()
	child.fields[key] = value
	return child
}

// WithFields 记录 WithFields 调用并返回包含新增字段的 fakeLogger 副本。
//
// 参数：
//   - fields: 字段映射。
//
// 返回：
//   - Logger: 包含新增字段的 fakeLogger 副本。
func (f *fakeLogger) WithFields(fields map[string]interface{}) Logger {
	copiedFields := copyFields(fields)
	f.calls = append(f.calls, fakeLoggerCall{
		method: "WithFields",
		fields: copiedFields,
	})
	child := f.clone()
	for key, value := range copiedFields {
		child.fields[key] = value
	}
	return child
}

// recordArgs 记录非格式化日志方法调用。
//
// 参数：
//   - method: 被调用的方法名。
//   - args: 方法接收到的日志参数。
func (f *fakeLogger) recordArgs(method string, args ...interface{}) {
	f.calls = append(f.calls, fakeLoggerCall{method: method, args: append([]interface{}(nil), args...)})
}

// recordFormat 记录格式化日志方法调用。
//
// 参数：
//   - method: 被调用的方法名。
//   - format: 方法接收到的格式字符串。
//   - args: 方法接收到的格式化参数。
func (f *fakeLogger) recordFormat(method string, format string, args ...interface{}) {
	f.calls = append(f.calls, fakeLoggerCall{method: method, format: format, args: append([]interface{}(nil), args...)})
}

// clone 复制 fakeLogger 的级别和字段状态。
//
// 返回：
//   - *fakeLogger: 拥有独立字段映射的 fakeLogger 副本。
func (f *fakeLogger) clone() *fakeLogger {
	return &fakeLogger{
		level:  f.level,
		fields: copyFields(f.fields),
	}
}

// copyFields 复制字段映射。
//
// 该辅助函数用于避免 fakeLogger 字段状态受调用方后续修改影响。
//
// 参数：
//   - fields: 待复制的字段映射。
//
// 返回：
//   - map[string]interface{}: 复制后的字段映射。
func copyFields(fields map[string]interface{}) map[string]interface{} {
	copied := make(map[string]interface{}, len(fields))
	for key, value := range fields {
		copied[key] = value
	}
	return copied
}
