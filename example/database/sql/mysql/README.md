# MySQL 数据库示例

本示例展示了如何使用 Kit 的 MySQL 数据库模块，实现基本的数据库 CRUD 操作和事务处理。

## 功能特性

- 支持单行和多行数据查询
- 支持数据插入、更新和删除操作
- 支持事务处理和自动回滚
- 支持结构化的错误处理
- 支持日志记录和监控

## 设计原理

Kit 的 MySQL 数据库示例采用了以下设计：

- 使用结构体定义数据模型
- 函数式编程风格，每个操作独立封装
- 统一的错误处理和日志记录
- 事务操作的安全处理机制
- 资源的自动释放（defer）

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
```

#### 单行数据查询

```go
// queryRow 查询单条数据示例。
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
```

#### 多行数据查询

```go
// queryMultiRow 查询多条数据示例。
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
```

#### 数据插入

```go
// insertRow 插入数据示例。
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
```

#### 数据更新

```go
// updateRow 更新数据示例。
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
```

#### 数据删除

```go
// deleteRow 删除数据示例。
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
```

#### 事务处理

```go
// transactionDemo 事务操作示例。
func transactionDemo(db *sql.DB, logger kitlog.Logger) error {
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

    // 更新第一条记录。
    sqlStr1 := "UPDATE `example_user` SET `age`=30 WHERE `id`=? LIMIT 1;"
    ret1, err := tx.Exec(sqlStr1, 2)
    if err != nil {
        return fmt.Errorf("执行表 `example_user` 第一次更新失败：%w", err)
    }
    // 获取第一次更新影响的行数。
    affRow1, err := ret1.RowsAffected()
    if err != nil {
        return fmt.Errorf("获取第一次更新影响行数失败：%w", err)
    }

    // 更新第二条记录。
    sqlStr2 := "UPDATE `example_user` SET `age`=40 WHERE `id`=? LIMIT 1;"
    ret2, err := tx.Exec(sqlStr2, 3)
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

    return fmt.Errorf("更新影响的行数不正确：第一次=%d，第二次=%d", affRow1, affRow2)
}
```

## 注意事项

- 所有 SQL 语句都使用参数绑定，防止 SQL 注入
- 查询结果集使用 defer 确保正确关闭
- 事务操作使用 defer 确保正确回滚
- 错误处理使用 fmt.Errorf 包装原始错误
- 更新和删除操作使用 LIMIT 1 限制影响范围
- 使用 RowsAffected 检查操作影响的行数

## 相关文档

- [MySQL 官方文档](https://dev.mysql.com/doc/)
- [Go database/sql 包文档](https://golang.org/pkg/database/sql/)
- [Kit 数据库模块文档](../../../../database/README.md)
- [Kit 日志模块文档](../../../../log/README.md)

## 许可证

本示例代码采用 MIT 许可证。详见 [LICENSE](../../../../LICENSE) 文件。