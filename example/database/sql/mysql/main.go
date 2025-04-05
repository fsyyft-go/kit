// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package main 演示了如何使用 fsyyft-go/kit 包中的 MySQL 数据库连接功能。
package main

import (
	"fmt"
	"time"

	kitmysql "github.com/fsyyft-go/kit/database/sql/mysql"
	kitlog "github.com/fsyyft-go/kit/log"
)

// main 函数展示了 MySQL 数据库连接的基本使用方法，包括连接初始化、查询执行和资源清理。
func main() {
	// 创建一个新的日志记录器实例。
	logger, err := kitlog.NewLogger()
	if err != nil {
		fmt.Println(err)
	}

	// 使用配置选项初始化 MySQL 数据库连接。
	// WithNamespace：设置命名空间为 "test"。
	// WithLogger：设置日志记录器。
	// WithLogError：启用错误日志记录。
	// WithSlowThreshold：设置慢查询阈值为 1 微秒。
	db, cleanup, err := kitmysql.NewMySQL(
		kitmysql.WithNamespace("test"),
		kitmysql.WithLogger(logger),
		kitmysql.WithLogError(true),
		kitmysql.WithSlowThreshold(time.Microsecond),
	)
	// 延迟执行资源清理函数。
	defer cleanup()

	// 检查数据库连接初始化是否成功。
	if err != nil {
		logger.Error(err)
		return
	}

	// 测试数据库连接是否正常。
	if err := db.Ping(); err != nil {
		logger.Error(err)
		return
	}

	// 执行 SHOW PROCESSLIST 命令查询当前数据库进程列表。
	rows, err := db.Query("SHOW PROCESSLIST")
	if err != nil {
		logger.Error(err)
		return
	}

	// 遍历查询结果集。
	for rows.Next() {
		// 定义变量用于存储进程信息。
		var id int
		var user string
		var host string
		var db string
		var command string
		var time int
		var state string
		var info string
		// 扫描当前行数据到变量中。
		if err := rows.Scan(&id, &user, &host, &db, &command, &time, &state, &info); err != nil {
			logger.Error(err)
			return
		}
		// 打印进程信息。
		fmt.Println(id, user, host, db, command, time, state)
	}

	// 关闭结果集。
	rows.Close() //nolint:errcheck

	// 关闭数据库连接。
	db.Close() //nolint:errcheck
}
