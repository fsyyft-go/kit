# cache

## 简介

cache 包提供了一个统一的缓存接口和多种缓存实现，适用于需要高性能内存缓存的 Go 应用程序。默认使用 Ristretto 作为底层实现，提供高效、可靠的缓存功能。

### 主要特性

- 支持多种缓存后端（默认使用 ristretto）
- 提供统一的缓存接口
- 支持泛型和类型安全
- 支持 TTL（生存时间）设置
- 支持全局缓存实例
- 线程安全
- 高并发性能

### 设计理念

cache 包的设计遵循以下原则：

1. **简单性**：提供简洁易用的 API，使缓存操作尽可能直观
2. **灵活性**：支持多种缓存用例，从简单到复杂
3. **类型安全**：通过泛型支持类型安全的缓存操作，减少运行时错误
4. **可扩展性**：通过统一接口支持不同的缓存后端
5. **性能优先**：所有实现都注重性能和效率

包内部使用了适配器模式将不同的缓存后端实现适配到统一的 `Cache` 接口，同时提供了函数选项模式进行灵活配置。此外，利用泛型特性提供了类型安全的缓存操作，可以避免手动类型断言的麻烦。

## 安装

### 前置条件

- Go 版本要求：Go 1.18+ (支持泛型)
- 依赖要求：
  - github.com/dgraph-io/ristretto v0.2.0

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/cache
```

## 快速开始

### 基础用法

```go
// 创建缓存实例
cache, err := cache.NewCache()  // 使用默认配置
if err != nil {
    panic(err)
}
defer cache.Close()

// 基本操作
cache.Set("key", "value")                    // 设置永不过期的值
cache.SetWithTTL("temp", "value", time.Hour) // 设置 1 小时后过期的值

// 获取值
if val, exists := cache.Get("key"); exists {
    fmt.Println(val)
}

// 获取带 TTL 的值
if val, exists, ttl := cache.GetWithTTL("temp"); exists {
    fmt.Printf("值：%v，剩余时间：%v\n", val, ttl)
}
```

### 配置选项

```go
// 使用自定义配置
cache, err := cache.NewCache(
    cache.WithNumCounters(1e7),  // 跟踪 1000 万个条目
    cache.WithMaxCost(1<<30),    // 最大内存使用 1GB
    cache.WithBufferItems(64),   // 默认缓冲区大小
)
if err != nil {
    panic(err)
}
defer cache.Close()
```

## 详细指南

### 核心概念

1. **缓存接口**：`Cache` 接口定义了所有缓存实现必须提供的基本操作
2. **过期时间**：通过 TTL (Time-To-Live) 控制缓存项的生存时间
3. **驱逐策略**：当缓存达到容量限制时，最不经常使用的项目将被自动移除
4. **类型安全**：`TypedCache[T]` 提供类型安全的缓存操作，避免类型断言
5. **全局缓存**：通过全局函数可以方便地访问共享的缓存实例

### 常见用例

#### 1. 使用全局缓存

```go
// 初始化全局缓存
if err := cache.InitCache(); err != nil {
    panic(err)
}
defer cache.Close()

// 使用全局缓存函数
cache.Set("key", "value")
if val, exists := cache.Get("key"); exists {
    fmt.Println(val)
}

// 清理全局缓存
cache.Clear()
```

#### 2. 类型安全的缓存操作

```go
// 定义结构体
type User struct {
    ID   int
    Name string
    Age  int
}

// 创建缓存实例
baseCache, _ := cache.NewCache()
userCache := cache.AsTypedCache[User](baseCache)

// 存储用户对象
user := User{ID: 1, Name: "张三", Age: 30}
userCache.Set("user:1", user)

// 获取用户对象（无需类型断言）
if user, exists := userCache.Get("user:1"); exists {
    fmt.Printf("找到用户: %s, 年龄: %d\n", user.Name, user.Age)
}
```

### 最佳实践

- 合理设置配置参数
  - NumCounters 建议设置为预期条目数的 10 倍
  - MaxCost 根据实际内存限制设置
  - BufferItems 默认值 64 适合大多数场景

- 使用类型安全的接口
  - 优先使用 TypedCache 避免类型断言
  - 为不同类型的数据创建专门的缓存实例

- 性能优化
  - 合理设置缓存大小避免频繁驱逐
  - 注意及时清理不需要的缓存项
  - 使用 Close 释放资源

- 错误处理
  - 总是检查 InitCache 的返回错误
  - 使用 defer Close() 确保资源释放
  - 检查 Get 操作的 exists 返回值

## API 文档

### 主要类型

```go
// Cache 定义了统一的缓存接口
type Cache interface {
    Get(key interface{}) (value interface{}, exists bool)
    GetWithTTL(key interface{}) (value interface{}, exists bool, remainingTTL time.Duration)
    Set(key interface{}, value interface{}) bool
    SetWithTTL(key interface{}, value interface{}, ttl time.Duration) bool
    Delete(key interface{})
    Clear()
    Close() error
}

