// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package main 演示了如何使用 fsyyft-go/kit 包中的 MySQL 数据库连接功能。
package main

import (
	"database/sql"
	"fmt"
	"time"

	kitmysql "github.com/fsyyft-go/kit/database/sql/mysql"
	kitlog "github.com/fsyyft-go/kit/log"
)

// User 用户表结构。
type User struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

// queryRow 查询单条数据示例。
func queryRow(db *sql.DB, _ kitlog.Logger, id int64) (*User, error) {
	sqlStr := "SELECT `id`, `name`, `age` FROM example_user WHERE `id` = ?"
	var u User
	err := db.QueryRow(sqlStr, id).Scan(&u.ID, &u.Name, &u.Age)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// queryMultiRow 查询多条数据示例。
func queryMultiRow(db *sql.DB, logger kitlog.Logger) ([]*User, error) {
	sqlStr := "SELECT `id`, `name`, `age` FROM example_user WHERE `id` > ?"
	rows, err := db.Query(sqlStr, 0)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error(err)
		}
	}()

	var users []*User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Name, &u.Age)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

// insertRow 插入数据示例。
func insertRow(db *sql.DB, _ kitlog.Logger, user *User) (int64, error) {
	sqlStr := "INSERT INTO example_user(`name`, `age`) VALUES (?, ?)"
	ret, err := db.Exec(sqlStr, user.Name, user.Age)
	if err != nil {
		return 0, err
	}
	theID, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return theID, nil
}

// updateRow 更新数据示例。
func updateRow(db *sql.DB, _ kitlog.Logger, user *User) (int64, error) {
	sqlStr := "UPDATE example_user SET `age`=? WHERE `id` = ?"
	ret, err := db.Exec(sqlStr, user.Age, user.ID)
	if err != nil {
		return 0, err
	}
	n, err := ret.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}

// deleteRow 删除数据示例。
func deleteRow(db *sql.DB, _ kitlog.Logger, id int64) (int64, error) {
	sqlStr := "DELETE FROM example_user WHERE `id` = ?"
	ret, err := db.Exec(sqlStr, id)
	if err != nil {
		return 0, err
	}
	n, err := ret.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}

// transactionDemo 事务操作示例。
func transactionDemo(db *sql.DB, logger kitlog.Logger) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}

	// 使用 defer 处理事务回滚，确保在返回错误时一定会回滚
	committed := false
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil {
				logger.Error(fmt.Errorf("rollback transaction failed: %w", err))
			}
		}
	}()

	sqlStr1 := "UPDATE example_user SET `age`=30 WHERE `id`=?"
	ret1, err := tx.Exec(sqlStr1, 2)
	if err != nil {
		return fmt.Errorf("execute first update failed: %w", err)
	}
	affRow1, err := ret1.RowsAffected()
	if err != nil {
		return fmt.Errorf("get first affected rows failed: %w", err)
	}

	sqlStr2 := "UPDATE example_user SET `age`=40 WHERE `id`=?"
	ret2, err := tx.Exec(sqlStr2, 3)
	if err != nil {
		return fmt.Errorf("execute second update failed: %w", err)
	}
	affRow2, err := ret2.RowsAffected()
	if err != nil {
		return fmt.Errorf("get second affected rows failed: %w", err)
	}

	if affRow1 == 1 && affRow2 == 1 {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit transaction failed: %w", err)
		}
		committed = true
		return nil
	}

	return fmt.Errorf("transaction affected rows not match: first=%d, second=%d", affRow1, affRow2)
}

// main 函数展示了 MySQL 数据库连接的基本使用方法，包括连接初始化、查询执行和资源清理。
func main() {
	// 创建一个新的日志记录器实例。
	logger, err := kitlog.NewLogger()
	if err != nil {
		fmt.Println(err)
		return
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
		logger.Error(fmt.Errorf("initialize database failed: %w", err))
		return
	}

	// 测试数据库连接是否正常。
	if err := db.Ping(); err != nil {
		logger.Error(fmt.Errorf("ping database failed: %w", err))
		return
	}

	// 创建测试表。
	createTableSQL := `CREATE TABLE IF NOT EXISTS example_user (
		id BIGINT(20) NOT NULL AUTO_INCREMENT,
		name VARCHAR(20) DEFAULT '',
		age INT(11) DEFAULT '0',
		PRIMARY KEY(id)
	)ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;`

	if _, err := db.Exec(createTableSQL); err != nil {
		logger.Error(fmt.Errorf("create table failed: %w", err))
		return
	}
	logger.Info("创建表成功")

	// 插入数据示例。
	user := &User{
		Name: "张三",
		Age:  25,
	}
	lastID, err := insertRow(db, logger, user)
	if err != nil {
		logger.Error(fmt.Errorf("insert data failed: %w", err))
		return
	}
	logger.Info(fmt.Sprintf("插入数据成功，ID：%d", lastID))

	// 查询单行数据示例。
	u, err := queryRow(db, logger, lastID)
	if err != nil {
		logger.Error(fmt.Errorf("query single row failed: %w", err))
		return
	}
	logger.Info(fmt.Sprintf("查询单行数据成功：%+v", u))

	// 更新数据示例。
	u.Age = 30
	affected, err := updateRow(db, logger, u)
	if err != nil {
		logger.Error(fmt.Errorf("update data failed: %w", err))
		return
	}
	logger.Info(fmt.Sprintf("更新数据成功，影响行数：%d", affected))

	// 查询多行数据示例。
	users, err := queryMultiRow(db, logger)
	if err != nil {
		logger.Error(fmt.Errorf("query multiple rows failed: %w", err))
		return
	}
	logger.Info(fmt.Sprintf("查询多行数据成功，数据条数：%d", len(users)))

	// 事务操作示例。
	if err := transactionDemo(db, logger); err != nil {
		logger.Error(fmt.Errorf("transaction demo failed: %w", err))
		return
	}
	logger.Info("事务操作成功")

	// 删除数据示例。
	affected, err = deleteRow(db, logger, lastID)
	if err != nil {
		logger.Error(fmt.Errorf("delete data failed: %w", err))
		return
	}
	logger.Info(fmt.Sprintf("删除数据成功，影响行数：%d", affected))

	// 关闭数据库连接。
	if err := db.Close(); err != nil {
		logger.Error(fmt.Errorf("close database failed: %w", err))
		return
	}
}
