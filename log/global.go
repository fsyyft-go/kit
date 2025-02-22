// Copyright 2024 fsyyft-go Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package log 提供了全局日志功能，支持多种日志后端的统一管理。
package log

import (
	"fmt"
	"sync"
)

type (
	// LogType 定义了支持的日志类型，用于在初始化时选择具体的日志实现。
	LogType string
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

// InitLogger 初始化全局日志实例。
// 使用可选的配置选项来配置日志行为。
// 如果没有提供任何选项，将使用默认配置：
//   - 日志类型：LogTypeStd
//   - 日志级别：InfoLevel
//   - 输出路径：标准输出
//
// 返回初始化过程中可能发生的错误。
func InitLogger(options ...Option) error {
	logger, err := NewLogger(options...)
	if err != nil {
		return fmt.Errorf("初始化日志实例失败：%v", err)
	}

	SetLogger(logger)
	return nil
}

// SetLevel 设置全局日志级别。
func SetLevel(level Level) {
	GetLogger().SetLevel(level)
}

// GetLevel 获取全局日志级别。
func GetLevel() Level {
	return GetLogger().GetLevel()
}

// SetLogger 设置全局日志实例。
func SetLogger(logger Logger) {
	globalLoggerLock.Lock()
	defer globalLoggerLock.Unlock()
	globalLogger = logger
}

// GetLogger 获取全局日志实例。
// 如果全局日志实例未设置，则返回一个默认的标准输出日志实例。
func GetLogger() Logger {
	globalLoggerLock.RLock()
	defer globalLoggerLock.RUnlock()

	if globalLogger == nil {
		stdLogger, err := NewLogger()
		if err != nil {
			panic(fmt.Sprintf("创建默认日志器失败：%v", err))
		}
		globalLogger = stdLogger
	}

	return globalLogger
}

// Debug 使用全局日志实例记录调试级别的日志。
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf 使用全局日志实例记录格式化的调试级别日志。
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Info 使用全局日志实例记录信息级别的日志。
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof 使用全局日志实例记录格式化的信息级别日志。
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn 使用全局日志实例记录警告级别的日志。
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf 使用全局日志实例记录格式化的警告级别日志。
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error 使用全局日志实例记录错误级别的日志。
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf 使用全局日志实例记录格式化的错误级别日志。
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal 使用全局日志实例记录致命错误级别的日志。
// 记录日志后会导致程序退出。
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf 使用全局日志实例记录格式化的致命错误级别日志。
// 记录日志后会导致程序退出。
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// WithField 使用全局日志实例添加一个结构化字段。
// 返回一个新的 Logger 实例，包含添加的字段。
func WithField(key string, value interface{}) Logger {
	return GetLogger().WithField(key, value)
}

// WithFields 使用全局日志实例添加多个结构化字段。
// 返回一个新的 Logger 实例，包含添加的字段。
func WithFields(fields map[string]interface{}) Logger {
	return GetLogger().WithFields(fields)
}
