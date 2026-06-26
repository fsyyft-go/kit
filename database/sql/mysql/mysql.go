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

// driverLocker 保护 driver 注册流程，避免同一进程内并发注册相同名称的 MySQL driver。
//
// sqlOpen、closeDB 和 newLogger 保存构造流程中的可替换入口，测试可替换这些变量以避免真实打开数据库、观察 cleanup 行为或模拟 logger 构造失败。
var (
	driverLocker sync.Mutex
	sqlOpen      = sql.Open
	closeDB      = func(db *sql.DB) error { return db.Close() }
	newLogger    = func() (kitlog.Logger, error) { return kitlog.NewLogger() }
)

var (
	// defaultDSN 是 NewMySQL 未显式配置 DSN 时使用的默认数据源名称。
	defaultDSN = "test:test@tcp(localhost:3306)/test?parseTime=true&loc=Local&allowNativePasswords=true&interpolateParams=true"
	// defaultPoolIdleTime 是连接可被复用的默认最大生命周期，最终传给 (*sql.DB).SetConnMaxLifetime。
	defaultPoolIdleTime = 10 * time.Second
	// defaultPoolMaxIdleTime 是空闲连接在连接池中保留的默认最长时间，最终传给 (*sql.DB).SetConnMaxIdleTime。
	defaultPoolMaxIdleTime = 10 * time.Second
	// defaultPoolMaxOpenConns 是连接池默认允许同时打开的最大连接数。
	defaultPoolMaxOpenConns = 100
	// defaultPoolMaxIdleConns 是连接池默认保留的最大空闲连接数。
	defaultPoolMaxIdleConns = 10
)

type (
	// MySQLOptions 保存 NewMySQL 的构造参数。
	//
	// MySQLOptions 通常通过 MySQLOption 函数组合填充，而不是由调用方直接操作
	// 内部字段。NewMySQL 会先建立一份默认配置，再按传入顺序应用这些选项。
	MySQLOptions struct {
		// dns 保存传递给 go-sql-driver/mysql 的 DSN 字符串。
		dns string
		// poolIdleTime 定义连接可被复用的最大生命周期。
		poolIdleTime time.Duration
		// poolMaxIdleTime 定义空闲连接在连接池中保留的最长时间。
		poolMaxIdleTime time.Duration
		// poolMaxOpenConns 定义连接池中允许的最大打开连接数。
		poolMaxOpenConns int
		// poolMaxIdleConns 定义连接池中允许的最大空闲连接数。
		poolMaxIdleConns int
		// hook 用于管理数据库操作的 Hook 链。
		hook *kitdriver.HookManager
		// namespace 定义当前配置对应的 driver 注册命名空间。
		namespace string
		// logger 用于记录构造阶段错误以及自动日志 Hook 产生的日志。
		logger kitlog.Logger
		// logError 控制是否在自动 HookManager 中安装错误日志 Hook。
		logError bool
		// slowThreshold 定义自动慢操作日志 Hook 的耗时阈值。
		slowThreshold time.Duration
	}

	// MySQLOption 定义按引用修改 MySQLOptions 的函数式选项。
	//
	// 多个选项按传入顺序应用，后面的设置会覆盖前面的同类配置。
	//
	// 参数：
	//   - *MySQLOptions: 待修改的 MySQL 构造配置；调用方不应直接传入 nil。
	MySQLOption func(*MySQLOptions)
)

// WithDSN 设置传递给 NewMySQL 的原始 DSN。
//
// DSN 会在 NewMySQL 中通过 go-sql-driver/mysql 的 ParseDSN 校验；
// 本选项本身不执行格式检查。
//
// 参数：
//   - dsn: 数据库连接字符串，例如 "user:password@tcp(host:port)/dbname?param=value"。
//
// 返回：
//   - MySQLOption: 设置 DSN 的配置函数。
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
//   - baseDSN: 基础 DSN；为空时使用包内默认 DSN。
//   - params: 要追加的 DSN 查询参数；为空时不追加查询串。
//
// 返回：
//   - MySQLOption: 设置 DSN 的配置函数。
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
// 该选项最终映射到 (*sql.DB).SetConnMaxLifetime。非正值的含义沿用标准库 database/sql。
//
// 参数：
//   - idleTime: 连接可被复用的最长时间。
//
// 返回：
//   - MySQLOption: 设置连接最大生命周期的配置函数。
func WithPoolIdleTime(idleTime time.Duration) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolIdleTime = idleTime
	}
}

// WithPoolMaxIdleTime 设置 NewMySQL 创建的 *sql.DB 的连接最大空闲时长。
//
// 该选项最终映射到 (*sql.DB).SetConnMaxIdleTime。非正值的含义沿用标准库 database/sql。
//
// 参数：
//   - maxIdleTime: 空闲连接在连接池中保留的最长时间。
//
// 返回：
//   - MySQLOption: 设置连接最大空闲时长的配置函数。
func WithPoolMaxIdleTime(maxIdleTime time.Duration) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolMaxIdleTime = maxIdleTime
	}
}

