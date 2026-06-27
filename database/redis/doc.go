// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package redis 提供对 go-redis/v9 客户端的轻量封装与常用扩展接口。
//
// NewRedis 创建基于 go-redis/v9 的客户端，并复用底层命令结果类型；返回的 Redis 实例持有底层连接资源，
// 调用方在不再使用时应调用 Close 释放连接。
//
// RedisExtension 在基础接口上补充常用 KV 与过期操作；ScriptFlush 和 ScriptKill 会按底层实现暴露的能力分派，
// 当通过 NewRedisExtension 包装的底层实现未提供对应方法时返回 nil，调用方需要显式处理。
package redis
