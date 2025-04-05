// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package mysql

import (
	"database/sql"
	"fmt"
	"slices"
	"sync"
	"time"

	goSqlDriver "github.com/go-sql-driver/mysql"

	kitdriver "github.com/fsyyft-go/kit/database/sql/driver"
	kitlog "github.com/fsyyft-go/kit/log"
)

var (
	driverLocker sync.Mutex
)

var (
	defaultDSN              = "root:root@tcp(localhost:3306)/test"
	defaultPoolSize         = 10
	defaultPoolIdleTime     = 10 * time.Second
	defaultPoolMaxIdleTime  = 10 * time.Second
	defaultPoolMaxOpenConns = 100
	defaultPoolMaxIdleConns = 10
)

type (
	// MySQLOptions 是 MySQL 数据库的配置选项。
	MySQLOptions struct {
		// 数据库连接字符串。
		dns string
		// 数据库连接池大小。
		poolSize int
		// 数据库连接池空闲时间。
		poolIdleTime time.Duration
		// 数据库连接池最大空闲时间。
		poolMaxIdleTime time.Duration
		// 数据库连接池最大连接数。
		poolMaxOpenConns int
		// 数据库连接池最大空闲连接数。
		poolMaxIdleConns int

		// 钩子管理器。
		hook *kitdriver.HookManager
		// 命名空间。
		namespace string
		// 日志记录器。
		logger kitlog.Logger
		// 日志记录错误。
		logError bool
		// 慢查询阈值。
		slowThreshold time.Duration
	}

	// MySQLOption 是 MySQL 数据库的配置选项。
	MySQLOption func(*MySQLOptions)
)

// WithDSN 设置数据库连接字符串。
func WithDSN(dsn string) MySQLOption {
	return func(o *MySQLOptions) {
		o.dns = dsn
	}
}

// WithPoolSize 设置数据库连接池大小。
func WithPoolSize(size int) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolSize = size
	}
}

// WithPoolIdleTime 设置数据库连接池空闲时间。
func WithPoolIdleTime(idleTime time.Duration) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolIdleTime = idleTime
	}
}

// WithPoolMaxIdleTime 设置数据库连接池最大空闲时间。
func WithPoolMaxIdleTime(maxIdleTime time.Duration) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolMaxIdleTime = maxIdleTime
	}
}

// WithPoolMaxOpenConns 设置数据库连接池最大连接数。
func WithPoolMaxOpenConns(maxOpenConns int) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolMaxOpenConns = maxOpenConns
	}
}

// WithPoolMaxIdleConns 设置数据库连接池最大空闲连接数。
func WithPoolMaxIdleConns(maxIdleConns int) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolMaxIdleConns = maxIdleConns
	}
}

// WithNamespace 设置命名空间。
func WithNamespace(namespace string) MySQLOption {
	return func(o *MySQLOptions) {
		o.namespace = namespace
	}
}

// WithLogger 设置日志记录器。
func WithLogger(logger kitlog.Logger) MySQLOption {
	return func(o *MySQLOptions) {
		o.logger = logger
	}
}

// WithLogError 设置是否记录错误。
func WithLogError(logError bool) MySQLOption {
	return func(o *MySQLOptions) {
		o.logError = logError
	}
}

// WithSlowThreshold 设置慢查询阈值。
func WithSlowThreshold(slowThreshold time.Duration) MySQLOption {
	return func(o *MySQLOptions) {
		o.slowThreshold = slowThreshold
	}
}

// NewMySQL 创建一个新的 MySQL 数据库实例。
func NewMySQL(opts ...MySQLOption) (*sql.DB, func(), error) {
	driverLocker.Lock()
	defer driverLocker.Unlock()

	options := &MySQLOptions{
		dns:              defaultDSN,
		poolSize:         defaultPoolSize,
		poolIdleTime:     defaultPoolIdleTime,
		poolMaxIdleTime:  defaultPoolMaxIdleTime,
		poolMaxOpenConns: defaultPoolMaxOpenConns,
		poolMaxIdleConns: defaultPoolMaxIdleConns,
		hook:             kitdriver.NewHookManager(),
	}
	for _, opt := range opts {
		opt(options)
	}

	var err error

	driverName := fmt.Sprintf("mysql-kit-%s", options.namespace)
	registered := sql.Drivers()
	if !slices.Contains(registered, driverName) {
		if options.logError {
			h := kitdriver.NewHookLogError(options.namespace, options.logger)
			options.hook.AddHook(h)
		}

		if options.slowThreshold > 0 {
			h := kitdriver.NewHookLogSlow(options.namespace, options.logger, options.slowThreshold)
			options.hook.AddHook(h)
		}

		originalDriver := goSqlDriver.MySQLDriver{}
		kitDriver := kitdriver.NewKitDriver(originalDriver, options.hook)
		sql.Register(driverName, kitDriver)
	}

	var db *sql.DB
	if db, err = sql.Open(driverName, options.dns); err != nil {
		return nil, nil, err
	} else {
		db.SetMaxOpenConns(options.poolMaxOpenConns)
		db.SetMaxIdleConns(options.poolMaxIdleConns)
		db.SetConnMaxIdleTime(options.poolMaxIdleTime)
		db.SetConnMaxLifetime(options.poolIdleTime)
	}

	cleanup := func() {
		if err := db.Close(); err != nil {
			options.logger.Error("mysql", "error", err)
		}
	}

	return db, cleanup, err
}
