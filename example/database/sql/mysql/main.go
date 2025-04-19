// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package main 演示了如何使用 fsyyft-go/kit 包中的 MySQL 数据库连接功能。
package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	kitmysql "github.com/fsyyft-go/kit/database/sql/mysql"
	kitlog "github.com/fsyyft-go/kit/log"
)

// User 用户表结构。
//
// 字段说明：
//   - ID：用户唯一标识，主键，自增。
//   - Name：用户名称。
//   - Age：用户年龄。
type User struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

// queryRow 查询单条数据示例。
//
// 参数：
//   - db：数据库连接实例。
//   - logger：日志记录器。
//   - id：要查询的用户 ID。
//
// 返回：
//   - *User：查询到的用户信息。
//   - error：查询过程中的错误信息。
func queryRow(db *sql.DB, _ kitlog.Logger, id int64) (*User, error) {
	// 构造查询 SQL 语句。
	sqlStr := "SELECT `id`, `name`, `age` FROM `example_user` WHERE `id` = ? LIMIT 1;"
	var u User
	// 执行查询并扫描结果到结构体。
	err := db.QueryRow(sqlStr, id).Scan(&u.ID, &u.Name, &u.Age)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// queryMultiRow 查询多条数据示例。
//
// 参数：
//   - db：数据库连接实例。
//   - logger：日志记录器。
//
// 返回：
//   - []*User：查询到的用户信息列表。
//   - error：查询过程中的错误信息。
func queryMultiRow(db *sql.DB, logger kitlog.Logger) ([]*User, error) {
	// 构造查询 SQL 语句。
	sqlStr := "SELECT `id`, `name`, `age` FROM `example_user` WHERE `id` > ? LIMIT 65535;"
	// 执行查询。
	rows, err := db.Query(sqlStr, 0)
	if err != nil {
		return nil, err
	}
	// 确保结果集被正确关闭。
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error(err)
		}
	}()

	var users []*User
	// 遍历结果集。
	for rows.Next() {
		var u User
		// 扫描当前行到结构体。
		err := rows.Scan(&u.ID, &u.Name, &u.Age)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}

// insertRow 插入数据示例。
//
// 参数：
//   - db：数据库连接实例。
//   - logger：日志记录器。
//   - user：要插入的用户信息。
//
// 返回：
//   - int64：插入记录的 ID。
//   - error：插入过程中的错误信息。
func insertRow(db *sql.DB, _ kitlog.Logger, user *User) (int64, error) {
	var sqlStr string
	var args []interface{}

	// 根据是否指定 ID 构造不同的插入语句。
	if user.ID > 0 {
		sqlStr = "INSERT INTO `example_user`(`id`, `name`, `age`) VALUES (?, ?, ?);"
		args = []interface{}{user.ID, user.Name, user.Age}
	} else {
		sqlStr = "INSERT INTO `example_user`(`name`, `age`) VALUES (?, ?);"
		args = []interface{}{user.Name, user.Age}
	}

	// 执行插入操作。
	ret, err := db.Exec(sqlStr, args...)
	if err != nil {
		return 0, err
	}
	// 获取插入记录的 ID。
	theID, err := ret.LastInsertId()
	if err != nil {
		return 0, err
	}
	return theID, nil
}

// updateRow 更新数据示例。
//
// 参数：
//   - db：数据库连接实例。
//   - logger：日志记录器。
//   - user：要更新的用户信息。
//
// 返回：
//   - int64：受影响的行数。
//   - error：更新过程中的错误信息。
func updateRow(db *sql.DB, _ kitlog.Logger, user *User) (int64, error) {
	// 构造更新 SQL 语句。
	sqlStr := "UPDATE `example_user` SET `age`=? WHERE `id` = ? LIMIT 1;"
	// 执行更新操作。
	ret, err := db.Exec(sqlStr, user.Age, user.ID)
	if err != nil {
		return 0, err
	}
	// 获取受影响的行数。
	n, err := ret.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}

