// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package redis

import (
	"context"
	"time"
)

type (
	// RedisExtension 定义在 Redis 基础能力之上的常用扩展操作。
	//
	// RedisExtension 通过 Do 组合 GET、SET、DEL 和过期命令，并保留 Redis 的脚本、管道和订阅能力。
	// 扩展方法返回的命令对象遵循 go-redis/v9 约定，执行错误由命令结果的 Err 方法承载。
	RedisExtension interface {
		Redis

		// Get 获取指定键的值。
		//
		// 参数：
		//   - ctx: 控制命令执行生命周期的上下文。
		//   - key: 要读取的 Redis 键名。
		//
		// 返回：
		//   - *Cmd: GET 命令结果；键不存在时 Err 通常为 ErrNil。
		Get(ctx context.Context, key string) *Cmd

		// Set 设置指定键的值和可选过期时间。
		//
		// 参数：
		//   - ctx: 控制命令执行生命周期的上下文。
		//   - key: 要写入的 Redis 键名。
		//   - value: 要写入的值，编码规则由底层 go-redis 客户端决定。
		//   - expiration: 键过期时间；非正值表示不设置过期时间，整秒使用 EX，非整秒向上取整为毫秒并使用 PX。
		//
		// 返回：
		//   - *Cmd: SET 命令结果；执行错误由返回值的 Err 方法承载。
		Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *Cmd

		// Del 删除指定键。
		//
		// 参数：
		//   - ctx: 控制命令执行生命周期的上下文。
		//   - key: 要删除的 Redis 键名。
		//
		// 返回：
		//   - *Cmd: DEL 命令结果；执行错误由返回值的 Err 方法承载。
		Del(ctx context.Context, key string) *Cmd

		// Expire 设置指定键的过期时间。
		//
		// 参数：
		//   - ctx: 控制命令执行生命周期的上下文。
		//   - key: 要设置过期语义的 Redis 键名。
		//   - expiration: 键过期时间；正整秒使用 EXPIRE，正非整秒向上取整为毫秒并使用 PEXPIRE，非正值使用 EXPIRE 0。
		//
		// 返回：
		//   - *Cmd: 过期命令结果；执行错误由返回值的 Err 方法承载。
		Expire(ctx context.Context, key string, expiration time.Duration) *Cmd

		// ScriptFlush 按底层客户端能力清空脚本缓存。
		//
		// 当底层 Redis 实现提供 ScriptFlush(context.Context) *StatusCmd 时委托调用；否则返回 nil，调用方应处理 nil 返回值。
		//
		// 参数：
		//   - ctx: 控制命令执行生命周期的上下文。
		//
		// 返回：
		//   - *StatusCmd: 底层支持脚本缓存管理时返回对应命令；否则返回 nil。
		ScriptFlush(ctx context.Context) *StatusCmd

		// ScriptKill 按底层客户端能力终止当前正在执行的脚本。
		//
		// 当底层 Redis 实现提供 ScriptKill(context.Context) *StatusCmd 时委托调用；否则返回 nil，调用方应处理 nil 返回值。
		//
		// 参数：
		//   - ctx: 控制命令执行生命周期的上下文。
		//
		// 返回：
		//   - *StatusCmd: 底层支持脚本终止能力时返回对应命令；否则返回 nil。
		ScriptKill(ctx context.Context) *StatusCmd
	}

	// redisExtension 将 Redis 基础实现包装为 RedisExtension。
	redisExtension struct {
		// redis 是被包装的 Redis 基础接口实例。
		redis Redis
	}
)

// NewRedisExtension 创建一个 Redis 扩展实例。
//
// 参数：
//   - redis: 被包装的基础 Redis 实现；调用方应传入非 nil 实例。
//
// 返回：
//   - RedisExtension: 基于 redis 转发基础命令并补充常用 KV 操作的扩展实例。
func NewRedisExtension(redis Redis) RedisExtension {
	return &redisExtension{redis: redis}
}

