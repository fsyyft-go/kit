# context

## 简介

context 包提供了对 Go 标准库 `context.Context` 的扩展功能，主要实现了 `WithoutCancel` 函数，用于创建忽略父 context 取消信号的新 context。该功能相当于 Go 1.21+ 中的 `context.WithoutCancel`，用于低版本 Go 的兼容性实现。

本文档详细解释了 Go context 的超时、取消等功能的原理，包括自上而下传递机制以及 `WithoutCancel` 如何阻断这些信号。

### 主要特性

- 提供 `WithoutCancel` 函数，忽略父 context 的取消信号和超时
- 继承父 context 的值传递功能
- 轻量级实现，无额外性能开销
- 完全兼容 Go 标准库 context 接口
- 详细的原理说明和使用示例

### 设计理念

context 包的设计遵循以下原则：

1. **兼容性**：提供标准库缺失功能的兼容实现
2. **轻量级**：最小化实现，避免不必要的复杂性
3. **透明性**：清晰解释 context 的工作原理
4. **安全性**：确保阻断机制不影响值的传递

包内部通过包装器模式实现 `WithoutCancel`，确保只阻断取消行为而保留值继承。

## 快速开始

### 基础用法

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    kitctx "github.com/fsyyft-go/kit/context"
)

func main() {
    // 创建可取消的 context
    ctx, cancel := context.WithCancel(context.Background())
    
    // 创建带超时的子 context
    ctxTimeout, _ := context.WithTimeout(ctx, 100*time.Millisecond)
    
    // 使用 WithoutCancel 包装
    ctxWithoutCancel := kitctx.WithoutCancel(ctxTimeout)
    
    go func() {
        select {
        case <-ctxWithoutCancel.Done():
            fmt.Println("WithoutCancel context 被取消")
        case <-time.After(200 * time.Millisecond):
            fmt.Println("WithoutCancel context 正常完成")
        }
    }()
    
    // 取消父 context
    time.Sleep(50 * time.Millisecond)
    cancel()
    
    time.Sleep(200 * time.Millisecond)
}
```

输出：
```
WithoutCancel context 正常完成
```

## 详细指南

### 核心概念

1. **Context 接口**：Go 的 context.Context 定义了取消、超时和值传递的标准接口
2. **取消机制**：通过 Done() channel 自上而下传递取消信号
3. **超时机制**：基于定时器的自动取消功能
4. **WithoutCancel**：阻断取消信号但保留值传递的包装器

### 常见用例

#### 1. 忽略请求取消的清理工作

```go
func cleanup(ctx context.Context) {
    // 使用 WithoutCancel 确保清理工作完成
    ctx = kitctx.WithoutCancel(ctx)
    
    // 执行必须完成的清理操作
    // ...
}
```

#### 2. 长期后台任务

```go
func backgroundTask(parentCtx context.Context) {
    // 忽略父 context 的超时
    ctx := kitctx.WithoutCancel(parentCtx)
    
    // 执行长期运行的任务
    for {
        select {
        case <-ctx.Done():
            // 永不会到达，除非手动取消
            return
        default:
            // 执行任务逻辑
        }
    }
}
```

### 最佳实践

- 谨慎使用 `WithoutCancel`，确保不会导致资源泄漏
- 只在确实需要忽略取消的场景下使用
- 结合适当的超时或手动取消机制管理生命周期
- 在 goroutine 中使用时，确保有退出条件

## API 文档

### 主要类型

```go
// withoutCancelCtx 是一个包装的 context 实现，用于忽略父 context 的取消信号。
type withoutCancelCtx struct {
    parent stdctx.Context // 父 context，用于获取值
}
```

### 关键函数

#### WithoutCancel

返回一个新的 context，它继承父 context 的值，但忽略取消信号和超时。

```go
func WithoutCancel(parent stdctx.Context) stdctx.Context
```

参数：
- `parent`：父 context，不能为 nil

返回值：
- `stdctx.Context`：新的 context，忽略取消和超时，但保留值

示例：
```go
ctx := kitctx.WithoutCancel(parentCtx)
```

### 错误处理

`WithoutCancel` 函数会在 parent 为 nil 时 panic，其他情况下不会返回错误。

## 性能考虑

- `WithoutCancel` 创建的 context 非常轻量，没有额外的 goroutine 或定时器
- 只在调用 `Value()` 时访问父 context，避免不必要的开销
- 内存占用小，适合大量使用

## 注意事项

1. **生命周期管理**：`WithoutCancel` 创建的 context 不会自动取消，需要手动管理其生命周期
2. **值传递**：只继承值，不继承取消行为。确保在适当的时候处理清理工作
3. **Go 版本兼容**：该实现用于 Go 1.21 之前的版本，1.21+ 可直接使用标准库的 `context.WithoutCancel`

## 相关文档

- [context_test.go](context_test.go) - 单元测试和使用示例
- [Go 官方文档：Context](https://golang.org/pkg/context/)
- [Go 博客：Context](https://blog.golang.org/context)

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考我们的[贡献指南](../CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT License 许可证。查看 [LICENSE](../LICENSE) 文件了解更多信息。</content>