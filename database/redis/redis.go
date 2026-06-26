// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

type (
	// Redis 定义本包暴露的 Redis 客户端基础能力。
	//
	// Redis 复用 go-redis/v9 的命令结果类型，单命令执行错误通常保存在返回的 Cmd 对象中；
	// 管道和事务管道还会通过返回的 error 承载回调、上下文或执行错误，调用方仍应检查各 Cmder.Err。Redis 实例持有底层连接资源，不再使用时应调用 Close。
	Redis interface {
		Scripter

		// Do 执行任意 Redis 命令。
		//
		// 参数：
		//   - ctx: 控制命令执行生命周期的上下文。
		//   - args: 命令名称及其参数，按 Redis 协议顺序传递。
		//
		// 返回：
		//   - *Cmd: 通用命令结果；执行错误由返回值的 Err 方法承载。
		Do(ctx context.Context, args ...interface{}) *Cmd

		// Pipelined 在管道中执行多个命令。
		//
		// 参数：
		//   - ctx: 控制管道执行生命周期的上下文。
		//   - fn: 向管道追加命令的回调；返回错误会中止管道执行。
		//
		// 返回：
		//   - []Cmder: 管道内各命令的执行结果，顺序与追加顺序一致；调用方仍应检查各 Cmder.Err。
		//   - error: fn 返回错误、上下文取消、网络异常或管道执行失败时返回错误。
		Pipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error)

		// TxPipelined 在 Redis 事务管道中执行多个命令。
		//
		// 参数：
		//   - ctx: 控制事务管道执行生命周期的上下文。
		//   - fn: 向事务管道追加命令的回调；返回错误会中止事务管道执行。
		//
		// 返回：
		//   - []Cmder: 事务管道内各命令的执行结果，顺序与追加顺序一致；调用方仍应检查各 Cmder.Err。
		//   - error: fn 返回错误、上下文取消、网络异常或事务管道执行失败时返回错误。
		TxPipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error)

		// Subscribe 订阅一个或多个 Redis 频道。
		//
		// 参数：
		//   - ctx: 控制订阅创建过程的上下文。
		//   - channels: 要订阅的频道名称列表。
		//
		// 返回：
		//   - *PubSub: 发布订阅客户端；调用方在不再接收消息时应关闭该客户端。
		Subscribe(ctx context.Context, channels ...string) *PubSub

		// PSubscribe 按模式订阅一个或多个 Redis 频道。
		//
		// 参数：
		//   - ctx: 控制订阅创建过程的上下文。
		//   - channels: 要订阅的频道模式列表。
		//
		// 返回：
		//   - *PubSub: 发布订阅客户端；调用方在不再接收消息时应关闭该客户端。
		PSubscribe(ctx context.Context, channels ...string) *PubSub

		// Close 关闭 Redis 客户端并释放底层连接资源。
		//
		// 参数：无。
		//
		// 返回：
		//   - error: 底层 go-redis 客户端关闭失败时返回错误。
		Close() error
	}

	// redisClient 是基于 go-redis/v9 Client 的 Redis 实现。
	redisClient struct {
		// client 是实际执行 Redis 命令的底层客户端。
		client *goredis.Client

		// addr 是 Redis 服务器地址。
		addr string
		// password 是连接 Redis 服务器时使用的认证密码。
		password string
	}
)

// NewRedis 创建一个 Redis 客户端。
//
// 未传入选项时，NewRedis 使用包内默认地址和默认密码创建 go-redis/v9 客户端。返回实例持有底层连接资源，
// 调用方在不再使用时应调用 Close。
//
// 参数：
//   - opts: 可选配置项，按传入顺序应用；后传入的同类配置会覆盖先前配置。
//
// 返回：
//   - Redis: 初始化完成的 Redis 客户端接口。
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
//   - ctx: 控制命令执行生命周期的上下文。
//   - args: 命令名称及其参数，按 Redis 协议顺序传递。
//
// 返回：
//   - *Cmd: 通用命令结果；执行错误由返回值的 Err 方法承载。
func (c *redisClient) Do(ctx context.Context, args ...interface{}) *Cmd {
	return c.client.Do(ctx, args...)
}

// Pipelined 在管道中执行多个命令。
//
// 参数：
//   - ctx: 控制管道执行生命周期的上下文。
//   - fn: 向管道追加命令的回调；返回错误会中止管道执行。
//
// 返回：
//   - []Cmder: 管道内各命令的执行结果，顺序与追加顺序一致；调用方仍应检查各 Cmder.Err。
//   - error: fn 返回错误、上下文取消、网络异常或管道执行失败时返回错误。
func (c *redisClient) Pipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error) {
	return c.client.Pipelined(ctx, fn)
}

