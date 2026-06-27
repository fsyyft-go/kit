// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCacheOptionsApply 验证函数式选项会按预期写入 CacheOptions。
func TestCacheOptionsApply(t *testing.T) {
	tests := []struct {
		name        string
		description string
		option      Option
		want        CacheOptions
	}{
		{
			name:        "num counters",
			description: "WithNumCounters 应只覆盖 NumCounters 字段。",
			option:      WithNumCounters(128),
			want:        CacheOptions{NumCounters: 128},
		},
		{
			name:        "max cost",
			description: "WithMaxCost 应只覆盖 MaxCost 字段。",
			option:      WithMaxCost(256),
			want:        CacheOptions{MaxCost: 256},
		},
		{
			name:        "buffer items",
			description: "WithBufferItems 应只覆盖 BufferItems 字段。",
			option:      WithBufferItems(32),
			want:        CacheOptions{BufferItems: 32},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			var got CacheOptions
			tt.option(&got)

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestNewCacheWithCustomOptions 验证 NewCache 能使用自定义选项创建可读写缓存。
func TestNewCacheWithCustomOptions(t *testing.T) {
	cache := newTestCache(t)

	require.True(t, cache.Set("custom-option-key", "custom-option-value"))

	got, exists := cache.Get("custom-option-key")
	require.True(t, exists)
	assert.Equal(t, "custom-option-value", got)

	got, exists, ttl := cache.GetWithTTL("custom-option-key")
	require.True(t, exists)
	assert.Equal(t, "custom-option-value", got)
	assert.Equal(t, time.Duration(-1), ttl)
}

// TestNewCacheInvalidOptions 验证无效配置会使 NewCache 返回错误且不返回缓存实例。
func TestNewCacheInvalidOptions(t *testing.T) {
	tests := []struct {
		name        string
		description string
		options     []Option
	}{
		{
			name:        "zero num counters",
			description: "NumCounters 为 0 时 ristretto 应拒绝创建缓存。",
			options:     []Option{WithNumCounters(0)},
		},
		{
			name:        "zero max cost",
			description: "MaxCost 为 0 时 ristretto 应拒绝创建缓存。",
			options:     []Option{WithMaxCost(0)},
		},
		{
			name:        "zero buffer items",
			description: "BufferItems 为 0 时 ristretto 应拒绝创建缓存。",
			options:     []Option{WithBufferItems(0)},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			cache, err := NewCache(tt.options...)
			if cache != nil {
				t.Cleanup(func() {
					require.NoError(t, cache.Close())
				})
			}

			require.Error(t, err)
			assert.Nil(t, cache)
		})
	}
}

// TestGlobalCacheUninitializedDefaultBehavior 验证全局缓存未初始化时的安全默认行为。
func TestGlobalCacheUninitializedDefaultBehavior(t *testing.T) {
	resetGlobalCacheForTest(t)

	got, exists := Get("missing")
	assert.Nil(t, got)
	assert.False(t, exists)

	got, exists, ttl := GetWithTTL("missing")
	assert.Nil(t, got)
	assert.False(t, exists)
	assert.Zero(t, ttl)

	assert.False(t, Set("key", "value"))
	assert.False(t, SetWithTTL("key", "value", time.Second))
	assert.NotPanics(t, func() { Delete("key") })
	assert.NotPanics(t, Clear)
	assert.NoError(t, Close())
}

// TestGlobalCacheSetWithTTLAfterInit 验证全局缓存初始化后的 SetWithTTL 成功分支。
func TestGlobalCacheSetWithTTLAfterInit(t *testing.T) {
	resetGlobalCacheForTest(t)

	require.NoError(t, InitCache(testCacheOptions()...))
	require.True(t, SetWithTTL("global-ttl-key", "global-ttl-value", time.Hour))

	got, exists, ttl := GetWithTTL("global-ttl-key")
	require.True(t, exists)
	assert.Equal(t, "global-ttl-value", got)
	assert.Positive(t, ttl)
}

// TestInitCacheOnceIgnoresSecondInvalidConfiguration 验证 InitCache 只执行首次初始化。
func TestInitCacheOnceIgnoresSecondInvalidConfiguration(t *testing.T) {
	resetGlobalCacheForTest(t)

	require.NoError(t, InitCache(testCacheOptions()...))
	require.NotNil(t, defaultCache)
	firstCache := defaultCache

	require.True(t, Set("once-key", "once-value"))
	require.NoError(t, InitCache(WithNumCounters(0), WithMaxCost(0), WithBufferItems(0)))
	assert.True(t, defaultCache == firstCache, "InitCache 不应在第二次调用时替换全局缓存实例")

	got, exists := Get("once-key")
	require.True(t, exists)
	assert.Equal(t, "once-value", got)
}

// TestTypedCacheControlOperations 验证类型安全缓存会正确委托删除、清空和关闭操作。
func TestTypedCacheControlOperations(t *testing.T) {
	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "delete",
			description: "Delete 应删除目标键且不影响其他键。",
			assert: func(t *testing.T) {
				cache := AsTypedCache[string](newTestCache(t))
				require.True(t, cache.Set("delete-key", "delete-value"))
				require.True(t, cache.Set("keep-key", "keep-value"))

				cache.Delete("delete-key")

				got, exists := cache.Get("delete-key")
				assert.Empty(t, got)
				assert.False(t, exists)

				got, exists = cache.Get("keep-key")
				require.True(t, exists)
				assert.Equal(t, "keep-value", got)
			},
		},
		{
			name:        "clear",
			description: "Clear 应清空类型安全缓存中所有已写入键。",
			assert: func(t *testing.T) {
				cache := AsTypedCache[string](newTestCache(t))
				require.True(t, cache.Set("clear-key-1", "value-1"))
				require.True(t, cache.Set("clear-key-2", "value-2"))

				cache.Clear()

				got, exists := cache.Get("clear-key-1")
				assert.Empty(t, got)
				assert.False(t, exists)

				got, exists = cache.Get("clear-key-2")
				assert.Empty(t, got)
				assert.False(t, exists)
			},
		},
		{
			name:        "close",
			description: "Close 应关闭底层缓存且返回 nil 错误。",
			assert: func(t *testing.T) {
				cache, err := NewCache(testCacheOptions()...)
				require.NoError(t, err)

				typedCache := AsTypedCache[string](cache)
				assert.NoError(t, typedCache.Close())
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			tt.assert(t)
		})
	}
}

