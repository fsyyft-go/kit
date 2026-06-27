// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package gorm 提供 kit/log 与 GORM logger.Interface 之间的日志适配器。
//
// NewLogger 根据底层 kit logger 的当前级别初始化 GORM 日志级别，并通过
// gorm logger.Interface 的 Info、Warn、Error 和 Trace 输出 SQL、影响行数、
// 执行错误与慢查询信息。
//
// 适配器会异步调用底层 logger，因此不保证日志在当前方法返回前已经完成写出。
// 本包只负责日志桥接，不负责创建 gorm.DB、配置迁移或管理数据库连接。
package gorm
