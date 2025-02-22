# 日志包使用示例

这个目录包含了日志包的使用示例，展示了各种日志功能的使用方法。

## 功能特性

- 支持多种日志后端（标准输出、Logrus）
- 提供统一的日志接口
- 支持结构化日志记录
- 支持多个日志级别
- 支持文件和标准输出
- 支持函数式配置选项

## 使用方法

### 1. 使用默认配置

```go
// 使用默认配置初始化日志
if err := log.InitLogger(); err != nil {
    panic(err)
}

// 记录不同级别的日志
log.Debug("这是一条调试日志")
log.Info("这是一条信息日志")
log.Warn("这是一条警告日志")
log.Error("这是一条错误日志")
```

### 2. 使用自定义配置

```go
// 使用自定义配置初始化日志
if err := log.InitLogger(
    log.WithLogType(log.LogTypeLogrus),
    log.WithOutput("/var/log/app.log"),
    log.WithLevel(log.DebugLevel),
); err != nil {
    panic(err)
}
```

### 3. 创建独立的日志实例

```go
// 创建独立的日志实例
logger, err := log.NewLogger(
    log.WithLogType(log.LogTypeStd),
    log.WithLevel(log.DebugLevel),
)
if err != nil {
    panic(err)
}

// 使用独立的日志实例
logger.Info("使用独立的日志实例")
```

### 4. 结构化日志

```go
// 添加单个字段
log.WithField("user", "admin").Info("用户登录")

// 添加多个字段
log.WithFields(map[string]interface{}{
    "ip":      "192.168.1.1",
    "method":  "POST",
    "latency": "20ms",
}).Info("收到HTTP请求")
```

### 5. 格式化日志

```go
log.Debugf("当前时间是: %v", time.Now())
log.Infof("用户 %s 执行了 %s 操作", "admin", "登录")
```

## 配置选项

日志包提供了以下配置选项：

- `WithLogType(LogType)`: 设置日志类型
  - `LogTypeConsole`: 控制台日志
  - `LogTypeStd`: 标准库日志
  - `LogTypeLogrus`: Logrus 日志

- `WithLevel(Level)`: 设置日志级别
  - `DebugLevel`: 调试级别
  - `InfoLevel`: 信息级别
  - `WarnLevel`: 警告级别
  - `ErrorLevel`: 错误级别
  - `FatalLevel`: 致命错误级别

- `WithOutput(string)`: 设置日志输出路径
  - 空字符串: 输出到标准输出
  - 文件路径: 输出到指定文件

## 运行示例

```bash
go run main.go
```

示例程序会依次展示：
1. 默认配置的使用
2. 不同日志级别的输出
3. 结构化日志的使用
4. 自定义配置的应用
5. 独立日志实例的创建和使用

## 注意事项

1. Fatal 级别的日志会导致程序退出
2. 文件日志需要确保目录存在且有写入权限
3. 默认使用 InfoLevel 日志级别
4. 默认使用标准输出作为日志输出

## 示例输出

运行示例程序后，你将在标准输出中看到类似下面的输出：

```
2025/02/24 20:03:42 [DEBUG] 这是一条调试日志
2025/02/24 20:03:42 [INFO] 这是一条信息日志
2025/02/24 20:03:42 [WARN] 这是一条警告日志
2025/02/24 20:03:42 [ERROR] 这是一条错误日志
2025/02/24 20:03:42 [DEBUG] 当前时间是: 2025-02-24 20:03:42
2025/02/24 20:03:42 [INFO] 程序运行在: /path/to/your/workspace
2025/02/24 20:03:42 [INFO] [user=admin] 用户登录
2025/02/24 20:03:42 [INFO] [latency=20ms ip=192.168.1.1 method=POST] 收到HTTP请求
2025/02/24 20:03:42 [ERROR] [error=示例错误] 操作失败
```

同时，在 `example/log/app.log` 文件中，你将看到 logrus 格式的日志：

```
time="2025-02-24 20:03:42" level=info msg="已切换到 logrus 日志器"
time="2025-02-24 20:03:42" level=info msg="服务器启动" component=server status=starting
```

## 说明

1. 标准输出日志器
   - 使用 `log.InitLogger(log.WithLogType(log.LogTypeConsole))` 初始化
   - 日志直接输出到标准输出
   - 适合开发环境使用
   - 支持结构化字段，以 `[key=value]` 的格式显示

2. Logrus 日志器
   - 使用 `log.InitLogger(log.WithLogType(log.LogTypeLogrus), log.WithOutput("path/to/file.log"))` 初始化
   - 日志输出到指定的文件
   - 使用 logrus 的默认格式输出
   - 支持结构化字段，以 key=value 的格式显示
   - 适合生产环境使用

3. 日志级别
   - Debug：调试信息
   - Info：一般信息
   - Warn：警告信息
   - Error：错误信息
   - Fatal：致命错误（会导致程序退出）

4. 结构化日志
   - 使用 `WithField` 添加单个字段
   - 使用 `WithFields` 添加多个字段
   - 字段以键值对的形式展示在日志中
   - 不同的日志器可能有不同的字段展示格式 