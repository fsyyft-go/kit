// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package bloom

import (
	"context"
	"errors"
	"fmt"

	kitredis "github.com/fsyyft-go/kit/database/redis"
)

// ErrResultTypeNotArray 表示 Redis 脚本返回结果类型不是数组时的错误。
// 当期望返回数组类型但实际类型不符时会返回该错误。
var (
	ErrResultTypeNotArray = errors.New("result type is not array")
)

const (
	// redisKeyFormat 定义了 Redis 中布隆过滤器键名的格式模板。
	// 通过格式化字符串，将布隆过滤器名称嵌入到 Redis 键名中，避免键名冲突。
	redisKeyFormat = "kit:bloom:%s"

	// existed 表示布隆过滤器位数组中某一位已被设置（即为 1）。
	// 用于判断哈希位是否已存在。
	existed = int64(1)

	// bloomSetScript 是用于批量设置 Redis 位数组中指定位置为 1 的 Lua 脚本。
	// 该脚本遍历所有哈希值，将对应位设置为 1，并返回每次 setbit 操作的结果。
	// KEYS[1]：布隆过滤器的 Redis 键名
	// ARGV：所有需要设置的哈希位索引
	bloomSetScript = `
		local key = KEYS[1]
		local result = {}
		for i = 1, #ARGV ,1 do
			local hash = ARGV[i]
			local r = redis.call("setbit", key, hash, 1)
			table.insert(result, r);
		end
		return result
	`

	// bloomGetScript 是用于批量获取 Redis 位数组中指定位置值的 Lua 脚本。
	// 该脚本遍历所有哈希值，获取对应位的值，并返回每次 getbit 操作的结果。
	// KEYS[1]：布隆过滤器的 Redis 键名
	// ARGV：所有需要查询的哈希位索引
	bloomGetScript = `
		local key = KEYS[1]
		local result = {}
		for i = 1, #ARGV ,1 do
			local hash = ARGV[i]
			local r = redis.call("getbit", key, hash)
			table.insert(result, r);
		end
		return result
	`
)

// redisStore 实现了 Store 接口，基于 Redis 实现布隆过滤器的底层存储。
// 该结构体封装了 Redis 客户端及相关脚本哈希，用于高效地进行批量位操作。
type redisStore struct {
	// redis 是 Redis 客户端实例，用于与 Redis 服务进行通信。
	redis kitredis.Redis

	// setScriptHash 是 bloomSetScript 脚本在 Redis 中的哈希值，用于脚本缓存和高效调用。
	setScriptHash string

	// getScriptHash 是 bloomGetScript 脚本在 Redis 中的哈希值，用于脚本缓存和高效调用。
	getScriptHash string
}

// Exist 判断指定 key 对应的所有 hash 值是否都已存在（即所有位都为 1）。
//
// 参数：
//   - ctx：上下文对象，用于控制请求的生命周期。
//   - key：存储键名，对应 Redis 中的布隆过滤器键。
//   - hash：要判断的哈希值列表，每个值对应位数组中的一个位置。
//
// 返回值：
//   - bool：所有哈希值是否都已存在。
//   - false：至少有一个哈希值不存在（对应位为 0）。
//   - true：所有哈希值都存在（对应位均为 1）。
//   - error：查询过程中发生的错误。
func (s *redisStore) Exist(ctx context.Context, key string, hash []uint64) (bool, error) {
	// 生成 Redis 脚本所需的 KEYS 和 ARGS 参数。
	keys, args := s.generateKeysAndArgs(key, hash)
	// 执行 Lua 脚本，批量获取所有哈希位的值。
	result, err := s.redis.EvalSha(ctx, s.getScriptHash, keys, args...).Result()
	if err != nil {
		return false, err
	}
	// 检查返回结果类型是否为数组。
	res, ok := result.([]any)
	if !ok {
		return false, ErrResultTypeNotArray
	}
	// 遍历所有返回值，只要有一位不是 1，则说明至少有一个哈希值不存在。
	for _, v := range res {
		if v != existed {
			return false, nil
		}
	}
	return true, nil
}

// Add 将一组 hash 值添加到指定 key 对应的存储中（即将对应位设置为 1）。
//
// 参数：
//   - ctx：上下文对象，用于控制请求的生命周期。
//   - key：存储键名，对应 Redis 中的布隆过滤器键。
//   - hash：要添加的哈希值列表，每个值对应位数组中的一个位置。
//
// 返回值：
//   - error：添加过程中发生的错误。
func (s *redisStore) Add(ctx context.Context, key string, hash []uint64) error {
	// 生成 Redis 脚本所需的 KEYS 和 ARGS 参数。
	keys, args := s.generateKeysAndArgs(key, hash)
	// 执行 Lua 脚本，批量设置所有哈希位为 1。
	_, err := s.redis.EvalSha(ctx, s.setScriptHash, keys, args...).Result()
	if err != nil {
		return err
	}
	return nil
}

// generateKeysAndArgs 生成 Redis 脚本调用所需的 KEYS 和 ARGS 参数列表。
//
// 参数：
//   - name：布隆过滤器名称，用于生成 Redis 键名。
//   - hash：哈希值列表，每个值对应位数组中的一个位置。
//
// 返回值：
//   - []string：包含一个元素的 KEYS 切片（即 Redis 键名）。
//   - []any：包含所有哈希值的 ARGS 切片。
func (r *redisStore) generateKeysAndArgs(name string, hash []uint64) ([]string, []any) {
	// 构造 Redis 键名。
	keys := make([]string, 1)
	keys[0] = fmt.Sprintf(redisKeyFormat, name)
	// 构造 ARGS 参数。
	args := make([]any, len(hash))
	for i, v := range hash {
		args[i] = v
	}
	return keys, args
}

// NewRedisStore 创建一个基于 Redis 的布隆过滤器存储实例。
//
// 参数：
//   - redis：Redis 客户端实例，用于与 Redis 服务进行通信。
//
// 返回值：
//   - *redisStore：Redis 存储实现的实例指针。
//   - error：初始化过程中发生的错误。
func NewRedisStore(redis kitredis.Redis) (*redisStore, error) {
	var setScriptHash string
	var getScriptHash string

	// 预加载 set 脚本到 Redis，获取脚本哈希值。
	if ssh, err := redis.ScriptLoad(context.Background(), bloomSetScript).Result(); err != nil {
		return nil, err
	} else {
		setScriptHash = ssh
	}

	// 预加载 get 脚本到 Redis，获取脚本哈希值。
	if gsh, err := redis.ScriptLoad(context.Background(), bloomGetScript).Result(); err != nil {
		return nil, err
	} else {
		getScriptHash = gsh
	}

	// 构造 redisStore 实例并返回。
	s := &redisStore{
		redis:         redis,
		setScriptHash: setScriptHash,
		getScriptHash: getScriptHash,
	}

	return s, nil
}
