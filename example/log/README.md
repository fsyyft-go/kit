 # 日志功能示例

本示例展示了如何使用 Kit 的日志模块，实现统一的日志记录、多级别日志控制和结构化日志输出。

## 功能特性

- 支持多种日志级别（Debug、Info、Warn、Error、Fatal）
- 支持结构化日志记录（WithField、WithFields）
- 支持多种日志后端（标准库、Logrus）
- 支持文件输出和日志滚动
- 支持格式化日志输出
- 支持独立日志实例创建

## 设计原理

Kit 的日志模块设计采用了统一接口 + 多种实现的方式，主要组件包括：

- 统一的 `Logger` 接口：定义了所有日志实现必须支持的方法
- 多种后端实现：支持标准库（StdLogger）和 Logrus（LogrusLogger）
- 全局日志实例：方便在应用的不同部分使用相同的日志配置
- 函数式选项模式：灵活配置日志行为

## 使用方法

### 1. 编译和运行

在 Unix/Linux/macOS 系统上：

```bash
# 添加执行权限
chmod +x build.sh

# 构建和运行
./build.sh
```

在 Windows 系统上：

```cmd
# 使用 Go 命令直接构建和运行
go build -o bin/example/log/log.exe main.go
./bin/example/log/log.exe
```

### 2. 代码示例

#### 基本日志记录

```go
// 初始化默认日志器
if err := log.InitLogger(); err != nil {
    panic(err)
}

// 设置日志级别
log.SetLevel(log.DebugLevel)

// 记录不同级别的日志
log.Debug("这是一条调试日志")
log.Info("这是一条信息日志")
log.Warn("这是一条警告日志")
log.Error("这是一条错误日志")
```

#### 格式化日志

```go
log.Debugf("当前时间是: %v", time.Now().Format("2006-01-02 15:04:05"))
log.Infof("程序运行在: %s", os.Getenv("PWD"))
```

#### 结构化日志

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

#### 自定义日志配置

```go
// 配置日志输出到文件，使用 Logrus 后端
logFile := filepath.Join("example", "log", "app.log")
if err := log.InitLogger(
    log.WithLogType(log.LogTypeLogrus),
    log.WithOutput(logFile),
    log.WithLevel(log.InfoLevel),
); err != nil {
    panic(err)
}
```

#### 创建独立日志实例

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
logger.Debug("这是独立日志实例的调试信息")
logger.WithField("module", "cache").Info("缓存已初始化")
```

### 3. 输出示例

使用默认配置（标准输出）：
```
2024/03/15 10:00:00 [DEBUG] 这是一条调试日志
2024/03/15 10:00:00 [INFO] 这是一条信息日志
2024/03/15 10:00:00 [WARN] 这是一条警告日志
2024/03/15 10:00:00 [ERROR] 这是一条错误日志
2024/03/15 10:00:00 [DEBUG] 当前时间是: 2024-03-15 10:00:00
2024/03/15 10:00:00 [INFO] 程序运行在: /path/to/example/log
2024/03/15 10:00:00 [INFO] [user=admin] 用户登录
2024/03/15 10:00:00 [INFO] [ip=192.168.1.1 method=POST latency=20ms] 收到HTTP请求
```

使用 Logrus 配置（JSON 格式）：
```json
{"level":"info","msg":"已切换到 logrus 日志器（默认启用日志滚动功能）","time":"2024-03-15T10:00:00+08:00"}
{"component":"server","level":"info","msg":"服务器启动","status":"starting","time":"2024-03-15T10:00:00+08:00"}
```

### 4. 在其他项目中使用

```go
package main

import (
    "github.com/fsyyft-go/kit/log"
)

func main() {
    // 初始化日志
    if err := log.InitLogger(
        log.WithLogType(log.LogTypeLogrus),
        log.WithOutput("/var/log/myapp.log"),
        log.WithLevel(log.InfoLevel),
    ); err != nil {
        panic(err)
    }
    
    // 使用日志
    log.Info("应用启动")
    
    // 记录带上下文的日志
    log.WithFields(map[string]interface{}{
        "module": "api",
        "method": "GET",
        "path":   "/users",
    }).Info("处理请求")
    
    // 记录错误
    if err := someFunction(); err != nil {
        log.WithField("error", err).Error("操作失败")
    }
}
```

## 注意事项

- 日志级别控制：只有大于或等于设置级别的日志才会被记录
- 文件路径：使用 `WithOutput` 指定日志文件路径时，会自动创建所需的目录
- 日志滚动：使用 Logrus 后端时默认启用日志滚动功能，按小时滚动
- 性能考虑：结构化日志在高并发场景下可能影响性能，建议适度使用
- 敏感信息：避免在日志中记录密码、令牌等敏感信息

## 相关文档

- [Kit 日志模块文档](../../log/README.md)
- [Logrus 官方文档](https://github.com/sirupsen/logrus)
- [Go 标准库 log 包](https://golang.org/pkg/log/)

## 许可证

本示例代码采用 MIT 许可证。详见 [LICENSE](../../LICENSE) 文件。