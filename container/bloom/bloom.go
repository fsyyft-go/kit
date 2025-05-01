// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package bloom 提供了布隆过滤器的接口定义和实现。
// 布隆过滤器是一种空间效率很高的概率型数据结构，用于判断一个元素是否在集合中。
// 它通过多个哈希函数将元素映射到位数组中的多个位置，从而实现高效的成员查询。
package bloom

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/spaolacci/murmur3"
	"github.com/spf13/cast"

	kitlog "github.com/fsyyft-go/kit/log"
)

var (
	// bloomNames 保存所有布隆过滤器名称，创建时避免名称重复。
	bloomNames = make(map[string]string)
)

var (
	ErrBloomNameEmpty                = errors.New("bloom: bloom name can't be empty")
	ErrBloomNameRepeated             = errors.New("bloom: bloom name can't repeated")
	ErrBloomFalseProbabilityThanOne  = errors.New("bloom: bloom false probability can't than 1")
	ErrBloomFalseProbabilityNegative = errors.New("bloom: bloom false probability can't be negative")
)

type (
	// Bloom 定义了布隆过滤器的核心接口。
	// 该接口提供了基本的元素判断和添加功能，以及分组操作的支持。
	// 布隆过滤器的主要特点是空间效率高，但可能存在误判（假阳性）。
	Bloom interface {
		// Contain 判断指定元素是否可能存在于布隆过滤器中。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制请求的生命周期
		//   - value：要判断是否存在的元素值
		//
		// 返回值：
		//   - bool：元素是否可能存在于布隆过滤器中
		//     - false：元素肯定不存在
		//     - true：元素可能存在（存在误判可能）
		//   - error：操作过程中发生的错误
		Contain(ctx context.Context, value string) (bool, error)

		// Put 将指定元素添加到布隆过滤器中。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制请求的生命周期
		//   - value：要添加到布隆过滤器中的元素值
		//
		// 返回值：
		//   - error：添加过程中发生的错误
		Put(ctx context.Context, value string) error

		// GroupContain 判断指定分组中是否可能包含指定元素。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制请求的生命周期
		//   - group：分组名称，用于区分不同的数据集合
		//   - value：要判断是否存在的元素值
		//
		// 返回值：
		//   - bool：元素是否可能存在于指定分组中
		//     - false：元素在指定分组中肯定不存在
		//     - true：元素在指定分组中可能存在（存在误判可能）
		//   - error：操作过程中发生的错误
		GroupContain(ctx context.Context, group string, value string) (bool, error)

		// GroupPut 将指定元素添加到指定分组的布隆过滤器中。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制请求的生命周期
		//   - group：分组名称，用于区分不同的数据集合
		//   - value：要添加到布隆过滤器中的元素值
		//
		// 返回值：
		//   - error：添加过程中发生的错误
		GroupPut(ctx context.Context, group string, value string) error
	}

	// Store 定义了布隆过滤器底层数据存储的接口。
	// 该接口负责实际的数据存储和查询操作。
	// 不同的存储实现可以支持不同的后端存储系统，如内存、Redis 等。
	Store interface {
		// Exist 判断指定 key 对应的所有 hash 值是否都已存在。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制请求的生命周期
		//   - key：存储键名
		//   - hash：要判断的哈希值列表
		//
		// 返回值：
		//   - bool：所有哈希值是否都已存在
		//     - false：至少有一个哈希值不存在
		//     - true：所有哈希值都存在
		//   - error：查询过程中发生的错误
		Exist(ctx context.Context, key string, hash []uint64) (bool, error)

		// Add 将一组 hash 值添加到指定 key 对应的存储中。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制请求的生命周期
		//   - key：存储键名
		//   - hash：要添加的哈希值列表
		//
		// 返回值：
		//   - error：添加过程中发生的错误
		Add(ctx context.Context, key string, hash []uint64) error
	}

	// bloom 是布隆过滤器的具体实现结构体。
	// 它包含了布隆过滤器所需的所有配置参数和存储接口。
	bloom struct {
		name   string // name 是布隆过滤器的名称，用于区分不同的过滤器实例
		store  Store  // store 是底层存储接口的实现
		logger kitlog.Logger

		n uint64  // n 是预计要存储的元素数量
		m uint64  // m 是位数组的大小（二进制位的总数）
		p float64 // p 是期望的误判率
		k uint    // k 是使用的哈希函数的数量
	}
)

