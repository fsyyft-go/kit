# Context 包说明文档

## 概述

`context` 包提供了对 Go 标准库 `context.Context` 的扩展功能，主要实现了 `WithoutCancel` 函数，用于创建忽略父 context 取消信号的新 context。该功能相当于 Go 1.21+ 中的 `context.WithoutCancel`，用于低版本 Go 的兼容性实现。

本文档详细解释了 Go context 的超时、取消等功能的原理，包括自上而下传递机制以及 `WithoutCancel` 如何阻断这些信号。

## Context 基础概念

### Context 接口定义

Go 的 `context.Context` 是一个接口，定义如下：

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key any) any
}
```

- `Deadline()`：返回 context 的截止时间，如果没有则返回零时间和 false。
- `Done()`：返回一个 channel，当 context 被取消或超时后，该 channel 会关闭。
- `Err()`：返回取消的原因，如果未取消则返回 nil。
- `Value()`：用于存储和传递请求范围内的值。

### Context 的树状结构

Context 通常形成一个树状结构：

```
Background Context (根)
    ├── WithCancel (可取消)
    │   ├── WithTimeout (带超时)
    │   └── WithValue (带值)
    └── WithDeadline (带截止时间)
```

子 context 继承父 context 的属性，但可以添加额外的行为（如取消、超时）。

## 取消机制原理

### 自上而下传递

Context 的取消信号通过以下方式自上而下传递：

1. **根源**：通常通过 `context.WithCancel(parent)` 创建可取消的 context，返回 context 和 cancel 函数。

2. **传播**：当调用 cancel 函数时，会关闭根 context 的 Done channel。

3. **传递**：所有子 context 的 Done() 方法都会返回同一个底层 channel，因此当根 context 被取消时，所有子 context 都会收到信号。

### 底层实现

Go 标准库使用 `cancelCtx` 结构体来实现取消功能：

```go
type cancelCtx struct {
    Context
    mu       sync.Mutex
    done     chan struct{}
    children map[canceler]struct{}  // 子 context 列表
    err      error
}
```

- `done`：共享的 channel，用于通知取消。
- `children`：记录所有子 context，当父 context 取消时，会递归取消所有子 context。
- `err`：存储取消原因。

### 取消流程

1. 调用 `cancel()` 函数。
2. 设置 `err` 为取消原因。
3. 关闭 `done` channel。
4. 遍历 `children`，递归调用子 context 的 cancel 方法。

这样确保了取消信号能够快速传播到整个 context 树。

## 超时机制原理

### WithTimeout 和 WithDeadline

- `WithTimeout(parent, duration)`：在指定时间后自动取消。
- `WithDeadline(parent, deadline)`：在指定时间点自动取消。

### 实现原理

超时 context 使用 `timerCtx` 结构体：

```go
type timerCtx struct {
    cancelCtx
    timer *time.Timer
    deadline time.Time
}
```

- 继承 `cancelCtx`，具有取消功能。
- `timer`：定时器，到期时自动调用 cancel。
- `deadline`：截止时间。

### 超时流程

1. 创建定时器，设置为 deadline。
2. 当定时器到期时，调用 cancel 方法。
3. 取消信号通过 cancelCtx 的机制传播。

## WithoutCancel 阻断机制

### 设计目的

`WithoutCancel` 创建一个新的 context，它：

- 继承父 context 的值（通过 `Value()` 方法）。
- 忽略父 context 的取消信号和超时。
- 永远不会被取消（`Done()` 返回 nil，`Err()` 返回 nil，`Deadline()` 返回无截止时间）。

### 实现原理

`withoutCancelCtx` 结构体包装了父 context：

```go
type withoutCancelCtx struct {
    parent Context
}
```

- `Deadline()`：总是返回零时间和 false，忽略父 context 的截止时间。
- `Done()`：返回 nil，表示永远不会关闭。
- `Err()`：返回 nil，表示没有错误。
- `Value(key)`：委托给父 context，返回对应的值。

### 阻断效果

1. **取消阻断**：即使父 context 被取消，`withoutCancelCtx` 的 `Done()` 永远不会关闭，`Err()` 永远返回 nil。
2. **超时阻断**：忽略父 context 的截止时间，永远不会因超时而取消。
3. **值继承**：仍然可以获取父 context 中存储的值。

### 应用场景

- 在 goroutine 中执行必须完成的清理工作，即使父 context 已取消。
- 创建长期运行的后台任务，不受请求 context 的超时影响。
- 在错误处理或日志记录中，确保关键操作完成。

## 代码示例

### 基本使用

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

### 测试验证

参考 `context_test.go` 中的测试用例：

- **超时停止**：使用带超时的 context，goroutine 会因超时而退出。
- **正常停止**：使用可取消的 context，goroutine 会因取消而退出。
- **阻断停止**：使用 `WithoutCancel` 包装的 context，goroutine 会正常完成，忽略取消信号。

## 性能考虑

- `WithoutCancel` 创建的 context 非常轻量，没有额外的 goroutine 或定时器。
- 只在调用 `Value()` 时访问父 context，避免不必要的开销。
- 内存占用小，适合大量使用。

## 注意事项

1. **生命周期管理**：`WithoutCancel` 创建的 context 不会自动取消，需要手动管理其生命周期。
2. **值传递**：只继承值，不继承取消行为。确保在适当的时候处理清理工作。
3. **Go 版本兼容**：该实现用于 Go 1.21 之前的版本，1.21+ 可直接使用标准库的 `context.WithoutCancel`。

## 参考资料

- [Go 官方文档：Context](https://golang.org/pkg/context/)
- [Go 博客：Context](https://blog.golang.org/context)
- [Go 1.21 发布说明：context.WithoutCancel](https://go.dev/doc/go1.21#context)</content>