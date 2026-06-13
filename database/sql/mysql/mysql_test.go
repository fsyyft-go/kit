// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	kitdriver "github.com/fsyyft-go/kit/database/sql/driver"
	kitlog "github.com/fsyyft-go/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMySQLOptions_Apply 验证 MySQL 函数式选项能够准确写入目标配置字段。
//
// 该测试通过表驱动用例覆盖公开选项函数，确保 DSN、连接池、命名空间、日志、慢查询阈值与钩子管理器配置稳定生效。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestMySQLOptions_Apply(t *testing.T) {
	logger, err := kitlog.NewLogger()
	require.NoError(t, err)

	hook := kitdriver.NewHookManager()

	tests := []struct {
		name        string
		description string
		giveOption  MySQLOption
		assert      func(t *testing.T, got *MySQLOptions)
	}{
		{
			name:        "success/dsn",
			description: "验证 WithDSN 将完整数据源名称写入配置。",
			giveOption:  WithDSN("user:pass@tcp(127.0.0.1:3306)/app"),
			assert: func(t *testing.T, got *MySQLOptions) {
				assert.Equal(t, "user:pass@tcp(127.0.0.1:3306)/app", got.dns)
			},
		},
		{
			name:        "success/pool-idle-time",
			description: "验证 WithPoolIdleTime 将连接最大生命周期写入配置。",
			giveOption:  WithPoolIdleTime(3 * time.Second),
			assert: func(t *testing.T, got *MySQLOptions) {
				assert.Equal(t, 3*time.Second, got.poolIdleTime)
			},
		},
		{
			name:        "success/pool-max-idle-time",
			description: "验证 WithPoolMaxIdleTime 将连接最大空闲时间写入配置。",
			giveOption:  WithPoolMaxIdleTime(4 * time.Second),
			assert: func(t *testing.T, got *MySQLOptions) {
				assert.Equal(t, 4*time.Second, got.poolMaxIdleTime)
			},
		},
		{
			name:        "success/pool-max-open-conns",
			description: "验证 WithPoolMaxOpenConns 将最大打开连接数写入配置。",
			giveOption:  WithPoolMaxOpenConns(17),
			assert: func(t *testing.T, got *MySQLOptions) {
				assert.Equal(t, 17, got.poolMaxOpenConns)
			},
		},
		{
			name:        "success/pool-max-idle-conns",
			description: "验证 WithPoolMaxIdleConns 将最大空闲连接数写入配置。",
			giveOption:  WithPoolMaxIdleConns(5),
			assert: func(t *testing.T, got *MySQLOptions) {
				assert.Equal(t, 5, got.poolMaxIdleConns)
			},
		},
		{
			name:        "success/namespace",
			description: "验证 WithNamespace 将驱动命名空间写入配置。",
			giveOption:  WithNamespace("unit-app"),
			assert: func(t *testing.T, got *MySQLOptions) {
				assert.Equal(t, "unit-app", got.namespace)
			},
		},
		{
			name:        "success/logger",
			description: "验证 WithLogger 将传入的日志记录器实例写入配置。",
			giveOption:  WithLogger(logger),
			assert: func(t *testing.T, got *MySQLOptions) {
				assert.Same(t, logger, got.logger)
			},
		},
		{
			name:        "success/log-error",
			description: "验证 WithLogError 将错误日志开关写入配置。",
			giveOption:  WithLogError(true),
			assert: func(t *testing.T, got *MySQLOptions) {
				assert.True(t, got.logError)
			},
		},
		{
			name:        "success/slow-threshold",
			description: "验证 WithSlowThreshold 将慢查询阈值写入配置。",
			giveOption:  WithSlowThreshold(250 * time.Millisecond),
			assert: func(t *testing.T, got *MySQLOptions) {
				assert.Equal(t, 250*time.Millisecond, got.slowThreshold)
			},
		},
		{
			name:        "success/hook-manager",
			description: "验证 WithHookManager 将外部钩子管理器写入配置。",
			giveOption:  WithHookManager(hook),
			assert: func(t *testing.T, got *MySQLOptions) {
				assert.Same(t, hook, got.hook)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := &MySQLOptions{}
			tt.giveOption(got)

			tt.assert(t, got)
		})
	}
}

