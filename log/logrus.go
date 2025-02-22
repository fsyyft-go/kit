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
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
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

// LogrusLoggerOptions 包含了 LogrusLogger 的所有配置选项。
type LogrusLoggerOptions struct {
	// 输出文件路径。
	OutputPath string
	// 日志格式化器。
	Formatter logrus.Formatter
	// 日志级别。
	Level logrus.Level
	// 文件权限。
	FileMode os.FileMode
	// 目录权限。
	DirMode os.FileMode
	// 是否启用日志滚动。
	EnableRotate bool
	// 日志滚动时间间隔。
	RotateTime time.Duration
	// 日志保留时间。
	MaxAge time.Duration
}

// LogrusOption 定义了 LogrusLogger 的配置选项函数类型。
type LogrusOption func(*LogrusLoggerOptions)

// 默认选项。
var defaultOptions = LogrusLoggerOptions{
	Formatter: &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	},
	Level:        logrus.InfoLevel,
	FileMode:     0666,
	DirMode:      0755,
	EnableRotate: true,               // 默认启用日志滚动
	RotateTime:   time.Hour,          // 默认每小时滚动一次
	MaxAge:       time.Hour * 24 * 7, // 默认保留7天
}

// WithOutputPath 设置日志输出路径。
func WithOutputPath(path string) LogrusOption {
	return func(o *LogrusLoggerOptions) {
		o.OutputPath = path
	}
}

// WithFormatter 设置日志格式化器。
func WithFormatter(formatter logrus.Formatter) LogrusOption {
	return func(o *LogrusLoggerOptions) {
		o.Formatter = formatter
	}
}

// WithLogrusLevel 设置日志级别。
func WithLogrusLevel(level Level) LogrusOption {
	return func(o *LogrusLoggerOptions) {
		if logrusLevel, ok := logrusLevelMap[level]; ok {
			o.Level = logrusLevel
		}
	}
}

// WithFileMode 设置日志文件权限。
func WithFileMode(mode os.FileMode) LogrusOption {
	return func(o *LogrusLoggerOptions) {
		o.FileMode = mode
	}
}

// WithDirMode 设置日志目录权限。
func WithDirMode(mode os.FileMode) LogrusOption {
	return func(o *LogrusLoggerOptions) {
		o.DirMode = mode
	}
}

// WithLogrusEnableRotate 设置是否启用日志滚动。
func WithLogrusEnableRotate(enable bool) LogrusOption {
	return func(o *LogrusLoggerOptions) {
		o.EnableRotate = enable
	}
}

// WithLogrusRotateTime 设置日志滚动时间间隔。
func WithLogrusRotateTime(duration time.Duration) LogrusOption {
	return func(o *LogrusLoggerOptions) {
		o.RotateTime = duration
	}
}

// WithLogrusMaxAge 设置日志保留时间。
func WithLogrusMaxAge(duration time.Duration) LogrusOption {
	return func(o *LogrusLoggerOptions) {
		o.MaxAge = duration
	}
}

// NewLogrusLogger 创建一个新的 LogrusLogger 实例。
// 使用可选的 LogrusOption 函数来配置 logger。
func NewLogrusLogger(opts ...LogrusOption) (Logger, error) {
	// 使用默认选项。
	options := defaultOptions

	// 应用自定义选项。
	for _, opt := range opts {
		opt(&options)
	}

	log := logrus.New()

	// 如果指定了输出目录，配置文件输出。
	if options.OutputPath != "" {
		// 确保日志文件所在的目录存在。
		if err := os.MkdirAll(filepath.Dir(options.OutputPath), options.DirMode); err != nil {
			return nil, err
		}

		if options.EnableRotate {
			// 获取文件名和扩展名
			ext := filepath.Ext(options.OutputPath)
			base := options.OutputPath[:len(options.OutputPath)-len(ext)]

			// 配置日志滚动
			writer, err := rotatelogs.New(
				base+"-%Y%m%d%H"+ext,
				rotatelogs.WithLinkName(options.OutputPath),
				rotatelogs.WithRotationTime(options.RotateTime),
				rotatelogs.WithMaxAge(options.MaxAge),
			)
			if err != nil {
				return nil, err
			}
			log.SetOutput(writer)
		} else {
			// 打开或创建日志文件。
			file, err := os.OpenFile(options.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, options.FileMode)
			if err != nil {
				return nil, err
			}
			log.SetOutput(file)
		}
	}

	// 配置日志格式。
	log.SetFormatter(options.Formatter)

	// 设置日志级别。
	log.SetLevel(options.Level)

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