// TypedCache 是一个泛型包装器，提供类型安全的缓存操作
type TypedCache[T any] struct {
    // 内部字段
}

// CacheOptions 定义了缓存的配置选项
type CacheOptions struct {
    NumCounters int64
    MaxCost     int64
    BufferItems int64
}
```

### 关键函数

#### NewCache

创建一个新的缓存实例，默认使用 Ristretto 实现。

```go
func NewCache(options ...Option) (Cache, error)
```

示例：
```go
cache, err := cache.NewCache(
    cache.WithNumCounters(1e7),
    cache.WithMaxCost(1<<30),
)
```

#### InitCache

初始化全局缓存实例，确保只被初始化一次。

```go
func InitCache(options ...Option) error
```

示例：
```go
if err := cache.InitCache(); err != nil {
    panic(err)
}
```

#### AsTypedCache

将缓存转换为类型安全的包装器。

```go
func AsTypedCache[T any](cache Cache) *TypedCache[T]
```

示例：
```go
strCache := cache.AsTypedCache[string](baseCache)
```

### 错误处理

cache 包的大多数操作都会返回一个 bool 值表示操作是否成功，但创建和初始化函数会返回详细的错误信息。建议始终检查这些错误：

```go
cache, err := cache.NewCache()
if err != nil {
    // 处理错误：日志记录、返回错误或中止
    log.Fatalf("创建缓存失败: %v", err)
}
```

对于获取操作，应始终检查 `exists` 返回值：

```go
if value, exists := cache.Get(key); exists {
    // 使用 value
} else {
    // 处理键不存在的情况
}
```

## 性能指标

基于最近的基准测试结果：

| 操作 | 性能指标 | 说明 |
|------|----------|------|
| Set | ~1930 ns/op, 442 B/op, 6 allocs/op | 设置缓存项的基本性能 |
| Get | ~46.43 ns/op, 17 B/op, 1 allocs/op | 获取缓存项的基本性能，非常快 |
| SetWithTTL | ~2414 ns/op, 674 B/op, 6 allocs/op | 设置带TTL的性能，稍慢于基本Set |
| GetWithTTL | ~278.1 ns/op, 22 B/op, 1 allocs/op | 获取带TTL的性能，慢于基本Get |
| TypedCache Set | ~1947 ns/op, 615 B/op, 7 allocs/op | 类型安全的Set操作 |
| TypedCache Get | ~62.22 ns/op, 18 B/op, 1 allocs/op | 类型安全的Get操作 |
| Global Cache Set | ~78.10 ns/op, 40 B/op, 2 allocs/op | 全局缓存Set性能 |
| Global Cache Get | ~14.66 ns/op, 16 B/op, 1 allocs/op | 全局缓存Get性能，非常快 |

## 测试覆盖率

| 包 | 覆盖率 |
|------|--------|
| github.com/fsyyft-go/kit/cache | 71.1% |

## 调试指南

### 日志级别

cache 包本身不输出日志，建议在应用程序中集成适当的日志记录：

- 记录缓存初始化和配置
- 记录缓存相关的重要操作
- 监控缓存命中率和性能指标

### 常见问题排查

#### 内存使用过高

确保设置了合理的 MaxCost 值，这会限制缓存使用的最大内存。

#### 项目过早被驱逐

如果缓存项被过早驱逐，考虑增加 NumCounters 和 MaxCost 值。NumCounters 应该是预期独特项数的约 10 倍。

#### 性能问题

如果遇到性能瓶颈，请确保：
- 缓存配置适合您的使用场景
- 避免在热路径上频繁调用 SetWithTTL
- 考虑使用批量操作而不是多次单个操作

## 相关文档

- [示例代码](../example/cache/README.md)
- [Ristretto 文档](https://github.com/dgraph-io/ristretto)
- [Go 泛型指南](https://go.dev/blog/intro-generics)

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考我们的[贡献指南](../CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT License 许可证。查看 [LICENSE](../LICENSE) 文件了解更多信息。