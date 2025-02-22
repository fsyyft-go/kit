# Goroutine ID 获取工具

这个包提供了在 Go 程序中获取 goroutine ID 的功能。虽然 Go 语言本身并不推荐依赖 goroutine ID，但在某些特殊场景下（如调试、日志追踪等），获取 goroutine ID 可能会很有用。

## 功能特性

- 支持多种 CPU 架构（AMD64、ARM64）的优化实现
- 提供通用的降级实现方案
- 高性能设计，针对不同平台进行优化
- 简单易用的 API

## 实现原理

该包提供了两种获取 goroutine ID 的方式：

1. **快速路径**：
   - 在支持的架构（AMD64、ARM64）上，通过直接访问 TLS（Thread Local Storage）获取 goroutine ID
   - 性能最优，几乎没有额外开销

2. **慢速路径**：
   - 适用于不支持快速路径的平台
   - 通过解析 goroutine 堆栈信息来获取 ID
   - 性能相对较低，但具有更好的兼容性

## 使用方法

```go
package main

import (
    "fmt"
    "github.com/fsyyft-go/kit/runtime/goroutine"
)

func main() {
    // 获取当前 goroutine 的 ID
    goid := goroutine.GetGoID()
    fmt.Printf("Current goroutine ID: %d\n", goid)

    // 在多个 goroutine 中使用
    go func() {
        goid := goroutine.GetGoID()
        fmt.Printf("Goroutine ID in goroutine: %d\n", goid)
    }()
}
```

## 性能考虑

- 在支持的架构上（AMD64、ARM64），`GetGoID()` 性能接近于直接内存访问
- 在不支持的架构上，会自动降级使用 `getGoIDSlow()`，可能会有一定性能开销
- 如果确实需要在不支持的架构上频繁获取 goroutine ID，建议缓存结果而不是重复获取

## 支持的平台

### 快速路径支持：
- AMD64 架构
- ARM64 架构

### 通用支持（慢速路径）：
- 所有 Go 支持的平台

## 注意事项

1. Go 语言官方不推荐依赖 goroutine ID 进行业务逻辑处理
2. 该功能主要用于调试、日志追踪等辅助场景
3. 在不同的 Go 版本中，获取 goroutine ID 的实现细节可能会发生变化

## 许可证

该项目采用 MIT 许可证。详见 [LICENSE](https://github.com/fsyyft-go/kit/blob/main/LICENSE) 文件。 