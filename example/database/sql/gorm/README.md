# GORM 数据库示例

本示例展示了如何使用 Kit 的 GORM 数据库模块，实现基于 GORM 的数据库操作，包括连接配置、表结构定义和基本的 CRUD 操作。

## 功能特性

- 支持 MySQL 数据库连接和配置
- 支持 GORM 模型定义和表结构映射
- 支持基本的 CRUD 操作（创建、读取、更新、删除）
- 支持自定义日志记录
- 支持慢查询监控
- 支持命名空间隔离

## 设计原理

Kit 的 GORM 数据库示例采用了以下设计：

- 使用 GORM 作为 ORM 框架，提供类型安全的数据库操作
- 通过结构体标签定义表结构和字段映射
- 使用函数式选项模式配置数据库连接
- 集成 Kit 的日志模块进行日志记录
- 支持自定义命名策略和表名

## 使用方法

### 1. 编译和运行

在 Unix/Linux/macOS 系统上：

```bash
# 添加执行权限
chmod +x build.sh

# 构建和运行
./build.sh
```

### 2. 代码示例

#### 数据模型定义

```go
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
func (User) TableName() string {
    return "example_user"
}
```

#### 数据库连接配置

```go
// 设置数据库连接参数
dsnparams := kitmysql.WithDSNParams("test:test@tcp(localhost:3306)/test", map[string]string{
    "interpolateParams":    "true",  // 启用参数插值
    "parseTime":            "true",  // 启用时间解析
    "loc":                  "Local", // 使用本地时区
    "allowNativePasswords": "true",  // 允许原生密码认证
})

// 初始化数据库连接
db, cleanup, err := kitmysql.NewMySQL(
    dsnparams,
    kitmysql.WithNamespace("example.gorm"),       // 设置命名空间
    kitmysql.WithLogger(logger),                  // 设置日志记录器
    kitmysql.WithLogError(true),                  // 启用错误日志记录
    kitmysql.WithSlowThreshold(time.Microsecond), // 设置慢查询阈值
)
defer cleanup()
```

#### GORM 配置和初始化

```go
// 创建 MySQL 配置
mysqlconfig := mysql.Config{
    Conn: db,
}

// 创建 MySQL 方言
mysqldialector := mysql.New(mysqlconfig)

// 配置并初始化 gorm
gormDB, err := gorm.Open(mysqldialector, &gorm.Config{
    NamingStrategy: schema.NamingStrategy{
        SingularTable: true, // 使用单数表名
    },
    Logger: kitgorm.NewLogger(logger), // 使用 kit 包中的 gorm logger 适配器
})
```

#### 基本 CRUD 操作

```go
// 创建用户
user := &User{
    Name: "张大山",
    Age:  25,
}
if err := gormDB.Create(user).Error; err != nil {
    logger.Error(fmt.Errorf("创建用户失败：%w", err))
    return
}

// 查询单个用户
var queryUser User
if err := gormDB.First(&queryUser, user.ID).Error; err != nil {
    logger.Error(fmt.Errorf("查询用户失败：%w", err))
    return
}

// 更新用户年龄
if err := gormDB.Model(&queryUser).Update("age", 30).Error; err != nil {
    logger.Error(fmt.Errorf("更新用户年龄失败：%w", err))
    return
}

// 查询所有用户
var users []User
if err := gormDB.Find(&users).Error; err != nil {
    logger.Error(fmt.Errorf("查询所有用户失败：%w", err))
    return
}

// 删除用户
if err := gormDB.Delete(&user).Error; err != nil {
    logger.Error(fmt.Errorf("删除用户失败：%w", err))
    return
}
```

### 3. 输出示例

```
[INFO] [module=gorm] 创建用户成功，ID：1
[INFO] [module=gorm] 查询用户成功：{ID:1 Name:张大山 Age:25}
[INFO] [module=gorm] 更新用户年龄成功
[INFO] [module=gorm] 查询所有用户成功，共 1 条记录
[INFO] [module=gorm] 删除用户成功
```

### 4. 在其他项目中使用

```go
package main

import (
    "github.com/fsyyft-go/kit/database/sql/gorm"
    "github.com/fsyyft-go/kit/database/sql/mysql"
    "github.com/fsyyft-go/kit/log"
)

func main() {
    // 初始化日志
    logger, err := log.NewLogger(log.WithLevel(log.DebugLevel))
    if err != nil {
        panic(err)
    }

    // 配置数据库连接
    db, cleanup, err := mysql.NewMySQL(
        mysql.WithDSNParams("user:pass@tcp(localhost:3306)/dbname", nil),
        mysql.WithLogger(logger),
    )
    if err != nil {
        panic(err)
    }
    defer cleanup()

    // 初始化 GORM
    gormDB, err := gorm.NewGORM(db, logger)
    if err != nil {
        panic(err)
    }

    // 使用 GORM 进行数据库操作
    // ...
}
```

## 注意事项

- 确保数据库连接参数正确，包括用户名、密码、主机地址和数据库名
- 使用 defer cleanup() 确保数据库连接正确关闭
- 使用结构体标签正确定义表结构和字段映射
- 在生产环境中，建议使用环境变量或配置文件管理数据库连接信息
- 注意处理数据库操作的错误，并进行适当的日志记录
- 使用事务确保数据一致性，特别是在需要多个操作时

## 相关文档

- [GORM 官方文档](https://gorm.io/docs/)
- [Kit MySQL 模块文档](../../../../database/sql/mysql/README.md)
- [Kit 日志模块文档](../../../../log/README.md)

## 许可证

本示例代码采用 MIT 许可证。详见 [LICENSE](../../../../LICENSE) 文件。 