// TestWithDSNParams_Compose 验证 WithDSNParams 对基础 DSN 与查询参数的组合行为。
//
// 该测试通过表驱动用例覆盖空基础 DSN、空参数、既有查询串和需要 URL 编码的参数，确保生成的 DSN 稳定且可重复。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestWithDSNParams_Compose(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveBaseDSN string
		giveParams  map[string]string
		wantDSN     string
	}{
		{
			name:        "success/empty-base-empty-params-uses-default",
			description: "验证基础 DSN 与参数均为空时使用包默认 DSN。",
			giveBaseDSN: "",
			giveParams:  nil,
			wantDSN:     defaultDSN,
		},
		{
			name:        "success/non-empty-base-empty-params-keeps-base",
			description: "验证参数为空且基础 DSN 非空时保留原始基础 DSN。",
			giveBaseDSN: "user:pass@tcp(127.0.0.1:3306)/app",
			giveParams:  map[string]string{},
			wantDSN:     "user:pass@tcp(127.0.0.1:3306)/app",
		},
		{
			name:        "success/appends-params-to-base-without-query",
			description: "验证基础 DSN 无查询串时使用问号追加稳定排序后的参数。",
			giveBaseDSN: "user:pass@tcp(127.0.0.1:3306)/app",
			giveParams: map[string]string{
				"parseTime": "true",
				"loc":       "Asia/Shanghai",
			},
			wantDSN: "user:pass@tcp(127.0.0.1:3306)/app?loc=Asia%2FShanghai&parseTime=true",
		},
		{
			name:        "success/appends-params-to-base-with-query",
			description: "验证基础 DSN 已有查询串时使用与号追加额外参数。",
			giveBaseDSN: "user:pass@tcp(127.0.0.1:3306)/app?charset=utf8mb4",
			giveParams: map[string]string{
				"timeout":     "10s",
				"readTimeout": "1s",
			},
			wantDSN: "user:pass@tcp(127.0.0.1:3306)/app?charset=utf8mb4&readTimeout=1s&timeout=10s",
		},
		{
			name:        "success/empty-base-with-params-uses-default",
			description: "验证基础 DSN 为空但参数非空时基于包默认 DSN 追加参数。",
			giveBaseDSN: "",
			giveParams: map[string]string{
				"charset": "utf8mb4",
				"timeout": "3s",
			},
			wantDSN: defaultDSN + "&charset=utf8mb4&timeout=3s",
		},
		{
			name:        "success/url-encodes-values",
			description: "验证参数值包含斜杠、空格和加号时按 URL 查询串规则编码。",
			giveBaseDSN: "user:pass@tcp(127.0.0.1:3306)/app",
			giveParams: map[string]string{
				"time_zone": "Asia/Shanghai +08:00",
			},
			wantDSN: "user:pass@tcp(127.0.0.1:3306)/app?time_zone=Asia%2FShanghai+%2B08%3A00",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := &MySQLOptions{}
			WithDSNParams(tt.giveBaseDSN, tt.giveParams)(got)

			assert.Equal(t, tt.wantDSN, got.dns)
		})
	}
}

