// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package log

import (
	"fmt"
	"time"
)

const (
	// DebugLevel 表示调试级别，用于记录详细的调试信息。
	// 这个级别的日志通常只在开发环境启用。
	DebugLevel Level = iota

	// InfoLevel 表示信息级别，用于记录正常的操作信息。
	// 这个级别的日志用于跟踪应用的正常运行状态。
	InfoLevel

	// WarnLevel 表示警告级别，用于记录可能的问题或异常情况。
	// 这个级别的日志表示出现了值得注意的情况，但不影响系统的正常运行。
	WarnLevel

	// ErrorLevel 表示错误级别，用于记录错误信息。
	// 这个级别的日志表示出现了影响系统正常运行的错误。
	ErrorLevel

	// FatalLevel 表示致命错误级别，记录后会导致程序退出。
	// 这个级别的日志表示出现了无法恢复的严重错误。
	FatalLevel
)

const (
	// TextFormat 表示文本格式的日志输出。
	TextFormat LoggerFormatType = "text"
	// JSONFormat 表示 JSON 格式的日志输出。
	JSONFormat LoggerFormatType = "json"
)

var (
	// timestampFormat 定义了日志时间戳的格式。
	timestampFormat = "2006-01-02 15:04:05.000"
	// disableColors 表示是否禁用颜色输出，true 表示禁用。
	disableColors = true
	// fullTimestamp 表示是否显示完整时间戳，true 表示显示。
	fullTimestamp = true
	// prettyPrint 表示是否美化 JSON 输出，false 表示不美化。
	prettyPrint = false
)

