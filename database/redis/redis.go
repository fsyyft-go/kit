// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package redis 提供了 Redis 客户端的基础类型和命令定义。
package redis

import (
	"context"
)

// Redis 接口定义了 Redis 客户端的基本操作。
type (
	Redis interface {
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
)