// TestConfigureHooks_LoggingOptions 验证 configureHooks 根据日志相关配置初始化钩子管理器。
//
// 该测试通过表驱动用例覆盖禁用日志、自动创建日志记录器、复用外部日志记录器和慢查询阈值组合，确保钩子配置过程不依赖外部服务。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestConfigureHooks_LoggingOptions(t *testing.T) {
	logger, err := kitlog.NewLogger()
	require.NoError(t, err)

	tests := []struct {
		name             string
		description      string
		giveOptions      MySQLOptions
		wantLogger       bool
		wantSameLogger   bool
		wantMinimumHooks int
	}{
		{
			name:             "success/disabled-logging-keeps-logger-empty",
			description:      "验证未启用错误日志与慢查询监控时不创建默认日志记录器。",
			giveOptions:      MySQLOptions{namespace: testNamespace(t, "hooks-disabled")},
			wantLogger:       false,
			wantMinimumHooks: 0,
		},
		{
			name:             "success/log-error-creates-default-logger",
			description:      "验证启用错误日志且未提供日志记录器时自动创建默认日志记录器。",
			giveOptions:      MySQLOptions{namespace: testNamespace(t, "hooks-log-error"), logError: true},
			wantLogger:       true,
			wantMinimumHooks: 1,
		},
		{
			name:             "success/log-error-reuses-given-logger",
			description:      "验证启用错误日志且已提供日志记录器时复用既有实例。",
			giveOptions:      MySQLOptions{namespace: testNamespace(t, "hooks-given-logger"), logger: logger, logError: true},
			wantLogger:       true,
			wantSameLogger:   true,
			wantMinimumHooks: 1,
		},
		{
			name:             "success/slow-threshold-without-error-logging",
			description:      "验证仅启用慢查询阈值时会创建默认日志记录器并注册慢查询钩子，避免慢查询 Hook 触发 nil logger panic。",
			giveOptions:      MySQLOptions{namespace: testNamespace(t, "hooks-slow-only"), slowThreshold: time.Millisecond},
			wantLogger:       true,
			wantMinimumHooks: 1,
		},
		{
			name:             "success/log-error-and-slow-threshold",
			description:      "验证错误日志与慢查询阈值同时启用时可完成钩子配置。",
			giveOptions:      MySQLOptions{namespace: testNamespace(t, "hooks-log-and-slow"), logError: true, slowThreshold: time.Millisecond},
			wantLogger:       true,
			wantMinimumHooks: 2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			hook := kitdriver.NewHookManager()
			gotOptions := tt.giveOptions

			require.NoError(t, configureHooks(hook, &gotOptions))

			if tt.wantLogger {
				assert.NotNil(t, gotOptions.logger)
			} else {
				assert.Nil(t, gotOptions.logger)
			}
			if tt.wantSameLogger {
				assert.Equal(t, logger, gotOptions.logger)
			}
			assert.GreaterOrEqual(t, countHookManagerHooks(t, hook), tt.wantMinimumHooks)
		})
	}
}

// TestNewMySQL_ConnectionLifecycle 验证 NewMySQL 创建的数据库句柄和清理函数在无真实数据库服务时的稳定行为。
//
// 该测试通过表驱动用例覆盖默认连接池、定制连接池、外部钩子管理器和日志钩子配置，确保 sql.Open 懒连接语义下不访问外部 MySQL 服务。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewMySQL_ConnectionLifecycle(t *testing.T) {
	tests := []struct {
		name             string
		description      string
		giveOptions      func(t *testing.T) []MySQLOption
		wantMaxOpenConns int
	}{
		{
			name:        "success/default-pool",
			description: "验证默认配置创建数据库句柄并设置默认最大打开连接数。",
			giveOptions: func(t *testing.T) []MySQLOption {
				return []MySQLOption{WithNamespace(testNamespace(t, "default-pool"))}
			},
			wantMaxOpenConns: defaultPoolMaxOpenConns,
		},
		{
			name:        "success/custom-pool-and-dsn",
			description: "验证定制 DSN 与连接池配置创建数据库句柄并写入最大打开连接数。",
			giveOptions: func(t *testing.T) []MySQLOption {
				return []MySQLOption{
					WithNamespace(testNamespace(t, "custom-pool")),
					WithDSN("user:pass@tcp(127.0.0.1:3306)/unit?parseTime=true"),
					WithPoolMaxOpenConns(7),
					WithPoolMaxIdleConns(3),
					WithPoolIdleTime(2 * time.Second),
					WithPoolMaxIdleTime(time.Second),
				}
			},
			wantMaxOpenConns: 7,
		},
		{
			name:        "success/custom-hook-manager",
			description: "验证提供外部钩子管理器时 NewMySQL 可注册驱动并返回数据库句柄。",
			giveOptions: func(t *testing.T) []MySQLOption {
				return []MySQLOption{
					WithNamespace(testNamespace(t, "custom-hook")),
					WithHookManager(kitdriver.NewHookManager()),
				}
			},
			wantMaxOpenConns: defaultPoolMaxOpenConns,
		},
		{
			name:        "success/logging-hooks",
			description: "验证启用错误日志与慢查询阈值时 NewMySQL 可完成钩子配置并返回数据库句柄。",
			giveOptions: func(t *testing.T) []MySQLOption {
				return []MySQLOption{
					WithNamespace(testNamespace(t, "logging-hooks")),
					WithLogError(true),
					WithSlowThreshold(time.Millisecond),
				}
			},
			wantMaxOpenConns: defaultPoolMaxOpenConns,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			db, cleanup, err := NewMySQL(tt.giveOptions(t)...)

			require.NoError(t, err)
			require.NotNil(t, db)
			require.NotNil(t, cleanup)
			assert.Equal(t, tt.wantMaxOpenConns, db.Stats().MaxOpenConnections)

			cleanup()
			assertDatabaseClosed(t, db)
			cleanup()
		})
	}
}

