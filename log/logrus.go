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

// Package log 提供了基于 Logrus 的日志实现。
package log

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// LogrusLogger 实现了 Logger 接口，使用 Logrus 作为底层日志库。
// 这个实现提供了丰富的日志功能，包括：
// - 结构化日志记录。
// - 多种输出格式（文本、JSON）。
// - 灵活的日志级别控制。
// - 支持同时输出到多个目标。
type LogrusLogger struct {
	// logger 是 Logrus 的日志实例，包含了所有的上下文信息。
	logger *logrus.Entry
}

// logrusLevelMap 定义了自定义日志级别到 Logrus 日志级别的映射。
var logrusLevelMap = map[Level]logrus.Level{
	DebugLevel: logrus.DebugLevel,
	InfoLevel:  logrus.InfoLevel,
	WarnLevel:  logrus.WarnLevel,
	ErrorLevel: logrus.ErrorLevel,
	FatalLevel: logrus.FatalLevel,
}

// NewLogrusLogger 创建一个新的 LogrusLogger 实例。
// 参数 output 指定日志文件的路径，如果为空则只输出到标准输出。
// 返回一个实现了 Logger 接口的实例和可能的错误。
func NewLogrusLogger(output string) (Logger, error) {
	log := logrus.New()

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

		// 只输出到文件。
		log.SetOutput(file)
	}

	// 配置日志格式为文本格式。
	// 启用完整时间戳，使用标准的日期时间格式。
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 默认设置为 InfoLevel。
	log.SetLevel(logrus.InfoLevel)

	return &LogrusLogger{
		logger: logrus.NewEntry(log),
	}, nil
}

// SetLevel 实现 Logger 接口的日志级别设置方法。
func (l *LogrusLogger) SetLevel(level Level) {
	if logrusLevel, ok := logrusLevelMap[level]; ok {
		l.logger.Logger.SetLevel(logrusLevel)
	}
}

// GetLevel 实现 Logger 接口的日志级别获取方法。
func (l *LogrusLogger) GetLevel() Level {
	logrusLevel := l.logger.Logger.GetLevel()
	for level, lLevel := range logrusLevelMap {
		if lLevel == logrusLevel {
			return level
		}
	}
	return InfoLevel
}

// Debug 实现 Logger 接口的调试级别日志记录。
func (l *LogrusLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

// Debugf 实现 Logger 接口的格式化调试级别日志记录。
func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

// Info 实现 Logger 接口的信息级别日志记录。
func (l *LogrusLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

// Infof 实现 Logger 接口的格式化信息级别日志记录。
func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

// Warn 实现 Logger 接口的警告级别日志记录。
func (l *LogrusLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

// Warnf 实现 Logger 接口的格式化警告级别日志记录。
func (l *LogrusLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

// Error 实现 Logger 接口的错误级别日志记录。
func (l *LogrusLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

// Errorf 实现 Logger 接口的格式化错误级别日志记录。
func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

// Fatal 实现 Logger 接口的致命错误级别日志记录。
// 记录日志后会导致程序以状态码 1 退出。
func (l *LogrusLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

// Fatalf 实现 Logger 接口的格式化致命错误级别日志记录。
// 记录日志后会导致程序以状态码 1 退出。
func (l *LogrusLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

// WithField 实现 Logger 接口的单字段添加方法。
// 返回一个包含新字段的新 Logger 实例。
func (l *LogrusLogger) WithField(key string, value interface{}) Logger {
	return &LogrusLogger{
		logger: l.logger.WithField(key, value),
	}
}

// WithFields 实现 Logger 接口的多字段添加方法。
// 返回一个包含新字段的新 Logger 实例。
func (l *LogrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &LogrusLogger{
		logger: l.logger.WithFields(fields),
	}
}
