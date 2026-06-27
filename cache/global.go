// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cache

import (
	"sync"
	"time"
)

var (
	// defaultCache 是包级默认缓存实例。
	//
	// defaultCache 由 InitCache 首次初始化，之后被 Get、Set、Delete、Clear 和 Close 等包级函数复用。
	defaultCache Cache

	// once 控制 defaultCache 只执行一次初始化流程。
	//
	// once 使 InitCache 的首次调用固定默认缓存配置，后续调用不会重新创建或替换 defaultCache。
	once sync.Once
)

// InitCache 初始化包级默认缓存实例。
//
// InitCache 通过 sync.Once 只尝试初始化一次默认缓存。首次初始化失败时，只有当前调用会返回该错误；
// 后续调用不会重试，且当前实现也不会再次返回首次错误。无论首次是否成功，后续调用都不会替换默认实例，
// 也不会重新应用新的 Option。
//
// 参数：
//   - options: 用于首次初始化默认缓存的可选配置项；后续调用即使传入新选项也不会重新应用。
//
// 返回：
//   - error: 仅在首次初始化调用中返回 NewCache 失败产生的错误。
func InitCache(options ...Option) error {
	var err error
	once.Do(func() {
		defaultCache, err = NewCache(options...)
	})
	return err
}

// Get 从包级默认缓存中获取 key 对应的值。
//
// 参数：
//   - key: 待查询的缓存键，具体可接受类型由默认缓存实现决定。
//
// 返回：
//   - value: 默认缓存已初始化且 key 命中、未过期时返回缓存值；未初始化、未命中或已过期时返回 nil。
//   - exists: 默认缓存已初始化且 key 存在、未过期时为 true。
func Get(key interface{}) (interface{}, bool) {
	if nil == defaultCache {
		return nil, false
	}
	return defaultCache.Get(key)
}

// GetWithTTL 从包级默认缓存中获取 key 对应的值及剩余过期时间。
//
// 参数：
//   - key: 待查询的缓存键，具体可接受类型由默认缓存实现决定。
//
// 返回：
//   - value: 默认缓存已初始化且 key 命中、未过期时返回缓存值；未初始化、未命中或已过期时返回 nil。
//   - exists: 默认缓存已初始化且 key 存在、未过期时为 true。
//   - remainingTTL: 剩余过期时间，0 表示默认缓存未初始化、key 不存在或已过期，-1 表示永不过期，正值表示实际剩余时间。
func GetWithTTL(key interface{}) (interface{}, bool, time.Duration) {
	if nil == defaultCache {
		return nil, false, 0
	}
	return defaultCache.GetWithTTL(key)
}

// Set 向包级默认缓存写入永不过期的值。
//
// 参数：
//   - key: 待写入的缓存键，具体可接受类型由默认缓存实现决定。
//   - value: 待缓存的值，可以为任意默认缓存实现支持的类型。
//
// 返回：
//   - bool: 默认缓存已初始化且底层实现接受或排队该写入请求时返回 true；默认缓存未初始化时返回 false。
//     true 是否保证最终保留由底层实现决定。
func Set(key interface{}, value interface{}) bool {
	if nil == defaultCache {
		return false
	}
	return defaultCache.Set(key, value)
}

// SetWithTTL 向包级默认缓存写入带过期时间的值。
//
// 参数：
//   - key: 待写入的缓存键，具体可接受类型由默认缓存实现决定。
//   - value: 待缓存的值，可以为任意默认缓存实现支持的类型。
//   - ttl: 缓存有效期；ttl 小于等于 0 时表示永不过期。
//
// 返回：
//   - bool: 默认缓存已初始化且底层实现接受或排队该写入请求时返回 true；默认缓存未初始化时返回 false。
//     true 是否保证最终保留由底层实现决定。
func SetWithTTL(key interface{}, value interface{}, ttl time.Duration) bool {
	if nil == defaultCache {
		return false
	}
	return defaultCache.SetWithTTL(key, value, ttl)
}

// Delete 从包级默认缓存中删除 key 对应的缓存项。
//
// 参数：
//   - key: 待删除的缓存键，具体可接受类型由默认缓存实现决定；默认缓存未初始化或 key 不存在时该操作无效果。
func Delete(key interface{}) {
	if nil == defaultCache {
		return
	}
	defaultCache.Delete(key)
}

// Clear 清空包级默认缓存中的所有缓存项。
//
// 默认缓存未初始化时该操作无效果。默认缓存已初始化时，该操作委托底层 Cache；Clear 非原子，调用方应避免
// 将其与读写操作并发调用，除非底层实现明确提供更强保证。
//
// 参数：无。
func Clear() {
	if nil == defaultCache {
		return
	}
	defaultCache.Clear()
}

// Close 关闭包级默认缓存并释放相关资源。
//
// 默认缓存未初始化时 Close 直接返回 nil。Close 不会重置 sync.Once 或清空 defaultCache 引用，且不应与其它
// 默认缓存操作并发调用；关闭后的包级默认缓存不应继续用于读写操作。
//
// 参数：无。
//
// 返回：
//   - error: 默认缓存已初始化且底层 Close 失败时返回错误；默认缓存未初始化时返回 nil。
func Close() error {
	if nil == defaultCache {
		return nil
	}
	return defaultCache.Close()
}
