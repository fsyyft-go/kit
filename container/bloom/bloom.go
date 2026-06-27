// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

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
	// bloomNames 保存已登记的布隆过滤器名称，用于 NewBloom 检测名称重复。
	//
	// 当前实现只读取该注册表，不会在 NewBloom 成功后自动写入或由 cleanup 清理。
	bloomNames = make(map[string]string)
)

var (
	// ErrBloomNameEmpty 表示布隆过滤器名称为空或仅包含空白字符。
	ErrBloomNameEmpty = errors.New("bloom: bloom name can't be empty")

	// ErrBloomNameRepeated 表示 bloomNames 注册表中已经存在相同的布隆过滤器名称。
	ErrBloomNameRepeated = errors.New("bloom: bloom name can't repeated")

	// ErrBloomFalseProbabilityThanOne 表示期望误判率大于 1。
	ErrBloomFalseProbabilityThanOne = errors.New("bloom: bloom false probability can't than 1")

	// ErrBloomFalseProbabilityNegative 表示期望误判率小于 0。
	ErrBloomFalseProbabilityNegative = errors.New("bloom: bloom false probability can't be negative")
)

type (
	// Bloom 定义可追加写入的布隆过滤器接口。
	//
	// Bloom 以字符串值为输入，通过底层 Store 保存多重 hash 对应的位图位置。Contain 或
	// GroupContain 返回 false 时表示元素一定不存在，返回 true 时表示元素可能存在且可能包含误判。
	Bloom interface {
		// Contain 判断元素是否可能存在于当前 Bloom 名称对应的位图中。
		//
		// 参数：
		//   - ctx: 调用上下文，会传递给底层 Store；是否生效由具体 Store 实现决定。
		//   - value: 待判断的元素值，按原始字符串内容参与 hash 计算。
		//
		// 返回：
		//   - bool: 元素存在性判断结果：
		//     - false: 元素一定不存在。
		//     - true: 元素可能存在，调用方需要接受布隆过滤器的误判可能。
		//   - error: 底层 Store 查询失败时返回错误。
		Contain(ctx context.Context, value string) (bool, error)

		// Put 将元素添加到当前 Bloom 名称对应的位图中。
		//
		// 参数：
		//   - ctx: 调用上下文，会传递给底层 Store；是否生效由具体 Store 实现决定。
		//   - value: 待添加的元素值，按原始字符串内容参与 hash 计算。
		//
		// 返回：
		//   - error: 底层 Store 写入失败时返回错误。
		Put(ctx context.Context, value string) error

		// GroupContain 判断元素是否可能存在于指定分组对应的位图中。
		//
		// 参数：
		//   - ctx: 调用上下文，会传递给底层 Store；是否生效由具体 Store 实现决定。
		//   - group: 分组名称，会按 "name:group" 与 Bloom 名称拼接成最终 Store key；调用方应避免
		//     name 或 group 中的冒号导致 key 冲突。
		//   - value: 待判断的元素值，按原始字符串内容参与 hash 计算。
		//
		// 返回：
		//   - bool: 元素在分组中的存在性判断结果：
		//     - false: 元素在该分组中一定不存在。
		//     - true: 元素在该分组中可能存在，调用方需要接受布隆过滤器的误判可能。
		//   - error: 底层 Store 查询失败时返回错误。
		GroupContain(ctx context.Context, group string, value string) (bool, error)

		// GroupPut 将元素添加到指定分组对应的位图中。
		//
		// 参数：
		//   - ctx: 调用上下文，会传递给底层 Store；是否生效由具体 Store 实现决定。
		//   - group: 分组名称，会按 "name:group" 与 Bloom 名称拼接成最终 Store key；调用方应避免
		//     name 或 group 中的冒号导致 key 冲突。
		//   - value: 待添加的元素值，按原始字符串内容参与 hash 计算。
		//
		// 返回：
		//   - error: 底层 Store 写入失败时返回错误。
		GroupPut(ctx context.Context, group string, value string) error
	}

	// Store 定义布隆过滤器底层位图存储接口。
	//
	// key 标识一个位图命名空间；只有最终 Store key 不同时，对应位图才彼此隔离。
	// 分组操作使用 "name:group" 直接拼接 Store key，不转义或校验分隔符；name 或
	// group 含有冒号时可能与其它 name/group 组合产生相同 Store key。
	// ctx 是否生效由具体实现决定，例如内存实现会忽略它，Redis 实现会将其传递给底层命令。
	Store interface {
		// Exist 判断指定 key 对应的所有 hash 值是否都已存在。
		//
		// 参数：
		//   - ctx: 调用上下文，是否生效由具体 Store 实现决定。
		//   - key: 位图命名空间标识；不同 key 之间互不影响。
		//   - hash: 要判断的哈希值列表；为空时按全称判断语义表示没有缺失位。
		//
		// 返回：
		//   - bool: 所有哈希值是否都已存在：
		//     - false: 至少有一个哈希值不存在。
		//     - true: 所有哈希值都存在，或 hash 为空且实现未返回错误。
		//   - error: 查询过程中发生的错误。
		Exist(ctx context.Context, key string, hash []uint64) (bool, error)

		// Add 将一组 hash 值添加到指定 key 对应的存储中。
		//
		// 参数：
		//   - ctx: 调用上下文，是否生效由具体 Store 实现决定。
		//   - key: 位图命名空间标识；不同 key 之间互不影响。
		//   - hash: 要添加的哈希值列表；为空时不设置任何位。
		//
		// 返回：
		//   - error: 添加过程中发生的错误。
		Add(ctx context.Context, key string, hash []uint64) error
	}

	// bloom 是 Bloom 接口的默认实现。
	//
	// bloom 保存名称、存储、预计元素数、位图规模、期望误判率和 hash 次数等运行参数。
	// 实例本身不加锁；并发读写能力由底层 Store 实现保证。
	bloom struct {
		name   string // name 标识当前 Bloom 实例的 Store key 前缀。
		store  Store  // store 保存位图读写的底层实现。
		logger kitlog.Logger

		n uint64  // n 是预计要存储的元素数量。
		m uint64  // m 是位数组的大小，以 bit 为单位。
		p float64 // p 是期望的误判率。
		k uint    // k 是使用的哈希函数数量。
	}
)

