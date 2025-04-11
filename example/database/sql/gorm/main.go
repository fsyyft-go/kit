// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package main 演示了如何使用 gorm 进行基本的数据库操作，包括连接配置、表结构定义和基本的 CRUD 操作。
package main

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	kitgorm "github.com/fsyyft-go/kit/database/sql/gorm"
	kitmysql "github.com/fsyyft-go/kit/database/sql/mysql"
	kitlog "github.com/fsyyft-go/kit/log"
)

// User 定义用户表的结构体。
//
// 字段说明：
//   - ID：用户唯一标识，主键，自增。
//   - Name：用户名称，默认为空字符串。
//   - Age：用户年龄，默认为 0。
type User struct {
	// ID 用户唯一标识，使用 BIGINT 类型，自增主键。
	ID int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	// Name 用户名称，使用 VARCHAR(20) 类型，默认为空字符串。
	Name string `gorm:"column:name;type:varchar(20);default:''" json:"name"`
	// Age 用户年龄，使用 INT 类型，默认为 0。
	Age int `gorm:"column:age;type:int;default:0" json:"age"`
}

// TableName 指定用户表的表名。
//
// 返回：
//   - string：表名 example_user。
func (User) TableName() string {
	return "example_user"
}

// main 函数演示了使用 gorm 进行数据库操作的完整流程。
//
// 主要步骤：
//  1. 初始化日志记录器。
//  2. 配置并建立数据库连接。
//  3. 配置 gorm。
//  4. 执行基本的 CRUD 操作。
func main() {
	// 创建一个新的日志记录器实例，设置日志级别为 Debug。
	logger, err := kitlog.NewLogger(
		kitlog.WithLevel(kitlog.DebugLevel),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 为日志记录器添加模块标识。
	logger = logger.WithField("module", "gorm")

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
		kitmysql.WithNamespace("example.gorm"),       // 设置命名空间。
		kitmysql.WithLogger(logger),                  // 设置日志记录器。
		kitmysql.WithLogError(true),                  // 启用错误日志记录。
		kitmysql.WithSlowThreshold(time.Microsecond), // 设置慢查询阈值。
	)

	// 延迟执行资源清理函数。
	defer cleanup()

	// 检查数据库连接是否成功。
	if nil != err {
		logger.Error(fmt.Errorf("初始化数据库失败：%w", err))
		return
	}

	// 创建 MySQL 配置。
	mysqlconfig := mysql.Config{
		Conn: db,
	}

	// 创建 MySQL 方言。
	mysqldialector := mysql.New(mysqlconfig)

	// 配置并初始化 gorm。
	gormDB, err := gorm.Open(mysqldialector, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名。
		},
		Logger: kitgorm.NewLogger(logger), // 使用 kit 包中的 gorm logger 适配器。
	})
	if err != nil {
		logger.Error(fmt.Errorf("初始化数据库失败：%w", err))
		return
	}

	// 创建用户示例。
	user := &User{
		Name: "张大山",
		Age:  25,
	}
	if err := gormDB.Create(user).Error; err != nil {
		logger.Error(fmt.Errorf("创建用户失败：%w", err))
		return
	}
	logger.Info(fmt.Sprintf("创建用户成功，ID：%d", user.ID))

	// 查询单个用户示例。
	var queryUser User
	if err := gormDB.First(&queryUser, user.ID).Error; err != nil {
		logger.Error(fmt.Errorf("查询用户失败：%w", err))
		return
	}
	logger.Info(fmt.Sprintf("查询用户成功：%+v", queryUser))

	// 更新用户年龄示例。
	if err := gormDB.Model(&queryUser).Update("age", 30).Error; err != nil {
		logger.Error(fmt.Errorf("更新用户年龄失败：%w", err))
		return
	}
	logger.Info("更新用户年龄成功")

	// 查询所有用户示例。
	var users []User
	if err := gormDB.Find(&users).Error; err != nil {
		logger.Error(fmt.Errorf("查询所有用户失败：%w", err))
		return
	}
	logger.Info(fmt.Sprintf("查询所有用户成功，共 %d 条记录", len(users)))

	// 删除用户示例。
	if err := gormDB.Delete(&user).Error; err != nil {
		logger.Error(fmt.Errorf("删除用户失败：%w", err))
		return
	}
	logger.Info("删除用户成功")
}
