// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package mysql

import (
	"database/sql"
	"fmt"
	"net/url"
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
	sqlOpen      = sql.Open
	closeDB      = func(db *sql.DB) error { return db.Close() }
	newLogger    = func() (kitlog.Logger, error) { return kitlog.NewLogger() }
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
	// MySQLOptions 保存 NewMySQL 的构造参数。
	//
	// MySQLOptions 通常通过 MySQLOption 函数组合填充，而不是由调用方直接操作
	// 内部字段。NewMySQL 会先建立一份默认配置，再按传入顺序应用这些选项。
	MySQLOptions struct {
		// dns 定义数据库的连接字符串。
		dns string
		// poolIdleTime 定义连接的最大生命周期。
		poolIdleTime time.Duration
		// poolMaxIdleTime 定义连接的最大空闲时长。
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

	// MySQLOption 定义按引用修改 MySQLOptions 的函数式选项。
	//
	// 多个选项按传入顺序应用，后面的设置会覆盖前面的同类配置。
	MySQLOption func(*MySQLOptions)
)

// WithDSN 设置传递给 NewMySQL 的原始 DSN。
//
// DSN 会在 NewMySQL 中通过 go-sql-driver/mysql 的 ParseDSN 校验；
// 本选项本身不执行格式检查。
//
// 参数：
//   - dsn：数据库连接字符串，例如 "user:password@tcp(host:port)/dbname?param=value"。
//
// 返回值：
//   - MySQLOption：设置 DSN 的配置函数。
func WithDSN(dsn string) MySQLOption {
	return func(o *MySQLOptions) {
		o.dns = dsn
	}
}

// WithDSNParams 基于基础 DSN 追加一组查询参数。
//
// baseDSN 为空时使用包内默认 DSN。params 中的值会按 URL 查询串规则编码；
// 当基础 DSN 已包含查询参数时继续使用 "&" 追加，否则使用 "?"。params
// 为空时保持基础 DSN 原样不变。
//
// 参数：
//   - baseDSN：基础 DSN；为空时使用包内默认 DSN。
//   - params：要追加的 DSN 查询参数。
//
// 返回值：
//   - MySQLOption：设置 DSN 的配置函数。
func WithDSNParams(baseDSN string, params map[string]string) MySQLOption {
	return func(o *MySQLOptions) {
		if baseDSN == "" {
			baseDSN = defaultDSN
		}
		if len(params) == 0 {
			o.dns = baseDSN
			return
		}

		values := url.Values{}
		for key, value := range params {
			values.Set(key, value)
		}

		separator := "?"
		if strings.Contains(baseDSN, "?") {
			separator = "&"
		}

		o.dns = baseDSN + separator + values.Encode()
	}
}

// WithPoolIdleTime 设置 NewMySQL 创建的 *sql.DB 的连接最大生命周期。
//
// 该选项最终映射到 (*sql.DB).SetConnMaxLifetime。
func WithPoolIdleTime(idleTime time.Duration) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolIdleTime = idleTime
	}
}

// WithPoolMaxIdleTime 设置 NewMySQL 创建的 *sql.DB 的连接最大空闲时长。
//
// 该选项最终映射到 (*sql.DB).SetConnMaxIdleTime。
func WithPoolMaxIdleTime(maxIdleTime time.Duration) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolMaxIdleTime = maxIdleTime
	}
}

// WithPoolMaxOpenConns 设置 NewMySQL 创建的 *sql.DB 的最大打开连接数。
func WithPoolMaxOpenConns(maxOpenConns int) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolMaxOpenConns = maxOpenConns
	}
}

// WithPoolMaxIdleConns 设置 NewMySQL 创建的 *sql.DB 的最大空闲连接数。
func WithPoolMaxIdleConns(maxIdleConns int) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolMaxIdleConns = maxIdleConns
	}
}

// WithNamespace 设置当前配置对应的驱动注册命名空间。
//
// NewMySQL 会使用 "mysql-kit-<namespace>" 作为 driver 名称；同一 namespace
// 的后续调用会复用已注册的 driver 与其初次注册时确定的 Hook 配置。
func WithNamespace(namespace string) MySQLOption {
	return func(o *MySQLOptions) {
		o.namespace = namespace
	}
}

// WithLogger 设置构造阶段和自动日志 Hook 使用的 logger。
//
// 该 logger 用于记录 DSN 校验失败、sql.Open 失败和 cleanup 关闭失败，
// 也会在自动安装 HookLogError 或 HookLogSlow 时复用。
func WithLogger(logger kitlog.Logger) MySQLOption {
	return func(o *MySQLOptions) {
		o.logger = logger
	}
}

// WithLogError 请求为自动创建的 HookManager 安装 HookLogError。
//
// 该选项只在 NewMySQL 首次为某个 namespace 注册 driver 且未显式提供
// WithHookManager 时生效。
func WithLogError(logError bool) MySQLOption {
	return func(o *MySQLOptions) {
		o.logError = logError
	}
}

// WithSlowThreshold 设置自动慢操作日志 Hook 的耗时阈值。
//
// 该选项只在 NewMySQL 首次为某个 namespace 注册 driver 且未显式提供
// WithHookManager 时生效；当操作耗时大于等于该阈值时会记录 Warn 日志。
func WithSlowThreshold(slowThreshold time.Duration) MySQLOption {
	return func(o *MySQLOptions) {
		o.slowThreshold = slowThreshold
	}
}

// WithHookManager 指定一个自定义 HookManager 供驱动包装使用。
//
// 提供该选项后，NewMySQL 不会再为当前调用自动安装 HookLogError 或 HookLogSlow；
// 调用方需要自行向该管理器注册所需 Hook。
func WithHookManager(hook *kitdriver.HookManager) MySQLOption {
	return func(o *MySQLOptions) {
		o.hook = hook
	}
}

// NewMySQL 基于 go-sql-driver/mysql 构造一个 *sql.DB 和清理函数。
//
// NewMySQL 会先应用默认配置与传入选项，使用 ParseDSN 校验 DSN，然后以
// "mysql-kit-<namespace>" 为名称按需注册带 Hook 的包装 driver，最后调用
// sql.Open 创建 *sql.DB 并设置连接池参数。该函数只调用 sql.Open，不会主动
// Ping 数据库，调用方需要在需要时自行验证连通性。
//
// 同一 namespace 的 driver 只会注册一次，后续调用会复用既有 driver 和其
// 初次注册时确定的 Hook 配置；新的 WithHookManager、WithLogError 和
// WithSlowThreshold 不会重新装配已注册 driver。
//
// 参数：
//   - opts：按顺序应用的 MySQL 构造选项。
//
// 返回值：
//   - *sql.DB：创建成功的数据库句柄。
//   - func()：清理函数，内部调用 db.Close；调用方在不再使用时应负责调用 cleanup 或等价的 db.Close。
//   - error：DSN 校验失败、自动创建 logger 失败、sql.Open 失败，或驱动注册前 Hook 配置失败时返回错误。
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
	if _, err = gosqldriver.ParseDSN(options.dns); nil != err {
		if nil != options.logger {
			options.logger.Error("mysql", "error", err)
		}
		return nil, nil, err
	}

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
	if db, err = sqlOpen(driverName, options.dns); nil != err {
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
		if err := closeDB(db); nil != err && nil != options.logger {
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
	if opts.logError || opts.slowThreshold > 0 {
		if nil == opts.logger {
			opts.logger, err = newLogger()
			if nil != err {
				return err
			}
		}
	}

	// 配置错误日志钩子。
	if opts.logError {
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
