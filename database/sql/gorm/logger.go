// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package gorm

import (
	"context"
	"fmt"
	"time"

	gormlogger "gorm.io/gorm/logger"

	kitlog "github.com/fsyyft-go/kit/log"
)

// getGormLevel 将 kit 日志级别转换为 gorm 日志级别。
//
// 参数：
//   - level：kit 日志级别。
//
// 返回值：
//   - gormlogger.LogLevel：返回对应的 gorm 日志级别。
func getGormLevel(level kitlog.Level) gormlogger.LogLevel {
	switch level {
	case kitlog.DebugLevel:
		return gormlogger.Info
	case kitlog.InfoLevel:
		return gormlogger.Info
	case kitlog.WarnLevel:
		return gormlogger.Warn
	case kitlog.ErrorLevel:
		return gormlogger.Error
	case kitlog.FatalLevel:
		return gormlogger.Silent
	default:
		return gormlogger.Info
	}
}

type (
	// gormLogger 实现了 gorm logger.Interface 接口，使用 kit logger 作为底层实现。
	// 这个适配器提供了以下功能：
	// - 将 kit logger 的日志级别映射到 gorm 的日志级别。
	// - 支持 SQL 执行的跟踪和性能监控。
	// - 支持错误日志记录和慢查询警告。
	// - 保持与 kit logger 统一的日志格式和管理。
	gormLogger struct {
		// logger 是 kit 的日志实例。
		logger kitlog.Logger
		// level 存储当前的 gorm 日志级别。
		level gormlogger.LogLevel
	}
)

// NewLogger 创建一个新的 gorm logger 适配器。
//
// 参数：
//   - logger：kit 日志实例，用于实际的日志记录。
//
// 返回值：
//   - gormlogger.Interface：返回一个实现了 gorm logger.Interface 的适配器实例。
func NewLogger(logger kitlog.Logger) gormlogger.Interface {
	return &gormLogger{
		logger: logger,
		level:  getGormLevel(logger.GetLevel()),
	}
}

// LogMode 设置日志级别。
//
// 参数：
//   - level：要设置的 gorm 日志级别。
//
// 返回值：
//   - gormlogger.Interface：返回一个新的日志实例，原实例不会被修改。
func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.level = level
	return &newLogger
}

// Info 打印信息级别的日志。
//
// 参数：
//   - ctx：上下文信息，目前未使用。
//   - msg：日志消息。
//   - data：可选的格式化参数。
func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Info {
		l.logger.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn 打印警告级别的日志。
//
// 参数：
//   - ctx：上下文信息，目前未使用。
//   - msg：日志消息。
//   - data：可选的格式化参数。
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Warn {
		l.logger.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error 打印错误级别的日志。
//
// 参数：
//   - ctx：上下文信息，目前未使用。
//   - msg：日志消息。
//   - data：可选的格式化参数。
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Error {
		l.logger.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace 打印 SQL 执行的跟踪日志。
//
// 参数：
//   - ctx：上下文信息，目前未使用。
//   - begin：SQL 执行的开始时间。
//   - fc：返回 SQL 语句和影响行数的函数。
//   - err：SQL 执行过程中可能发生的错误。
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 构建日志信息
	logMsg := fmt.Sprintf("[%.3fms] [rows:%d] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)

	switch {
	case err != nil && l.level >= gormlogger.Error:
		l.logger.Error(logMsg, "err", err)
	case elapsed > time.Second && l.level >= gormlogger.Warn:
		l.logger.Warn(logMsg, "elapsed", elapsed)
	case l.level >= gormlogger.Info:
		l.logger.Info(logMsg)
	}
}
