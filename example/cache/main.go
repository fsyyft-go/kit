// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package main

import (
	"time"

	"github.com/fsyyft-go/kit/cache"
	"github.com/fsyyft-go/kit/log"
)

// User 定义了一个示例结构体，用于展示类型安全的缓存操作。
type User struct {
	ID   int
	Name string
	Age  int
}

func main() {
	// 初始化日志
	if err := log.InitLogger(); err != nil {
		panic(err)
	}

	// 示例1：使用全局缓存。
	if err := cache.InitCache(); err != nil {
		log.WithField("error", err).Fatal("初始化缓存失败")
	}
	defer cache.Close() //nolint:errcheck

	log.Info("开始缓存示例演示...")

	// 基本的缓存操作。
	cache.Set("string_key", "这是一个字符串值")
	cache.Set("int_key", 42)
	cache.SetWithTTL("temp_key", "这个值将在 1 秒后过期", time.Second)

	// 获取缓存值。
	if value, exists := cache.Get("string_key"); exists {
		log.WithField("value", value).Info("获取 string_key 的值")
	}

	if value, exists := cache.Get("int_key"); exists {
		log.WithField("value", value).Info("获取 int_key 的值")
	}

	// 获取带 TTL 的缓存值。
	if value, exists, ttl := cache.GetWithTTL("temp_key"); exists {
		log.WithFields(map[string]interface{}{
			"value": value,
			"ttl":   ttl,
		}).Info("获取 temp_key 的值")
	}

	// 等待一段时间后检查过期的值。
	time.Sleep(1500 * time.Millisecond)
	if _, exists := cache.Get("temp_key"); !exists {
		log.Info("temp_key 已经过期")
	}

	// 示例2：使用自定义配置的缓存实例。
	customCache, err := cache.NewCache(
		cache.WithNumCounters(1e5), // 跟踪 10 万个条目
		cache.WithMaxCost(1<<28),   // 最大内存使用 256MB
		cache.WithBufferItems(64),  // 默认的缓冲区大小
	)
	if err != nil {
		log.WithField("error", err).Fatal("创建自定义缓存失败")
	}
	defer customCache.Close() //nolint:errcheck

	// 使用自定义缓存实例。
	customCache.Set("custom_key", "这是自定义缓存中的值")
	if value, exists := customCache.Get("custom_key"); exists {
		log.WithField("value", value).Info("获取 custom_key 的值")
	}

	// 示例3：使用类型安全的缓存。
	log.Info("开始类型安全的缓存操作示例...")
	// 创建一个专门用于存储 User 类型的缓存包装器。
	userCache := cache.AsTypedCache[User](customCache)

	// 存储用户对象。
	user := User{
		ID:   1,
		Name: "张三",
		Age:  30,
	}
	userCache.Set("user:1", user)

	// 获取用户对象（无需类型断言）。
	if user, exists := userCache.Get("user:1"); exists {
		log.WithFields(map[string]interface{}{
			"id":   user.ID,
			"name": user.Name,
			"age":  user.Age,
		}).Info("找到用户")
	}

	// 使用 TTL 存储用户对象。
	tempUser := User{
		ID:   2,
		Name: "李四",
		Age:  25,
	}
	userCache.SetWithTTL("user:2", tempUser, time.Second)

	// 获取带 TTL 的用户对象。
	if user, exists, ttl := userCache.GetWithTTL("user:2"); exists {
		log.WithFields(map[string]interface{}{
			"id":   user.ID,
			"name": user.Name,
			"age":  user.Age,
			"ttl":  ttl,
		}).Info("找到临时用户")
	}

	// 等待用户对象过期。
	time.Sleep(1500 * time.Millisecond)
	if _, exists := userCache.Get("user:2"); !exists {
		log.Info("用户 2 的缓存已过期")
	}

	// 示例4：删除和清空操作。
	log.Info("开始删除和清空操作示例...")
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// 删除单个键。
	cache.Delete("key1")
	if _, exists := cache.Get("key1"); !exists {
		log.Info("key1 已被删除")
	}

	// 清空所有缓存。
	cache.Clear()
	if _, exists := cache.Get("key2"); !exists {
		log.Info("缓存已被清空")
	}

	log.Info("缓存示例演示完成")
}
