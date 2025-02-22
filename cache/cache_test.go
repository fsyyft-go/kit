// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	// 创建缓存实例
	cache, err := NewCache()
	if err != nil {
		t.Fatalf("创建缓存失败: %v", err)
	}
	defer cache.Close()

	// 测试设置和获取
	t.Run("Set and Get", func(t *testing.T) {
		key := "test_key"
		value := "test_value"

		if !cache.Set(key, value) {
			t.Error("设置缓存失败")
		}

		// 测试不带 TTL 的 Get
		if val, exists := cache.Get(key); !exists {
			t.Error("获取缓存失败")
		} else if val != value {
			t.Errorf("缓存值不匹配，期望 %v，实际 %v", value, val)
		}

		// 测试带 TTL 的 Get
		if val, exists, ttl := cache.GetWithTTL(key); !exists {
			t.Error("获取缓存失败")
		} else if val != value {
			t.Errorf("缓存值不匹配，期望 %v，实际 %v", value, val)
		} else if ttl != -1 {
			t.Errorf("永不过期的缓存项的 TTL 应该为 -1，实际为 %v", ttl)
		}
	})

	// 测试过期时间
	t.Run("TTL", func(t *testing.T) {
		key := "ttl_key"
		value := "ttl_value"
		ttl := 100 * time.Millisecond

		if !cache.SetWithTTL(key, value, ttl) {
			t.Error("设置缓存失败")
		}

		// 测试不带 TTL 的 Get
		if val, exists := cache.Get(key); !exists {
			t.Error("获取缓存失败")
		} else if val != value {
			t.Errorf("缓存值不匹配，期望 %v，实际 %v", value, val)
		}

		// 测试带 TTL 的 Get
		if val, exists, remainingTTL := cache.GetWithTTL(key); !exists {
			t.Error("获取缓存失败")
		} else if val != value {
			t.Errorf("缓存值不匹配，期望 %v，实际 %v", value, val)
		} else if remainingTTL <= 0 {
			t.Errorf("剩余过期时间应该大于 0，实际为 %v", remainingTTL)
		}

		// 等待缓存过期
		time.Sleep(200 * time.Millisecond)

		// 测试不带 TTL 的 Get
		if _, exists := cache.Get(key); exists {
			t.Error("缓存应该已经过期")
		}

		// 测试带 TTL 的 Get
		if _, exists, _ := cache.GetWithTTL(key); exists {
			t.Error("缓存应该已经过期")
		}
	})

	// 测试删除
	t.Run("Delete", func(t *testing.T) {
		key := "delete_key"
		value := "delete_value"

		if !cache.Set(key, value) {
			t.Error("设置缓存失败")
		}

		cache.Delete(key)

		if _, exists := cache.Get(key); exists {
			t.Error("缓存应该已被删除")
		}
	})

	// 测试清空
	t.Run("Clear", func(t *testing.T) {
		keys := []string{"key1", "key2", "key3"}
		for _, key := range keys {
			if !cache.Set(key, "value") {
				t.Error("设置缓存失败")
			}
		}

		cache.Clear()

		for _, key := range keys {
			if _, exists := cache.Get(key); exists {
				t.Error("缓存应该已被清空")
			}
		}
	})
}

// TestGlobalCache 测试全局缓存的基本功能。
func TestGlobalCache(t *testing.T) {
	// 在测试开始时重置全局缓存
	defaultCache = nil
	once = sync.Once{}

	t.Run("Global Cache Operations", func(t *testing.T) {
		if err := InitCache(); err != nil {
			t.Fatalf("初始化全局缓存失败: %v", err)
		}
		defer func() {
			if err := Close(); err != nil {
				t.Errorf("关闭缓存失败: %v", err)
			}
		}()

		key := "test_key"
		value := "test_value"

		if !Set(key, value) {
			t.Error("设置全局缓存失败")
		}

		if val, exists := Get(key); !exists {
			t.Error("获取全局缓存失败")
		} else if val != value {
			t.Errorf("全局缓存值不匹配，期望 %v，实际 %v", value, val)
		}

		if val, exists, ttl := GetWithTTL(key); !exists {
			t.Error("获取全局缓存失败")
		} else if val != value {
			t.Errorf("全局缓存值不匹配，期望 %v，实际 %v", value, val)
		} else if ttl != -1 {
			t.Errorf("永不过期的缓存项的 TTL 应该为 -1，实际为 %v", ttl)
		}

		Delete(key)

		if _, exists := Get(key); exists {
			t.Error("全局缓存应该已被删除")
		}

		Clear()
	})
}