// TestTypedCacheGetWithTTLTypeMismatch 验证类型不匹配时 GetWithTTL 返回零值、false 和 0 TTL。
func TestTypedCacheGetWithTTLTypeMismatch(t *testing.T) {
	baseCache := newTestCache(t)
	require.True(t, baseCache.SetWithTTL("typed-mismatch-key", 42, time.Hour))

	cache := AsTypedCache[string](baseCache)
	got, exists, ttl := cache.GetWithTTL("typed-mismatch-key")
	assert.Empty(t, got)
	assert.False(t, exists)
	assert.Zero(t, ttl)
}

// TestTypedCacheSetWithTTLNonPositiveNeverExpires 验证非正 TTL 会按永不过期写入类型安全缓存。
func TestTypedCacheSetWithTTLNonPositiveNeverExpires(t *testing.T) {
	tests := []struct {
		name        string
		description string
		ttl         time.Duration
	}{
		{
			name:        "zero ttl",
			description: "TTL 为 0 时应写入永不过期缓存项。",
			ttl:         0,
		},
		{
			name:        "negative ttl",
			description: "TTL 为负数时应写入永不过期缓存项。",
			ttl:         -time.Second,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			cache := AsTypedCache[string](newTestCache(t))
			require.True(t, cache.SetWithTTL("non-positive-ttl-key", "non-positive-ttl-value", tt.ttl))

			got, exists, ttl := cache.GetWithTTL("non-positive-ttl-key")
			require.True(t, exists)
			assert.Equal(t, "non-positive-ttl-value", got)
			assert.Equal(t, time.Duration(-1), ttl)
		})
	}
}

// TestNormalizeRistrettoTTL 验证 ristretto TTL 结果会被转换为 Cache 接口约定。
func TestNormalizeRistrettoTTL(t *testing.T) {
	tests := []struct {
		name        string
		description string
		value       interface{}
		exists      bool
		ttl         time.Duration
		wantValue   interface{}
		wantExists  bool
		wantTTL     time.Duration
	}{
		{
			name:        "ttl missing after value hit",
			description: "当值存在但 TTL 查询失败时，应按缓存项不存在返回。",
			value:       "value",
			exists:      false,
			ttl:         0,
			wantValue:   nil,
			wantExists:  false,
			wantTTL:     0,
		},
		{
			name:        "never expires",
			description: "ristretto TTL 为 0 时，应转换为 Cache 接口的 -1 永不过期标记。",
			value:       "value",
			exists:      true,
			ttl:         0,
			wantValue:   "value",
			wantExists:  true,
			wantTTL:     -1,
		},
		{
			name:        "positive ttl",
			description: "正 TTL 应原样作为剩余过期时间返回。",
			value:       "value",
			exists:      true,
			ttl:         time.Minute,
			wantValue:   "value",
			wantExists:  true,
			wantTTL:     time.Minute,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			gotValue, gotExists, gotTTL := normalizeRistrettoTTL(tt.value, tt.exists, tt.ttl)
			assert.Equal(t, tt.wantValue, gotValue)
			assert.Equal(t, tt.wantExists, gotExists)
			assert.Equal(t, tt.wantTTL, gotTTL)
		})
	}
}

// TestRistrettoCacheGetWithTTLMissing 验证底层 ristretto 缓存缺失键时的 TTL 返回语义。
func TestRistrettoCacheGetWithTTLMissing(t *testing.T) {
	cache, err := newRistrettoCache(CacheOptions{
		NumCounters: 1_000,
		MaxCost:     100,
		BufferItems: 64,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, cache.Close())
	})

	got, exists, ttl := cache.GetWithTTL("missing-key")
	assert.Nil(t, got)
	assert.False(t, exists)
	assert.Zero(t, ttl)
}

// newTestCache 创建使用较小自定义配置的测试缓存，并在测试结束时关闭它。
func newTestCache(t *testing.T) Cache {
	t.Helper()

	cache, err := NewCache(testCacheOptions()...)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, cache.Close())
	})

	return cache
}

// testCacheOptions 返回适合单元测试使用的轻量缓存配置选项。
func testCacheOptions() []Option {
	return []Option{
		WithNumCounters(1_000),
		WithMaxCost(100),
		WithBufferItems(64),
	}
}

// resetGlobalCacheForTest 重置包级全局缓存状态，并在测试结束后恢复未初始化状态。
func resetGlobalCacheForTest(t *testing.T) {
	t.Helper()

	if defaultCache != nil {
		require.NoError(t, defaultCache.Close())
	}
	defaultCache = nil
	once = sync.Once{}

	t.Cleanup(func() {
		if defaultCache != nil {
			require.NoError(t, defaultCache.Close())
		}
		defaultCache = nil
		once = sync.Once{}
	})
}

// waitForCacheMiss 轮询等待缓存项变为不存在，用于避免固定 sleep 带来的脆弱 TTL 测试。
func waitForCacheMiss(t *testing.T, exists func() bool) {
	t.Helper()

	require.Eventually(t, func() bool { return !exists() }, time.Second, 10*time.Millisecond)
}
