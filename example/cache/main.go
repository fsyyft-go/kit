// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package main

import (
	"time"

	kitcache "github.com/fsyyft-go/kit/cache"
	kitlog "github.com/fsyyft-go/kit/log"
)

// User 定义了一个示例结构体，用于展示类型安全的缓存操作。
type User struct {
	ID   int
	Name string
	Age  int
}

func main() {
	// 初始化日志
	if err := kitlog.InitLogger(); err != nil {
		panic(err)
	}

	// 示例1：使用全局缓存。
	if err := kitcache.InitCache(); err != nil {
		kitlog.WithField("error", err).Fatal("初始化缓存失败")
	}
	defer kitcache.Close() //nolint:errcheck

	kitlog.Info("开始缓存示例演示...")

	// 基本的缓存操作。
	kitcache.Set("string_key", "这是一个字符串值")
	kitcache.Set("int_key", 42)
	kitcache.SetWithTTL("temp_key", "这个值将在 1 秒后过期", time.Second)

	// 获取缓存值。
	if value, exists := kitcache.Get("string_key"); exists {
		kitlog.WithField("value", value).Info("获取 string_key 的值")
	}

	if value, exists := kitcache.Get("int_key"); exists {
		kitlog.WithField("value", value).Info("获取 int_key 的值")
	}

	// 获取带 TTL 的缓存值。
	if value, exists, ttl := kitcache.GetWithTTL("temp_key"); exists {
		kitlog.WithFields(map[string]interface{}{
			"value": value,
			"ttl":   ttl,
		}).Info("获取 temp_key 的值")
	}

	// 等待一段时间后检查过期的值。
	time.Sleep(1500 * time.Millisecond)
	if _, exists := kitcache.Get("temp_key"); !exists {
		kitlog.Info("temp_key 已经过期")
	}

	// 示例2：使用自定义配置的缓存实例。
	customCache, err := kitcache.NewCache(
		kitcache.WithNumCounters(1e5), // 跟踪 10 万个条目
		kitcache.WithMaxCost(1<<28),   // 最大内存使用 256MB
		kitcache.WithBufferItems(64),  // 默认的缓冲区大小
	)
	if err != nil {
		kitlog.WithField("error", err).Fatal("创建自定义缓存失败")
	}
	defer customCache.Close() //nolint:errcheck

	// 使用自定义缓存实例。
	customCache.Set("custom_key", "这是自定义缓存中的值")
	if value, exists := customCache.Get("custom_key"); exists {
		kitlog.WithField("value", value).Info("获取 custom_key 的值")
	}

	// 示例3：使用类型安全的缓存。
	kitlog.Info("开始类型安全的缓存操作示例...")
	// 创建一个专门用于存储 User 类型的缓存包装器。
	userCache := kitcache.AsTypedCache[User](customCache)

	// 存储用户对象。
	user := User{
		ID:   1,
		Name: "张三",
		Age:  30,
	}
	userCache.Set("user:1", user)

	// 获取用户对象（无需类型断言）。
	if user, exists := userCache.Get("user:1"); exists {
		kitlog.WithFields(map[string]interface{}{
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
		kitlog.WithFields(map[string]interface{}{
			"id":   user.ID,
			"name": user.Name,
			"age":  user.Age,
			"ttl":  ttl,
		}).Info("找到临时用户")
	}

	// 等待用户对象过期。
	time.Sleep(1500 * time.Millisecond)
	if _, exists := userCache.Get("user:2"); !exists {
		kitlog.Info("用户 2 的缓存已过期")
	}

	// 示例4：删除和清空操作。
	kitlog.Info("开始删除和清空操作示例...")
	kitcache.Set("key1", "value1")
	kitcache.Set("key2", "value2")

	// 删除单个键。
	kitcache.Delete("key1")
	if _, exists := kitcache.Get("key1"); !exists {
		kitlog.Info("key1 已被删除")
	}

	// 清空所有缓存。
	kitcache.Clear()
	if _, exists := kitcache.Get("key2"); !exists {
		kitlog.Info("缓存已被清空")
	}

	kitlog.Info("缓存示例演示完成")
}
