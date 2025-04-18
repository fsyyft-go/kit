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
	// maxCost 定义了默认的缓存最大成本，默认为 1GB。
	maxCost int64 = 1 << 30
	// bufferItems 定义了默认的写入操作缓冲大小，默认为 64。
	bufferItems int64 = 64
)

// 缓存的默认配置。
var ()

// Cache 定义了统一的缓存接口。
// 这个接口提供了基本的缓存操作功能，可以通过不同的实现来支持不同的缓存后端。
// 所有的实现都必须保证线程安全。
type Cache interface {
	// Get 获取缓存中的值。
	// 如果键不存在或已过期，exists 将返回 false。
	// 参数：
	//   - key：要获取的缓存键，可以是任意类型。
	// 返回值：
	//   - value：缓存的值，如果不存在则为 nil。
	//   - exists：值是否存在且未过期。
	Get(key interface{}) (value interface{}, exists bool)

	// GetWithTTL 获取缓存中的值及其剩余过期时间。
	// 如果键不存在或已过期，exists 将返回 false。
	// 参数：
	//   - key：要获取的缓存键，可以是任意类型。
	// 返回值：
	//   - value：缓存的值，如果不存在则为 nil。
	//   - exists：值是否存在且未过期。
	//   - remainingTTL：剩余过期时间：
	//     * 如果值不存在或已过期，返回 0。
	//     * 如果值永不过期，返回 -1。
	//     * 否则返回实际的剩余时间。
	GetWithTTL(key interface{}) (value interface{}, exists bool, remainingTTL time.Duration)

	// Set 设置缓存值，该值永不过期。
	// 参数：
	//   - key：缓存键，可以是任意类型。
	//   - value：要缓存的值，可以是任意类型。
	// 返回值：
	//   - bool：是否设置成功。
	Set(key interface{}, value interface{}) bool

	// SetWithTTL 设置带过期时间的缓存值。
	// 参数：
	//   - key：缓存键，可以是任意类型。
	//   - value：要缓存的值，可以是任意类型。
	//   - ttl：过期时间，如果 <= 0 则表示永不过期。
	// 返回值：
	//   - bool：是否设置成功。
	SetWithTTL(key interface{}, value interface{}, ttl time.Duration) bool

	// Delete 从缓存中删除指定的键。
	// 如果键不存在，该操作也会成功返回。
	// 删除后，该键的所有后续访问都将返回不存在。
	// 参数：
	//   - key：要删除的缓存键，可以是任意类型。
	Delete(key interface{})

	// Clear 清空缓存中的所有内容。
	// 这个操作会立即使所有缓存项失效。
	// 在清空后，所有之前的键都将返回不存在。
	Clear()

	// Close 关闭缓存，释放相关资源。
	// 关闭后的缓存不应该再被使用。
	// 重复调用 Close 是安全的。
	// 返回值：
	//   - error：如果关闭过程中发生错误则返回错误。
	Close() error
}

// CacheOptions 定义了缓存的配置选项。
// 这些配置项会影响缓存的性能和资源使用。
type CacheOptions struct {
	// NumCounters 定义了缓存跟踪的最大条目数。
	// 这个数字应该是预期独特条目数的大约 10 倍。
	// 例如，如果预计会有 1 万个不同的键，则应设置为 10 万。
	NumCounters int64

	// MaxCost 定义了缓存的最大成本。
	// 对于简单的缓存使用场景，这可以理解为最大条目数。
	// 当缓存达到这个限制时，最少使用的项目将被驱逐。
	MaxCost int64

	// BufferItems 定义了在写入操作时的缓冲大小。
	// 更大的缓冲区会导致更好的并发性能，但会使用更多的内存。
	// 对于大多数场景，默认值 64 是合适的。
	BufferItems int64
}

// Option 定义了缓存配置的函数选项。
// 这是一种函数式选项模式，用于灵活配置缓存参数。
type Option func(*CacheOptions)

// TypedCache 是一个泛型包装器，提供类型安全的缓存操作。
// 通过泛型参数 T 指定缓存值的类型，避免了手动类型断言。
// 这个包装器适用于需要类型安全的场景，例如：
//   - 存储特定类型的数据
//   - 避免运行时类型错误
//   - 提供更好的 IDE 支持
type TypedCache[T any] struct {
	// cache 是底层的缓存实现。
	cache Cache
}

// WithNumCounters 设置缓存跟踪的最大条目数。
// numCounters 应该是预期独特条目数的大约 10 倍。
// 参数：
//   - numCounters：要设置的计数器数量。
//
// 返回值：
//   - Option：用于配置缓存的函数选项。
func WithNumCounters(numCounters int64) Option {
	return func(opts *CacheOptions) {
		opts.NumCounters = numCounters
	}
}

// WithMaxCost 设置缓存的最大成本。
// 对于简单的缓存使用场景，这可以理解为最大条目数。
// 参数：
//   - maxCost：要设置的最大成本。
//
// 返回值：
//   - Option：用于配置缓存的函数选项。
func WithMaxCost(maxCost int64) Option {
	return func(opts *CacheOptions) {
		opts.MaxCost = maxCost
	}
}

