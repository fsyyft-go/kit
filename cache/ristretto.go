// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cache

import (
	"time"

	"github.com/dgraph-io/ristretto"
)

// ristrettoCache 使用 Ristretto 实现 Cache 接口。
//
// ristrettoCache 的常规 Get、GetWithTTL、Set、SetWithTTL 和 Delete 继承 Ristretto 的并发能力。Clear 非原子，
// 调用方应避免将其与读写操作并发执行；Close 后不得继续使用缓存，且不应与其它操作并发调用。写入方法会在
// Set 或 SetWithTTL 后调用 Wait，确保返回前写入请求已从缓冲中处理，但不保证该缓存项最终通过准入策略并被保留。
type ristrettoCache struct {
	// cache 是底层 Ristretto 缓存实例，由 newRistrettoCache 创建并由 Close 释放。
	cache *ristretto.Cache
}

// Get 获取 key 对应的缓存值。
//
// 参数：
//   - key: 待查询的缓存键，具体可接受类型遵循 Ristretto 的键约束。
//
// 返回：
//   - value: 命中且未过期时返回缓存值；未命中或已过期时返回 nil。
//   - exists: key 存在且未过期时为 true。
func (c *ristrettoCache) Get(key interface{}) (interface{}, bool) {
	return c.cache.Get(key)
}

// GetWithTTL 获取 key 对应的缓存值及剩余过期时间。
//
// Ristretto 使用 0 表示缓存项永不过期，本方法会转换为 Cache 接口约定的 -1。
//
// 参数：
//   - key: 待查询的缓存键，具体可接受类型遵循 Ristretto 的键约束。
//
// 返回：
//   - value: 命中且未过期时返回缓存值；未命中或 TTL 查询失败时返回 nil。
//   - exists: key 存在、未过期且 TTL 查询成功时为 true。
//   - remainingTTL: 剩余过期时间，0 表示 key 不存在或已过期，-1 表示永不过期，正值表示实际剩余时间。
func (c *ristrettoCache) GetWithTTL(key interface{}) (interface{}, bool, time.Duration) {
	value, exists := c.cache.Get(key)
	if !exists {
		return nil, false, 0
	}

	// 获取剩余过期时间；若底层 TTL 查询失败，后续归一化会按缓存未命中处理。
	ttl, exists := c.cache.GetTTL(key)
	return normalizeRistrettoTTL(value, exists, ttl)
}

// normalizeRistrettoTTL 将 Ristretto 的 TTL 查询结果转换为 Cache 接口约定。
//
// 参数：
//   - value: 已从 Ristretto 读取到的缓存值。
//   - exists: Ristretto GetTTL 返回的存在状态。
//   - ttl: Ristretto GetTTL 返回的剩余过期时间；0 表示永不过期。
//
// 返回：
//   - interface{}: exists 为 true 时返回 value，否则返回 nil。
//   - bool: TTL 查询成功时为 true，查询失败时为 false。
//   - time.Duration: 0 表示不存在或已过期，-1 表示永不过期，正值表示实际剩余时间。
func normalizeRistrettoTTL(value interface{}, exists bool, ttl time.Duration) (interface{}, bool, time.Duration) {
	if !exists {
		// 如果获取 TTL 失败，说明键不存在或已过期。
		return nil, false, 0
	}

	// 如果 TTL 为 0，表示永不过期。
	if ttl == 0 {
		return value, true, -1
	}

	return value, true, ttl
}

// Set 写入永不过期的缓存值。
//
// 本实现写入时传入成本 1，但未设置 IgnoreInternalCost，Ristretto 可能计入内部存储成本；
// 因此 CacheOptions.MaxCost 表示底层成本容量，不能作为严格最大条目数。
//
// 参数：
//   - key: 待写入的缓存键，具体可接受类型遵循 Ristretto 的键约束。
//   - value: 待缓存的值，可以为任意 Ristretto 支持保存的类型。
//
// 返回：
//   - bool: Ristretto 未立即丢弃并将该写入请求排入缓冲时返回 true；返回 true 后仍可能被准入策略拒绝。
func (c *ristrettoCache) Set(key interface{}, value interface{}) bool {
	ok := c.cache.Set(key, value, 1)
	// 等待缓冲写入请求被处理；是否通过准入策略并最终可被 Get 命中仍由 Ristretto 决定。
	c.cache.Wait()
	return ok
}

// SetWithTTL 写入带过期时间的缓存值。
//
// ttl 小于等于 0 时使用永不过期写入；ttl 大于 0 时使用 Ristretto 的 TTL 写入。本实现写入时传入成本 1，
// 但未设置 IgnoreInternalCost，Ristretto 可能计入内部存储成本；因此 CacheOptions.MaxCost 表示底层成本容量，
// 不能作为严格最大条目数。
//
// 参数：
//   - key: 待写入的缓存键，具体可接受类型遵循 Ristretto 的键约束。
//   - value: 待缓存的值，可以为任意 Ristretto 支持保存的类型。
//   - ttl: 缓存有效期；ttl 小于等于 0 时表示永不过期。
//
// 返回：
//   - bool: Ristretto 未立即丢弃并将该写入请求排入缓冲时返回 true；返回 true 后仍可能被准入策略拒绝。
func (c *ristrettoCache) SetWithTTL(key interface{}, value interface{}, ttl time.Duration) bool {
	var ok bool
	// 如果 ttl <= 0，则表示永不过期。
	if ttl <= 0 {
		ok = c.cache.Set(key, value, 1)
	} else {
		// 设置带正 TTL 的缓存项，返回值仍只表示写入请求是否进入缓冲。
		ok = c.cache.SetWithTTL(key, value, 1, ttl)
	}
	// 等待缓冲写入请求被处理；是否通过准入策略并最终可被 Get 命中仍由 Ristretto 决定。
	c.cache.Wait()
	return ok
}

// Delete 删除 key 对应的缓存项。
//
// 参数：
//   - key: 待删除的缓存键，具体可接受类型遵循 Ristretto 的键约束；key 不存在时该操作无效果。
func (c *ristrettoCache) Delete(key interface{}) {
	c.cache.Del(key)
}

// Clear 清空当前 Ristretto 缓存中的所有缓存项。
//
// Ristretto 的 Clear 非原子；调用方应避免将其与 Get、Set、Delete 等读写操作并发执行。
//
// 参数：无。
func (c *ristrettoCache) Clear() {
	c.cache.Clear()
}

// Close 关闭底层 Ristretto 缓存并释放相关资源。
//
// Close 会停止底层 goroutine 并关闭通道，不应与其它缓存操作并发调用；关闭后的缓存不应继续使用。
//
// 参数：无。
//
// 返回：
//   - error: 当前实现始终返回 nil。
func (c *ristrettoCache) Close() error {
	c.cache.Close()
	return nil
}

// newRistrettoCache 创建基于 Ristretto 的 Cache 实例。
//
// 参数：
//   - options: Ristretto 初始化配置，NumCounters、MaxCost 和 BufferItems 通常应使用正值；0 会导致当前 Ristretto
//     初始化失败，其它非法值由底层 Ristretto 返回错误或决定具体表现。
//
// 返回：
//   - Cache: 创建成功后的 Ristretto 缓存实现。
//   - error: Ristretto 初始化失败时返回错误，通常由无效配置触发。
func newRistrettoCache(options CacheOptions) (Cache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: options.NumCounters,
		MaxCost:     options.MaxCost,
		BufferItems: options.BufferItems,
	})
	if nil != err {
		return nil, err
	}

	return &ristrettoCache{
		cache: cache,
	}, nil
}
