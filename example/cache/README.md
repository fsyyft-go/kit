 # 高性能缓存示例

本示例展示了如何使用 Kit 的缓存模块，实现高性能的进程内缓存功能，支持 TTL 过期时间和类型安全操作。

## 功能特性

- 全局缓存实例的基本操作（设置、获取、删除、清空）
- 自定义配置的缓存实例创建和使用
- 支持设置缓存项的过期时间（TTL）
- 类型安全的缓存操作（使用泛型）
- 线程安全的缓存访问

## 设计原理

本示例基于 ristretto 缓存库实现，具有以下核心设计特点：

- 统一的缓存接口：提供简洁一致的 API，支持多种缓存操作
- 全局缓存实例：方便在应用程序不同部分共享缓存
- 类型安全：利用 Go 泛型特性，提供类型安全的缓存操作，避免类型断言
- 资源管理：自动处理缓存资源的分配和释放
- 高性能：基于 LFU（最不经常使用）的驱逐策略，优化内存使用

## 使用方法

### 1. 编译和运行

在 Unix/Linux/macOS 系统上：

```bash
# 添加执行权限
chmod +x build.sh

# 构建和运行
./build.sh
```

### 2. 代码示例

#### 基本缓存操作

```go
// 初始化全局缓存
if err := cache.InitCache(); err != nil {
    log.WithField("error", err).Fatal("初始化缓存失败")
}
defer cache.Close()

// 设置缓存值
cache.Set("string_key", "这是一个字符串值")
cache.Set("int_key", 42)

// 设置带过期时间的缓存值
cache.SetWithTTL("temp_key", "这个值将在 1 秒后过期", time.Second)

// 获取缓存值
if value, exists := cache.Get("string_key"); exists {
    log.WithField("value", value).Info("获取 string_key 的值")
}

// 获取带 TTL 的缓存值
if value, exists, ttl := cache.GetWithTTL("temp_key"); exists {
    log.WithFields(map[string]interface{}{
        "value": value,
        "ttl":   ttl,
    }).Info("获取 temp_key 的值")
}

// 删除缓存项
cache.Delete("key1")

// 清空所有缓存
cache.Clear()
```

#### 自定义缓存配置

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
```

#### 类型安全的缓存操作

```go
// 定义结构体类型
type User struct {
    ID   int
    Name string
    Age  int
}

// 创建类型安全的缓存包装器
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
```

### 3. 输出示例

```
INFO[0000] 开始缓存示例演示...                             
INFO[0000] 获取 string_key 的值                        value="这是一个字符串值"
INFO[0000] 获取 int_key 的值                           value=42
INFO[0000] 获取 temp_key 的值                          ttl=1s value="这个值将在 1 秒后过期"
INFO[0001] temp_key 已经过期                            
INFO[0001] 获取 custom_key 的值                        value="这是自定义缓存中的值"
INFO[0001] 开始类型安全的缓存操作示例...                       
INFO[0001] 找到用户                                    age=30 id=1 name=张三
INFO[0001] 找到临时用户                                  age=25 id=2 name=李四 ttl=1s
INFO[0003] 用户 2 的缓存已过期                            
INFO[0003] 开始删除和清空操作示例...                         
INFO[0003] key1 已被删除                               
INFO[0003] 缓存已被清空                                 
INFO[0003] 缓存示例演示完成                              
```

### 4. 在其他项目中使用

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/fsyyft-go/kit/cache"
    "github.com/fsyyft-go/kit/log"
)

func main() {
    // 初始化日志
    if err := log.InitLogger(); err != nil {
        panic(err)
    }
    
    // 初始化缓存
    if err := cache.InitCache(); err != nil {
        log.WithField("error", err).Fatal("初始化缓存失败")
    }
    defer cache.Close()
    
    // 使用缓存存储会话数据
    cache.Set("session:123", map[string]interface{}{
        "user_id": 42,
        "role":    "admin",
        "created": time.Now(),
    })
    
    // 设置带过期时间的配置项
    cache.SetWithTTL("config:theme", "dark", time.Hour*24)
    
    // 获取缓存数据
    if session, exists := cache.Get("session:123"); exists {
        sessionData := session.(map[string]interface{})
        fmt.Printf("用户 ID: %v, 角色: %s\n", 
            sessionData["user_id"], 
            sessionData["role"])
    }
}
```

## 注意事项

- 缓存模块默认使用 ristretto 作为后端，它是一个基于 LFU（最不经常使用）策略的缓存实现
- 缓存容量受到 `WithMaxCost` 参数的限制，超出限制时会自动驱逐最不常用的项
- 对于需要频繁访问的数据，建议使用全局缓存实例
- 对于特定类型的数据，建议使用类型安全的缓存包装器，避免类型断言错误
- 在高并发场景下，适当增加 `BufferItems` 参数可以提高性能
- 缓存使用完毕后必须调用 `Close()` 方法释放资源

## 相关文档

- [Ristretto 缓存库](https://github.com/dgraph-io/ristretto)
- [Go 泛型介绍](https://go.dev/doc/tutorial/generics)
- [缓存模块 API 文档](../../cache/README.md)

## 许可证

本示例代码采用 MIT 许可证。详见 [LICENSE](../../LICENSE) 文件。