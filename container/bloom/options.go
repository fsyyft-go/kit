// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package bloom

import (
	kitlog "github.com/fsyyft-go/kit/log"
)

type (
	// Option 定义了布隆过滤器的配置选项类型。
	// 用于在创建布隆过滤器时进行自定义配置。
	Option func(*bloom)
)

// 以下为布隆过滤器的默认参数配置。
// 可通过 Option 机制覆盖。
var (
	// nameDefault 为布隆过滤器默认名称。
	nameDefault = "default"
	// storeDefault 为布隆过滤器默认存储实现。
	storeDefault = NewMemoryStore(0)
	// expectedElementsDefault 为布隆过滤器默认预计元素数量。
	expectedElementsDefault uint64 = 0
	// falsePositiveRateDefault 为布隆过滤器默认误判率。
	falsePositiveRateDefault float64 = 0.01
)

// WithName 设置布隆过滤器的名称。
//
// 参数：
//   - name：布隆过滤器的名称。
//
// 返回值：
//   - Option：配置选项函数。
func WithName(name string) Option {
	return func(b *bloom) {
		b.name = name
	}
}

// WithStore 设置布隆过滤器的存储接口。
//
// 参数：
//   - store：存储接口实现。
//
// 返回值：
//   - Option：配置选项函数。
func WithStore(store Store) Option {
	return func(b *bloom) {
		b.store = store
	}
}

// WithLogger 设置布隆过滤器的日志记录器。
//
// 参数：
//   - logger：日志记录器实例。
//
// 返回值：
//   - Option：配置选项函数。
func WithLogger(logger kitlog.Logger) Option {
	return func(b *bloom) {
		b.logger = logger
	}
}

// WithExpectedElements 设置布隆过滤器预计要存储的元素数量。
//
// 参数：
//   - n：预计元素数量。
//
// 返回值：
//   - Option：配置选项函数。
func WithExpectedElements(n uint64) Option {
	return func(b *bloom) {
		b.n = n
	}
}

// WithFalsePositiveRate 设置布隆过滤器的期望误判率。
//
// 参数：
//   - p：期望误判率。
//
// 返回值：
//   - Option：配置选项函数。
func WithFalsePositiveRate(p float64) Option {
	return func(b *bloom) {
		b.p = p
	}
}
