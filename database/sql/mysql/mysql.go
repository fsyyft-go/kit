// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package mysql

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	gosqldriver "github.com/go-sql-driver/mysql"

	kitdriver "github.com/fsyyft-go/kit/database/sql/driver"
	kitlog "github.com/fsyyft-go/kit/log"
)

// 用于保护驱动注册过程的互斥锁。
var (
	driverLocker sync.Mutex
)

// MySQL 连接的默认配置值。
var (
	// 默认的数据源名称（DSN）。
	defaultDSN = "test:test@tcp(localhost:3306)/test?parseTime=true&loc=Local&allowNativePasswords=true&interpolateParams=true"
	// 默认的连接空闲超时时间。
	defaultPoolIdleTime = 10 * time.Second
	// 默认的连接最大空闲时间。
	defaultPoolMaxIdleTime = 10 * time.Second
	// 默认的最大打开连接数。
	defaultPoolMaxOpenConns = 100
	// 默认的最大空闲连接数。
	defaultPoolMaxIdleConns = 10
)

type (
	// MySQLOptions 定义了 MySQL 数据库连接的配置选项。
	MySQLOptions struct {
		// dns 定义数据库的连接字符串。
		dns string
		// poolIdleTime 定义连接在连接池中的空闲超时时间。
		poolIdleTime time.Duration
		// poolMaxIdleTime 定义连接在连接池中的最大空闲时间。
		poolMaxIdleTime time.Duration
		// poolMaxOpenConns 定义连接池中允许的最大打开连接数。
		poolMaxOpenConns int
		// poolMaxIdleConns 定义连接池中允许的最大空闲连接数。
		poolMaxIdleConns int

		// hook 用于管理数据库操作的钩子函数。
		hook *kitdriver.HookManager
		// namespace 定义数据库连接的命名空间。
		namespace string
		// logger 用于记录数据库操作日志。
		logger kitlog.Logger
		// logError 控制是否记录错误日志。
		logError bool
		// slowThreshold 定义慢查询的时间阈值。
		slowThreshold time.Duration
	}

	// MySQLOption 定义了用于配置 MySQL 选项的函数类型。
	MySQLOption func(*MySQLOptions)
)

// WithDSN 设置 MySQL 数据源名称（DSN）。
func WithDSN(dsn string) MySQLOption {
	return func(o *MySQLOptions) {
		o.dns = dsn
	}
}

// WithDSNParams 使用基础 DSN 和额外参数设置 MySQL 数据源名称。
//
// 参数：
//   - baseDSN: 基础数据源名称字符串，如果为空则使用默认 DSN。
//   - params: DSN 参数映射，key 为参数名，value 为参数值。
//
// 示例：
//
//	WithDSNParams("user:pass@tcp(host:port)/dbname", map[string]string{
//	    "parseTime": "true",
//	    "loc": "Local",
//	})
func WithDSNParams(baseDSN string, params map[string]string) MySQLOption {
	return func(o *MySQLOptions) {
		if len(params) == 0 {
			o.dns = baseDSN
			return
		}

		if baseDSN == "" {
			baseDSN = defaultDSN
		}

		paramStr := ""
		if !strings.Contains(baseDSN, "?") {
			paramStr = "?"
		} else {
			paramStr = "&"
		}

		for key, value := range params {
			paramStr += fmt.Sprintf("%s=%s&", key, value)
		}
		paramStr = strings.TrimRight(paramStr, "&")

		o.dns = baseDSN + paramStr
	}
}

// WithPoolIdleTime 设置连接的空闲超时时间。
func WithPoolIdleTime(idleTime time.Duration) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolIdleTime = idleTime
	}
}

// WithPoolMaxIdleTime 设置连接的最大空闲时间。
func WithPoolMaxIdleTime(maxIdleTime time.Duration) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolMaxIdleTime = maxIdleTime
	}
}

// WithPoolMaxOpenConns 设置最大打开连接数。
func WithPoolMaxOpenConns(maxOpenConns int) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolMaxOpenConns = maxOpenConns
	}
}

// WithPoolMaxIdleConns 设置最大空闲连接数。
func WithPoolMaxIdleConns(maxIdleConns int) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolMaxIdleConns = maxIdleConns
	}
}

