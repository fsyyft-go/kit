 # mysql

## 简介

mysql 包提供了一个增强的 MySQL 数据库连接管理工具，基于 Go 标准库的 `database/sql` 和 `go-sql-driver/mysql` 驱动。该包通过函数式选项模式提供灵活的配置，并支持连接池管理、错误日志记录和慢查询监控等功能。

### 主要特性

- 基于标准库 `database/sql` 的 MySQL 连接管理
- 支持连接池配置和优化
- 内置错误日志记录功能
- 支持慢查询监控和日志记录
- 命名空间隔离的连接管理
- 函数式选项的配置方式
- 支持自定义日志记录器
- 连接生命周期管理

### 设计理念

该包的设计遵循以下原则：

1. **简单易用**：提供简洁的 API，使数据库连接管理变得简单
2. **灵活配置**：通过函数式选项模式支持灵活的配置
3. **性能优化**：内置连接池管理，优化数据库连接性能
4. **可观测性**：支持错误日志和慢查询监控
5. **安全性**：支持命名空间隔离，避免连接混用

## 安装

### 前置条件

- Go 版本要求：>= 1.18
- 依赖要求：
  - github.com/go-sql-driver/mysql
  - github.com/fsyyft-go/kit/database/sql/driver
  - github.com/fsyyft-go/kit/log

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/database/sql/mysql
```

## 快速开始

### 基础用法

```go
package main

import (
    "github.com/fsyyft-go/kit/database/sql/mysql"
)

func main() {
    // 创建数据库连接
    db, cleanup, err := mysql.NewMySQL(
        mysql.WithDSN("user:password@tcp(localhost:3306)/dbname"),
        mysql.WithPoolMaxOpenConns(100),
        mysql.WithPoolMaxIdleConns(10),
    )
    if err != nil {
        panic(err)
    }
    defer cleanup()

    // 使用数据库连接
    // db.Query()...
}
```

### 配置选项

```go
// 完整配置示例
db, cleanup, err := mysql.NewMySQL(
    // 设置数据源名称
    mysql.WithDSN("user:password@tcp(localhost:3306)/dbname"),
    // 设置连接池参数
    mysql.WithPoolIdleTime(10 * time.Second),
    mysql.WithPoolMaxIdleTime(10 * time.Second),
    mysql.WithPoolMaxOpenConns(100),
    mysql.WithPoolMaxIdleConns(10),
    // 设置命名空间
    mysql.WithNamespace("app"),
    // 配置日志
    mysql.WithLogger(logger),
    mysql.WithLogError(true),
    // 设置慢查询监控
    mysql.WithSlowThreshold(200 * time.Millisecond),
)
```

## 详细指南

### 核心概念

1. **连接池管理**
   - 控制最大连接数
   - 管理空闲连接
   - 优化连接复用

2. **日志记录**
   - 错误日志记录
   - 慢查询监控
   - 自定义日志格式

3. **命名空间**
   - 隔离不同的数据库连接
   - 避免配置混淆
   - 支持多数据库管理

### 常见用例

#### 1. 基本数据库连接

```go
db, cleanup, err := mysql.NewMySQL(
    mysql.WithDSN("user:password@tcp(localhost:3306)/dbname"),
)
if err != nil {
    panic(err)
}
defer cleanup()
```

#### 2. 配置连接池

```go
db, cleanup, err := mysql.NewMySQL(
    mysql.WithDSN("user:password@tcp(localhost:3306)/dbname"),
    mysql.WithPoolMaxOpenConns(100),
    mysql.WithPoolMaxIdleConns(10),
    mysql.WithPoolIdleTime(10 * time.Second),
    mysql.WithPoolMaxIdleTime(10 * time.Second),
)
```

#### 3. 启用错误日志和慢查询监控

```go
logger, _ := kitlog.NewLogger()
db, cleanup, err := mysql.NewMySQL(
    mysql.WithDSN("user:password@tcp(localhost:3306)/dbname"),
    mysql.WithLogger(logger),
    mysql.WithLogError(true),
    mysql.WithSlowThreshold(200 * time.Millisecond),
)
```

### 最佳实践

- 合理配置连接池参数
  - 根据应用负载设置最大连接数
  - 适当配置空闲连接数和超时时间
  - 避免连接资源浪费

- 使用命名空间隔离连接
  - 为不同的业务模块使用不同的命名空间
  - 避免配置混淆
  - 便于问题定位

- 配置监控和日志
  - 启用错误日志记录
  - 设置合适的慢查询阈值
  - 定期检查日志

## API 文档

### 主要类型

#### MySQLOptions

```go
type MySQLOptions struct {
    dns              string          // 数据源名称
    poolIdleTime     time.Duration  // 连接空闲超时时间
    poolMaxIdleTime  time.Duration  // 最大空闲时间
    poolMaxOpenConns int            // 最大打开连接数
    poolMaxIdleConns int            // 最大空闲连接数
    hook            *HookManager    // 钩子管理器
    namespace       string          // 命名空间
    logger          kitlog.Logger   // 日志记录器
    logError        bool           // 是否记录错误
    slowThreshold   time.Duration  // 慢查询阈值
}
```

### 关键函数

#### NewMySQL

创建新的 MySQL 数据库连接。

```go
func NewMySQL(opts ...MySQLOption) (*sql.DB, func(), error)
```

#### 配置选项函数

- WithDSN：设置数据源名称
- WithPoolIdleTime：设置连接空闲超时时间
- WithPoolMaxIdleTime：设置最大空闲时间
- WithPoolMaxOpenConns：设置最大打开连接数
- WithPoolMaxIdleConns：设置最大空闲连接数
- WithNamespace：设置命名空间
- WithLogger：设置日志记录器
- WithLogError：设置是否记录错误
- WithSlowThreshold：设置慢查询阈值

### 默认值

```go
defaultDSN = "test:test@tcp(localhost:3306)/test"
defaultPoolIdleTime = 10 * time.Second
defaultPoolMaxIdleTime = 10 * time.Second
defaultPoolMaxOpenConns = 100
defaultPoolMaxIdleConns = 10
```

## 性能指标

| 配置项 | 默认值 | 建议范围 | 说明 |
|--------|---------|----------|------|
| MaxOpenConns | 100 | 50-500 | 根据服务器配置调整 |
| MaxIdleConns | 10 | 10-50 | 通常为 MaxOpenConns 的 10-20% |
| IdleTime | 10s | 10s-60s | 根据业务峰值调整 |
| MaxIdleTime | 10s | 10s-300s | 避免空闲连接占用资源 |

## 调试指南

### 常见问题排查

#### 1. 连接数过多

问题：数据库连接数持续增长
解决方案：
- 检查 MaxOpenConns 设置
- 确保正确调用 cleanup 函数
- 检查是否有连接泄漏

#### 2. 慢查询频繁

问题：出现大量慢查询日志
解决方案：
- 调整 slowThreshold 阈值
- 优化数据库索引
- 检查 SQL 语句性能

#### 3. 连接获取超时

问题：无法获取数据库连接
解决方案：
- 增加 MaxOpenConns 值
- 优化连接池配置
- 检查数据库负载

## 相关文档

- [Go database/sql 文档](https://golang.org/pkg/database/sql/)
- [go-sql-driver/mysql 文档](https://github.com/go-sql-driver/mysql)
- [MySQL 官方文档](https://dev.mysql.com/doc/)

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考我们的[贡献指南](../../../CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT 许可证。查看 [LICENSE](../../../LICENSE) 文件了解更多信息。