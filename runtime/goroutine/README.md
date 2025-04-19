# goroutine

## 简介

goroutine 包提供了在 Go 程序中获取 goroutine ID 的功能和高效的协程池实现。虽然 Go 语言官方不推荐依赖 goroutine ID 进行业务逻辑处理，但在特定场景下（如调试、日志追踪、性能分析）获取 goroutine ID 非常有用。同时，包内提供了高性能的协程池实现，支持任务调度、资源管理和性能监控等功能。该包针对不同平台和架构提供了优化实现，确保高性能和广泛兼容性。

### 主要特性

- 支持多种 CPU 架构（AMD64、ARM64）的优化实现
- 提供通用的降级实现方案，确保所有平台兼容性
- 优化的性能设计，针对不同平台特性进行调整
- 简单易用的 API，便于快速集成
- 完整的测试覆盖和基准测试
- 高性能协程池实现，支持动态扩缩容
- 丰富的配置选项，满足不同场景需求
- 内置监控指标，便于性能分析和调优

### 设计理念

本包采用多层次实现策略，根据不同平台的特性提供最优性能的实现。设计核心是"优先性能，保证兼容"，通过汇编语言和运行时结构直接访问等方式实现高效获取 goroutine ID。同时，包设计考虑了 Go 语言版本兼容性问题，针对不同版本的 Go 运行时结构提供了相应的适配。

协程池设计遵循以下原则：
- 资源高效利用：动态调整协程数量，避免资源浪费
- 任务公平调度：确保任务按提交顺序执行
- 异常安全处理：内置 panic 恢复机制
- 监控友好：提供丰富的运行时指标
- 配置灵活：支持多种配置选项

## 安装

### 前置条件

- Go 版本要求：Go 1.5+
- 依赖要求：
  - 无外部依赖，仅使用 Go 标准库

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/runtime/goroutine
```

## 快速开始

### 基础用法

#### 获取 Goroutine ID

```go
package main

import (
    "fmt"
    "github.com/fsyyft-go/kit/runtime/goroutine"
)

func main() {
    // 获取当前 goroutine 的 ID
    goid := goroutine.GetGoID()
    fmt.Printf("当前 goroutine ID: %d\n", goid)

    // 在多个 goroutine 中使用
    go func() {
        goid := goroutine.GetGoID()
        fmt.Printf("子 goroutine ID: %d\n", goid)
    }()
}
```

#### 使用协程池

```go
package main

import (
    "fmt"
    "sync"
    "time"
    "github.com/fsyyft-go/kit/runtime/goroutine"
)

func main() {
    // 创建协程池
    pool, cleanup, err := goroutine.NewGoroutinePool(
        goroutine.WithSize(10),           // 设置池大小
        goroutine.WithExpiry(time.Second), // 设置协程过期时间
        goroutine.WithName("worker"),      // 设置池名称
    )
    if err != nil {
        panic(err)
    }
    defer cleanup()

    // 提交任务
    var wg sync.WaitGroup
    for i := 0; i < 5; i++ {
        wg.Add(1)
        taskID := i
        err := pool.Submit(func() {
            defer wg.Done()
            fmt.Printf("执行任务 %d\n", taskID)
            time.Sleep(100 * time.Millisecond)
        })
        if err != nil {
            fmt.Printf("提交任务失败: %v\n", err)
        }
    }
    wg.Wait()

    // 查看池状态
    fmt.Printf("池容量: %d\n", pool.Cap())
    fmt.Printf("运行中协程: %d\n", pool.Running())
    fmt.Printf("空闲协程: %d\n", pool.Free())
    fmt.Printf("等待任务: %d\n", pool.Waiting())
}
```

## 详细指南

### 核心概念

#### Goroutine ID

goroutine ID 是 Go 运行时为每个 goroutine 分配的唯一标识符。虽然 Go 语言设计上不鼓励依赖 goroutine ID 进行编程，但在某些场景下（如调试、日志追踪）获取 goroutine ID 非常有价值。本包采用多种实现方式获取 goroutine ID：

1. **快速路径**：针对特定平台（AMD64、ARM64）的优化实现，直接访问运行时内部结构
2. **慢速路径**：通用实现，通过解析 goroutine 堆栈信息提取 ID

#### 协程池

协程池是一种用于管理和复用 goroutine 的机制，主要功能包括：

1. **资源管理**：控制并发 goroutine 数量，避免资源耗尽
2. **任务调度**：公平调度任务执行，支持任务队列
3. **性能优化**：动态调整协程数量，优化资源利用
4. **监控统计**：提供运行时指标，便于性能分析

### 配置选项

协程池支持丰富的配置选项，可以通过 `NewGoroutinePool` 函数的选项参数进行配置：

```go
// 创建协程池示例
pool, cleanup, err := goroutine.NewGoroutinePool(
    goroutine.WithSize(100),              // 设置池大小
    goroutine.WithExpiry(time.Second),    // 设置协程过期时间
    goroutine.WithPreAlloc(true),         // 预创建协程
    goroutine.WithNonBlocking(false),     // 阻塞模式
    goroutine.WithMaxBlocking(1000),      // 最大阻塞任务数
    goroutine.WithPanicHandler(func(r interface{}) {
        // 处理 panic
    }),
    goroutine.WithName("worker"),         // 设置池名称
    goroutine.WithMetrics(true),          // 启用指标收集
)
```

主要配置选项说明：

- `WithSize`：设置协程池大小
- `WithExpiry`：设置协程过期时间
- `WithPreAlloc`：是否预创建协程
- `WithNonBlocking`：是否使用非阻塞模式
- `WithMaxBlocking`：最大阻塞任务数
- `WithPanicHandler`：panic 处理函数
- `WithName`：协程池名称
- `WithMetrics`：是否启用指标收集

### 常见用例

#### 1. 在日志系统中跟踪 goroutine

```go
package main