// redisSetExpirationArgs 根据过期时间构造 SET 命令的过期参数。
//
// 该辅助函数确保 Redis 收到整数秒或整数毫秒 TTL；非正过期时间表示不设置过期参数。
//
// 参数：
//   - expiration: 业务侧传入的过期时间。
//
// 返回：
//   - []interface{}: 可追加到 SET 命令后的过期参数；不需要设置过期时间时返回 nil。
func redisSetExpirationArgs(expiration time.Duration) []interface{} {
	if expiration <= 0 {
		return nil
	}

	unit, ttl := redisExpirationUnitAndTTL(expiration)
	return []interface{}{unit, ttl}
}

// redisExpireArgs 根据过期时间构造过期相关命令参数。
//
// 该辅助函数对整秒过期使用 EXPIRE，对非整秒过期使用 PEXPIRE；非正过期时间使用 EXPIRE 0 保持 Redis 立即过期语义。
//
// 参数：
//   - key: 要设置过期语义的键名。
//   - expiration: 业务侧传入的过期时间。
//
// 返回：
//   - []interface{}: 可直接传给 Redis Do 方法的命令参数。
func redisExpireArgs(key string, expiration time.Duration) []interface{} {
	if expiration <= 0 {
		return []interface{}{"EXPIRE", key, int64(0)}
	}

	unit, ttl := redisExpirationUnitAndTTL(expiration)
	if unit == "EX" {
		return []interface{}{"EXPIRE", key, ttl}
	}
	return []interface{}{"PEXPIRE", key, ttl}
}

// redisExpirationUnitAndTTL 将过期时间转换为 Redis 可接受的整数 TTL。
//
// 该辅助函数优先使用整秒语义；非整秒过期时间向上取整到至少 1 毫秒，避免正过期时间被截断为 0。
//
// 参数：
//   - expiration: 正数过期时间。
//
// 返回：
//   - string: Redis 过期单位，取值为 EX 或 PX。
//   - int64: Redis 可接受的整数 TTL。
func redisExpirationUnitAndTTL(expiration time.Duration) (string, int64) {
	if expiration%time.Second == 0 {
		return "EX", int64(expiration / time.Second)
	}

	ttl := expiration / time.Millisecond
	if expiration%time.Millisecond != 0 {
		ttl++
	}
	if ttl < 1 {
		ttl = 1
	}
	return "PX", int64(ttl)
}

// Do 执行任意 Redis 命令。
//
// 参数：
//   - ctx: 控制命令执行生命周期的上下文。
//   - args: 命令名称及其参数，按 Redis 协议顺序传递。
//
// 返回：
//   - *Cmd: 通用命令结果；执行错误由返回值的 Err 方法承载。
func (r *redisExtension) Do(ctx context.Context, args ...interface{}) *Cmd {
	return r.redis.Do(ctx, args...)
}

// Pipelined 在管道中执行多个命令。
//
// 参数：
//   - ctx: 控制管道执行生命周期的上下文。
//   - fn: 向管道追加命令的回调；返回错误会中止管道执行。
//
// 返回：
//   - []Cmder: 管道内各命令的执行结果，顺序与追加顺序一致。
//   - error: fn 返回错误、上下文取消或底层 Redis 管道执行失败时返回错误。
func (r *redisExtension) Pipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error) {
	return r.redis.Pipelined(ctx, fn)
}

// TxPipelined 在 Redis 事务管道中执行多个命令。
//
// 参数：
//   - ctx: 控制事务管道执行生命周期的上下文。
//   - fn: 向事务管道追加命令的回调；返回错误会中止事务管道执行。
//
// 返回：
//   - []Cmder: 事务管道内各命令的执行结果，顺序与追加顺序一致。
//   - error: fn 返回错误、上下文取消或底层 Redis 事务管道执行失败时返回错误。
func (r *redisExtension) TxPipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error) {
	return r.redis.TxPipelined(ctx, fn)
}