func TestTypedCache(t *testing.T) {
	// 创建基础缓存实例
	baseCache, err := NewCache()
	if err != nil {
		t.Fatalf("创建缓存失败: %v", err)
	}
	defer baseCache.Close()

	// 创建字符串类型的缓存
	strCache := AsTypedCache[string](baseCache)

	// 测试设置和获取
	t.Run("Set and Get String", func(t *testing.T) {
		key := "test_key"
		value := "test_value"

		if !strCache.Set(key, value) {
			t.Error("设置缓存失败")
		}

		// 测试不带 TTL 的 Get
		if val, exists := strCache.Get(key); !exists {
			t.Error("获取缓存失败")
		} else if val != value {
			t.Errorf("缓存值不匹配，期望 %v，实际 %v", value, val)
		}

		// 测试带 TTL 的 Get
		if val, exists, ttl := strCache.GetWithTTL(key); !exists {
			t.Error("获取缓存失败")
		} else if val != value {
			t.Errorf("缓存值不匹配，期望 %v，实际 %v", value, val)
		} else if ttl != -1 {
			t.Errorf("永不过期的缓存项的 TTL 应该为 -1，实际为 %v", ttl)
		}
	})

	// 测试类型不匹配
	t.Run("Type Mismatch", func(t *testing.T) {
		key := "int_key"
		value := 42

		// 使用基础缓存设置整数值
		if !baseCache.Set(key, value) {
			t.Error("设置缓存失败")
		}

		// 尝试使用字符串缓存获取整数值
		if _, exists := strCache.Get(key); exists {
			t.Error("应该无法获取类型不匹配的值")
		}

		if _, exists, _ := strCache.GetWithTTL(key); exists {
			t.Error("应该无法获取类型不匹配的值")
		}
	})

	// 测试过期时间
	t.Run("TTL", func(t *testing.T) {
		key := "ttl_key"
		value := "ttl_value"
		ttl := 100 * time.Millisecond

		if !strCache.SetWithTTL(key, value, ttl) {
			t.Error("设置缓存失败")
		}

		// 测试不带 TTL 的 Get
		if val, exists := strCache.Get(key); !exists {
			t.Error("获取缓存失败")
		} else if val != value {
			t.Errorf("缓存值不匹配，期望 %v，实际 %v", value, val)
		}

		// 测试带 TTL 的 Get
		if val, exists, remainingTTL := strCache.GetWithTTL(key); !exists {
			t.Error("获取缓存失败")
		} else if val != value {
			t.Errorf("缓存值不匹配，期望 %v，实际 %v", value, val)
		} else if remainingTTL <= 0 {
			t.Errorf("剩余过期时间应该大于 0，实际为 %v", remainingTTL)
		}

		// 等待缓存过期
		time.Sleep(200 * time.Millisecond)

		if _, exists := strCache.Get(key); exists {
			t.Error("缓存应该已经过期")
		}
	})

	// 测试结构体类型
	type Person struct {
		Name string
		Age  int
	}
	personCache := AsTypedCache[Person](baseCache)

	t.Run("Struct Type", func(t *testing.T) {
		key := "person_key"
		value := Person{Name: "Alice", Age: 30}

		if !personCache.Set(key, value) {
			t.Error("设置缓存失败")
		}

		if val, exists := personCache.Get(key); !exists {
			t.Error("获取缓存失败")
		} else if val != value {
			t.Errorf("缓存值不匹配，期望 %v，实际 %v", value, val)
		}
	})
}

