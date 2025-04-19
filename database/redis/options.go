// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package redis

type (
	// Option 定义了 Redis 客户端的配置选项类型。
	// 用于在创建 Redis 客户端时进行自定义配置。
	Option func(*redisClient)
)

// WithAddr 设置 Redis 服务器的地址。
//
// 参数：
//   - addr：Redis 服务器的地址，格式为 "host:port"
//
// 返回值：
//   - Option：配置选项函数
func WithAddr(addr string) Option {
	return func(o *redisClient) {
		o.addr = addr
	}
}

// WithPassword 设置连接 Redis 服务器的密码。
//
// 参数：
//   - password：Redis 服务器的认证密码
//
// 返回值：
//   - Option：配置选项函数
func WithPassword(password string) Option {
	return func(o *redisClient) {
		o.password = password
	}
}