// WithPoolMaxOpenConns 设置 NewMySQL 创建的 *sql.DB 的最大打开连接数。
//
// 该选项最终映射到 (*sql.DB).SetMaxOpenConns。非正值的含义沿用标准库 database/sql。
//
// 参数：
//   - maxOpenConns: 连接池允许同时打开的最大连接数。
//
// 返回：
//   - MySQLOption: 设置最大打开连接数的配置函数。
func WithPoolMaxOpenConns(maxOpenConns int) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolMaxOpenConns = maxOpenConns
	}
}

// WithPoolMaxIdleConns 设置 NewMySQL 创建的 *sql.DB 的最大空闲连接数。
//
// 该选项最终映射到 (*sql.DB).SetMaxIdleConns。非正值与超过最大打开连接数时的行为沿用标准库 database/sql。
//
// 参数：
//   - maxIdleConns: 连接池允许保留的最大空闲连接数。
//
// 返回：
//   - MySQLOption: 设置最大空闲连接数的配置函数。
func WithPoolMaxIdleConns(maxIdleConns int) MySQLOption {
	return func(o *MySQLOptions) {
		o.poolMaxIdleConns = maxIdleConns
	}
}

// WithNamespace 设置当前配置对应的 driver 注册命名空间。
//
// NewMySQL 会使用 "mysql-kit-<namespace>" 作为 driver 名称；同一 namespace
// 的后续调用会复用已注册的 driver 与其初次注册时确定的 Hook 配置。
//
// 参数：
//   - namespace: driver 注册命名空间；空字符串会生成 "mysql-kit-"。
//
// 返回：
//   - MySQLOption: 设置命名空间的配置函数。
func WithNamespace(namespace string) MySQLOption {
	return func(o *MySQLOptions) {
		o.namespace = namespace
	}
}

// WithLogger 设置构造阶段和自动日志 Hook 使用的 logger。
//
// 该 logger 用于记录 DSN 校验失败、sql.Open 失败和 cleanup 关闭失败，
// 也会在自动安装 HookLogError 或 HookLogSlow 时复用。
//
// 参数：
//   - logger: 用于输出日志的 kit logger；传入 nil 时仅在需要自动日志 Hook 时尝试创建默认 logger。
//
// 返回：
//   - MySQLOption: 设置 logger 的配置函数。
func WithLogger(logger kitlog.Logger) MySQLOption {
	return func(o *MySQLOptions) {
		o.logger = logger
	}
}

// WithLogError 请求为自动创建的 HookManager 安装 HookLogError。
//
// 该选项只在 NewMySQL 首次为某个 namespace 注册 driver 且未显式提供
// WithHookManager 时生效。
//
// 参数：
//   - logError: 为 true 时启用错误日志 Hook；为 false 时不自动安装。
//
// 返回：
//   - MySQLOption: 设置错误日志 Hook 开关的配置函数。
func WithLogError(logError bool) MySQLOption {
	return func(o *MySQLOptions) {
		o.logError = logError
	}
}

// WithSlowThreshold 设置自动慢操作日志 Hook 的耗时阈值。
//
// 该选项只在 NewMySQL 首次为某个 namespace 注册 driver 且未显式提供
// WithHookManager 时生效；当操作耗时大于等于该阈值时会记录 Warn 日志。
// 非正值表示不自动安装慢操作日志 Hook。
//
// 参数：
//   - slowThreshold: 慢操作耗时阈值。
//
// 返回：
//   - MySQLOption: 设置慢操作阈值的配置函数。
func WithSlowThreshold(slowThreshold time.Duration) MySQLOption {
	return func(o *MySQLOptions) {
		o.slowThreshold = slowThreshold
	}
}

// WithHookManager 指定一个自定义 HookManager 供驱动包装使用。
//
// 传入非 nil HookManager 后，NewMySQL 不会再为当前调用自动安装 HookLogError 或 HookLogSlow；
// 调用方需要自行向该管理器注册所需 Hook。传入 nil 等价于未指定。
//
// 参数：
//   - hook: 用于包装 driver 的 HookManager；调用方应传入非 nil 实例。
//
// 返回：
//   - MySQLOption: 设置 HookManager 的配置函数。
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
//   - opts: 按顺序应用的 MySQL 构造选项。
//
// 返回：
//   - *sql.DB: 创建成功的数据库句柄，由调用方负责在不再使用时关闭。
//   - func(): 清理函数，内部调用 db.Close；关闭失败时吞掉错误，并在 logger 非 nil 时记录错误；需要感知关闭错误的调用方应直接调用 db.Close。
//   - error: DSN 校验失败、自动创建 logger 失败、sql.Open 失败，或驱动注册前 Hook 配置失败时返回错误。
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

// configureHooks 根据 MySQL 选项配置 HookManager。
//
// 参数：
//   - hook: 需要配置的 HookManager。
//   - opts: MySQL 构造配置；当需要自动日志 Hook 且 logger 为空时会写入新建的 logger。
//
// 返回：
//   - error: 创建默认 logger 失败时返回错误；不需要默认 logger 或配置成功时返回 nil。
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