// deleteRow 删除数据示例。
//
// 参数：
//   - db：数据库连接实例。
//   - logger：日志记录器。
//   - id：要删除的用户 ID。
//
// 返回：
//   - int64：受影响的行数。
//   - error：删除过程中的错误信息。
func deleteRow(db *sql.DB, _ kitlog.Logger, id int64) (int64, error) {
	// 构造删除 SQL 语句。
	sqlStr := "DELETE FROM `example_user` WHERE `id` = ? LIMIT 1;"
	// 执行删除操作。
	ret, err := db.Exec(sqlStr, id)
	if err != nil {
		return 0, err
	}
	// 获取受影响的行数。
	n, err := ret.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}

// transactionDemo 事务操作示例。
//
// 参数：
//   - db：数据库连接实例。
//   - logger：日志记录器。
//
// 返回：
//   - error：事务执行过程中的错误信息。
func transactionDemo(db *sql.DB, logger kitlog.Logger) error {
	// 创建一个本地随机数生成器。
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 开始事务。
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败：%w", err)
	}

	// 使用 defer 处理事务回滚，确保在返回错误时一定会回滚。
	committed := false
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil {
				logger.Error(fmt.Errorf("回滚事务失败：%w", err))
			}
		}
	}()

	// 首先检查记录是否存在。
	var exists1, exists2 bool
	// 检查 ID 为 2 的记录是否存在。
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM `example_user` WHERE `id` = ?)", 2).Scan(&exists1)
	if err != nil {
		return fmt.Errorf("检查表 `example_user` 中第一条记录是否存在失败：%w", err)
	}
	// 检查 ID 为 3 的记录是否存在。
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM `example_user` WHERE `id` = ?)", 3).Scan(&exists2)
	if err != nil {
		return fmt.Errorf("检查表 `example_user` 中第二条记录是否存在失败：%w", err)
	}

	// 如果任一记录不存在，返回错误。
	if !exists1 || !exists2 {
		return fmt.Errorf("表 `example_user` 中部分记录不存在：`id=2` 存在=%v，`id=3` 存在=%v", exists1, exists2)
	}

	// 生成两个随机年龄。
	age1 := r.Intn(61) + 10 // 生成 10 ~ 70 之间的随机数。
	age2 := r.Intn(61) + 10 // 生成 10 ~ 70 之间的随机数。

	// 更新第一条记录。
	sqlStr := "UPDATE `example_user` SET `age`=? WHERE `id`=? LIMIT 1;"
	ret1, err := tx.Exec(sqlStr, age1, 2)
	if err != nil {
		return fmt.Errorf("执行表 `example_user` 第一次更新失败：%w", err)
	}
	// 获取第一次更新影响的行数。
	affRow1, err := ret1.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取第一次更新影响行数失败：%w", err)
	}
	ret2, err := tx.Exec(sqlStr, age2, 3)
	if err != nil {
		return fmt.Errorf("执行表 `example_user` 第二次更新失败：%w", err)
	}
	// 获取第二次更新影响的行数。
	affRow2, err := ret2.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取第二次更新影响行数失败：%w", err)
	}

	// 检查更新是否都成功，如果成功则提交事务。
	if affRow1 == 1 && affRow2 == 1 {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("提交事务失败：%w", err)
		}
		committed = true
		return nil
	}

	return fmt.Errorf("事务影响行数不匹配：第一次=%d，第二次=%d", affRow1, affRow2)
}