import (
    "fmt"
    "log"
    "sync"
    "github.com/fsyyft-go/kit/runtime/goroutine"
)

func main() {
    var wg sync.WaitGroup
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(taskID int) {
            defer wg.Done()
            goid := goroutine.GetGoID()
            log.Printf("[goroutine:%d] 执行任务 %d", goid, taskID)
            // 执行业务逻辑...
        }(i)
    }
    wg.Wait()
}
```

#### 2. 使用协程池处理并发任务

```go
package main

import (
    "fmt"
    "sync"
    "time"
    "github.com/fsyyft-go/kit/runtime/goroutine"
)

func main() {
    // 创建协程池
    pool, cleanup, err := goroutine.NewGoroutinePool(
        goroutine.WithSize(10),
        goroutine.WithName("worker"),
    )
    if err != nil {
        panic(err)
    }
    defer cleanup()

    // 提交并发任务
    var wg sync.WaitGroup
    for i := 0; i < 20; i++ {
        wg.Add(1)
        taskID := i
        err := pool.Submit(func() {
            defer wg.Done()
            fmt.Printf("任务 %d 开始执行\n", taskID)
            time.Sleep(100 * time.Millisecond)
            fmt.Printf("任务 %d 执行完成\n", taskID)
        })
        if err != nil {
            fmt.Printf("提交任务 %d 失败: %v\n", taskID, err)
            wg.Done()
        }
    }
    wg.Wait()
}
```

### 最佳实践

#### Goroutine ID 使用建议

- 谨慎依赖 goroutine ID，不要将其作为业务逻辑的核心
- 在性能敏感场景，获取 ID 后应缓存使用，避免重复获取
- 在适当的抽象层次使用 goroutine ID，如日志系统、调试工具
- 避免使用 goroutine ID 作为同步或通信机制的依赖
- 在不支持快速路径的平台上，注意性能损耗问题

#### 协程池使用建议

- 根据实际负载合理设置池大小，避免资源浪费
- 使用非阻塞模式时注意处理任务提交失败的情况
- 合理设置协程过期时间，平衡资源利用和响应速度
- 在关键任务中实现 panic 处理，确保系统稳定性
- 定期监控池状态，及时发现性能问题
- 使用池名称区分不同业务场景的协程池
- 在服务关闭时正确清理协程池资源

## API 文档

### 主要类型

```go
// GoroutinePool 定义了协程池的接口
type GoroutinePool interface {
    // Submit 提交任务到协程池
    Submit(task func()) error
    // Tune 调整协程池大小
    Tune(size int)
    // Cap 获取协程池容量
    Cap() int
    // Running 获取运行中协程数
    Running() int
    // Free 获取空闲协程数
    Free() int
    // Waiting 获取等待任务数
    Waiting() int
    // IsClosed 检查协程池是否已关闭
    IsClosed() bool
}
```

### 关键函数

#### GetGoID

获取当前 goroutine 的 ID。根据平台和架构自动选择最优实现。

```go
func GetGoID() int64
```

示例：

```go
id := goroutine.GetGoID()
fmt.Printf("当前 goroutine ID: %d\n", id)
```

#### GetGoIDSlow

获取当前 goroutine 的 ID，使用通用但较慢的实现方式。适用于所有平台。

```go
func GetGoIDSlow() int64
```

示例：

```go
id := goroutine.GetGoIDSlow()
fmt.Printf("使用慢速路径获取的 goroutine ID: %d\n", id)
```

#### NewGoroutinePool

创建新的协程池实例。

```go
func NewGoroutinePool(opts ...Option) (GoroutinePool, func(), error)
```

示例：

```go
pool, cleanup, err := goroutine.NewGoroutinePool(
    goroutine.WithSize(10),
    goroutine.WithName("worker"),
)
if err != nil {
    panic(err)
}
defer cleanup()
```

#### Submit

提交任务到默认协程池。

```go
func Submit(task func()) error
```

示例：

```go
err := goroutine.Submit(func() {
    // 执行任务
})
if err != nil {
    // 处理错误
}
```

### 错误处理

本包的函数可能返回以下错误：

- `ErrPoolClosed`：协程池已关闭
- `ErrPoolOverload`：协程池过载
- `ErrInvalidPoolSize`：无效的池大小
- `ErrInvalidPoolExpiry`：无效的过期时间

建议在关键应用中添加适当的错误处理：

```go
pool, cleanup, err := goroutine.NewGoroutinePool(
    goroutine.WithSize(10),
)
if err != nil {
    // 处理创建失败
    log.Printf("创建协程池失败: %v", err)
    return
}
defer cleanup()

