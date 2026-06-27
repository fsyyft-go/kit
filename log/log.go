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

	// LoggerFormatType 定义 Logrus 日志输出格式。
	//
	// 可选值包括：
	//   - TextFormat：使用 logrus.TextFormatter 输出文本日志。
	//   - JSONFormat：使用 logrus.JSONFormatter 输出 JSON 日志。
	LoggerFormatType string

	// Logger 定义了统一的日志接口。
	// 该接口提供了以下功能：
	//   - 支持多个日志级别（Debug、Info、Warn、Error、Fatal）。
	//   - 提供格式化和非格式化的日志记录方法。
	//   - 支持结构化日志记录。
	//   - 支持日志级别的动态调整。
	//   - 提供上下文信息的添加和管理。
	Logger interface {
		// SetLevel 设置日志级别。
		// 只有大于或等于设置级别的日志才会被记录。
		//
		// 参数：
		//   - level：要设置的日志级别。
		SetLevel(level Level)

		// GetLevel 获取当前的日志级别。
		//
		// 返回：
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
		// 返回：
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
		// 返回：
		//   - Logger：新的日志实例。
		WithFields(fields map[string]interface{}) Logger
	}

	// LoggerOptions 定义 NewLogger 使用的日志配置。
	//
	// 零值会在 NewLogger 中与默认配置合并使用。Output 为空时输出到标准输出；
	// EnableRotate、RotateTime、MaxAge 和 FormatType 主要影响 Logrus 实现。
	LoggerOptions struct {
		// Type 指定日志实现类型。可选值包括：
		//   - LogTypeConsole：使用标准库日志并输出到标准输出。
		//   - LogTypeStd：使用标准库日志，可写入标准输出或指定文件。
		//   - LogTypeLogrus：使用 Logrus 日志实现，支持格式化和文件轮转配置。
		Type LogType
		// Level 指定日志过滤级别。可选值包括：
		//   - DebugLevel：输出调试及以上级别日志。
		//   - InfoLevel：输出信息及以上级别日志。
		//   - WarnLevel：输出警告及以上级别日志。
		//   - ErrorLevel：输出错误及以上级别日志。
		//   - FatalLevel：输出致命错误日志。
		Level Level
		// Output 指定日志文件路径。空字符串表示写入标准输出。
		Output string
		// EnableRotate 控制 Logrus 文件输出是否启用日志轮转。
		EnableRotate bool
		// RotateTime 指定 Logrus 日志轮转周期。零值会传递给底层轮转实现。
		RotateTime time.Duration
		// MaxAge 指定 Logrus 轮转日志的最大保留时间。零值表示由底层实现处理。
		MaxAge time.Duration
		// FormatType 指定 Logrus 输出格式。
		//
		// 可选值包括：
		//   - TextFormat：使用文本格式输出。
		//   - JSONFormat：使用 JSON 格式输出。
		FormatType LoggerFormatType
	}

	// Option 定义日志配置修改函数。
	//
	// 参数：
	//   - *LoggerOptions：待修改的日志配置，调用方应保证其非 nil。
	Option func(*LoggerOptions)
)

// String 返回日志级别的字符串表示。
//
// 参数：无。
//
// 返回：
//   - string：已定义级别返回 debug、info、warn、error 或 fatal；未知级别返回 unknown。
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
//
// 参数：
//   - level：日志级别字符串，支持 debug、info、warn、error 和 fatal。
//
// 返回：
//   - Level：解析成功时返回对应日志级别；解析失败时返回 InfoLevel 作为回退值。
//   - error：level 不属于支持值时返回错误，错误消息包含未知级别字符串。
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

// WithLogType 设置 NewLogger 使用的日志实现类型。
//
// 参数：
//   - logType：日志实现类型，可选值包括：
//   - LogTypeConsole：使用标准库日志并强制输出到标准输出。
//   - LogTypeStd：使用标准库日志，可按 Output 写入文件或标准输出。
//   - LogTypeLogrus：使用 Logrus 日志实现，支持格式化和文件轮转配置。
//
// 返回：
//   - Option：应用于 LoggerOptions 的配置选项。
func WithLogType(logType LogType) Option {
	return func(opts *LoggerOptions) {
		opts.Type = logType
	}
}

// WithFormatType 设置 Logrus 日志输出格式类型。
//
// 参数：
//   - formatType：日志输出格式类型，可选值包括：
//   - TextFormat：使用文本格式输出日志。
//   - JSONFormat：使用 JSON 格式输出日志。
//
// 返回：
//   - Option：应用于 LoggerOptions 的配置选项。
func WithFormatType(formatType LoggerFormatType) Option {
	return func(opts *LoggerOptions) {
		opts.FormatType = formatType
	}
}

// WithLevel 设置日志过滤级别。
//
// 参数：
//   - level：日志过滤级别，可选值包括：
//   - DebugLevel：输出调试及以上级别日志。
//   - InfoLevel：输出信息及以上级别日志。
//   - WarnLevel：输出警告及以上级别日志。
//   - ErrorLevel：输出错误及以上级别日志。
//   - FatalLevel：输出致命错误日志。
//
// 返回：
//   - Option：应用于 LoggerOptions 的配置选项。
func WithLevel(level Level) Option {
	return func(opts *LoggerOptions) {
		opts.Level = level
	}
}

// WithOutput 设置日志输出路径。
//
// 参数：
//   - output：日志文件路径；空字符串表示输出到标准输出。
//
// 返回：
//   - Option：应用于 LoggerOptions 的配置选项。
func WithOutput(output string) Option {
	return func(opts *LoggerOptions) {
		opts.Output = output
	}
}

// WithEnableRotate 设置 Logrus 文件输出是否启用日志轮转。
//
// 参数：
//   - enable：是否启用日志轮转；true 表示启用，false 表示禁用。
//
// 返回：
//   - Option：应用于 LoggerOptions 的配置选项。
func WithEnableRotate(enable bool) Option {
	return func(opts *LoggerOptions) {
		opts.EnableRotate = enable
	}
}

// WithRotateTime 设置 Logrus 文件输出的日志轮转周期。
//
// 参数：
//   - duration：日志轮转周期，传递给底层 rotatelogs 实现。
//
// 返回：
//   - Option：应用于 LoggerOptions 的配置选项。
func WithRotateTime(duration time.Duration) Option {
	return func(opts *LoggerOptions) {
		opts.RotateTime = duration
	}
}

// WithMaxAge 设置 Logrus 轮转日志文件的最大保留时间。
//
// 参数：
//   - duration：轮转日志文件的最大保留时间，传递给底层 rotatelogs 实现。
//
// 返回：
//   - Option：应用于 LoggerOptions 的配置选项。
func WithMaxAge(duration time.Duration) Option {
	return func(opts *LoggerOptions) {
		opts.MaxAge = duration
	}
}

// NewLogger 创建一个新的日志实例。
//
// 未传入 options 时使用标准库日志实现、InfoLevel、标准输出、JSONFormat 以及
// Logrus 轮转默认值。LogTypeConsole 会忽略 Output 并写入标准输出。
//
// 参数：
//   - options：可选配置项，按传入顺序应用；未传入时使用默认配置。
//
// 返回：
//   - Logger：初始化完成的日志实例。
//   - error：日志类型不受支持、文件输出路径创建失败、文件打开失败或 Logrus 轮转 writer 创建失败时返回错误。
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