// TxPipelined 在 Redis 事务管道中执行多个命令。
//
// 参数：
//   - ctx: 控制事务管道执行生命周期的上下文。
//   - fn: 向事务管道追加命令的回调；返回错误会中止事务管道执行。
//
// 返回：
//   - []Cmder: 事务管道内各命令的执行结果，顺序与追加顺序一致；调用方仍应检查各 Cmder.Err。
//   - error: fn 返回错误、上下文取消、网络异常或事务管道执行失败时返回错误。
func (c *redisClient) TxPipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error) {
	return c.client.TxPipelined(ctx, fn)
}

// Subscribe 订阅一个或多个 Redis 频道。
//
// 参数：
//   - ctx: 控制订阅创建过程的上下文。
//   - channels: 要订阅的频道名称列表。
//
// 返回：
//   - *PubSub: 发布订阅客户端；调用方在不再接收消息时应关闭该客户端。
func (c *redisClient) Subscribe(ctx context.Context, channels ...string) *PubSub {
	return c.client.Subscribe(ctx, channels...)
}

// PSubscribe 按模式订阅一个或多个 Redis 频道。
//
// 参数：
//   - ctx: 控制订阅创建过程的上下文。
//   - channels: 要订阅的频道模式列表。
//
// 返回：
//   - *PubSub: 发布订阅客户端；调用方在不再接收消息时应关闭该客户端。
func (c *redisClient) PSubscribe(ctx context.Context, channels ...string) *PubSub {
	return c.client.PSubscribe(ctx, channels...)
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
func (c *redisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *Cmd {
	return c.client.Eval(ctx, script, keys, args...)
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
func (c *redisClient) EvalRO(ctx context.Context, script string, keys []string, args ...interface{}) *Cmd {
	return c.client.EvalRO(ctx, script, keys, args...)
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
func (c *redisClient) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *Cmd {
	return c.client.EvalSha(ctx, sha1, keys, args...)
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
func (c *redisClient) EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...interface{}) *Cmd {
	return c.client.EvalShaRO(ctx, sha1, keys, args...)
}

// ScriptExists 检查脚本缓存中是否存在指定 SHA1。
//
// 参数：
//   - ctx: 控制命令执行生命周期的上下文。
//   - hashes: 要检查的脚本 SHA1 摘要列表。
//
// 返回：
//   - *BoolSliceCmd: 每个元素对应同位置 SHA1 是否存在；执行错误由返回值的 Err 方法承载。
func (c *redisClient) ScriptExists(ctx context.Context, hashes ...string) *BoolSliceCmd {
	return c.client.ScriptExists(ctx, hashes...)
}

// ScriptLoad 将 Lua 脚本加载到 Redis 脚本缓存。
//
// 参数：
//   - ctx: 控制命令执行生命周期的上下文。
//   - script: 要加载的 Lua 脚本文本。
//
// 返回：
//   - *StringCmd: 成功时包含脚本 SHA1 摘要；执行错误由返回值的 Err 方法承载。
func (c *redisClient) ScriptLoad(ctx context.Context, script string) *StringCmd {
	return c.client.ScriptLoad(ctx, script)
}

// ScriptFlush 清空 Redis 脚本缓存。
//
// 参数：
//   - ctx: 控制命令执行生命周期的上下文。
//
// 返回：
//   - *StatusCmd: 脚本缓存清理命令结果；执行错误由返回值的 Err 方法承载。
func (c *redisClient) ScriptFlush(ctx context.Context) *StatusCmd {
	return c.client.ScriptFlush(ctx)
}

// ScriptKill 终止当前正在执行的 Lua 脚本。
//
// 参数：
//   - ctx: 控制命令执行生命周期的上下文。
//
// 返回：
//   - *StatusCmd: 脚本终止命令结果；执行错误由返回值的 Err 方法承载。
func (c *redisClient) ScriptKill(ctx context.Context) *StatusCmd {
	return c.client.ScriptKill(ctx)
}

// Close 关闭 Redis 客户端并释放底层连接资源。
//
// 参数：无。
//
// 返回：
//   - error: 底层 go-redis 客户端关闭失败时返回错误。
func (c *redisClient) Close() error {
	return c.client.Close()
}
