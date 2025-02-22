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

## 平台实现对照表

| 操作系统 | CPU 架构 | 实现文件 | 实现方法 | 说明 |
|---------|---------|---------|---------|------|
| Windows | AMD64 | goid_amd64.s | 汇编实现 | 通过 TLS 直接访问，使用偏移量获取 |
| Windows | ARM64 | goid_windows_arm64.go | getGoIDSlow | 使用堆栈信息解析方式获取 |
| Linux | AMD64 | goid_amd64.s | 汇编实现 | 通过 TLS 直接访问，使用偏移量获取 |
| Linux | ARM64 | goid_arm64.go | getg().goid | 通过 TLS 直接访问 g 结构体 |
| Darwin | AMD64 | goid_amd64.s | 汇编实现 | 通过 TLS 直接访问，使用偏移量获取 |
| Darwin | ARM64 | goid_arm64.go | getg().goid | 通过 TLS 直接访问 g 结构体 |
| 所有 | 386 | goid.go | getGoIDSlow | 使用堆栈信息解析方式获取 |
| 所有 | ARM(32位) | goid.go | getGoIDSlow | 使用堆栈信息解析方式获取 |
| 所有 | MIPS/MIPS64 | goid.go | getGoIDSlow | 使用堆栈信息解析方式获取 |
| 所有 | PPC64/PPC64LE | goid.go | getGoIDSlow | 使用堆栈信息解析方式获取 |
| 所有 | S390X | goid.go | getGoIDSlow | 使用堆栈信息解析方式获取 |
| 所有 | RISC-V 64 | goid.go | getGoIDSlow | 使用堆栈信息解析方式获取 |

### 实现说明

1. **AMD64 平台**：
   - 使用汇编实现，通过 TLS（Thread Local Storage）直接访问
   - 使用 offset 变量存储不同 Go 版本的 goid 偏移量
   - 性能最优
   - 支持 Windows、Linux、Darwin 等主流操作系统

2. **ARM64 平台**：
   - 非 Windows 系统：直接通过 getg() 获取 g 结构体
   - Windows 系统：使用 getGoIDSlow 方法
   - 性能次优（非 Windows）
   - M1/M2 芯片的 Mac 设备使用此实现

3. **其他平台**：
   - 统一使用 getGoIDSlow 方法
   - 通过解析堆栈信息获取 goroutine ID
   - 性能较低但通用性好
   - 支持所有 Go 语言支持的操作系统和 CPU 架构组合

### 性能说明

从性能角度，实现方式按照以下顺序从高到低排列：
1. AMD64 汇编实现：性能最优，接近直接内存访问速度
2. ARM64 非 Windows 实现：性能次优，通过 getg() 直接访问
3. 通用 getGoIDSlow 实现：性能最低，但兼容性最好

## 注意事项

1. Go 语言官方不推荐依赖 goroutine ID 进行业务逻辑处理
2. 该功能主要用于调试、日志追踪等辅助场景
3. 在不同的 Go 版本中，获取 goroutine ID 的实现细节可能会发生变化

## 许可证

该项目采用 MIT 许可证。详见 [LICENSE](https://github.com/fsyyft-go/kit/blob/main/LICENSE) 文件。 