// main 函数展示了 MySQL 数据库连接的基本使用方法，包括连接初始化、查询执行和资源清理。
func main() {
	// 创建一个新的日志记录器实例。
	logger, err := kitlog.NewLogger(
		kitlog.WithLevel(kitlog.DebugLevel),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 设置数据库连接参数。
	dsnparams := kitmysql.WithDSNParams("test:test@tcp(localhost:3306)/test", map[string]string{
		"interpolateParams":    "true",  // 启用参数插值。
		"parseTime":            "true",  // 启用时间解析。
		"loc":                  "Local", // 使用本地时区。
		"allowNativePasswords": "true",  // 允许原生密码认证。
	})

	// 使用配置选项初始化 MySQL 数据库连接。
	db, cleanup, err := kitmysql.NewMySQL(
		dsnparams,
		kitmysql.WithNamespace("test"),               // 设置命名空间为 "test"。
		kitmysql.WithLogger(logger),                  // 设置日志记录器。
		kitmysql.WithLogError(true),                  // 启用错误日志记录。
		kitmysql.WithSlowThreshold(time.Microsecond), // 设置慢查询阈值为 1 微秒。
	)
	// 延迟执行资源清理函数。
	defer cleanup()

	// 检查数据库连接初始化是否成功。
	if err != nil {
		logger.Error(fmt.Errorf("初始化数据库失败：%w", err))
		return
	}

	// 测试数据库连接是否正常。
	if err := db.Ping(); err != nil {
		logger.Error(fmt.Errorf("数据库连接测试失败：%w", err))
		return
	}

	// 创建测试表。
	createTableSQL := "CREATE TABLE IF NOT EXISTS `example_user` (" +
		"`id` BIGINT(20) NOT NULL AUTO_INCREMENT," +
		"`name` VARCHAR(20) DEFAULT ''," +
		"`age` INT(11) DEFAULT '0'," +
		"PRIMARY KEY(`id`)" +
		")ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;"

	// 执行建表语句。
	if _, err := db.Exec(createTableSQL); err != nil {
		logger.Error(fmt.Errorf("创建表 `example_user` 失败：%w", err))
		return
	}
	logger.Info("创建表成功")

	// 插入数据示例。
	user := &User{
		Name: "张三",
		Age:  25,
	}
	// 执行插入操作。
	lastID, err := insertRow(db, logger, user)
	if err != nil {
		logger.Error(fmt.Errorf("向表 `example_user` 插入数据失败：%w", err))
		return
	}
	logger.Info(fmt.Sprintf("插入数据成功，ID：%d", lastID))

	// 查询单行数据示例。
	u, err := queryRow(db, logger, lastID)
	if err != nil {
		logger.Error(fmt.Errorf("从表 `example_user` 查询单行数据失败：%w", err))
		return
	}
	logger.Info(fmt.Sprintf("查询单行数据成功：%+v", u))

	// 更新数据示例。
	u.Age = 30
	// 执行更新操作。
	affected, err := updateRow(db, logger, u)
	if err != nil {
		logger.Error(fmt.Errorf("更新表 `example_user` 中 `id=%d` 的数据失败：%w", u.ID, err))
		return
	}
	logger.Info(fmt.Sprintf("更新数据成功，影响行数：%d", affected))

	// 查询多行数据示例。
	users, err := queryMultiRow(db, logger)
	if err != nil {
		logger.Error(fmt.Errorf("从表 `example_user` 查询多行数据失败：%w", err))
		return
	}
	logger.Info(fmt.Sprintf("查询多行数据成功，数据条数：%d", len(users)))

	// 准备事务测试数据。
	testUsers := []User{
		{ID: 2, Name: "李四", Age: 20},
		{ID: 3, Name: "王五", Age: 25},
	}

	// 遍历测试用户数据。
	for _, u := range testUsers {
		var exists bool
		// 检查用户是否已存在。
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM `example_user` WHERE `id` = ?)", u.ID).Scan(&exists)
		if err != nil {
			logger.Error(fmt.Errorf("检查表 `example_user` 中 `id=%d` 的记录是否存在失败：%w", u.ID, err))
			return
		}

		if !exists {
			// 使用通用的 insertRow 函数插入数据。
			_, err = insertRow(db, logger, &u)
			if err != nil {
				logger.Error(fmt.Errorf("向表 `example_user` 插入测试用户失败：%w", err))
				return
			}
			logger.Info(fmt.Sprintf("插入测试数据成功，ID：%d", u.ID))
		} else {
			logger.Info(fmt.Sprintf("测试数据已存在，ID：%d", u.ID))
		}
	}

	// 执行事务操作示例。
	if err := transactionDemo(db, logger); err != nil {
		logger.Error(fmt.Errorf("执行表 `example_user` 的事务操作示例失败：%w", err))
		return
	}
	logger.Info("事务操作成功")

	// 删除数据示例。
	affected, err = deleteRow(db, logger, lastID)
	if err != nil {
		logger.Error(fmt.Errorf("从表 `example_user` 删除 `id=%d` 的数据失败：%w", lastID, err))
		return
	}
	logger.Info(fmt.Sprintf("删除数据成功，影响行数：%d", affected))

	// 关闭数据库连接。
	if err := db.Close(); err != nil {
		logger.Error(fmt.Errorf("关闭数据库连接失败：%w", err))
		return
	}
}
