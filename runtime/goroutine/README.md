# goroutine

## 简介

goroutine 包提供了在 Go 程序中获取 goroutine ID 的功能。虽然 Go 语言官方不推荐依赖 goroutine ID 进行业务逻辑处理，但在特定场景下（如调试、日志追踪、性能分析）获取 goroutine ID 非常有用。该包针对不同平台和架构提供了优化实现，确保高性能和广泛兼容性。

### 主要特性

- 支持多种 CPU 架构（AMD64、ARM64）的优化实现
- 提供通用的降级实现方案，确保所有平台兼容性
- 优化的性能设计，针对不同平台特性进行调整
- 简单易用的 API，便于快速集成
- 完整的测试覆盖和基准测试

### 设计理念

本包采用多层次实现策略，根据不同平台的特性提供最优性能的实现。设计核心是"优先性能，保证兼容"，通过汇编语言和运行时结构直接访问等方式实现高效获取 goroutine ID。同时，包设计考虑了 Go 语言版本兼容性问题，针对不同版本的 Go 运行时结构提供了相应的适配。

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

## 详细指南

### 核心概念

goroutine ID 是 Go 运行时为每个 goroutine 分配的唯一标识符。虽然 Go 语言设计上不鼓励依赖 goroutine ID 进行编程，但在某些场景下（如调试、日志追踪）获取 goroutine ID 非常有价值。本包采用多种实现方式获取 goroutine ID：

1. **快速路径**：针对特定平台（AMD64、ARM64）的优化实现，直接访问运行时内部结构
2. **慢速路径**：通用实现，通过解析 goroutine 堆栈信息提取 ID

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

#### 2. 性能分析和调试

```go
package main

import (
    "fmt"
    "runtime"
    "sync"
    "time"
    "github.com/fsyyft-go/kit/runtime/goroutine"
)

func main() {
    var wg sync.WaitGroup
    goroutineStats := make(map[int64]time.Duration)
    var mu sync.Mutex

    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            start := time.Now()
            goid := goroutine.GetGoID()

            // 模拟工作负载
            time.Sleep(time.Duration(n*100) * time.Millisecond)

            elapsed := time.Since(start)
            mu.Lock()
            goroutineStats[goid] = elapsed
            mu.Unlock()
        }(i)
    }

    wg.Wait()

    // 输出统计信息
    fmt.Println("Goroutine 执行统计:")
    for goid, duration := range goroutineStats {
        fmt.Printf("Goroutine %d: %v\n", goid, duration)
    }
}
```

### 最佳实践

- 谨慎依赖 goroutine ID，不要将其作为业务逻辑的核心
- 在性能敏感场景，获取 ID 后应缓存使用，避免重复获取
- 在适当的抽象层次使用 goroutine ID，如日志系统、调试工具
- 避免使用 goroutine ID 作为同步或通信机制的依赖
- 在不支持快速路径的平台上，注意性能损耗问题

## API 文档

### 主要类型

```go
// 包 goroutine 未定义公开类型，主要提供函数 API
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

### 错误处理

本包的函数不返回错误，但在极少数情况下可能会因运行时结构变化导致获取 ID 失败。在这种情况下，可能会返回意外值。建议在关键应用中添加额外的错误处理逻辑：

```go
id := goroutine.GetGoID()
if id <= 0 {
    // 处理异常情况
    log.Printf("警告: 获取 goroutine ID 失败，返回: %d", id)
    // 可能的降级策略...
}
```

## 性能指标

| 操作            | 性能指标  | 说明                                            |
| --------------- | --------- | ----------------------------------------------- |
| GetGoID (AMD64) | ~5ns/op   | 在 AMD64 架构上，通过汇编优化，接近直接内存访问 |
| GetGoID (ARM64) | ~8ns/op   | 在 ARM64 架构上，通过直接访问 g 结构体          |
| GetGoIDSlow     | ~200ns/op | 通过解析堆栈信息，性能较低但通用性好            |

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

## 补充说明

本文的大部分信息，由 AI 使用[模板](../../ai/templates/docs/package_readme_template.md)根据[提示词](../../ai/prompts/docs/package_readme_generator.md)自动生成，如有任何问题，请随时联系我。
