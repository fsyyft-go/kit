// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package cache 提供了一个统一的缓存接口和多种缓存实现。

主要特性：

  - 支持多种缓存后端（默认使用 Ristretto）
  - 提供统一的缓存接口
  - 支持泛型和类型安全
  - 支持 TTL（生存时间）设置
  - 支持全局缓存实例
  - 线程安全
  - 高性能并发操作
  - 内存使用优化

配置选项：

  - NumCounters：缓存跟踪的最大条目数（建议为预期条目数的 10 倍）
  - MaxCost：缓存的最大成本（可理解为最大条目数或最大内存使用量）
  - BufferItems：写入操作的缓冲大小（影响并发性能）

基本使用：

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

类型安全的缓存使用：

	// 创建类型安全的缓存包装器
	userCache := cache.AsTypedCache[User](cache)

	// 存储特定类型的值
	user := User{ID: 1, Name: "admin"}
	userCache.Set("user:1", user)

	// 获取特定类型的值（无需类型断言）
	if user, exists := userCache.Get("user:1"); exists {
	    fmt.Printf("用户名：%s\n", user.Name)
	}

自定义配置：

	// 使用自定义配置创建缓存
	cache, err := cache.NewCache(
	    cache.WithNumCounters(1e7),    // 1000万个计数器
	    cache.WithMaxCost(1<<30),      // 1GB 最大内存
	    cache.WithBufferItems(64),      // 64 个缓冲项
	)
	if err != nil {
	    panic(err)
	}
	defer cache.Close()

性能优化建议：

1. NumCounters 设置：
  - 设置为预期独特条目数的 10 倍
  - 例如：预期 100 万个键，设置为 1000 万

2. MaxCost 设置：
  - 根据可用内存和数据大小设置
  - 可以使用字节数或条目数作为度量

3. BufferItems 设置：
  - 影响写入性能和内存使用
  - 默认值 64 适合大多数场景
  - 增加可提高并发性能，但会使用更多内存

注意事项：

1. 在程序退出前调用 Close() 释放资源
2. 合理设置 TTL 避免内存泄漏
3. 注意类型安全包装器的使用场景
4. 监控缓存命中率和内存使用情况

更多示例请参考 example/cache 目录。
*/
package cache
