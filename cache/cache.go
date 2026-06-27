// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cache

import (
	"time"
)

const (
	// numCounters 定义了默认的缓存跟踪最大条目数，默认为 1000 万。
	numCounters int64 = 1e7
	// maxCost 定义默认的 Ristretto 成本容量，默认值为 1 << 30。
	maxCost int64 = 1 << 30
	// bufferItems 定义了默认的写入操作缓冲大小，默认为 64。
	bufferItems int64 = 64
)

// 缓存的默认配置。
var ()

// Cache 定义缓存访问接口。
//
// 常规 Get、GetWithTTL、Set、SetWithTTL 和 Delete 操作遵循具体实现的并发能力；本包内置实现使用
// Ristretto 作为后端。Clear 非原子，调用方应避免将其与读写操作并发执行；Close 属于生命周期操作，
// 关闭后不得继续使用缓存，且不应与其它操作并发调用。键和值的可接受类型由具体实现决定。
type Cache interface {
	// Get 获取 key 对应的缓存值。
	//
	// 参数：
	//   - key: 待查询的缓存键，具体可接受类型由实现决定。
	//
	// 返回：
	//   - value: 命中且未过期时返回缓存值；未命中或已过期时返回 nil。
	//   - exists: key 存在且未过期时为 true。
	Get(key interface{}) (value interface{}, exists bool)

	// GetWithTTL 获取 key 对应的缓存值及剩余过期时间。
	//
	// 参数：
	//   - key: 待查询的缓存键，具体可接受类型由实现决定。
	//
	// 返回：
	//   - value: 命中且未过期时返回缓存值；未命中或已过期时返回 nil。
	//   - exists: key 存在且未过期时为 true。
	//   - remainingTTL: 剩余过期时间，0 表示 key 不存在或已过期，-1 表示永不过期，正值表示实际剩余时间。
	GetWithTTL(key interface{}) (value interface{}, exists bool, remainingTTL time.Duration)

	// Set 写入永不过期的缓存值。
	//
	// 参数：
	//   - key: 待写入的缓存键，具体可接受类型由实现决定。
	//   - value: 待缓存的值，可以为任意实现支持的类型。
	//
	// 返回：
	//   - bool: 底层实现接受或排队该写入请求时返回 true；是否最终保留由具体实现决定。
	Set(key interface{}, value interface{}) bool

	// SetWithTTL 写入带过期时间的缓存值。
	//
	// 参数：
	//   - key: 待写入的缓存键，具体可接受类型由实现决定。
	//   - value: 待缓存的值，可以为任意实现支持的类型。
	//   - ttl: 缓存有效期；ttl 小于等于 0 时表示永不过期。
	//
	// 返回：
	//   - bool: 底层实现接受或排队该写入请求时返回 true；是否最终保留由具体实现决定。
	SetWithTTL(key interface{}, value interface{}, ttl time.Duration) bool

	// Delete 删除 key 对应的缓存项。
	//
	// 参数：
	//   - key: 待删除的缓存键，具体可接受类型由实现决定；key 不存在时该操作无效果。
	Delete(key interface{})

	// Clear 清空当前缓存中的所有缓存项。
	//
	// Clear 非原子；调用方应避免将其与读写操作并发调用，除非具体实现明确提供更强保证。
	//
	// 参数：无。
	Clear()

	// Close 关闭缓存并释放相关资源。
	//
	// Close 不应与其它缓存操作并发调用，关闭后的缓存实例也不应继续用于读写操作。
	//
	// 参数：无。
	//
	// 返回：
	//   - error: 关闭底层资源失败时返回错误；具体错误语义由实现决定。
	Close() error
}