// NewBloom 创建一个新的布隆过滤器实例。
//
// 参数：
//   - opts: 可选配置项，按传入顺序应用；未传入时使用默认名称、默认内存 Store、默认预计元素数和默认误判率。调用方应保证每个 Option 非 nil。
//
// 返回：
//   - Bloom: 初始化完成的布隆过滤器实例。
//   - func(): 清理函数；当前实现返回空操作函数，不释放 Store 或 Redis 客户端资源。
//   - error: 名称为空、名称已存在于 bloomNames 注册表、误判率大于 1 或误判率小于 0 时返回对应错误。
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

// Contain 判断元素是否可能存在于当前 Bloom 名称对应的位图中。
//
// 参数：
//   - ctx: 调用上下文，会传递给底层 Store；是否生效由具体 Store 实现决定。
//   - value: 待判断的元素值，按原始字符串内容参与 hash 计算。
//
// 返回：
//   - bool: 元素存在性判断结果；false 表示一定不存在，true 表示可能存在且可能误判。
//   - error: 底层 Store 查询失败时返回错误。
func (b *bloom) Contain(ctx context.Context, value string) (bool, error) {
	hash := b.multiHash(value)
	return b.store.Exist(ctx, b.name, hash)
}

// Put 将元素添加到当前 Bloom 名称对应的位图中。
//
// 参数：
//   - ctx: 调用上下文，会传递给底层 Store；是否生效由具体 Store 实现决定。
//   - value: 待添加的元素值，按原始字符串内容参与 hash 计算。
//
// 返回：
//   - error: 底层 Store 写入失败时返回错误。
func (b *bloom) Put(ctx context.Context, value string) error {
	hash := b.multiHash(value)
	return b.store.Add(ctx, b.name, hash)
}

// GroupContain 判断元素是否可能存在于指定分组对应的位图中。
//
// 参数：
//   - ctx: 调用上下文，会传递给底层 Store；是否生效由具体 Store 实现决定。
//   - group: 分组名称，会按 "name:group" 与 Bloom 名称直接拼接成最终 Store key；调用方应避免
//     name 或 group 中的冒号导致 key 冲突。
//   - value: 待判断的元素值，按原始字符串内容参与 hash 计算。
//
// 返回：
//   - bool: 元素在分组中的存在性判断结果；false 表示一定不存在，true 表示可能存在且可能误判。
//   - error: 底层 Store 查询失败时返回错误。
func (b *bloom) GroupContain(ctx context.Context, group string, value string) (bool, error) {
	hash := b.multiHash(value)
	return b.store.Exist(ctx, b.buildGroupKey(group), hash)
}

// GroupPut 将元素添加到指定分组对应的位图中。
//
// 参数：
//   - ctx: 调用上下文，会传递给底层 Store；是否生效由具体 Store 实现决定。
//   - group: 分组名称，会按 "name:group" 与 Bloom 名称直接拼接成最终 Store key；调用方应避免
//     name 或 group 中的冒号导致 key 冲突。
//   - value: 待添加的元素值，按原始字符串内容参与 hash 计算。
//
// 返回：
//   - error: 底层 Store 写入失败时返回错误。
func (b *bloom) GroupPut(ctx context.Context, group string, value string) error {
	hash := b.multiHash(value)
	return b.store.Add(ctx, b.buildGroupKey(group), hash)
}

// multiHash 根据 value 生成当前 Bloom 实例需要写入或查询的多个位图位置。
//
// multiHash 使用 murmur3.Sum128 生成两个基础哈希值，再通过线性组合生成 b.k 个位置。
// 调用方应保证 b.m 和 b.k 已由 NewBloom 初始化完成。
//
// 参数：
//   - value: 待计算哈希位置的字符串。
//
// 返回：
//   - []uint64: 长度为 b.k 的哈希位置切片，每个值位于当前 Bloom 位数组范围内。
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

// buildGroupKey 构建分组位图命名空间 key。
//
// buildGroupKey 直接用冒号拼接 Bloom name 与 group，不转义或校验分隔符；只有最终
// Store key 不同时，对应位图才彼此隔离。调用方应避免 name 或 group 中的冒号导致 key 冲突。
//
// 参数：
//   - group: 分组名称，会与 Bloom name 直接拼接成最终 Store key。
//
// 返回：
//   - string: 格式为 "name:group" 的 Store key。
func (b *bloom) buildGroupKey(group string) string {
	return fmt.Sprintf("%s:%s", b.name, group)
}