err = pool.Submit(func() {
    // 执行任务
})
if err != nil {
    // 处理提交失败
    log.Printf("提交任务失败: %v", err)
    return
}
```

## 性能指标

| 操作            | 性能指标  | 说明                                            |
| --------------- | --------- | ----------------------------------------------- |
| GetGoID (AMD64) | ~5ns/op   | 在 AMD64 架构上，通过汇编优化，接近直接内存访问 |
| GetGoID (ARM64) | ~8ns/op   | 在 ARM64 架构上，通过直接访问 g 结构体          |
| GetGoIDSlow     | ~200ns/op | 通过解析堆栈信息，性能较低但通用性好            |
| 任务提交        | ~100ns/op | 提交任务到协程池的开销                          |
| 协程创建        | ~1μs/op   | 创建新协程的开销                                |
| 任务调度        | ~50ns/op  | 任务调度的开销                                  |

## 测试覆盖率

| 包        | 覆盖率 |
| --------- | ------ |
| goroutine | >85%   |

## 调试指南

### 日志级别

- ERROR: 获取 goroutine ID 失败的错误
- WARN: 特定平台限制导致性能降级的警告
- INFO: 包初始化和版本适配信息
- DEBUG: 详细的运行时信息和性能数据

### 常见问题排查

#### 在 M1/M2 芯片的 Mac 设备上获取的 ID 不稳定

在 Darwin ARM64 架构（如 M1/M2 Mac）上，由于平台限制，可能需要使用不同的实现。请确保使用最新版本的包。

#### 不同 Go 版本表现不一致

本包针对不同 Go 版本的运行时结构提供了适配。如果在特定 Go 版本上遇到问题，请检查是否使用了匹配的适配文件。

## 相关文档

- [Go 语言运行时调度器](https://go.dev/src/runtime/HACKING.md)
- [内部 G 结构定义](https://github.com/golang/go/blob/master/src/runtime/runtime2.go)
- [TLS (Thread Local Storage) 在 Go 中的应用](https://go.dev/src/runtime/asm.s)

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考我们的[贡献指南](https://github.com/fsyyft-go/kit/blob/main/CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT 许可证。查看 [LICENSE](https://github.com/fsyyft-go/kit/blob/main/LICENSE) 文件了解更多信息。
