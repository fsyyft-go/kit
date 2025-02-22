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

// Package log 提供了基于标准库的日志实现。
package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// StdLogger 实现了 Logger 接口，使用 Go 标准库的 log 包作为底层实现。
// 这个实现提供了基本的日志功能：
// - 支持不同的日志级别。
// - 支持结构化字段。
// - 支持文件输出。
// - 支持格式化日志。
type StdLogger struct {
	// logger 是标准库的日志实例。
	logger *log.Logger
	// fields 存储结构化字段信息。
	fields map[string]interface{}
	// level 存储当前的日志级别。
	level Level
}

// NewStdLogger 创建一个新的 StdLogger 实例。
// 参数 output 指定日志文件的路径，如果为空则输出到标准输出。
// 返回一个实现了 Logger 接口的实例和可能的错误。
func NewStdLogger(output string) (Logger, error) {
	var writer *os.File = os.Stdout

	// 如果指定了输出目录，配置文件输出。
	if output != "" {
		// 确保日志文件所在的目录存在。
		// 使用 0755 权限确保目录可读可执行，且所有者可写。
		if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
			return nil, err
		}

		// 打开或创建日志文件。
		// 使用 0666 权限确保文件可读可写。
		file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		writer = file
	}

	return &StdLogger{
		// 创建标准库日志实例，启用时间戳。
		logger: log.New(writer, "", log.LstdFlags),
		// 初始化结构化字段映射。
		fields: make(map[string]interface{}),
		// 默认使用 InfoLevel。
		level: InfoLevel,
	}, nil
}

// SetLevel 实现 Logger 接口的日志级别设置方法。
func (l *StdLogger) SetLevel(level Level) {
	l.level = level
}

// GetLevel 实现 Logger 接口的日志级别获取方法。
func (l *StdLogger) GetLevel() Level {
	return l.level
}

// shouldLog 检查给定的日志级别是否应该被记录。
func (l *StdLogger) shouldLog(level Level) bool {
	return level >= l.level
}

// formatFields 格式化结构化字段为字符串。
// 返回格式化后的字段字符串，如果没有字段则返回空字符串。
func (l *StdLogger) formatFields() string {
	if len(l.fields) == 0 {
		return ""
	}
	fields := "["
	for k, v := range l.fields {
		fields += fmt.Sprintf("%s=%v ", k, v)
	}
	return fields[:len(fields)-1] + "]"
}

// log 记录指定级别的日志。
// 参数 level 是日志级别标识，args 是要记录的内容。
func (l *StdLogger) log(logLevel Level, levelStr string, args ...interface{}) {
	if !l.shouldLog(logLevel) {
		return
	}
	fields := l.formatFields()
	if fields != "" {
		l.logger.Printf("%s %s %v", levelStr, fields, fmt.Sprint(args...))
	} else {
		l.logger.Printf("%s %v", levelStr, fmt.Sprint(args...))
	}
}

// logf 记录指定级别的格式化日志。
// 参数 level 是日志级别标识，format 是格式化字符串，args 是格式化参数。
func (l *StdLogger) logf(logLevel Level, levelStr string, format string, args ...interface{}) {
	if !l.shouldLog(logLevel) {
		return
	}
	fields := l.formatFields()
	if fields != "" {
		l.logger.Printf("%s %s "+format, append([]interface{}{levelStr, fields}, args...)...)
	} else {
		l.logger.Printf("%s "+format, append([]interface{}{levelStr}, args...)...)
	}
}

// Debug 实现 Logger 接口的调试级别日志记录。
func (l *StdLogger) Debug(args ...interface{}) {
	l.log(DebugLevel, "[DEBUG]", args...)
}

// Debugf 实现 Logger 接口的格式化调试级别日志记录。
func (l *StdLogger) Debugf(format string, args ...interface{}) {
	l.logf(DebugLevel, "[DEBUG]", format, args...)
}

// Info 实现 Logger 接口的信息级别日志记录。
func (l *StdLogger) Info(args ...interface{}) {
	l.log(InfoLevel, "[INFO]", args...)
}

// Infof 实现 Logger 接口的格式化信息级别日志记录。
func (l *StdLogger) Infof(format string, args ...interface{}) {
	l.logf(InfoLevel, "[INFO]", format, args...)
}

// Warn 实现 Logger 接口的警告级别日志记录。
func (l *StdLogger) Warn(args ...interface{}) {
	l.log(WarnLevel, "[WARN]", args...)
}

// Warnf 实现 Logger 接口的格式化警告级别日志记录。
func (l *StdLogger) Warnf(format string, args ...interface{}) {
	l.logf(WarnLevel, "[WARN]", format, args...)
}

// Error 实现 Logger 接口的错误级别日志记录。
func (l *StdLogger) Error(args ...interface{}) {
	l.log(ErrorLevel, "[ERROR]", args...)
}

// Errorf 实现 Logger 接口的格式化错误级别日志记录。
func (l *StdLogger) Errorf(format string, args ...interface{}) {
	l.logf(ErrorLevel, "[ERROR]", format, args...)
}

// Fatal 实现 Logger 接口的致命错误级别日志记录。
// 记录日志后会导致程序以状态码 1 退出。
func (l *StdLogger) Fatal(args ...interface{}) {
	l.log(FatalLevel, "[FATAL]", args...)
	os.Exit(1)
}

// Fatalf 实现 Logger 接口的格式化致命错误级别日志记录。
// 记录日志后会导致程序以状态码 1 退出。
func (l *StdLogger) Fatalf(format string, args ...interface{}) {
	l.logf(FatalLevel, "[FATAL]", format, args...)
	os.Exit(1)
}

// WithField 实现 Logger 接口的单字段添加方法。
// 创建一个新的日志实例，包含当前实例的所有字段和新添加的字段。
// 返回包含所有字段的新 Logger 实例。
func (l *StdLogger) WithField(key string, value interface{}) Logger {
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value
	return &StdLogger{
		logger: l.logger,
		fields: newFields,
		level:  l.level,
	}
}

// WithFields 实现 Logger 接口的多字段添加方法。
// 创建一个新的日志实例，包含当前实例的所有字段和新添加的字段。
// 返回包含所有字段的新 Logger 实例。
func (l *StdLogger) WithFields(fields map[string]interface{}) Logger {
	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}
	return &StdLogger{
		logger: l.logger,
		fields: newFields,
		level:  l.level,
	}
}