// TestNewMySQL_SlowThresholdInitializesDefaultLogger 验证 NewMySQL 仅启用慢查询阈值时会初始化默认日志记录器。
//
// 该测试通过包内 newLogger 注入点观测默认日志记录器创建路径，确保慢查询钩子不会持有 nil logger 且不依赖真实 MySQL 服务。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewMySQL_SlowThresholdInitializesDefaultLogger(t *testing.T) {
	// 仅配置慢查询阈值时应创建默认 logger，并安全完成慢查询 Hook 构造。
	logger, err := kitlog.NewLogger()
	require.NoError(t, err)

	newLoggerCalls := 0
	restoreNewLogger := replaceNewLogger(t, func() (kitlog.Logger, error) {
		newLoggerCalls++
		return logger, nil
	})
	defer restoreNewLogger()

	db, cleanup, err := NewMySQL(
		WithNamespace(testNamespace(t, "slow-threshold-default-logger")),
		WithSlowThreshold(time.Millisecond),
	)

	require.NoError(t, err)
	require.NotNil(t, db)
	require.NotNil(t, cleanup)
	assert.Equal(t, 1, newLoggerCalls)

	cleanup()
	assertDatabaseClosed(t, db)
}

// TestNewMySQL_ConfigureHooksError 验证 NewMySQL 在自动配置日志 Hook 失败时返回错误。
//
// 该测试通过可恢复的包内函数注入模拟默认日志记录器创建失败，确保构造函数在注册驱动或打开连接前返回错误和空资源。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewMySQL_ConfigureHooksError(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveOptions func(t *testing.T) []MySQLOption
		wantErr     error
	}{
		{
			name:        "error/log-error-new-logger",
			description: "验证启用错误日志且默认日志记录器创建失败时 NewMySQL 返回该错误且不创建数据库句柄。",
			giveOptions: func(t *testing.T) []MySQLOption {
				return []MySQLOption{
					WithNamespace(testNamespace(t, "newmysql-log-error")),
					WithLogError(true),
				}
			},
			wantErr: errors.New("new mysql logger failed for error log"),
		},
		{
			name:        "error/slow-threshold-new-logger",
			description: "验证仅启用慢查询阈值且默认日志记录器创建失败时 NewMySQL 返回该错误且不创建数据库句柄。",
			giveOptions: func(t *testing.T) []MySQLOption {
				return []MySQLOption{
					WithNamespace(testNamespace(t, "newmysql-slow-threshold")),
					WithSlowThreshold(time.Millisecond),
				}
			},
			wantErr: errors.New("new mysql logger failed for slow log"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			restoreNewLogger := replaceNewLogger(t, func() (kitlog.Logger, error) {
				return nil, tt.wantErr
			})
			defer restoreNewLogger()

			db, cleanup, err := NewMySQL(tt.giveOptions(t)...)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, db)
			assert.Nil(t, cleanup)
		})
	}
}