// CacheOptions 定义创建缓存实例时使用的容量与缓冲配置。
//
// 这些配置会直接传递给底层 Ristretto 构造过程。零值 CacheOptions 会使当前 Ristretto 初始化失败，
// 调用方通常应通过 NewCache 的默认值或 WithNumCounters、WithMaxCost、WithBufferItems 组合生成正值配置；
// 其它非法值由底层 Ristretto 返回错误或决定具体表现。
type CacheOptions struct {
	// NumCounters 指定用于跟踪访问频率的计数器数量，通常应约为预期最大独立键数量的 10 倍。
	//
	// 该值应使用正值；0 会导致当前 Ristretto 初始化失败。值越大，驱逐准确度通常越高，但会占用更多内存。
	NumCounters int64

	// MaxCost 指定缓存允许的最大成本。
	//
	// 本包内置实现写入每个缓存项时传入成本 1，但未设置 IgnoreInternalCost，底层可能计入内部存储成本，
	// 因此该值不能作为严格最大条目数。该值应使用正值；0 会导致当前 Ristretto 初始化失败。
	MaxCost int64

	// BufferItems 指定 Ristretto 读写缓冲使用的条目数量。
	//
	// 该值应使用正值；0 会导致当前 Ristretto 初始化失败。较大的缓冲区可能提升并发性能，但会增加内存使用。
	BufferItems int64
}

// Option 定义修改 CacheOptions 的函数式选项。
//
// 参数：
//   - *CacheOptions: 待修改的缓存配置实例，NewCache 和 InitCache 在应用选项时传入非 nil 指针。
type Option func(*CacheOptions)

// TypedCache 在 Cache 上提供泛型类型安全包装。
//
// TypedCache 与底层 Cache 共享同一批缓存项和生命周期。零值 TypedCache 不可直接使用；调用方应通过
// AsTypedCache 传入非 nil Cache 创建实例。读取时如果底层值无法断言为 T，会按缓存未命中处理。
type TypedCache[T any] struct {
	// cache 是底层缓存实现，必须由 AsTypedCache 注入非 nil 值。
	cache Cache
}

// WithNumCounters 设置缓存跟踪访问频率使用的计数器数量。
//
// 参数：
//   - numCounters: 计数器数量，通常应使用正值并约为预期最大独立键数量的 10 倍；0 会导致当前 Ristretto 初始化失败。
//
// 返回：
//   - Option: 应用于 CacheOptions.NumCounters 的函数式选项。
func WithNumCounters(numCounters int64) Option {
	return func(opts *CacheOptions) {
		opts.NumCounters = numCounters
	}
}

// WithMaxCost 设置缓存允许的最大成本。
//
// 参数：
//   - maxCost: Ristretto 成本容量，通常应使用正值；0 会导致当前 Ristretto 初始化失败。内置实现写入成本为 1，
//     但底层可能计入内部存储成本，因此该值不能作为严格最大条目数。
//
// 返回：
//   - Option: 应用于 CacheOptions.MaxCost 的函数式选项。
func WithMaxCost(maxCost int64) Option {
	return func(opts *CacheOptions) {
		opts.MaxCost = maxCost
	}
}

// WithBufferItems 设置 Ristretto 读写缓冲使用的条目数量。
//
// 参数：
//   - bufferItems: 缓冲条目数量，通常应使用正值；0 会导致当前 Ristretto 初始化失败，默认值 64 适用于大多数场景。
//
// 返回：
//   - Option: 应用于 CacheOptions.BufferItems 的函数式选项。
func WithBufferItems(bufferItems int64) Option {
	return func(opts *CacheOptions) {
		opts.BufferItems = bufferItems
	}
}

// NewCache 使用当前内置的 Ristretto 后端创建独立缓存实例。
//
// 未提供 Option 时会使用包内默认的 NumCounters、MaxCost 和 BufferItems。多个 Option 会按传入顺序应用，
// 后传入的选项可以覆盖先前写入的同一字段。调用方在实例不再使用时应调用 Close。
//
// 参数：
//   - options: 可选配置项；为空时使用默认的 NumCounters、MaxCost 和 BufferItems。
//
// 返回：
//   - Cache: 创建成功后的缓存实例。
//   - error: 底层 Ristretto 初始化失败时返回错误，通常由无效配置触发。
func NewCache(options ...Option) (Cache, error) {
	// 使用默认配置
	opts := &CacheOptions{
		NumCounters: numCounters,
		MaxCost:     maxCost,
		BufferItems: bufferItems,
	}

	// 应用自定义选项
	for _, option := range options {
		option(opts)
	}

	// 创建缓存实例
	return newRistrettoCache(*opts)
}