// NewBloom 创建一个新的布隆过滤器实例。
//
// 参数：
//   - opts：配置选项列表。
//
// 返回值：
//   - Bloom：实现了 Bloom 接口的布隆过滤器实例。
func NewBloom(opts ...Option) (Bloom, func(), error) {
	b := &bloom{
		name:   nameDefault,
		store:  storeDefault,
		n:      expectedElementsDefault,
		m:      0,
		p:      falsePositiveRateDefault, // 默认误判率为 1%。
		k:      0,
		logger: kitlog.GetLogger(),
	}

	// 应用用户提供的配置选项。
	for _, opt := range opts {
		opt(b)
	}
	if strings.TrimSpace(b.name) == "" {
		return nil, nil, ErrBloomNameEmpty
	}

	_, ok := bloomNames[b.name]
	if ok {
		return nil, nil, ErrBloomNameRepeated
	}
	if b.p > 1 {
		return nil, nil, ErrBloomFalseProbabilityThanOne
	}
	if b.p < 0 {
		return nil, nil, ErrBloomFalseProbabilityNegative
	}

	// 根据元素总量，误判率计算最优二进制位总量和哈希次数。
	mm := -cast.ToFloat64(b.n) * math.Log(b.p) / (math.Log(2) * math.Log(2)) //nolint:mnd
	m := cast.ToUint64(mm)
	kk := math.Max(1, math.Round(mm/cast.ToFloat64(b.n)*math.Log(2))) //nolint:mnd
	k := cast.ToUint(kk)

	b.k = k
	b.m = m

	return b, func() {}, nil
}

// Contain 实现了 Bloom 接口的 Contain 方法，用于判断指定元素是否可能存在于布隆过滤器中。
//
// 参数：
//   - ctx：上下文对象，用于控制请求的生命周期
//   - value：要判断是否存在的元素值
//
// 返回值：
//   - bool：元素是否可能存在于布隆过滤器中
//   - false：元素肯定不存在
//   - true：元素可能存在（存在误判可能）
//   - error：操作过程中发生的错误
func (b *bloom) Contain(ctx context.Context, value string) (bool, error) {
	hash := b.multiHash(value)
	return b.store.Exist(ctx, b.name, hash)
}

// Put 实现了 Bloom 接口的 Put 方法，用于将指定元素添加到布隆过滤器中。
//
// 参数：
//   - ctx：上下文对象，用于控制请求的生命周期
//   - value：要添加到布隆过滤器中的元素值
//
// 返回值：
//   - error：添加过程中发生的错误
func (b *bloom) Put(ctx context.Context, value string) error {
	hash := b.multiHash(value)
	return b.store.Add(ctx, b.name, hash)
}

// GroupContain 实现了 Bloom 接口的 GroupContain 方法，用于判断指定分组中是否可能包含指定元素。
//
// 参数：
//   - ctx：上下文对象，用于控制请求的生命周期
//   - group：分组名称，用于区分不同的数据集合
//   - value：要判断是否存在的元素值
//
// 返回值：
//   - bool：元素是否可能存在于指定分组中
//   - false：元素在指定分组中肯定不存在
//   - true：元素在指定分组中可能存在（存在误判可能）
//   - error：操作过程中发生的错误
func (b *bloom) GroupContain(ctx context.Context, group string, value string) (bool, error) {
	hash := b.multiHash(value)
	return b.store.Exist(ctx, b.buildGroupKey(group), hash)
}

// GroupPut 实现了 Bloom 接口的 GroupPut 方法，用于将指定元素添加到指定分组的布隆过滤器中。
//
// 参数：
//   - ctx：上下文对象，用于控制请求的生命周期
//   - group：分组名称，用于区分不同的数据集合
//   - value：要添加到布隆过滤器中的元素值
//
// 返回值：
//   - error：添加过程中发生的错误
func (b *bloom) GroupPut(ctx context.Context, group string, value string) error {
	hash := b.multiHash(value)
	return b.store.Add(ctx, b.buildGroupKey(group), hash)
}

// multiHash 使用多个哈希函数计算输入值的哈希值。
// 该方法使用 murmur3 哈希算法生成两个基础哈希值，然后通过线性组合生成 k 个不同的哈希值。
//
// 参数：
//   - value：要计算哈希值的字符串
//
// 返回值：
//   - []uint64：包含 k 个哈希值的切片，每个哈希值对应位数组中的一个位置
func (b *bloom) multiHash(value string) []uint64 {
	hash := make([]uint64, b.k)
	hash1, hash2 := murmur3.Sum128([]byte(value))
	for i := uint(0); i < b.k; i++ {
		k := uint64(i)
		h := hash1 + k*hash2 + k*k
		hash[i] = (h & math.MaxUint64) % b.m
	}
	return hash
}

// buildGroupKey 构建分组键名，将布隆过滤器名称和分组名称组合成一个唯一的键名。
//
// 参数：
//   - group：分组名称，用于区分不同的数据集合
//
// 返回值：
//   - string：组合后的键名，格式为 "name:group"
func (b *bloom) buildGroupKey(group string) string {
	return fmt.Sprintf("%s:%s", b.name, group)
}