// TestNewMySQL_OpenError 验证 NewMySQL 在 sql.Open 失败时返回错误。
//
// 该测试通过可恢复的包内函数注入模拟 sql.Open 失败，确保错误路径可稳定验证且不会访问真实 MySQL 服务。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewMySQL_OpenError(t *testing.T) {
	logger, err := kitlog.NewLogger()
	require.NoError(t, err)

	tests := []struct {
		name        string
		description string
		giveLogger  kitlog.Logger
		wantErr     error
	}{
		{
			name:        "error/open-without-logger",
			description: "验证 sql.Open 失败且未提供日志记录器时返回原始错误。",
			wantErr:     errors.New("open failed"),
		},
		{
			name:        "error/open-with-logger",
			description: "验证 sql.Open 失败且提供日志记录器时仍返回原始错误。",
			giveLogger:  logger,
			wantErr:     errors.New("open failed with logger"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			restoreSQLOpen := replaceSQLOpen(t, func(driverName string, dataSourceName string) (*sql.DB, error) {
				assert.Contains(t, driverName, "mysql-kit-")
				assert.NotEmpty(t, dataSourceName)
				return nil, tt.wantErr
			})
			defer restoreSQLOpen()

			options := []MySQLOption{WithNamespace(testNamespace(t, "open-error"))}
			if tt.giveLogger != nil {
				options = append(options, WithLogger(tt.giveLogger))
			}

			db, cleanup, err := NewMySQL(options...)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, db)
			assert.Nil(t, cleanup)
		})
	}
}

// TestNewMySQL_CleanupCloseError 验证 NewMySQL 返回的清理函数会调用数据库关闭入口并安全处理关闭失败。
//
// 该测试通过可恢复的包内函数注入模拟 Close 失败，并断言 cleanup 确实触发关闭入口，确保清理失败路径可观测且不会依赖真实数据库服务。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewMySQL_CleanupCloseError(t *testing.T) {
	logger, err := kitlog.NewLogger()
	require.NoError(t, err)

	tests := []struct {
		name        string
		description string
		giveLogger  kitlog.Logger
	}{
		{
			name:        "success/close-error-without-logger",
			description: "验证清理函数关闭数据库失败且未提供日志记录器时仍可安全返回。",
		},
		{
			name:        "success/close-error-with-logger",
			description: "验证清理函数关闭数据库失败且提供日志记录器时仍可安全返回。",
			giveLogger:  logger,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			closeCalls := 0
			closeErr := errors.New("close failed")
			restoreCloseDB := replaceCloseDB(t, func(db *sql.DB) error {
				closeCalls++
				assert.NotNil(t, db)
				return closeErr
			})
			defer restoreCloseDB()

			options := []MySQLOption{WithNamespace(testNamespace(t, "cleanup-close-error"))}
			if tt.giveLogger != nil {
				options = append(options, WithLogger(tt.giveLogger))
			}

			db, cleanup, err := NewMySQL(options...)
			require.NoError(t, err)
			require.NotNil(t, db)
			require.NotNil(t, cleanup)

			cleanup()

			assert.Equal(t, 1, closeCalls)
			require.NoError(t, db.Close())
		})
	}
}

