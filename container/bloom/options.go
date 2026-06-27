// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package bloom

import (
	kitredis "github.com/fsyyft-go/kit/database/redis"
	kitlog "github.com/fsyyft-go/kit/log"
)

type (
	// Option 定义 NewBloom 的配置修改函数。
	//
	// 参数：
	//   - *bloom: 待修改的布隆过滤器配置实例，由 NewBloom 在初始化阶段传入。
	Option func(*bloom)
)

// 以下为布隆过滤器的默认参数配置。
// 可通过 Option 机制覆盖。
var (
	// nameDefault 是未显式配置名称时使用的默认布隆过滤器名称。
	nameDefault = "default"

	// storeDefault 是未显式配置 Store 时使用的默认内存存储实例。
	storeDefault = NewMemoryStore(0)

	// expectedElementsDefault 是未显式配置预计元素数量时使用的默认值。
	expectedElementsDefault uint64 = 65536

	// falsePositiveRateDefault 是未显式配置误判率时使用的默认值。
	falsePositiveRateDefault float64 = 0.01
)

// WithName 设置布隆过滤器名称。
//
// 参数：
//   - name: 布隆过滤器名称；NewBloom 会拒绝空白名称，并使用该名称作为默认位图命名空间 key。
//
// 返回：
//   - Option: 应用于 NewBloom 的名称配置项。
func WithName(name string) Option {
	return func(b *bloom) {
		b.name = name
	}
}

// WithStore 设置布隆过滤器的底层存储实现。
//
// store 会接收 Bloom 名称或 "name:group" 形式的派生 key，用于区分不同位图命名空间。
// 当前实现不校验 nil；调用方传入 nil 后再执行读写会触发运行时 panic。
//
// 参数：
//   - store: 布隆过滤器底层存储实现，负责保存和查询 hash 对应的位图位置。
//
// 返回：
//   - Option: 应用于 NewBloom 的存储配置项。
func WithStore(store Store) Option {
	return func(b *bloom) {
		b.store = store
	}
}

// WithRedis 设置布隆过滤器使用基于 Redis 的存储实现。
//
// Redis 存储会把 Bloom 名称或分组派生 key 转换为实际 Redis key，并透传调用时的 ctx。
// 初始化脚本加载失败时，当前实现会 panic。
//
// 参数：
//   - redis: Redis 客户端实例，用于初始化 Redis Store 并执行后续 Lua 脚本。
//
// 返回：
//   - Option: 应用于 NewBloom 的 Redis 存储配置项。
func WithRedis(redis kitredis.Redis) Option {
	return func(b *bloom) {
		store, err := NewRedisStore(redis)
		if err != nil {
			panic(err)
		}
		b.store = store
	}
}

// WithLogger 设置布隆过滤器持有的日志记录器。
//
// 参数：
//   - logger: 日志记录器实例；当前核心读写路径不主动写日志，但会保存在实例配置中。
//
// 返回：
//   - Option: 应用于 NewBloom 的日志配置项。
func WithLogger(logger kitlog.Logger) Option {
	return func(b *bloom) {
		b.logger = logger
	}
}

// WithExpectedElements 设置布隆过滤器预计要存储的元素数量。
//
// 参数：
//   - n: 预计元素数量，用于 NewBloom 计算位数组大小和 hash 次数；调用方应传入正数，当前实现不单独校验 0。
//
// 返回：
//   - Option: 应用于 NewBloom 的预计元素数配置项。
func WithExpectedElements(n uint64) Option {
	return func(b *bloom) {
		b.n = n
	}
}

// WithFalsePositiveRate 设置布隆过滤器的期望误判率。
//
// 参数：
//   - p: 期望误判率；NewBloom 会拒绝小于 0 或大于 1 的值。
//
// 返回：
//   - Option: 应用于 NewBloom 的误判率配置项。
func WithFalsePositiveRate(p float64) Option {
	return func(b *bloom) {
		b.p = p
	}
}
