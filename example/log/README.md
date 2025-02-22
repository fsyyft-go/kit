# Log 包使用示例

这个示例展示了如何使用 `fsyyft-go/kit/log` 包进行日志记录。示例包含了以下主要功能的演示：

- 初始化不同类型的日志器（标准输出和 logrus）
- 设置日志级别
- 记录不同级别的日志
- 使用格式化日志
- 使用结构化日志
- 错误处理日志记录

## 运行示例

```bash
# 进入项目根目录
cd $GOPATH/src/github.com/fsyyft-go/kit

# 运行示例程序
go run example/log/main.go
```

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
   - 使用 `log.InitLogger(log.LogTypeConsole, "")` 初始化
   - 日志直接输出到标准输出
   - 适合开发环境使用
   - 支持结构化字段，以 `[key=value]` 的格式显示

2. Logrus 日志器
   - 使用 `log.InitLogger(log.LogTypeLogrus, "path/to/file.log")` 初始化
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