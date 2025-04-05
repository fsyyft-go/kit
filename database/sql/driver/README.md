# Database SQL Driver Hook

这个包提供了一个 Go 语言数据库驱动的扩展，允许在数据库操作的前后添加钩子（Hook）函数。通过这个扩展，你可以轻松地实现数据库操作的监控、日志记录、性能分析等功能。

## 设计思路

### 核心组件

1. **操作类型（OpType）**
   - 定义了所有可能的数据库操作类型
   - 包括：连接、事务、预处理语句、执行、查询等
   - 每个操作类型都有对应的字符串表示

2. **钩子上下文（HookContext）**
   - 包含操作的完整上下文信息
   - 原始 Context
   - 操作类型
   - SQL 语句
   - 参数列表
   - 开始时间
   - 结束时间
   - 操作结果
   - 错误信息
   - 自定义数据存储（hookMap）

3. **钩子接口（Hook）**
   - Before：操作执行前调用
   - After：操作执行后调用
   - 可以访问完整的操作上下文

4. **钩子管理器（HookManager）**
   - 管理多个钩子的执行
   - 支持添加多个钩子
   - Before 按正序执行
   - After 按倒序执行

5. **驱动包装器（KitDriver）**
   - 包装原始数据库驱动
   - 在所有操作前后调用钩子
   - 保持与原始驱动完全兼容

### 工作流程

1. 创建钩子实例
2. 创建钩子管理器
3. 将钩子添加到管理器
4. 使用管理器包装原始驱动
5. 注册包装后的驱动
6. 正常使用数据库，钩子会自动执行

## 使用方法

### 基本用法

```go
import (
    "database/sql"
    "github.com/fsyyft-go/kit/database/sql/driver"
)

// 1. 实现自定义钩子
type MyHook struct{}

func (h *MyHook) Before(ctx *driver.HookContext) error {
    // 操作执行前的逻辑
    return nil
}

func (h *MyHook) After(ctx *driver.HookContext) error {
    // 操作执行后的逻辑
    return nil
}

// 2. 创建并注册包装的驱动
func main() {
    // 创建原始驱动
    originalDriver := &mysql.MySQLDriver{}

    // 创建钩子
    hook := &MyHook{}

    // 创建钩子管理器
    hookManager := driver.NewHookManager()
    hookManager.AddHook(hook)

    // 创建包装的驱动
    wrappedDriver := driver.NewKitDriver(originalDriver, hookManager)

    // 注册驱动
    sql.Register("mysql-with-hook", wrappedDriver)

    // 使用包装的驱动
    db, err := sql.Open("mysql-with-hook", "user:password@/dbname")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // 正常使用数据库
    // 所有操作都会触发钩子
}
```

### 示例：性能监控钩子

```go
type TimingHook struct{}

func (h *TimingHook) Before(ctx *driver.HookContext) error {
    log.Printf("[%s] 开始执行，SQL: %s，参数: %v", 
        ctx.OpType(), ctx.Query(), ctx.Args())
    return nil
}

func (h *TimingHook) After(ctx *driver.HookContext) error {
    duration := ctx.EndTime().Sub(ctx.StartTime())
    log.Printf("[%s] 执行完成，耗时: %v，错误: %v", 
        ctx.OpType(), duration, ctx.OriginError())
    return nil
}
```

## 支持的操作类型

- `OpConnect`: 连接数据库
- `OpBegin`: 开始事务
- `OpCommit`: 提交事务
- `OpRollback`: 回滚事务
- `OpPrepare`: 预处理语句
- `OpStmtExec`: 执行预处理语句
- `OpStmtQuery`: 查询预处理语句
- `OpStmtClose`: 关闭预处理语句
- `OpExec`: 执行 SQL
- `OpQuery`: 查询 SQL
- `OpPing`: Ping 操作

## 注意事项

1. **钩子执行顺序**
   - Before 钩子按添加顺序执行
   - After 钩子按添加顺序的反序执行
   - 任何钩子返回错误都会中断执行

2. **上下文数据**
   - HookContext 实现了 context.Context 接口
   - 可以通过 GetHookValue/SetHookValue 在钩子间共享数据
   - 所有时间相关的字段都是只读的

3. **性能考虑**
   - 钩子的执行会增加一定的开销
   - 建议在钩子中避免耗时操作
   - 可以使用 goroutine 处理异步任务

4. **错误处理**
   - Before 钩子的错误会阻止操作执行
   - After 钩子的错误会覆盖操作的错误
   - 建议在 After 钩子中避免返回错误

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

使用 BSD 许可证 - 查看 [LICENSE](../../../LICENSE) 文件了解更多信息。 