// TestConfigureHooks_NewLoggerError 验证 configureHooks 在默认日志记录器创建失败时返回错误。
//
// 该测试通过可恢复的包内函数注入模拟日志初始化失败，确保错误日志和慢查询日志钩子配置不会静默吞掉初始化错误。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestConfigureHooks_NewLoggerError(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveOptions MySQLOptions
		wantErr     error
	}{
		{
			name:        "error/log-error-new-logger",
			description: "验证启用错误日志但默认日志记录器创建失败时返回该错误。",
			giveOptions: MySQLOptions{logError: true},
			wantErr:     errors.New("new logger failed for error log"),
		},
		{
			name:        "error/slow-threshold-new-logger",
			description: "验证仅启用慢查询阈值但默认日志记录器创建失败时返回该错误，避免注册 nil logger 慢查询钩子。",
			giveOptions: MySQLOptions{slowThreshold: time.Millisecond},
			wantErr:     errors.New("new logger failed for slow log"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			restoreNewLogger := replaceNewLogger(t, func() (kitlog.Logger, error) {
				return nil, tt.wantErr
			})
			defer restoreNewLogger()

			hook := kitdriver.NewHookManager()
			options := tt.giveOptions
			options.namespace = testNamespace(t, "new-logger-error")

			err := configureHooks(hook, &options)

			require.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, options.logger)
			assert.Equal(t, 0, countHookManagerHooks(t, hook))
		})
	}
}

// TestNewMySQL_InvalidDSN 验证 NewMySQL 在数据源名称无效时返回错误。
//
// 该测试使用 go-sql-driver/mysql 在 sql.Open 阶段即可识别的非法 DSN，确保错误分支不依赖真实 MySQL 服务或外部网络。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewMySQL_InvalidDSN(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveOptions []MySQLOption
		wantErrText string
	}{
		{
			name:        "error/invalid-dsn",
			description: "验证 DSN 格式非法时返回解析错误并且不创建数据库句柄或清理函数。",
			giveOptions: []MySQLOption{
				WithNamespace(testNamespace(t, "invalid-dsn")),
				WithDSN("%"),
			},
			wantErrText: "invalid DSN",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			db, cleanup, err := NewMySQL(tt.giveOptions...)

			require.Error(t, err)
			assert.Nil(t, db)
			assert.Nil(t, cleanup)
			assert.ErrorContains(t, err, tt.wantErrText)
		})
	}
}

// TestNewMySQL_RegisteredDriverReuse 验证 NewMySQL 在命名空间对应驱动已注册时复用既有注册项。
//
// 该测试在同一命名空间下连续创建数据库句柄，确保重复调用不会再次注册驱动或影响返回的数据库句柄生命周期。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewMySQL_RegisteredDriverReuse(t *testing.T) {
	// 使用同一命名空间触发第二次调用的驱动复用分支，且不访问真实 MySQL 服务。
	namespace := testNamespace(t, "reuse")
	driverName := fmt.Sprintf("mysql-kit-%s", namespace)

	firstDB, firstCleanup, err := NewMySQL(WithNamespace(namespace), WithPoolMaxOpenConns(11))
	require.NoError(t, err)
	require.NotNil(t, firstDB)
	require.NotNil(t, firstCleanup)
	assert.Contains(t, sql.Drivers(), driverName)
	assert.Equal(t, 11, firstDB.Stats().MaxOpenConnections)

	secondDB, secondCleanup, err := NewMySQL(WithNamespace(namespace), WithPoolMaxOpenConns(13))
	require.NoError(t, err)
	require.NotNil(t, secondDB)
	require.NotNil(t, secondCleanup)
	assert.Equal(t, 13, secondDB.Stats().MaxOpenConnections)

	firstCleanup()
	secondCleanup()
	assertDatabaseClosed(t, firstDB)
	assertDatabaseClosed(t, secondDB)
}

// testNamespace 构造当前测试进程内稳定唯一的 MySQL 驱动命名空间。
//
// 该辅助函数将测试名与语义后缀组合，并替换不利于诊断的分隔符，避免全局 sql 驱动注册表在不同用例间发生命名冲突。
//
// 参数：
//   - t: 测试上下文，用于标记辅助函数调用栈并读取当前测试名。
//   - suffix: 命名空间的语义后缀，用于区分同一测试内的不同驱动配置。
//
// 返回：
//   - string: 可用于 WithNamespace 的稳定命名空间。
func testNamespace(t *testing.T, suffix string) string {
	t.Helper()

	name := strings.NewReplacer("/", "-", " ", "-", "_", "-").Replace(t.Name())
	return strings.ToLower(name + "-" + suffix)
}