// AsTypedCache 将已有 Cache 包装为类型安全缓存。
//
// 包装器不复制数据，也不改变底层缓存的关闭责任；类型安全读取仅在取出的值可断言为 T 时命中。
//
// 参数：
//   - cache: 待包装的底层缓存实例，调用方应保证其非 nil。
//
// 返回：
//   - *TypedCache[T]: 与 cache 共享存储和生命周期的类型安全缓存包装器。
func AsTypedCache[T any](cache Cache) *TypedCache[T] {
	return &TypedCache[T]{
		cache: cache,
	}
}

// Get 获取 key 对应的 T 类型缓存值。
//
// 参数：
//   - key: 待查询的缓存键，具体可接受类型由底层 Cache 决定。
//
// 返回：
//   - value: 命中且类型匹配时返回缓存值；未命中、已过期或类型不匹配时返回 T 的零值。
//   - exists: key 存在、未过期且底层值可断言为 T 时为 true。
func (tc *TypedCache[T]) Get(key interface{}) (value T, exists bool) {
	if v, ok := tc.cache.Get(key); ok {
		if typed, ok := v.(T); ok {
			return typed, true
		}
	}
	return value, false
}

// GetWithTTL 获取 key 对应的 T 类型缓存值及剩余过期时间。
//
// 参数：
//   - key: 待查询的缓存键，具体可接受类型由底层 Cache 决定。
//
// 返回：
//   - value: 命中且类型匹配时返回缓存值；未命中、已过期或类型不匹配时返回 T 的零值。
//   - exists: key 存在、未过期且底层值可断言为 T 时为 true。
//   - remainingTTL: 剩余过期时间，0 表示 key 不存在、已过期或类型不匹配，-1 表示永不过期，正值表示实际剩余时间。
func (tc *TypedCache[T]) GetWithTTL(key interface{}) (value T, exists bool, remainingTTL time.Duration) {
	if v, ok, ttl := tc.cache.GetWithTTL(key); ok {
		if typed, ok := v.(T); ok {
			return typed, true, ttl
		}
	}
	return value, false, 0
}

// Set 写入永不过期的 T 类型缓存值。
//
// 参数：
//   - key: 待写入的缓存键，具体可接受类型由底层 Cache 决定。
//   - value: 待缓存的 T 类型值。
//
// 返回：
//   - bool: 底层 Cache 接受或排队该写入请求时返回 true；是否保证最终保留由底层实现决定。
func (tc *TypedCache[T]) Set(key interface{}, value T) bool {
	return tc.cache.Set(key, value)
}

// SetWithTTL 写入带过期时间的 T 类型缓存值。
//
// 参数：
//   - key: 待写入的缓存键，具体可接受类型由底层 Cache 决定。
//   - value: 待缓存的 T 类型值。
//   - ttl: 缓存有效期；ttl 小于等于 0 时表示永不过期。
//
// 返回：
//   - bool: 底层 Cache 接受或排队该写入请求时返回 true；是否保证最终保留由底层实现决定。
func (tc *TypedCache[T]) SetWithTTL(key interface{}, value T, ttl time.Duration) bool {
	return tc.cache.SetWithTTL(key, value, ttl)
}

// Delete 删除 key 对应的缓存项。
//
// 参数：
//   - key: 待删除的缓存键，具体可接受类型由底层 Cache 决定；key 不存在时该操作无效果。
func (tc *TypedCache[T]) Delete(key interface{}) {
	tc.cache.Delete(key)
}

// Clear 清空底层 Cache 中的所有缓存项。
//
// Clear 非原子；调用方应避免将其与读写操作并发调用，除非底层 Cache 明确提供更强保证。
//
// 参数：无。
func (tc *TypedCache[T]) Clear() {
	tc.cache.Clear()
}

// Close 关闭底层 Cache 并释放相关资源。
//
// Close 不会修改 TypedCache 自身状态，不应与其它缓存操作并发调用；关闭后的包装器不应继续用于读写操作。
//
// 参数：无。
//
// 返回：
//   - error: 底层 Cache 关闭失败时返回错误。
func (tc *TypedCache[T]) Close() error {
	return tc.cache.Close()
}
