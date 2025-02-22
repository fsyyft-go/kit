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

### 示例1：使用全局缓存

```go
// 初始化日志
if err := log.InitLogger(); err != nil {
    panic(err)
}

// 初始化缓存
if err := cache.InitCache(); err != nil {
    log.WithField("error", err).Fatal("初始化缓存失败")
}
defer cache.Close()

// 基本的缓存操作
cache.Set("string_key", "这是一个字符串值")
cache.Set("int_key", 42)
cache.SetWithTTL("temp_key", "这个值将在 1 秒后过期", time.Second)

// 获取缓存值
if value, exists := cache.Get("string_key"); exists {
    log.WithField("value", value).Info("获取 string_key 的值")
}

if value, exists := cache.Get("int_key"); exists {
    log.WithField("value", value).Info("获取 int_key 的值")
}

// 获取带 TTL 的缓存值
if value, exists, ttl := cache.GetWithTTL("temp_key"); exists {
    log.WithFields(map[string]interface{}{
        "value": value,
        "ttl":   ttl,
    }).Info("获取 temp_key 的值")
}

// 等待一段时间后检查过期的值
time.Sleep(1500 * time.Millisecond)
if _, exists := cache.Get("temp_key"); !exists {
    log.Info("temp_key 已经过期")
}
```

### 示例2：使用自定义配置的缓存实例

```go
// 创建自定义配置的缓存实例
customCache, err := cache.NewCache(
    cache.WithNumCounters(1e5), // 跟踪 10 万个条目
    cache.WithMaxCost(1<<28),   // 最大内存使用 256MB
    cache.WithBufferItems(64),  // 默认的缓冲区大小
)
if err != nil {
    log.WithField("error", err).Fatal("创建自定义缓存失败")
}
defer customCache.Close()

// 使用自定义缓存实例
customCache.Set("custom_key", "这是自定义缓存中的值")
if value, exists := customCache.Get("custom_key"); exists {
    log.WithField("value", value).Info("获取 custom_key 的值")
}
```

### 示例3：使用类型安全的缓存

```go
// 定义一个示例结构体
type User struct {
    ID   int
    Name string
    Age  int
}

// 创建一个专门用于存储 User 类型的缓存包装器
userCache := cache.AsTypedCache[User](customCache)

// 存储用户对象
user := User{
    ID:   1,
    Name: "张三",
    Age:  30,
}
userCache.Set("user:1", user)

// 获取用户对象（无需类型断言）
if user, exists := userCache.Get("user:1"); exists {
    log.WithFields(map[string]interface{}{
        "id":   user.ID,
        "name": user.Name,
        "age":  user.Age,
    }).Info("找到用户")
}

// 使用 TTL 存储用户对象
tempUser := User{
    ID:   2,
    Name: "李四",
    Age:  25,
}
userCache.SetWithTTL("user:2", tempUser, time.Second)

// 获取带 TTL 的用户对象
if user, exists, ttl := userCache.GetWithTTL("user:2"); exists {
    log.WithFields(map[string]interface{}{
        "id":   user.ID,
        "name": user.Name,
        "age":  user.Age,
        "ttl":  ttl,
    }).Info("找到临时用户")
}

// 等待用户对象过期
time.Sleep(1500 * time.Millisecond)
if _, exists := userCache.Get("user:2"); !exists {
    log.Info("用户 2 的缓存已过期")
}
```

### 示例4：删除和清空操作

```go
// 设置缓存值
cache.Set("key1", "value1")
cache.Set("key2", "value2")

// 删除单个键
cache.Delete("key1")
if _, exists := cache.Get("key1"); !exists {
    log.Info("key1 已被删除")
}

// 清空所有缓存
cache.Clear()
if _, exists := cache.Get("key2"); !exists {
    log.Info("缓存已被清空")
}
```

## 配置选项

以下是创建自定义缓存实例时可用的配置选项：

```go
// 自定义配置
customCache, err := cache.NewCache(
    cache.WithNumCounters(1e5),    // 跟踪 10 万个条目
    cache.WithMaxCost(1<<28),      // 最大内存使用 256MB
    cache.WithBufferItems(64),     // 默认的缓冲区大小
)
```

### 配置说明

- `NumCounters`：缓存跟踪的最大条目数，建议设置为预期独特条目数的 10 倍
- `MaxCost`：缓存的最大成本（可理解为最大内存使用量，单位为字节）
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
   - 总是检查 InitCache 和 NewCache 的返回错误
   - 使用 defer Close() 确保资源释放
   - 检查 Get 操作的 exists 返回值

## 贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个包。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../LICENSE) 文件。 