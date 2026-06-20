// Copyright 2025 fsyyft-go
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

// InitCache 初始化包级默认缓存实例。
//
// InitCache 通过 sync.Once 只尝试初始化一次默认缓存。首次初始化失败时，只有当前调用会返回该错误；
// 后续调用不会重试，且当前实现也不会再次返回首次错误。无论首次是否成功，后续调用都不会替换默认实例，
// 也不会重新应用新的 Option。
//
// 参数：
//   - options：用于首次初始化默认缓存的可选配置项；后续调用即使传入新选项也不会重新应用。
//
// 返回值：
//   - error：仅在首次初始化调用中返回 NewCache 失败产生的错误。
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
//   - key：要获取的缓存键，可以是任意类型。
//
// 返回值：
//   - value：缓存的值，如果不存在则为 nil。
//   - exists：值是否存在且未过期。
func Get(key interface{}) (interface{}, bool) {
	if nil == defaultCache {
		return nil, false
	}
	return defaultCache.Get(key)
}

// GetWithTTL 从全局缓存中获取值及其剩余过期时间。
// 如果全局缓存未初始化，将返回 (nil, false, 0)。
//
// 参数：
//   - key：要获取的缓存键，可以是任意类型。
//
// 返回值：
//   - value：缓存的值，如果不存在则为 nil。
//   - exists：值是否存在且未过期。
//   - remainingTTL：剩余过期时间：
//   - 如果值不存在或已过期，返回 0。
//   - 如果值永不过期，返回 -1。
//   - 否则返回实际的剩余时间。
func GetWithTTL(key interface{}) (interface{}, bool, time.Duration) {
	if nil == defaultCache {
		return nil, false, 0
	}
	return defaultCache.GetWithTTL(key)
}

// Set 设置全局缓存中的值。
// 如果全局缓存未初始化，将返回 false。
//
// 参数：
//   - key：缓存键，可以是任意类型。
//   - value：要缓存的值，可以是任意类型。
//
// 返回值：
//   - bool：是否设置成功。
func Set(key interface{}, value interface{}) bool {
	if nil == defaultCache {
		return false
	}
	return defaultCache.Set(key, value)
}

// SetWithTTL 设置全局缓存中的值，带过期时间。
// 如果全局缓存未初始化，将返回 false。
//
// 参数：
//   - key：缓存键，可以是任意类型。
//   - value：要缓存的值，可以是任意类型。
//   - ttl：过期时间，如果 <= 0 则表示永不过期。
//
// 返回值：
//   - bool：是否设置成功。
func SetWithTTL(key interface{}, value interface{}, ttl time.Duration) bool {
	if nil == defaultCache {
		return false
	}
	return defaultCache.SetWithTTL(key, value, ttl)
}

// Delete 从全局缓存中删除值。
// 如果全局缓存未初始化，该操作将被忽略。
//
// 参数：
//   - key：要删除的缓存键，可以是任意类型。
func Delete(key interface{}) {
	if nil == defaultCache {
		return
	}
	defaultCache.Delete(key)
}

// Clear 清空全局缓存。
// 如果全局缓存未初始化，该操作将被忽略。
// 这个操作会立即使所有缓存项失效。
func Clear() {
	if nil == defaultCache {
		return
	}
	defaultCache.Clear()
}

// Close 关闭全局缓存。
// 如果全局缓存未初始化，将返回 nil。
// 关闭后的全局缓存不应该再被使用。
//
// 返回值：
//   - error：如果关闭过程中发生错误则返回错误。
func Close() error {
	if nil == defaultCache {
		return nil
	}
	return defaultCache.Close()
}