// assertDatabaseClosed 验证数据库句柄已经被关闭。
//
// 该辅助函数通过关闭后的 Ping 行为确认 cleanup 生效；由于 database/sql 在关闭状态下会直接返回错误，因此不会访问外部数据库服务。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败并标记辅助函数调用栈。
//   - db: 需要验证关闭状态的数据库句柄。
func assertDatabaseClosed(t *testing.T, db *sql.DB) {
	t.Helper()

	err := db.Ping()
	require.Error(t, err)
	assert.ErrorContains(t, err, "database is closed")
}

// countHookManagerHooks 返回钩子管理器中已注册钩子的数量。
//
// 该辅助函数仅用于验证 mysql 包的 configureHooks 合约：启用日志能力时应向 HookManager 注册钩子。
//
// 参数：
//   - t: 测试上下文，用于报告无法识别钩子管理器结构时的断言失败。
//   - hook: 需要统计已注册钩子数量的钩子管理器。
//
// 返回：
//   - int: 钩子管理器中已注册钩子的数量。
func countHookManagerHooks(t *testing.T, hook *kitdriver.HookManager) int {
	t.Helper()

	manager := fmt.Sprintf("%#v", hook)
	return strings.Count(manager, "HookLog")
}

// replaceSQLOpen 临时替换 mysql 包内部的 sql.Open 调用入口。
//
// 该辅助函数用于稳定模拟 database/sql 打开数据库失败的分支，并通过返回的恢复函数保障全局状态隔离。
//
// 参数：
//   - t: 测试上下文，用于标记辅助函数调用栈并注册恢复逻辑。
//   - replacement: 测试期间替代 sql.Open 的函数。
//
// 返回：
//   - func(): 恢复原始 sql.Open 调用入口的函数。
func replaceSQLOpen(t *testing.T, replacement func(driverName string, dataSourceName string) (*sql.DB, error)) func() {
	t.Helper()

	original := sqlOpen
	sqlOpen = replacement
	restored := false
	restore := func() {
		if restored {
			return
		}
		sqlOpen = original
		restored = true
	}
	t.Cleanup(restore)
	return restore
}

// replaceCloseDB 临时替换 mysql 包内部的数据库关闭入口。
//
// 该辅助函数用于稳定模拟 cleanup 关闭数据库失败的分支，并通过返回的恢复函数保障全局状态隔离。
//
// 参数：
//   - t: 测试上下文，用于标记辅助函数调用栈并注册恢复逻辑。
//   - replacement: 测试期间替代关闭数据库的函数。
//
// 返回：
//   - func(): 恢复原始数据库关闭入口的函数。
func replaceCloseDB(t *testing.T, replacement func(db *sql.DB) error) func() {
	t.Helper()

	original := closeDB
	closeDB = replacement
	restored := false
	restore := func() {
		if restored {
			return
		}
		closeDB = original
		restored = true
	}
	t.Cleanup(restore)
	return restore
}

// replaceNewLogger 临时替换 mysql 包内部的默认日志记录器构造入口。
//
// 该辅助函数用于稳定模拟日志初始化失败的分支，并通过返回的恢复函数保障全局状态隔离。
//
// 参数：
//   - t: 测试上下文，用于标记辅助函数调用栈并注册恢复逻辑。
//   - replacement: 测试期间替代默认日志记录器构造的函数。
//
// 返回：
//   - func(): 恢复原始日志记录器构造入口的函数。
func replaceNewLogger(t *testing.T, replacement func() (kitlog.Logger, error)) func() {
	t.Helper()

	original := newLogger
	newLogger = replacement
	restored := false
	restore := func() {
		if restored {
			return
		}
		newLogger = original
		restored = true
	}
	t.Cleanup(restore)
	return restore
}
