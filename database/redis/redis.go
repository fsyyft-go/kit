// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package redis 提供了 Redis 客户端的基础类型和命令定义。
package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

// Redis 接口定义了 Redis 客户端的基本操作。
type (
	Redis interface {
		Scripter

		// Do 执行任意 Redis 命令。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制命令的执行
		//   - args：命令的参数列表
		//
		// 返回值：
		//   - *Cmd：通用的命令结果对象
		Do(ctx context.Context, args ...interface{}) *Cmd

		// Pipelined 在管道中执行多个命令。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制命令的执行
		//   - fn：定义要在管道中执行的命令的函数
		//
		// 返回值：
		//   - []Cmder：管道中所有命令的执行结果
		//   - error：执行过程中的错误信息
		Pipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error)

		// TxPipelined 在事务管道中执行多个命令。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制命令的执行
		//   - fn：定义要在事务管道中执行的命令的函数
		//
		// 返回值：
		//   - []Cmder：事务中所有命令的执行结果
		//   - error：执行过程中的错误信息
		TxPipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error)

		// Subscribe 订阅指定的频道。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制订阅的生命周期
		//   - channels：要订阅的频道列表
		//
		// 返回值：
		//   - *PubSub：发布订阅客户端对象
		Subscribe(ctx context.Context, channels ...string) *PubSub

		// PSubscribe 使用模式匹配订阅频道。
		//
		// 参数：
		//   - ctx：上下文对象，用于控制订阅的生命周期
		//   - channels：要订阅的频道模式列表
		//
		// 返回值：
		//   - *PubSub：发布订阅客户端对象
		PSubscribe(ctx context.Context, channels ...string) *PubSub
	}

	// redisClient 是 Redis 接口的具体实现。
	redisClient struct {
		// client 是底层的 Redis 客户端实例。
		client *goredis.Client

		// addr 是 Redis 服务器的地址。
		addr string
		// password 是连接 Redis 服务器的密码。
		password string
	}
)

// NewRedis 创建一个新的 Redis 客户端实例。
//
// 参数：
//   - opts：可选的配置选项
//
// 返回值：
//   - Redis：Redis 客户端接口实例
func NewRedis(opts ...Option) Redis {
	o := &redisClient{
		addr:     addrDefault,
		password: passwordDefault,
	}
	for _, opt := range opts {
		opt(o)
	}
	o.client = goredis.NewClient(&goredis.Options{
		Addr:     o.addr,
		Password: o.password,
	})
	return o
}

// Do 执行任意 Redis 命令。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//   - args：命令的参数列表
//
// 返回值：
//   - *Cmd：通用的命令结果对象
func (c *redisClient) Do(ctx context.Context, args ...interface{}) *Cmd {
	return c.client.Do(ctx, args...)
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
func (c *redisClient) Pipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error) {
	return c.client.Pipelined(ctx, fn)
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
func (c *redisClient) TxPipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error) {
	return c.client.TxPipelined(ctx, fn)
}

// Subscribe 订阅指定的频道。
//
// 参数：
//   - ctx：上下文对象，用于控制订阅的生命周期
//   - channels：要订阅的频道列表
//
// 返回值：
//   - *PubSub：发布订阅客户端对象
func (c *redisClient) Subscribe(ctx context.Context, channels ...string) *PubSub {
	return c.client.Subscribe(ctx, channels...)
}

// PSubscribe 使用模式匹配订阅频道。
//
// 参数：
//   - ctx：上下文对象，用于控制订阅的生命周期
//   - channels：要订阅的频道模式列表
//
// 返回值：
//   - *PubSub：发布订阅客户端对象
func (c *redisClient) PSubscribe(ctx context.Context, channels ...string) *PubSub {
	return c.client.PSubscribe(ctx, channels...)
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
func (c *redisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *Cmd {
	return c.client.Eval(ctx, script, keys, args...)
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
func (c *redisClient) EvalRO(ctx context.Context, script string, keys []string, args ...interface{}) *Cmd {
	return c.client.EvalRO(ctx, script, keys, args...)
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
func (c *redisClient) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *Cmd {
	return c.client.EvalSha(ctx, sha1, keys, args...)
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
func (c *redisClient) EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...interface{}) *Cmd {
	return c.client.EvalShaRO(ctx, sha1, keys, args...)
}

// ScriptExists 检查指定的脚本是否存在于脚本缓存中。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//   - hashes：要检查的脚本 SHA1 值列表
//
// 返回值：
//   - *BoolSliceCmd：检查结果，每个元素表示对应脚本是否存在
func (c *redisClient) ScriptExists(ctx context.Context, hashes ...string) *BoolSliceCmd {
	return c.client.ScriptExists(ctx, hashes...)
}

// ScriptLoad 将脚本加载到脚本缓存中。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//   - script：要加载的 Lua 脚本
//
// 返回值：
//   - *StringCmd：脚本的 SHA1 值
func (c *redisClient) ScriptLoad(ctx context.Context, script string) *StringCmd {
	return c.client.ScriptLoad(ctx, script)
}

// ScriptFlush 清空脚本缓存。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//
// 返回值：
//   - *StatusCmd：命令执行状态
func (c *redisClient) ScriptFlush(ctx context.Context) *StatusCmd {
	return c.client.ScriptFlush(ctx)
}

// ScriptKill 终止当前正在执行的脚本。
//
// 参数：
//   - ctx：上下文对象，用于控制命令的执行
//
// 返回值：
//   - *StatusCmd：命令执行状态
func (c *redisClient) ScriptKill(ctx context.Context) *StatusCmd {
	return c.client.ScriptKill(ctx)
}
