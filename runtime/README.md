# Runtime 包

`runtime` 包提供了应用程序运行时管理的基础设施。

## 特性

- 提供 Runner 接口，用于管理组件的生命周期
- 支持组件的启动和优雅停止
- 基于 context 的生命周期控制

## 快速开始

### 实现 Runner 接口

```go
package main

import (
    "context"
    "github.com/fsyyft-go/kit/runtime"
)

// MyComponent 实现了 Runner 接口
type MyComponent struct {
    // ... 组件字段
}

// Start 实现 Runner 接口的启动方法
func (c *MyComponent) Start(ctx context.Context) error {
    // 启动组件的逻辑
    return nil
}

// Stop 实现 Runner 接口的停止方法
func (c *MyComponent) Stop(ctx context.Context) error {
    // 优雅停止组件的逻辑
    return nil
}
```

### 使用组件

```go
func main() {
    // 创建组件
    component := &MyComponent{}
    
    // 启动组件
    ctx := context.Background()
    if err := component.Start(ctx); err != nil {
        panic(err)
    }
    
    // ... 应用程序逻辑
    
    // 停止组件
    if err := component.Stop(ctx); err != nil {
        panic(err)
    }
}
```

## 子包

### [goroutine](goroutine/README.md)

提供获取 goroutine ID 的功能。⚠️ 仅用于特殊调试场景，不建议在生产环境使用。[详细说明 →](goroutine/README.md)

## 最佳实践

1. 组件生命周期
   - 在 Start 方法中初始化资源
   - 在 Stop 方法中清理资源
   - 正确处理 context 取消信号

2. 错误处理
   - 返回有意义的错误信息
   - 在停止时尽可能清理资源
   - 避免在清理过程中产生新的错误

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../LICENSE) 文件。 