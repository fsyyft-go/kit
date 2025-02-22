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

// Package log 提供了一个统一的日志接口和多种日志实现。
//
// 这个包的主要特性包括：
//   - 支持多种日志后端（标准输出、Logrus）。
//   - 提供统一的日志接口。
//   - 支持结构化日志记录。
//   - 支持多个日志级别。
//   - 支持文件和标准输出。
//
// 基本使用示例：
//
//	if err := log.InitLogger(log.LogTypeStd, ""); err != nil {
//	    panic(err)
//	}
//	log.Info("应用启动")
//	log.WithField("user", "admin").Info("用户登录")
//
// 更多示例请参考 example/log 目录。
package log

import (
	"fmt"
)

// Level 定义了日志的级别类型，用于控制日志的输出粒度。
type Level int

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

// Logger 定义了统一的日志接口。
// 这个接口提供了基本的日志记录功能和结构化日志支持，可以通过不同的实现来支持不同的日志后端。
type Logger interface {
	// SetLevel 设置日志级别。
	// 只有大于或等于设置级别的日志才会被记录。
	SetLevel(level Level)

	// GetLevel 获取当前的日志级别。
	GetLevel() Level

	// Debug 记录调试级别的日志。
	// 参数 args 支持任意类型的值，这些值会被转换为字符串并连接。
	// 调试日志应该包含有助于诊断问题的详细信息。
	Debug(args ...interface{})

	// Debugf 记录格式化的调试级别日志。
	// 参数 format 是格式化字符串，args 是对应的参数。
	// 支持标准的 Printf 风格的格式化。
	Debugf(format string, args ...interface{})

	// Info 记录信息级别的日志。
	// 参数 args 支持任意类型的值，这些值会被转换为字符串并连接。
	// 信息日志应该记录系统的正常运行状态。
	Info(args ...interface{})

	// Infof 记录格式化的信息级别日志。
	// 参数 format 是格式化字符串，args 是对应的参数。
	// 支持标准的 Printf 风格的格式化。
	Infof(format string, args ...interface{})

	// Warn 记录警告级别的日志。
	// 参数 args 支持任意类型的值，这些值会被转换为字符串并连接。
	// 警告日志应该包含可能导致问题的情况。
	Warn(args ...interface{})

	// Warnf 记录格式化的警告级别日志。
	// 参数 format 是格式化字符串，args 是对应的参数。
	// 支持标准的 Printf 风格的格式化。
	Warnf(format string, args ...interface{})

	// Error 记录错误级别的日志。
	// 参数 args 支持任意类型的值，这些值会被转换为字符串并连接。
	// 错误日志应该包含错误的详细信息和上下文。
	Error(args ...interface{})

	// Errorf 记录格式化的错误级别日志。
	// 参数 format 是格式化字符串，args 是对应的参数。
	// 支持标准的 Printf 风格的格式化。
	Errorf(format string, args ...interface{})

	// Fatal 记录致命错误级别的日志。
	// 参数 args 支持任意类型的值，这些值会被转换为字符串并连接。
	// 记录日志后会导致程序以状态码 1 退出。
	// 这个方法应该只在程序无法继续运行时使用。
	Fatal(args ...interface{})

	// Fatalf 记录格式化的致命错误级别日志。
	// 参数 format 是格式化字符串，args 是对应的参数。
	// 支持标准的 Printf 风格的格式化。
	// 记录日志后会导致程序以状态码 1 退出。
	Fatalf(format string, args ...interface{})

	// WithField 添加一个字段到日志上下文。
	// 参数 key 是字段名，value 是字段值。
	// 返回一个新的 Logger 实例，原实例不会被修改。
	// 这个方法用于添加结构化的上下文信息到日志中。
	WithField(key string, value interface{}) Logger

	// WithFields 添加多个字段到日志上下文。
	// 参数 fields 是要添加的字段映射。
	// 返回一个新的 Logger 实例，原实例不会被修改。
	// 这个方法用于一次性添加多个结构化字段。
	WithFields(fields map[string]interface{}) Logger
}
