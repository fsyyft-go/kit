// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package redis 提供了 Redis 操作的扩展功能实现。
package redis

import (
	"context"
	"time"
)

type (
	// RedisExtension 定义了 Redis 扩展接口，继承自基础 Redis 接口，提供额外的功能扩展。
	RedisExtension interface {
		Redis

		// Get 获取指定键的值。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制命令的执行
		//   - key：要获取的键名
		//
		// 返回值：
		//   - *Cmd：命令执行结果
		Get(ctx context.Context, key string) *Cmd

		// Set 设置指定键的值。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制命令的执行
		//   - key：要设置的键名
		//   - value：要设置的值
		//   - expiration：键的过期时间
		//
		// 返回值：
		//   - *Cmd：命令执行结果
		Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *Cmd

		// Del 删除指定的键。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制命令的执行
		//   - key：要删除的键名
		//
		// 返回值：
		//   - *Cmd：命令执行结果
		Del(ctx context.Context, key string) *Cmd

		// Expire 设置指定键的过期时间。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制命令的执行
		//   - key：要设置过期时间的键名
		//   - expiration：过期时间
		//
		// 返回值：
		//   - *Cmd：命令执行结果
		Expire(ctx context.Context, key string, expiration time.Duration) *Cmd

		// ScriptFlush 清空脚本缓存。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制命令的执行
		//
		// 返回值：
		//   - *StatusCmd：命令执行状态
		ScriptFlush(ctx context.Context) *StatusCmd

		// ScriptKill 终止当前正在执行的脚本。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制命令的执行
		//
		// 返回值：
		//   - *StatusCmd：命令执行状态
		ScriptKill(ctx context.Context) *StatusCmd
	}

	// redisExtension 是 RedisExtension 接口的具体实现。
	redisExtension struct {
		// redis 是基础 Redis 接口的实例。
		redis Redis
	}
)

// NewRedisExtension 创建一个新的 Redis 扩展实例。
//
// 参数：
//   - redis：基础的 Redis 接口实现
//
// 返回值：
//   - RedisExtension：Redis 扩展接口实例
func NewRedisExtension(redis Redis) RedisExtension {
	return &redisExtension{redis: redis}
}

// Do 执行任意 Redis 命令。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//   - args：命令的参数列表
//
// 返回值：
//   - *Cmd：通用的命令结果对象
func (r *redisExtension) Do(ctx context.Context, args ...interface{}) *Cmd {
	return r.redis.Do(ctx, args...)
}

// Pipelined 在管道中执行多个命令。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//   - fn：定义要在管道中执行的命令的函数
//
// 返回值：
//   - []Cmder：管道中所有命令的执行结果
//   - error：执行过程中的错误信息
func (r *redisExtension) Pipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error) {
	return r.redis.Pipelined(ctx, fn)
}

// TxPipelined 在事务管道中执行多个命令。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//   - fn：定义要在事务管道中执行的命令的函数
//
// 返回值：
//   - []Cmder：事务中所有命令的执行结果
//   - error：执行过程中的错误信息
func (r *redisExtension) TxPipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error) {
	return r.redis.TxPipelined(ctx, fn)
}

// Subscribe 订阅指定的频道。
//
// 参数：
//   - ctx：上下文对象，用于控制订阅的生命周期
//   - channels：要订阅的频道列表
//
// 返回值：
//   - *PubSub：发布订阅客户端对象
func (r *redisExtension) Subscribe(ctx context.Context, channels ...string) *PubSub {
	return r.redis.Subscribe(ctx, channels...)
}

// PSubscribe 使用模式匹配订阅频道。
//
// 参数：
//   - ctx：上下文对象，用于控制订阅的生命周期
//   - channels：要订阅的频道模式列表
//
// 返回值：
//   - *PubSub：发布订阅客户端对象
func (r *redisExtension) PSubscribe(ctx context.Context, channels ...string) *PubSub {
	return r.redis.PSubscribe(ctx, channels...)
}

// Eval 执行 Lua 脚本。
//
// 参数：
//   - ctx：上下文对象，用于控制脚本的执行
//   - script：要执行的 Lua 脚本
//   - keys：脚本中使用的键列表
//   - args：脚本的参数列表
//
// 返回值：
//   - *Cmd：脚本执行的结果
func (r *redisExtension) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *Cmd {
	return r.redis.Eval(ctx, script, keys, args...)
}

