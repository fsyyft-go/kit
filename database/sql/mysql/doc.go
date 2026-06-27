// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package mysql 提供基于 go-sql-driver/mysql 的数据库连接构造器。
//
// NewMySQL 会校验 DSN，为给定 namespace 注册带 Hook 的包装驱动，创建
// *sql.DB 并配置连接池参数，然后返回数据库句柄和关闭用的 cleanup 函数。
// 同一 namespace 会复用已注册的驱动名称，便于在不同调用点共享同一组
// driver 包 Hook 规则。
//
// 当启用 WithLogError 或 WithSlowThreshold 时，本包会按需安装错误日志或
// 慢查询日志 Hook；若未显式提供 logger，则会在需要时创建默认 logger。
// NewMySQL 仅调用 sql.Open，不会主动 Ping 数据库，调用方需要在需要时
// 自行校验连通性。
package mysql
