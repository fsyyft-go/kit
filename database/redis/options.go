// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package redis

type (
	// Option 定义 NewRedis 的函数式配置项。
	//
	// 参数：
	//   - *redisClient: 待修改的 Redis 客户端配置实例，调用方不应直接传入 nil。
	Option func(*redisClient)
)

var (
	// addrDefault 是 NewRedis 未显式配置地址时使用的 Redis 服务地址。
	addrDefault = "127.0.0.1:6379"
	// passwordDefault 是 NewRedis 未显式配置密码时使用的 Redis 认证密码。
	passwordDefault = "redis*2025"
)

// WithAddr 设置 Redis 服务器地址。
//
// 参数：
//   - addr: Redis 服务器地址，格式通常为 "host:port"。
//
// 返回：
//   - Option: 应用于 NewRedis 的地址配置项。
func WithAddr(addr string) Option {
	return func(o *redisClient) {
		o.addr = addr
	}
}

// WithPassword 设置 Redis 服务器认证密码。
//
// 参数：
//   - password: Redis 服务器认证密码；为空字符串表示不发送密码。
//
// 返回：
//   - Option: 应用于 NewRedis 的密码配置项。
func WithPassword(password string) Option {
	return func(o *redisClient) {
		o.password = password
	}
}