// EvalRO 以只读模式执行 Lua 脚本。
//
// 参数：
//   - ctx：上下文对象，用于控制脚本的执行
//   - script：要执行的 Lua 脚本
//   - keys：脚本中使用的键列表
//   - args：脚本的参数列表
//
// 返回值：
//   - *Cmd：脚本执行的结果
func (r *redisExtension) EvalRO(ctx context.Context, script string, keys []string, args ...interface{}) *Cmd {
	return r.redis.EvalRO(ctx, script, keys, args...)
}

// EvalSha 使用脚本的 SHA1 值执行 Lua 脚本。
//
// 参数：
//   - ctx：上下文对象，用于控制脚本的执行
//   - sha1：脚本的 SHA1 值
//   - keys：脚本中使用的键列表
//   - args：脚本的参数列表
//
// 返回值：
//   - *Cmd：脚本执行的结果
func (r *redisExtension) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *Cmd {
	return r.redis.EvalSha(ctx, sha1, keys, args...)
}

// EvalShaRO 以只读模式使用脚本的 SHA1 值执行 Lua 脚本。
//
// 参数：
//   - ctx：上下文对象，用于控制脚本的执行
//   - sha1：脚本的 SHA1 值
//   - keys：脚本中使用的键列表
//   - args：脚本的参数列表
//
// 返回值：
//   - *Cmd：脚本执行的结果
func (r *redisExtension) EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...interface{}) *Cmd {
	return r.redis.EvalShaRO(ctx, sha1, keys, args...)
}

// ScriptExists 检查指定的脚本是否存在于脚本缓存中。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//   - hashes：要检查的脚本 SHA1 值列表
//
// 返回值：
//   - *BoolSliceCmd：检查结果，每个元素表示对应脚本是否存在
func (r *redisExtension) ScriptExists(ctx context.Context, hashes ...string) *BoolSliceCmd {
	return r.redis.ScriptExists(ctx, hashes...)
}

// ScriptLoad 将脚本加载到脚本缓存中。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//   - script：要加载的 Lua 脚本
//
// 返回值：
//   - *StringCmd：脚本的 SHA1 值
func (r *redisExtension) ScriptLoad(ctx context.Context, script string) *StringCmd {
	return r.redis.ScriptLoad(ctx, script)
}

// Get 获取指定键的值。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//   - key：要获取的键名
//
// 返回值：
//   - *Cmd：命令执行结果
func (r *redisExtension) Get(ctx context.Context, key string) *Cmd {
	return r.redis.Do(ctx, "GET", key)
}

// Set 设置指定键的值。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//   - key：要设置的键名
//   - value：要设置的值
//   - expiration：键的过期时间
//
// 返回值：
//   - *Cmd：命令执行结果
func (r *redisExtension) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *Cmd {
	return r.redis.Do(ctx, "SET", key, value, "EX", expiration.Seconds())
}

// Del 删除指定的键。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//   - key：要删除的键名
//
// 返回值：
//   - *Cmd：命令执行结果
func (r *redisExtension) Del(ctx context.Context, key string) *Cmd {
	return r.redis.Do(ctx, "DEL", key)
}

// Expire 设置指定键的过期时间。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//   - key：要设置过期时间的键名
//   - expiration：过期时间
//
// 返回值：
//   - *Cmd：命令执行结果
func (r *redisExtension) Expire(ctx context.Context, key string, expiration time.Duration) *Cmd {
	return r.redis.Do(ctx, "EXPIRE", key, expiration.Seconds())
}

// ScriptFlush 清空脚本缓存。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//
// 返回值：
//   - *StatusCmd：命令执行状态
func (r *redisExtension) ScriptFlush(ctx context.Context) *StatusCmd {
	// 通过类型断言调用底层实现，确保兼容接口
	if c, ok := r.redis.(*redisClient); ok {
		return c.ScriptFlush(ctx)
	}
	return nil
}

// ScriptKill 终止当前正在执行的脚本。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//
// 返回值：
//   - *StatusCmd：命令执行状态
func (r *redisExtension) ScriptKill(ctx context.Context) *StatusCmd {
	if c, ok := r.redis.(*redisClient); ok {
		return c.ScriptKill(ctx)
	}
	return nil
}

// Close 关闭 Redis 客户端。
//
// 返回值：
//   - error：关闭过程中发生的错误
func (r *redisExtension) Close() error {
	return r.redis.Close()
}
