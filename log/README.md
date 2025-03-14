# log

## 简介

`log` 包提供了一个统一的日志接口和多种日志实现，支持结构化日志记录、多种输出格式和灵活的配置选项。该包设计用于满足不同场景下的日志需求，从简单的控制台输出到复杂的生产环境日志记录都能胜任。

### 主要特性

- 统一的日志接口，支持多种日志后端（标准库、Logrus）
- 支持结构化日志记录，方便日志分析和处理
- 支持多个日志级别（Debug、Info、Warn、Error、Fatal）
- 支持文件输出和标准输出
- 支持日志文件自动滚动和保留期限设置
- 支持 JSON 和文本两种输出格式
- 支持字段注入和链式调用
- 线程安全的全局日志实例管理
- 完整的单元测试覆盖

### 设计理念

该包的设计遵循以下原则：

1. **统一接口**：通过 `Logger` 接口定义统一的日志行为，使不同的日志实现可以无缝切换。

2. **可扩展性**：采用接口设计和函数式配置，方便扩展新的日志实现和配置选项。

3. **易用性**：提供简单直观的 API，支持链式调用，使日志记录更加便捷。

4. **灵活性**：支持多种日志格式和输出方式，满足不同场景的需求。

5. **性能**：通过级别控制和格式优化，确保日志记录的性能开销最小化。

## 安装

### 前置条件

- Go 版本要求：>= 1.16
- 依赖要求：
  - github.com/sirupsen/logrus v1.8.1
  - github.com/lestrrat-go/file-rotatelogs v2.4.0

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/log
```

## 快速开始

### 基础用法

```go
// 使用默认配置初始化日志
if err := log.InitLogger(); err != nil {
    panic(err)
}

// 记录不同级别的日志
log.Info("应用启动")
log.WithField("user", "admin").Info("用户登录")
log.Error("发生错误")
```

### 配置选项

```go
// 使用自定义配置初始化日志
if err := log.InitLogger(
    log.WithLogType(log.LogTypeLogrus),
    log.WithLevel(log.DebugLevel),
    log.WithOutput("/var/log/app.log"),
    log.WithEnableRotate(true),
    log.WithRotateTime(time.Hour * 24),
    log.WithMaxAge(time.Hour * 24 * 7),
); err != nil {
    panic(err)
}
```

## 详细指南

### 核心概念

1. **日志级别**：日志分为 Debug、Info、Warn、Error、Fatal 五个级别，级别越高表示日志越重要。

2. **结构化字段**：支持添加键值对形式的结构化信息，方便日志分析。

3. **日志滚动**：支持按时间自动滚动日志文件，并可设置日志保留时间。

### 常见用例

#### 1. 使用结构化字段记录日志

```go
logger.WithFields(map[string]interface{}{
    "user_id": "12345",
    "action": "login",
    "status": "success",
}).Info("用户登录成功")
```

#### 2. 配置日志滚动

```go
if err := log.InitLogger(
    log.WithLogType(log.LogTypeLogrus),
    log.WithOutput("/var/log/app.log"),
    log.WithEnableRotate(true),
    log.WithRotateTime(time.Hour * 24),  // 每天滚动
    log.WithMaxAge(time.Hour * 24 * 7),  // 保留7天
); err != nil {
    panic(err)
}
```

### 最佳实践

- 合理设置日志级别，开发环境可使用 Debug 级别，生产环境建议使用 Info 级别
- 使用结构化字段记录关键信息，方便后续分析
- 在生产环境中启用日志滚动，防止日志文件过大
- 使用全局日志实例时注意并发安全
- 错误日志应包含足够的上下文信息

## API 文档

### 主要类型

```go
// Logger 接口定义了所有日志操作
type Logger interface {
    SetLevel(level Level)
    GetLevel() Level
    Debug(args ...interface{})
    Info(args ...interface{})
    Warn(args ...interface{})
    Error(args ...interface{})
    Fatal(args ...interface{})
    WithField(key string, value interface{}) Logger
    WithFields(fields map[string]interface{}) Logger
}
```

### 关键函数

#### InitLogger

初始化全局日志实例。

```go
func InitLogger(options ...Option) error
```

示例：
```go
err := log.InitLogger(
    log.WithLogType(log.LogTypeLogrus),
    log.WithLevel(log.InfoLevel),
)
```

#### NewLogger

创建新的日志实例。

```go
func NewLogger(options ...Option) (Logger, error)
```

示例：
```go
logger, err := log.NewLogger(
    log.WithLogType(log.LogTypeStd),
    log.WithOutput("app.log"),
)
```

### 错误处理

- 所有可能失败的操作都会返回 error
- 日志初始化失败会返回具体的错误原因
- Fatal 级别的日志会导致程序以状态码 1 退出

## 性能指标

| 操作 | 性能指标 | 说明 |
|------|----------|------|
| 普通日志写入 | O(1) | 内存中的操作，性能开销很小 |
| 文件日志写入 | O(1) | 取决于系统 IO 性能 |
| 结构化字段 | O(n) | n 为字段数量 |

## 测试覆盖率

| 包 | 覆盖率 |
|------|--------|
| log | >90% |

## 调试指南

### 日志级别

- Debug: 调试信息，用于开发和问题诊断
- Info: 普通信息，记录正常的操作流程
- Warn: 警告信息，表示可能的问题
- Error: 错误信息，表示操作失败
- Fatal: 致命错误，记录后程序会退出

### 常见问题排查

#### 日志文件无法创建

- 检查目录权限
- 确保目录路径存在
- 验证文件名格式是否正确

#### 日志级别不正确

- 检查 InitLogger 时的级别设置
- 确认是否调用了 SetLevel 修改了级别
- 验证日志调用使用了正确的方法

## 相关文档

- [Logrus 文档](https://github.com/sirupsen/logrus)
- [file-rotatelogs 文档](https://github.com/lestrrat-go/file-rotatelogs)
- [示例代码](../example/log)

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考我们的[贡献指南](../CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT 许可证。查看 [LICENSE](./LICENSE) 文件了解更多信息。

## 补充说明

本文的大部分信息，由 AI 使用[模板](../ai/templates/docs/package_readme_template.md)根据[提示词](../ai/prompts/docs/package_readme_generator.md)自动生成，如有任何问题，请随时联系我。