type (
	// Level 定义了日志的级别类型，用于控制日志的输出粒度。
	Level int

	// LoggerFormatType 定义了日志输出格式的类型。
	LoggerFormatType string

	// Logger 定义了统一的日志接口。
	// 该接口提供了以下功能：
	// - 支持多个日志级别（Debug、Info、Warn、Error、Fatal）。
	// - 提供格式化和非格式化的日志记录方法。
	// - 支持结构化日志记录。
	// - 支持日志级别的动态调整。
	// - 提供上下文信息的添加和管理。
	Logger interface {
		// SetLevel 设置日志级别。
		// 只有大于或等于设置级别的日志才会被记录。
		//
		// 参数：
		//   - level：要设置的日志级别。
		SetLevel(level Level)

		// GetLevel 获取当前的日志级别。
		//
		// 返回值：
		//   - Level：当前的日志级别。
		GetLevel() Level

		// Debug 记录调试级别的日志。
		// 参数 args 支持任意类型的值，这些值会被转换为字符串并连接。
		// 调试日志应该包含有助于诊断问题的详细信息。
		//
		// 参数：
		//   - args：要记录的日志内容，支持多个参数。
		Debug(args ...interface{})

		// Debugf 记录格式化的调试级别日志。
		// 参数 format 是格式化字符串，args 是对应的参数。
		// 支持标准的 Printf 风格的格式化。
		//
		// 参数：
		//   - format：格式化字符串。
		//   - args：格式化参数。
		Debugf(format string, args ...interface{})

		// Info 记录信息级别的日志。
		// 参数 args 支持任意类型的值，这些值会被转换为字符串并连接。
		// 信息日志应该记录系统的正常运行状态。
		//
		// 参数：
		//   - args：要记录的日志内容，支持多个参数。
		Info(args ...interface{})

		// Infof 记录格式化的信息级别日志。
		// 参数 format 是格式化字符串，args 是对应的参数。
		// 支持标准的 Printf 风格的格式化。
		//
		// 参数：
		//   - format：格式化字符串。
		//   - args：格式化参数。
		Infof(format string, args ...interface{})

		// Warn 记录警告级别的日志。
		// 参数 args 支持任意类型的值，这些值会被转换为字符串并连接。
		// 警告日志应该包含可能导致问题的情况。
		//
		// 参数：
		//   - args：要记录的日志内容，支持多个参数。
		Warn(args ...interface{})

		// Warnf 记录格式化的警告级别日志。
		// 参数 format 是格式化字符串，args 是对应的参数。
		// 支持标准的 Printf 风格的格式化。
		//
		// 参数：
		//   - format：格式化字符串。
		//   - args：格式化参数。
		Warnf(format string, args ...interface{})

		// Error 记录错误级别的日志。
		// 参数 args 支持任意类型的值，这些值会被转换为字符串并连接。
		// 错误日志应该包含错误的详细信息和上下文。
		//
		// 参数：
		//   - args：要记录的日志内容，支持多个参数。
		Error(args ...interface{})

		// Errorf 记录格式化的错误级别日志。
		// 参数 format 是格式化字符串，args 是对应的参数。
		// 支持标准的 Printf 风格的格式化。
		//
		// 参数：
		//   - format：格式化字符串。
		//   - args：格式化参数。
		Errorf(format string, args ...interface{})

		// Fatal 记录致命错误级别的日志。
		// 参数 args 支持任意类型的值，这些值会被转换为字符串并连接。
		// 记录日志后会导致程序以状态码 1 退出。
		// 这个方法应该只在程序无法继续运行时使用。
		//
		// 参数：
		//   - args：要记录的日志内容，支持多个参数。
		Fatal(args ...interface{})

		// Fatalf 记录格式化的致命错误级别日志。
		// 参数 format 是格式化字符串，args 是对应的参数。
		// 支持标准的 Printf 风格的格式化。
		// 记录日志后会导致程序以状态码 1 退出。
		//
		// 参数：
		//   - format：格式化字符串。
		//   - args：格式化参数。
		Fatalf(format string, args ...interface{})

		// WithField 添加一个字段到日志上下文。
		// 参数 key 是字段名，value 是字段值。
		// 返回一个新的 Logger 实例，原实例不会被修改。
		// 这个方法用于添加结构化的上下文信息到日志中。
		//
		// 参数：
		//   - key：字段名。
		//   - value：字段值。
		//
		// 返回值：
		//   - Logger：新的日志实例。
		WithField(key string, value interface{}) Logger

		// WithFields 添加多个字段到日志上下文。
		// 参数 fields 是要添加的字段映射。
		// 返回一个新的 Logger 实例，原实例不会被修改。
		// 这个方法用于一次性添加多个结构化字段。
		//
		// 参数：
		//   - fields：字段映射。
		//
		// 返回值：
		//   - Logger：新的日志实例。
		WithFields(fields map[string]interface{}) Logger
	}

	// LoggerOptions 定义了日志配置选项。
	// 该结构体提供了以下功能：
	// - 配置日志的基本行为。
	// - 控制日志的输出目标。
	// - 管理日志的滚动策略。
	// - 设置日志的格式类型。
	LoggerOptions struct {
		// Type 指定日志实现类型。
		Type LogType
		// Level 指定日志级别。
		Level Level
		// Output 指定日志输出路径。
		Output string
		// EnableRotate 是否启用日志滚动。
		EnableRotate bool
		// RotateTime 日志滚动时间间隔。
		RotateTime time.Duration
		// MaxAge 日志保留时间。
		MaxAge time.Duration
		// FormatType 指定日志输出格式类型。
		FormatType LoggerFormatType
	}

	// Option 定义了日志配置的函数选项。
	Option func(*LoggerOptions)
)

// String 返回日志级别的字符串表示。
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// ParseLevel 从字符串解析日志级别。
func ParseLevel(level string) (Level, error) {
	switch level {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	default:
		return InfoLevel, fmt.Errorf("unknown level: %s", level)
	}
}

// WithLogType 设置日志类型。
//
// 参数：
//   - logType：要设置的日志类型。
//
// 返回值：
//   - 返回一个配置选项函数，可用于配置日志实例。
func WithLogType(logType LogType) Option {
	return func(opts *LoggerOptions) {
		opts.Type = logType
	}
}

// WithFormatType 设置日志输出格式类型。
//
// 参数：
//   - formatType：日志输出格式类型，可选值包括 TextFormat、JSONFormat。
//
// 返回值：
//   - 返回一个配置选项函数，可用于配置日志实例。
func WithFormatType(formatType LoggerFormatType) Option {
	return func(opts *LoggerOptions) {
		opts.FormatType = formatType
	}
}

