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
	// gormLogger 实现 gormlogger.Interface，并将日志异步转发到 kit logger。
	//
	// 该适配器会把 GORM 的日志级别映射为 kit logger 的输出语义，并在当前
	// 方法返回后异步调用底层 logger，因此不保证日志在返回前已经写出。
	gormLogger struct {
		// logger 是 kit 的日志实例。
		logger kitlog.Logger
		// level 存储当前的 gorm 日志级别。
		level gormlogger.LogLevel
	}
)

// NewLogger 创建一个实现 gormlogger.Interface 的 kit logger 适配器。
//
// 返回的适配器会根据底层 logger 当前的 GetLevel 计算初始 GORM 日志级别，
// 并在 Info、Warn、Error 和 Trace 中异步调用底层 logger。
//
// 参数：
//   - logger：非 nil 的 kit 日志实例，用于实际输出日志。
//
// 返回值：
//   - gormlogger.Interface：基于 logger 的 GORM 日志适配器。
func NewLogger(logger kitlog.Logger) gormlogger.Interface {
	return &gormLogger{
		logger: logger,
		level:  getGormLevel(logger.GetLevel()),
	}
}

// LogMode 返回一份使用指定 GORM 日志级别的新适配器。
//
// LogMode 不会修改当前实例，而是复制一份适配器并复用原有底层 logger。
//
// 参数：
//   - level：要设置的新 GORM 日志级别。
//
// 返回值：
//   - gormlogger.Interface：级别调整后的新适配器。
func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.level = level
	return &newLogger
}

// Info 在当前级别允许时异步输出一条 GORM Info 日志。
//
// 参数：
//   - ctx：当前未参与日志内容生成的上下文。
//   - msg：按 fmt.Sprintf 规则格式化的日志模板。
//   - data：用于格式化 msg 的可选参数。
func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Info {
		go l.logger.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn 在当前级别允许时异步输出一条 GORM Warn 日志。
//
// 参数：
//   - ctx：当前未参与日志内容生成的上下文。
//   - msg：按 fmt.Sprintf 规则格式化的日志模板。
//   - data：用于格式化 msg 的可选参数。
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Warn {
		go l.logger.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error 在当前级别允许时异步输出一条 GORM Error 日志。
//
// 参数：
//   - ctx：当前未参与日志内容生成的上下文。
//   - msg：按 fmt.Sprintf 规则格式化的日志模板。
//   - data：用于格式化 msg 的可选参数。
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Error {
		go l.logger.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace 根据执行结果和耗时异步输出一条 SQL 跟踪日志。
//
// 当级别为 gormlogger.Silent 时，Trace 会直接返回且不会调用 fc。其他级别下，
// Trace 会先调用 fc 获取 SQL 和 rowsAffected，然后按以下优先级记录：
// 1. err 非 nil 且级别允许时输出 Error 日志，并附带 "err" 字段。
// 2. err 为 nil、耗时大于 1 秒且级别允许时输出 Warn 日志，并附带 "elapsed" 字段。
// 3. 其余级别允许的成功语句输出 Info 日志。
//
// 参数：
//   - ctx：当前未参与日志内容生成的上下文。
//   - begin：SQL 执行的开始时间。
//   - fc：返回 SQL 文本和 rows affected 的回调；除 Silent 级别外都会被调用。
//   - err：SQL 执行过程中返回的错误。
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
		go l.logger.Error(logMsg, "err", err)
	case elapsed > time.Second && l.level >= gormlogger.Warn:
		go l.logger.Warn(logMsg, "elapsed", elapsed)
	case l.level >= gormlogger.Info:
		go l.logger.Info(logMsg)
	}
}
