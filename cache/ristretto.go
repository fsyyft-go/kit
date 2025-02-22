// Copyright 2024 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cache

import (
	"time"

	"github.com/dgraph-io/ristretto"
)

// ristrettoCache 是使用 ristretto 实现的缓存。
// ristretto 是一个高性能的内存缓存，具有以下特点：
//   - 高并发性能
//   - 自动内存管理
//   - 基于 LFU（最不经常使用）的驱逐策略
//   - 支持 TTL
type ristrettoCache struct {
	// cache 是底层的 ristretto 缓存实例。
	cache *ristretto.Cache
}

// newRistrettoCache 创建一个新的 Ristretto 缓存实例。
// 默认使用 ristretto 作为缓存后端，它提供了高性能和可靠的缓存实现。
// 参数：
//   - options：缓存配置，包括计数器数量、最大成本和缓冲区大小
//
// 返回值：
//   - Cache：缓存接口实现
//   - error：如果创建失败则返回错误
func newRistrettoCache(options CacheOptions) (Cache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: options.NumCounters,
		MaxCost:     options.MaxCost,
		BufferItems: options.BufferItems,
	})
	if err != nil {
		return nil, err
	}

	return &ristrettoCache{
		cache: cache,
	}, nil
}

// Get 实现 Cache 接口的 Get 方法。
// 这个实现直接使用 ristretto 的 Get 方法，它是线程安全的。
func (c *ristrettoCache) Get(key interface{}) (interface{}, bool) {
	return c.cache.Get(key)
}

// GetWithTTL 实现 Cache 接口的 GetWithTTL 方法。
// 这个实现结合了 ristretto 的 Get 和 GetTTL 方法，提供了值和过期时间信息。
// 注意：ristretto 的 TTL 实现中，0 表示永不过期，这里会转换为 -1 以符合接口约定。
func (c *ristrettoCache) GetWithTTL(key interface{}) (interface{}, bool, time.Duration) {
	value, exists := c.cache.Get(key)
	if !exists {
		return nil, false, 0
	}

	// 获取剩余过期时间。
	ttl, exists := c.cache.GetTTL(key)
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

// Set 实现 Cache 接口的 Set 方法。
// 这个实现使用 ristretto 的 Set 方法，并等待写入完成。
// cost 参数固定为 1，因为我们使用 MaxCost 来表示最大条目数。
func (c *ristrettoCache) Set(key interface{}, value interface{}) bool {
	ok := c.cache.Set(key, value, 1)
	// 等待写入完成，确保后续的 Get 操作能够看到这次写入。
	c.cache.Wait()
	return ok
}

// SetWithTTL 实现 Cache 接口的 SetWithTTL 方法。
// 这个实现根据 ttl 参数选择使用 Set 或 SetWithTTL 方法。
// cost 参数固定为 1，因为我们使用 MaxCost 来表示最大条目数。
func (c *ristrettoCache) SetWithTTL(key interface{}, value interface{}, ttl time.Duration) bool {
	var ok bool
	// 如果 ttl <= 0，则表示永不过期。
	if ttl <= 0 {
		ok = c.cache.Set(key, value, 1)
	} else {
		// 设置带过期时间的缓存项。
		ok = c.cache.SetWithTTL(key, value, 1, ttl)
	}
	// 等待写入完成，确保后续的 Get 操作能够看到这次写入。
	c.cache.Wait()
	return ok
}

// Delete 实现 Cache 接口的 Delete 方法。
// 这个实现直接使用 ristretto 的 Del 方法，它是线程安全的。
func (c *ristrettoCache) Delete(key interface{}) {
	c.cache.Del(key)
}

// Clear 实现 Cache 接口的 Clear 方法。
// 这个实现直接使用 ristretto 的 Clear 方法，它是线程安全的。
func (c *ristrettoCache) Clear() {
	c.cache.Clear()
}

// Close 实现 Cache 接口的 Close 方法。
// 这个实现直接使用 ristretto 的 Close 方法，它会释放所有资源。
func (c *ristrettoCache) Close() error {
	c.cache.Close()
	return nil
}