// BenchmarkCache 对缓存的基本操作进行基准测试。
func BenchmarkCache(b *testing.B) {
	cache, err := NewCache()
	if err != nil {
		b.Fatalf("创建缓存失败: %v", err)
	}
	defer cache.Close()

	b.Run("Set", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key-%d", i)
			cache.Set(key, "value")
		}
	})

	b.Run("Get", func(b *testing.B) {
		// 预先设置一些数据
		key := "bench-key"
		cache.Set(key, "bench-value")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Get(key)
		}
	})

	b.Run("SetWithTTL", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("ttl-key-%d", i)
			cache.SetWithTTL(key, "value", time.Hour)
		}
	})

	b.Run("GetWithTTL", func(b *testing.B) {
		// 预先设置一些数据
		key := "bench-ttl-key"
		cache.SetWithTTL(key, "bench-value", time.Hour)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.GetWithTTL(key)
		}
	})

	b.Run("Delete", func(b *testing.B) {
		// 预先设置一些数据
		keys := make([]string, b.N)
		for i := 0; i < b.N; i++ {
			keys[i] = fmt.Sprintf("del-key-%d", i)
			cache.Set(keys[i], "value")
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Delete(keys[i])
		}
	})
}

// BenchmarkTypedCache 对类型安全缓存的基本操作进行基准测试。
func BenchmarkTypedCache(b *testing.B) {
	baseCache, err := NewCache()
	if err != nil {
		b.Fatalf("创建缓存失败: %v", err)
	}
	defer baseCache.Close()

	type User struct {
		ID   int
		Name string
	}

	cache := AsTypedCache[User](baseCache)
	testUser := User{ID: 1, Name: "test"}

	b.Run("Set", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("user-%d", i)
			cache.Set(key, testUser)
		}
	})

	b.Run("Get", func(b *testing.B) {
		// 预先设置一些数据
		key := "bench-user"
		cache.Set(key, testUser)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.Get(key)
		}
	})

	b.Run("SetWithTTL", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("ttl-user-%d", i)
			cache.SetWithTTL(key, testUser, time.Hour)
		}
	})

	b.Run("GetWithTTL", func(b *testing.B) {
		// 预先设置一些数据
		key := "bench-ttl-user"
		cache.SetWithTTL(key, testUser, time.Hour)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.GetWithTTL(key)
		}
	})
}

// BenchmarkGlobalCache 对全局缓存的基本操作进行基准测试。
func BenchmarkGlobalCache(b *testing.B) {
	// 在每个子测试前重新初始化全局缓存
	b.Run("Set", func(b *testing.B) {
		if err := InitCache(); err != nil {
			b.Fatalf("初始化全局缓存失败: %v", err)
		}
		defer func() {
			if err := Close(); err != nil {
				b.Errorf("关闭缓存失败: %v", err)
			}
		}()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("global-%d", i)
			Set(key, "value")
		}
	})

	b.Run("Get", func(b *testing.B) {
		if err := InitCache(); err != nil {
			b.Fatalf("初始化全局缓存失败: %v", err)
		}
		defer func() {
			if err := Close(); err != nil {
				b.Errorf("关闭缓存失败: %v", err)
			}
		}()

		// 预先设置一些数据
		key := "bench-global"
		Set(key, "value")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Get(key)
		}
	})

	b.Run("SetWithTTL", func(b *testing.B) {
		if err := InitCache(); err != nil {
			b.Fatalf("初始化全局缓存失败: %v", err)
		}
		defer func() {
			if err := Close(); err != nil {
				b.Errorf("关闭缓存失败: %v", err)
			}
		}()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("global-ttl-%d", i)
			SetWithTTL(key, "value", time.Hour)
		}
	})

	b.Run("GetWithTTL", func(b *testing.B) {
		if err := InitCache(); err != nil {
			b.Fatalf("初始化全局缓存失败: %v", err)
		}
		defer func() {
			if err := Close(); err != nil {
				b.Errorf("关闭缓存失败: %v", err)
			}
		}()

		// 预先设置一些数据
		key := "bench-global-ttl"
		SetWithTTL(key, "value", time.Hour)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GetWithTTL(key)
		}
	})
}
