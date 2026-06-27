// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package log

import (
	"fmt"
	"sync"
)

const (
	// LogTypeConsole 表示控制台日志类型。
	// 这种类型的日志会直接输出到标准输出，适合开发调试使用。
	LogTypeConsole LogType = "console"

	// LogTypeStd 表示标准库日志类型。
	// 使用 Go 标准库的 log 包实现，提供基本的日志功能。
	LogTypeStd LogType = "std"

	// LogTypeLogrus 表示 Logrus 日志类型。
	// 使用 Logrus 库实现，提供丰富的日志功能，包括结构化日志、多种输出格式等。
	LogTypeLogrus LogType = "logrus"
)

var (
	// globalLogger 是全局日志实例。
	globalLogger Logger
	// globalLoggerLock 用于保护全局日志实例的并发访问。
	globalLoggerLock sync.RWMutex
)

type (
	// LogType 定义 NewLogger 可选择的日志实现类型。
	//
	// 可选值包括：
	//   - LogTypeConsole：使用标准库日志并输出到标准输出。
	//   - LogTypeStd：使用标准库日志，可写入标准输出或指定文件。
	//   - LogTypeLogrus：使用 Logrus 日志实现，支持格式化和文件轮转配置。
	LogType string
)

// InitLogger 初始化包级全局日志实例。
//
// 未传入 options 时使用 NewLogger 的默认配置。初始化成功后会替换当前全局 Logger；
// 初始化失败时保持既有全局 Logger 不变。
//
// 参数：
//   - options：可选配置项，按传入顺序应用；未传入时使用默认配置。
//
// 返回：
//   - error：NewLogger 返回错误时包装并返回；成功时返回 nil。
func InitLogger(options ...Option) error {
	logger, err := NewLogger(options...)
	if nil != err {
		return fmt.Errorf("初始化日志实例失败：%v", err)
	}

	SetLogger(logger)
	return nil
}

// SetLevel 设置全局日志级别。
//
// 参数：
//   - level：日志过滤级别，可选值包括 DebugLevel、InfoLevel、WarnLevel、ErrorLevel 和 FatalLevel。
func SetLevel(level Level) {
	GetLogger().SetLevel(level)
}

// GetLevel 获取全局日志级别。
//
// 参数：无。
//
// 返回：
//   - Level：当前全局 Logger 的日志过滤级别。
func GetLevel() Level {
	return GetLogger().GetLevel()
}

// SetLogger 设置包级全局日志实例。
//
// 参数：
//   - logger：要设置为全局实例的日志记录器；传入 nil 会清空当前全局实例，并使下一次 GetLogger 重新创建默认实例。
func SetLogger(logger Logger) {
	globalLoggerLock.Lock()
	defer globalLoggerLock.Unlock()
	globalLogger = logger
}

// GetLogger 获取包级全局日志实例。
//
// 如果尚未设置全局 Logger，GetLogger 会惰性创建默认日志实例并缓存；默认实例写入标准输出。
// 默认日志实例创建失败时会 panic。
//
// 参数：无。
//
// 返回：
//   - Logger：当前全局日志实例。
func GetLogger() Logger {
	globalLoggerLock.RLock()
	defer globalLoggerLock.RUnlock()

	if nil == globalLogger {
		stdLogger, err := NewLogger()
		if nil != err {
			panic(fmt.Sprintf("创建默认日志器失败：%v", err))
		}
		globalLogger = stdLogger
	}

	return globalLogger
}

// Debug 使用全局日志实例记录调试级别的日志。
//
// 参数：
//   - args：要记录的内容，支持任意类型的值。
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf 使用全局日志实例记录格式化的调试级别日志。
//
// 参数：
//   - format：格式化字符串。
//   - args：格式化参数。
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Info 使用全局日志实例记录信息级别的日志。
//
// 参数：
//   - args：要记录的内容，支持任意类型的值。
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof 使用全局日志实例记录格式化的信息级别日志。
//
// 参数：
//   - format：格式化字符串。
//   - args：格式化参数。
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn 使用全局日志实例记录警告级别的日志。
//
// 参数：
//   - args：要记录的内容，支持任意类型的值。
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf 使用全局日志实例记录格式化的警告级别日志。
//
// 参数：
//   - format：格式化字符串。
//   - args：格式化参数。
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error 使用全局日志实例记录错误级别的日志。
//
// 参数：
//   - args：要记录的内容，支持任意类型的值。
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf 使用全局日志实例记录格式化的错误级别日志。
//
// 参数：
//   - format：格式化字符串。
//   - args：格式化参数。
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal 使用全局日志实例记录致命错误级别的日志。
// 记录日志后会导致程序退出。
//
// 参数：
//   - args：要记录的内容，支持任意类型的值。
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf 使用全局日志实例记录格式化的致命错误级别日志。
// 记录日志后会导致程序退出。
//
// 参数：
//   - format：格式化字符串。
//   - args：格式化参数。
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// WithField 使用全局日志实例添加一个结构化字段。
//
// 参数：
//   - key：字段名。
//   - value：字段值。
//
// 返回：
//   - Logger：返回一个新的 Logger 实例，包含添加的字段。
func WithField(key string, value interface{}) Logger {
	return GetLogger().WithField(key, value)
}

// WithFields 使用全局日志实例添加多个结构化字段。
//
// 参数：
//   - fields：要添加的字段映射。
//
// 返回：
//   - Logger：返回一个新的 Logger 实例，包含添加的字段。
func WithFields(fields map[string]interface{}) Logger {
	return GetLogger().WithFields(fields)
}