// WithLevel 设置日志级别。
//
// 参数：
//   - level：要设置的日志级别。
//
// 返回值：
//   - 返回一个配置选项函数，可用于配置日志实例。
func WithLevel(level Level) Option {
	return func(opts *LoggerOptions) {
		opts.Level = level
	}
}

// WithOutput 设置日志输出路径。
//
// 参数：
//   - output：日志文件的输出路径，空字符串表示输出到标准输出。
//
// 返回值：
//   - 返回一个配置选项函数，可用于配置日志实例。
func WithOutput(output string) Option {
	return func(opts *LoggerOptions) {
		opts.Output = output
	}
}

// WithEnableRotate 设置是否启用日志滚动。
//
// 参数：
//   - enable：是否启用日志滚动，true 表示启用，false 表示禁用。
//
// 返回值：
//   - 返回一个配置选项函数，可用于配置日志实例。
func WithEnableRotate(enable bool) Option {
	return func(opts *LoggerOptions) {
		opts.EnableRotate = enable
	}
}

// WithRotateTime 设置日志滚动时间间隔。
//
// 参数：
//   - duration：日志滚动的时间间隔。
//
// 返回值：
//   - 返回一个配置选项函数，可用于配置日志实例。
func WithRotateTime(duration time.Duration) Option {
	return func(opts *LoggerOptions) {
		opts.RotateTime = duration
	}
}

// WithMaxAge 设置日志保留时间。
//
// 参数：
//   - duration：日志文件的最大保留时间。
//
// 返回值：
//   - 返回一个配置选项函数，可用于配置日志实例。
func WithMaxAge(duration time.Duration) Option {
	return func(opts *LoggerOptions) {
		opts.MaxAge = duration
	}
}

// NewLogger 创建一个新的日志实例。
//
// 参数：
//   - options：可选的配置选项列表，用于自定义日志记录器的行为。
//
// 返回值：
//   - Logger：返回创建的日志实例。
//   - error：返回创建过程中可能发生的错误。
func NewLogger(options ...Option) (Logger, error) {
	// 默认配置。
	opts := &LoggerOptions{
		Type:         LogTypeStd,
		Level:        InfoLevel,
		Output:       "",
		EnableRotate: true,               // 默认启用日志滚动
		RotateTime:   time.Hour,          // 默认每小时滚动一次
		MaxAge:       time.Hour * 24 * 7, // 默认保留7天
		FormatType:   JSONFormat,         // 默认使用 JSON 格式
	}

	// 应用所有选项。
	for _, option := range options {
		option(opts)
	}

	var logger Logger
	var err error

	switch opts.Type {
	case LogTypeConsole:
		logger, err = NewStdLogger("")
	case LogTypeStd:
		logger, err = NewStdLogger(opts.Output)
	case LogTypeLogrus:
		// 使用 WithOutputPath 和其他选项创建 Logrus 日志实例。
		logrusOpts := []LogrusOption{
			WithOutputPath(opts.Output),
			WithLogrusLevel(opts.Level),
			WithLogrusEnableRotate(opts.EnableRotate),
			WithLogrusRotateTime(opts.RotateTime),
			WithLogrusMaxAge(opts.MaxAge),
		}

		// 根据格式类型设置格式化器。
		switch opts.FormatType {
		case TextFormat:
			logrusOpts = append(logrusOpts,
				WithTextFormatter(timestampFormat, fullTimestamp, disableColors),
			)
		case JSONFormat:
			logrusOpts = append(logrusOpts,
				WithJSONFormatter(timestampFormat, prettyPrint),
			)
		}

		logger, err = NewLogrusLogger(logrusOpts...)
	default:
		return nil, fmt.Errorf("不支持的日志类型：%s", opts.Type)
	}

	if nil != err {
		return nil, fmt.Errorf("创建日志实例失败：%v", err)
	}

	// 设置日志级别。
	logger.SetLevel(opts.Level)

	return logger, nil
}
