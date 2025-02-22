# Log 包

`log` 包提供了一个统一的日志接口和多种日志实现，支持结构化日志记录和多个日志级别。

## 特性

- 统一的日志接口
- 多种日志后端支持（标准输出、Logrus）
- 结构化日志记录
- 多个日志级别（Debug、Info、Warn、Error、Fatal）
- 支持文件和标准输出
- 支持字段注入和上下文
- 支持日志滚动（按时间自动切分）
- 支持多种输出格式（行日志、JSON）
- 支持编译时配置（如时间格式等）

## 快速开始

### 基本使用

```go
package main

import "github.com/fsyyft-go/kit/log"

func main() {
    // 初始化标准输出日志
    if err := log.InitLogger(log.WithLogType(log.LogTypeStd)); err != nil {
        panic(err)
    }

    // 记录不同级别的日志
    log.Debug("这是一条调试日志")
    log.Info("这是一条信息日志")
    log.Warn("这是一条警告日志")
    log.Error("这是一条错误日志")
}
```

### 结构化日志

```go
// 添加单个字段
log.WithField("user", "admin").Info("用户登录")

// 添加多个字段
log.WithFields(map[string]interface{}{
    "user": "admin",
    "ip":   "192.168.1.1",
}).Info("用户登录")
```

### 使用 Logrus 后端

Logrus 后端默认使用 JSON 格式输出，可以通过配置更改为行日志格式：

```go
// 初始化 Logrus 日志（默认 JSON 格式）
if err := log.InitLogger(
    log.WithLogType(log.LogTypeLogrus),
    log.WithOutput("/path/to/log/file.log"),
); err != nil {
    panic(err)
}

// 使用行日志格式
if err := log.InitLogger(
    log.WithLogType(log.LogTypeLogrus),
    log.WithOutput("/path/to/log/file.log"),
    log.WithFormat(log.FormatText),
); err != nil {
    panic(err)
}
```

注意：日志格式（JSON/行日志）和日志滚动功能仅在使用 Logrus 后端时可用。

### 编译时配置

在使用 Logrus 后端时，以下配置项可以在编译时通过 `-X` 链接标志指定：

```go
package log

var (
    // timestampFormat 定义时间戳格式
    timestampFormat = "2006-01-02 15:04:05"
    
    // disableColors 控制是否禁用颜色输出
    disableColors = false
    
    // fullTimestamp 控制是否使用完整时间戳
    fullTimestamp = true
    
    // prettyPrint 控制是否美化输出（仅在 JSON 格式下有效）
    prettyPrint = false
)
```

可以在编译时使用 `-X` 标志设置这些变量的值：

```bash
# 设置时间格式
go build -ldflags "-X 'github.com/fsyyft-go/kit/log.timestampFormat=2006/01/02-15:04:05'"

# 设置多个变量
go build -ldflags "
    -X 'github.com/fsyyft-go/kit/log.timestampFormat=2006/01/02-15:04:05'
    -X 'github.com/fsyyft-go/kit/log.disableColors=true'
    -X 'github.com/fsyyft-go/kit/log.fullTimestamp=true'
    -X 'github.com/fsyyft-go/kit/log.prettyPrint=true'
"
```

注意：这些配置项仅在使用 Logrus 后端时生效。其中：
- timestampFormat：定义日志中时间戳的格式
- disableColors：设置为 true 时禁用终端颜色输出
- fullTimestamp：设置为 true 时使用完整时间戳，false 时使用相对时间
- prettyPrint：设置为 true 时美化 JSON 输出格式（仅在 JSON 格式下有效）

### 使用日志滚动功能

日志滚动功能默认启用，配置如下：
- 默认每小时滚动一次
- 默认保留7天的日志
- 自动创建软链接到最新日志文件

日志文件命名规则：
- 原始文件名：`app.log`
- 滚动后的文件名：`app-2024031510.log`（表示2024年3月15日10点的日志）
- 软链接：始终保持原始文件名（`app.log`），指向最新的日志文件

如果需要自定义配置，可以使用以下选项：

```go
// 自定义日志滚动配置
if err := log.InitLogger(
    log.WithLogType(log.LogTypeLogrus),
    log.WithOutput("/path/to/log/app.log"),
    log.WithLevel(log.InfoLevel),
    log.WithRotateTime(time.Minute * 30),    // 每30分钟滚动一次
    log.WithMaxAge(time.Hour*24*30),         // 保留30天的日志
); err != nil {
    panic(err)
}
```

如果需要禁用日志滚动功能：
```go
if err := log.InitLogger(
    log.WithLogType(log.LogTypeLogrus),
    log.WithOutput("/path/to/log/app.log"),
    log.WithEnableRotate(false),             // 禁用日志滚动
); err != nil {
    panic(err)
}
```

日志滚动功能特性：
- 支持按时间自动切分日志文件（默认每小时一个）
- 可配置日志滚动时间间隔
- 可配置日志文件保留时间（默认7天）
- 自动清理过期日志文件
- 支持软链接到最新日志文件

注意：日志滚动功能仅在使用 Logrus 后端时可用。

## 日志级别

- `Debug`: 调试信息，用于开发环境
- `Info`: 一般信息，用于记录正常操作
- `Warn`: 警告信息，表示可能的问题
- `Error`: 错误信息，表示操作失败
- `Fatal`: 致命错误，记录后程序会退出

## 配置示例

### 设置日志级别

```go
logger := log.GetLogger()
logger.SetLevel(log.InfoLevel) // 只记录 Info 及以上级别的日志
```

### 获取当前日志级别

```go
level := logger.GetLevel()
```

## 更多示例

更多使用示例请参考 [example/log](../example/log) 目录，其中包含了：
- 不同输出格式（JSON/行日志）的使用
- 编译时配置的示例
- 日志滚动的完整示例
- 各种日志级别的使用方式

## 最佳实践

1. 合理使用日志级别
   - Debug: 仅在开发环境使用
   - Info: 记录重要的业务操作
   - Warn: 记录潜在问题
   - Error: 记录错误但不影响系统运行
   - Fatal: 仅在系统无法继续运行时使用

2. 结构化日志
   - 使用 WithField/WithFields 添加上下文信息
   - 保持字段名称的一致性
   - 避免在日志中包含敏感信息

3. 性能考虑
   - 在高性能场景下，使用 Debug 级别前先检查级别
   - 避免在热点代码路径中过度记录日志
   - JSON 格式输出可能比行日志格式有更多的性能开销

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 Apache License 2.0 许可证。详见 [LICENSE](../LICENSE) 文件。 