// Subscribe 订阅一个或多个 Redis 频道。
//
// 参数：
//   - ctx: 控制订阅创建过程的上下文。
//   - channels: 要订阅的频道名称列表。
//
// 返回：
//   - *PubSub: 发布订阅客户端；调用方在不再接收消息时应关闭该客户端。
func (r *redisExtension) Subscribe(ctx context.Context, channels ...string) *PubSub {
	return r.redis.Subscribe(ctx, channels...)
}

// PSubscribe 按模式订阅一个或多个 Redis 频道。
//
// 参数：
//   - ctx: 控制订阅创建过程的上下文。
//   - channels: 要订阅的频道模式列表。
//
// 返回：
//   - *PubSub: 发布订阅客户端；调用方在不再接收消息时应关闭该客户端。
func (r *redisExtension) PSubscribe(ctx context.Context, channels ...string) *PubSub {
	return r.redis.PSubscribe(ctx, channels...)
}

// Eval 执行 Lua 脚本。
//
// 参数：
//   - ctx: 控制脚本执行生命周期的上下文。
//   - script: 要执行的 Lua 脚本文本。
//   - keys: 脚本使用的 Redis 键名列表。
//   - args: 传递给脚本的参数列表。
//
// 返回：
//   - *Cmd: 脚本执行结果；执行错误由返回值的 Err 方法承载。
func (r *redisExtension) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *Cmd {
	return r.redis.Eval(ctx, script, keys, args...)
}

// EvalRO 以只读模式执行 Lua 脚本。
//
// 参数：
//   - ctx: 控制脚本执行生命周期的上下文。
//   - script: 要执行的只读 Lua 脚本文本。
//   - keys: 脚本使用的 Redis 键名列表。
//   - args: 传递给脚本的参数列表。
//
// 返回：
//   - *Cmd: 脚本执行结果；执行错误由返回值的 Err 方法承载。
func (r *redisExtension) EvalRO(ctx context.Context, script string, keys []string, args ...interface{}) *Cmd {
	return r.redis.EvalRO(ctx, script, keys, args...)
}

// EvalSha 使用脚本 SHA1 执行已缓存的 Lua 脚本。
//
// 参数：
//   - ctx: 控制脚本执行生命周期的上下文。
//   - sha1: 已加载脚本的 SHA1 摘要。
//   - keys: 脚本使用的 Redis 键名列表。
//   - args: 传递给脚本的参数列表。
//
// 返回：
//   - *Cmd: 脚本执行结果；执行错误由返回值的 Err 方法承载。
func (r *redisExtension) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *Cmd {
	return r.redis.EvalSha(ctx, sha1, keys, args...)
}

// EvalShaRO 以只读模式使用脚本 SHA1 执行已缓存的 Lua 脚本。
//
// 参数：
//   - ctx: 控制脚本执行生命周期的上下文。
//   - sha1: 已加载脚本的 SHA1 摘要。
//   - keys: 脚本使用的 Redis 键名列表。
//   - args: 传递给脚本的参数列表。
//
// 返回：
//   - *Cmd: 脚本执行结果；执行错误由返回值的 Err 方法承载。
func (r *redisExtension) EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...interface{}) *Cmd {
	return r.redis.EvalShaRO(ctx, sha1, keys, args...)
}

// ScriptExists 检查脚本缓存中是否存在指定 SHA1。
//
// 参数：
//   - ctx: 控制命令执行生命周期的上下文。
//   - hashes: 要检查的脚本 SHA1 摘要列表。
//
// 返回：
//   - *BoolSliceCmd: 每个元素对应同位置 SHA1 是否存在；执行错误由返回值的 Err 方法承载。
func (r *redisExtension) ScriptExists(ctx context.Context, hashes ...string) *BoolSliceCmd {
	return r.redis.ScriptExists(ctx, hashes...)
}

// ScriptLoad 将 Lua 脚本加载到 Redis 脚本缓存。
//
// 参数：
//   - ctx: 控制命令执行生命周期的上下文。
//   - script: 要加载的 Lua 脚本文本。
//
// 返回：
//   - *StringCmd: 成功时包含脚本 SHA1 摘要；执行错误由返回值的 Err 方法承载。
func (r *redisExtension) ScriptLoad(ctx context.Context, script string) *StringCmd {
	return r.redis.ScriptLoad(ctx, script)
}

