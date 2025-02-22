# Cache 包

`cache` 包提供了一个基于 ristretto 的高性能进程内缓存实现，支持基本的缓存操作和过期时间设置。

## 特性

- 高性能进程内缓存（基于 ristretto）
- 支持过期时间设置（TTL）
- 支持泛型和类型安全
- 支持全局缓存实例
- 线程安全
- 基于 LFU（最不经常使用）的驱逐策略
- 自动内存管理

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "time"
    "github.com/fsyyft-go/kit/cache"
)

func main() {
    // 初始化缓存
    if err := cache.InitCache(); err != nil {
        panic(err)
    }
    defer cache.Close()

    // 设置缓存
    cache.Set("key", "value")                    // 永不过期
    cache.SetWithTTL("temp", "value", time.Hour) // 1 小时后过期

    // 获取缓存
    if val, exists := cache.Get("key"); exists {
        fmt.Printf("value: %v\n", val)
    }

    // 获取带 TTL 的缓存
    if val, exists, ttl := cache.GetWithTTL("temp"); exists {
        fmt.Printf("值：%v，剩余时间：%v\n", val, ttl)
    }

    // 删除缓存
    cache.Delete("key")
}
```

### 使用泛型接口

```go
// 创建类型安全的缓存包装器
strCache := cache.AsTypedCache[string](cache.NewCache())
intCache := cache.AsTypedCache[int](cache.NewCache())

// 类型安全的操作
strCache.Set("str", "hello")
intCache.Set("int", 42)

// 无需类型断言
str, exists := strCache.Get("str")
num, exists := intCache.Get("int")
```

## 配置选项

```go
// 自定义配置
if err := cache.InitCache(
    cache.WithNumCounters(1e7),    // 跟踪的最大条目数（建议是实际条目数的 10 倍）
    cache.WithMaxCost(1<<30),      // 最大内存使用（字节）
    cache.WithBufferItems(64),     // 写入缓冲区大小
); err != nil {
    panic(err)
}
```

### 配置说明

- `NumCounters`：缓存跟踪的最大条目数，建议设置为预期独特条目数的 10 倍
- `MaxCost`：缓存的最大成本（可理解为最大条目数）
- `BufferItems`：写入操作的缓冲区大小，更大的缓冲区会提高并发性能，但会使用更多内存

## 最佳实践

1. 合理设置配置参数
   - NumCounters 建议设置为预期条目数的 10 倍
   - MaxCost 根据实际内存限制设置
   - BufferItems 默认值 64 适合大多数场景

2. 使用类型安全的接口
   - 优先使用 TypedCache 避免类型断言
   - 为不同类型的数据创建专门的缓存实例

3. 性能优化
   - 合理设置缓存大小避免频繁驱逐
   - 注意及时清理不需要的缓存项
   - 使用 Close 释放资源

4. 错误处理
   - 总是检查 InitCache 的返回错误
   - 使用 defer Close() 确保资源释放
   - 检查 Get 操作的 exists 返回值

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../LICENSE) 文件。 