// WithBufferItems 设置写入操作时的缓冲大小。
// 更大的缓冲区会导致更好的并发性能，但会使用更多的内存。
// 参数：
//   - bufferItems：要设置的缓冲区大小。
//
// 返回值：
//   - Option：用于配置缓存的函数选项。
func WithBufferItems(bufferItems int64) Option {
	return func(opts *CacheOptions) {
		opts.BufferItems = bufferItems
	}
}

// NewCache 创建一个新的缓存实例。
// 使用 Option 模式配置缓存参数，如果没有提供任何选项，将使用默认配置：
//   - NumCounters：1000 万（适合跟踪 100 万个不同的键）
//   - MaxCost：1GB（适合存储较大的数据集）
//   - BufferItems：64（提供良好的并发性能）
//
// 参数：
//   - options：可选的配置选项，如果不提供则使用默认配置。
//
// 返回值：
//   - Cache：缓存接口实现。
//   - error：如果创建失败则返回错误。
//
// 示例：
//
//	// 使用默认配置
//	cache, err := cache.NewCache()
//	if err != nil {
//	    panic(err)
//	}
//	defer cache.Close()
//
//	// 使用自定义配置
//	cache, err := cache.NewCache(
//	    cache.WithNumCounters(1e7),
//	    cache.WithMaxCost(1<<30),
//	    cache.WithBufferItems(64),
//	)
//	if err != nil {
//	    panic(err)
//	}
//	defer cache.Close()
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

// AsTypedCache 将缓存转换为类型安全的包装器。
// 参数：
//   - cache：底层的缓存实现。
//
// 返回值：
//   - *TypedCache[T]：类型安全的缓存包装器。
//
// 示例：
//
//	baseCache := cache.NewCache()
//	strCache := cache.AsTypedCache[string](baseCache)
//	intCache := cache.AsTypedCache[int](baseCache)
func AsTypedCache[T any](cache Cache) *TypedCache[T] {
	return &TypedCache[T]{
		cache: cache,
	}
}

// Get 获取缓存中的值，并进行类型转换。
// 如果键不存在、已过期或类型不匹配，exists 将返回 false。
// 返回的值已经是正确的类型，无需进行类型断言。
// 参数：
//   - key：要获取的缓存键，可以是任意类型。
//
// 返回值：
//   - value：缓存的值，已转换为类型 T。
//   - exists：值是否存在且未过期且类型匹配。
func (tc *TypedCache[T]) Get(key interface{}) (value T, exists bool) {
	if v, ok := tc.cache.Get(key); ok {
		if typed, ok := v.(T); ok {
			return typed, true
		}
	}
	return value, false
}

// GetWithTTL 获取缓存中的值及其剩余过期时间，并进行类型转换。
// 如果键不存在、已过期或类型不匹配，exists 将返回 false。
// 返回的值已经是正确的类型，无需进行类型断言。
// 参数：
//   - key：要获取的缓存键，可以是任意类型。
//
// 返回值：
//   - value：缓存的值，已转换为类型 T。
//   - exists：值是否存在且未过期且类型匹配。
//   - remainingTTL：剩余过期时间：
//   - 如果值不存在或已过期，返回 0。
//   - 如果值永不过期，返回 -1。
//   - 否则返回实际的剩余时间。
func (tc *TypedCache[T]) GetWithTTL(key interface{}) (value T, exists bool, remainingTTL time.Duration) {
	if v, ok, ttl := tc.cache.GetWithTTL(key); ok {
		if typed, ok := v.(T); ok {
			return typed, true, ttl
		}
	}
	return value, false, 0
}

// Set 设置缓存中的值。
// 这个方法会将值保存在缓存中，永不过期。
// 参数：
//   - key：缓存键，可以是任意类型。
//   - value：要缓存的值，类型为 T。
//
// 返回值：
//   - bool：是否设置成功。
func (tc *TypedCache[T]) Set(key interface{}, value T) bool {
	return tc.cache.Set(key, value)
}

// SetWithTTL 设置缓存中的值，带过期时间。
// 这个方法会将值保存在缓存中，在指定的过期时间后自动失效。
// 参数：
//   - key：缓存键，可以是任意类型。
//   - value：要缓存的值，类型为 T。
//   - ttl：过期时间，如果 <= 0 则表示永不过期。
//
// 返回值：
//   - bool：是否设置成功。
func (tc *TypedCache[T]) SetWithTTL(key interface{}, value T, ttl time.Duration) bool {
	return tc.cache.SetWithTTL(key, value, ttl)
}

// Delete 从缓存中删除指定的键。
// 这个方法会立即使指定的键失效。
// 参数：
//   - key：要删除的缓存键，可以是任意类型。
func (tc *TypedCache[T]) Delete(key interface{}) {
	tc.cache.Delete(key)
}

// Clear 清空缓存中的所有内容。
// 这个方法会立即使所有缓存项失效。
func (tc *TypedCache[T]) Clear() {
	tc.cache.Clear()
}

// Close 关闭缓存，释放相关资源。
// 这个方法会关闭底层的缓存实现。
// 返回值：
//   - error：如果关闭过程中发生错误则返回错误。
func (tc *TypedCache[T]) Close() error {
	return tc.cache.Close()
}