// Get 获取指定键的值。
//
// 参数：
//   - ctx: 控制命令执行生命周期的上下文。
//   - key: 要读取的 Redis 键名。
//
// 返回：
//   - *Cmd: GET 命令结果；键不存在时 Err 通常为 ErrNil。
func (r *redisExtension) Get(ctx context.Context, key string) *Cmd {
	return r.redis.Do(ctx, "GET", key)
}

// Set 设置指定键的值和可选过期时间。
//
// 参数：
//   - ctx: 控制命令执行生命周期的上下文。
//   - key: 要写入的 Redis 键名。
//   - value: 要写入的值，编码规则由底层 go-redis 客户端决定。
//   - expiration: 键过期时间；非正值表示不设置过期时间，整秒使用 EX，非整秒向上取整为毫秒并使用 PX。
//
// 返回：
//   - *Cmd: SET 命令结果；执行错误由返回值的 Err 方法承载。
func (r *redisExtension) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *Cmd {
	args := []interface{}{"SET", key, value}
	args = append(args, redisSetExpirationArgs(expiration)...)
	return r.redis.Do(ctx, args...)
}

// Del 删除指定键。
//
// 参数：
//   - ctx: 控制命令执行生命周期的上下文。
//   - key: 要删除的 Redis 键名。
//
// 返回：
//   - *Cmd: DEL 命令结果；执行错误由返回值的 Err 方法承载。
func (r *redisExtension) Del(ctx context.Context, key string) *Cmd {
	return r.redis.Do(ctx, "DEL", key)
}

// Expire 设置指定键的过期时间。
//
// 参数：
//   - ctx: 控制命令执行生命周期的上下文。
//   - key: 要设置过期语义的 Redis 键名。
//   - expiration: 键过期时间；正整秒使用 EXPIRE，正非整秒向上取整为毫秒并使用 PEXPIRE，非正值使用 EXPIRE 0。
//
// 返回：
//   - *Cmd: 过期命令结果；执行错误由返回值的 Err 方法承载。
func (r *redisExtension) Expire(ctx context.Context, key string, expiration time.Duration) *Cmd {
	return r.redis.Do(ctx, redisExpireArgs(key, expiration)...)
}

// ScriptFlush 按底层客户端能力清空脚本缓存。
//
// 当底层 Redis 实现提供 ScriptFlush(context.Context) *StatusCmd 时委托调用；否则返回 nil，调用方应处理 nil 返回值。
//
// 参数：
//   - ctx: 控制命令执行生命周期的上下文。
//
// 返回：
//   - *StatusCmd: 底层支持脚本缓存管理时返回对应命令；否则返回 nil。
func (r *redisExtension) ScriptFlush(ctx context.Context) *StatusCmd {
	if scriptFlusher, ok := r.redis.(interface {
		ScriptFlush(context.Context) *StatusCmd
	}); ok {
		return scriptFlusher.ScriptFlush(ctx)
	}
	return nil
}

// ScriptKill 按底层客户端能力终止当前正在执行的脚本。
//
// 当底层 Redis 实现提供 ScriptKill(context.Context) *StatusCmd 时委托调用；否则返回 nil，调用方应处理 nil 返回值。
//
// 参数：
//   - ctx: 控制命令执行生命周期的上下文。
//
// 返回：
//   - *StatusCmd: 底层支持脚本终止能力时返回对应命令；否则返回 nil。
func (r *redisExtension) ScriptKill(ctx context.Context) *StatusCmd {
	if scriptKiller, ok := r.redis.(interface {
		ScriptKill(context.Context) *StatusCmd
	}); ok {
		return scriptKiller.ScriptKill(ctx)
	}
	return nil
}

// Close 关闭底层 Redis 客户端并释放连接资源。
//
// 参数：无。
//
// 返回：
//   - error: 底层 Redis 实现关闭失败时返回错误。
func (r *redisExtension) Close() error {
	return r.redis.Close()
}