// WithNamespace 设置数据库连接的命名空间。
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

// WithLogError 设置是否启用错误日志记录。
func WithLogError(logError bool) MySQLOption {
	return func(o *MySQLOptions) {
		o.logError = logError
	}
}

// WithSlowThreshold 设置慢查询的时间阈值。
func WithSlowThreshold(slowThreshold time.Duration) MySQLOption {
	return func(o *MySQLOptions) {
		o.slowThreshold = slowThreshold
	}
}

// WithHookManager 设置钩子管理器。
func WithHookManager(hook *kitdriver.HookManager) MySQLOption {
	return func(o *MySQLOptions) {
		o.hook = hook
	}
}

// NewMySQL 创建并返回一个新的 MySQL 数据库连接实例。
//
// 参数：
//   - opts: MySQL 配置选项的可变参数列表。
//
// 返回值：
//   - *sql.DB: 数据库连接实例。
//   - func(): 清理函数，用于关闭数据库连接。
//   - error: 如果创建过程中发生错误，返回相应的错误信息。
func NewMySQL(opts ...MySQLOption) (*sql.DB, func(), error) {
	// 加锁保护驱动注册过程。
	driverLocker.Lock()
	defer driverLocker.Unlock()

	// 初始化默认配置选项。
	options := &MySQLOptions{
		dns:              defaultDSN,
		poolIdleTime:     defaultPoolIdleTime,
		poolMaxIdleTime:  defaultPoolMaxIdleTime,
		poolMaxOpenConns: defaultPoolMaxOpenConns,
		poolMaxIdleConns: defaultPoolMaxIdleConns,
	}
	// 应用用户提供的配置选项。
	for _, opt := range opts {
		opt(options)
	}

	var err error

	// 生成唯一的驱动名称。
	driverName := fmt.Sprintf("mysql-kit-%s", options.namespace)
	// 检查驱动是否已注册。
	registered := sql.Drivers()
	if !slices.Contains(registered, driverName) {
		// 配置钩子。
		if nil == options.hook {
			// 如果钩子管理器为空，则创建一个新的钩子管理器。
			options.hook = kitdriver.NewHookManager()
			// 配置钩子。
			if err = configureHooks(options.hook, options); nil != err {
				return nil, nil, err
			}
		}

		// 创建并注册带有钩子的 MySQL 驱动。
		originalDriver := gosqldriver.MySQLDriver{}
		kitDriver := kitdriver.NewKitDriver(originalDriver, options.hook)
		sql.Register(driverName, kitDriver)
	}

	// 打开数据库连接。
	var db *sql.DB
	if db, err = sql.Open(driverName, options.dns); nil != err {
		if nil != options.logger {
			options.logger.Error("mysql", "error", err)
		}
		return nil, nil, err
	} else {
		// 配置连接池参数。
		db.SetMaxOpenConns(options.poolMaxOpenConns)
		db.SetMaxIdleConns(options.poolMaxIdleConns)
		db.SetConnMaxIdleTime(options.poolMaxIdleTime)
		db.SetConnMaxLifetime(options.poolIdleTime)
	}

	// 定义清理函数。
	cleanup := func() {
		if err := db.Close(); nil != err && nil != options.logger {
			options.logger.Error("mysql", "error", err)
		}
	}

	return db, cleanup, err
}

// configureHooks 根据 MySQL 选项配置钩子管理器。
//
// 参数：
//   - hook: 需要配置的钩子管理器。
//   - opts: MySQL 配置选项。
//
// 返回值：
//   - error: 如果配置过程中发生错误，返回相应的错误信息。
func configureHooks(hook *kitdriver.HookManager, opts *MySQLOptions) error {
	var err error
	// 配置错误日志钩子。
	if opts.logError {
		if nil == opts.logger {
			opts.logger, err = kitlog.NewLogger()
			if nil != err {
				return err
			}
		}
		h := kitdriver.NewHookLogError(opts.namespace, opts.logger)
		hook.AddHook(h)
	}

	// 配置慢查询日志钩子。
	if opts.slowThreshold > 0 {
		h := kitdriver.NewHookLogSlow(opts.namespace, opts.logger, opts.slowThreshold)
		hook.AddHook(h)
	}

	return nil
}
