// Copyright 2024 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cache

import (
	"sync"
	"time"
)

var (
	// defaultCache 是默认的全局缓存实例。
	// 这个实例在首次调用 InitCache 时初始化，之后可以通过全局函数访问。
	defaultCache Cache

	// once 确保全局缓存只被初始化一次。
	// 使用 sync.Once 保证线程安全的单例模式实现。
	once sync.Once
)

// InitCache 初始化全局缓存实例。
// 这个函数使用 sync.Once 确保全局缓存只被初始化一次。
// 如果已经初始化过，该函数将不会执行任何操作。
//
// 参数：
//   - options：可选的配置选项，如果不提供则使用默认配置
//
// 返回值：
//   - error：如果初始化失败则返回错误
//
// 示例：
//
//	// 使用默认配置
//	if err := cache.InitCache(); err != nil {
//	    panic(err)
//	}
//	defer cache.Close()
//
//	// 使用自定义配置
//	if err := cache.InitCache(
//	    cache.WithNumCounters(1e7),
//	    cache.WithMaxCost(1<<30),
//	    cache.WithBufferItems(64),
//	); err != nil {
//	    panic(err)
//	}
//	defer cache.Close()
//
//	// 使用 CacheOptions 配置（迁移方式）
//	if err := cache.InitCache(cache.WithOptions(cache.DefaultConfig())...); err != nil {
//	    panic(err)
//	}
//	defer cache.Close()
func InitCache(options ...Option) error {
	var err error
	once.Do(func() {
		defaultCache, err = NewCache(options...)
	})
	return err
}

// Get 从全局缓存中获取值。
// 如果全局缓存未初始化，将返回 (nil, false)。
//
// 参数：
//   - key：要获取的缓存键
//
// 返回值：
//   - value：缓存的值，如果不存在则为 nil
//   - exists：值是否存在且未过期
func Get(key interface{}) (interface{}, bool) {
	if defaultCache == nil {
		return nil, false
	}
	return defaultCache.Get(key)
}

// GetWithTTL 从全局缓存中获取值及其剩余过期时间。
// 如果全局缓存未初始化，将返回 (nil, false, 0)。
//
// 参数：
//   - key：要获取的缓存键
//
// 返回值：
//   - value：缓存的值，如果不存在则为 nil
//   - exists：值是否存在且未过期
//   - remainingTTL：剩余过期时间：
//   - 如果值不存在或已过期，返回 0
//   - 如果值永不过期，返回 -1
//   - 否则返回实际的剩余时间
func GetWithTTL(key interface{}) (interface{}, bool, time.Duration) {
	if defaultCache == nil {
		return nil, false, 0
	}
	return defaultCache.GetWithTTL(key)
}

// Set 设置全局缓存中的值。
// 如果全局缓存未初始化，将返回 false。
//
// 参数：
//   - key：缓存键
//   - value：要缓存的值
//
// 返回值：
//   - bool：是否设置成功
func Set(key interface{}, value interface{}) bool {
	if defaultCache == nil {
		return false
	}
	return defaultCache.Set(key, value)
}

// SetWithTTL 设置全局缓存中的值，带过期时间。
// 如果全局缓存未初始化，将返回 false。
//
// 参数：
//   - key：缓存键
//   - value：要缓存的值
//   - ttl：过期时间，如果 <= 0 则表示永不过期
//
// 返回值：
//   - bool：是否设置成功
func SetWithTTL(key interface{}, value interface{}, ttl time.Duration) bool {
	if defaultCache == nil {
		return false
	}
	return defaultCache.SetWithTTL(key, value, ttl)
}

// Delete 从全局缓存中删除值。
// 如果全局缓存未初始化，该操作将被忽略。
//
// 参数：
//   - key：要删除的缓存键
func Delete(key interface{}) {
	if defaultCache == nil {
		return
	}
	defaultCache.Delete(key)
}

// Clear 清空全局缓存。
// 如果全局缓存未初始化，该操作将被忽略。
// 这个操作会立即使所有缓存项失效。
func Clear() {
	if defaultCache == nil {
		return
	}
	defaultCache.Clear()
}

// Close 关闭全局缓存。
// 如果全局缓存未初始化，将返回 nil。
// 关闭后的全局缓存不应该再被使用。
//
// 返回值：
//   - error：如果关闭过程中发生错误则返回错误
func Close() error {
	if defaultCache == nil {
		return nil
	}
	return defaultCache.Close()
}
