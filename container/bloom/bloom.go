// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package bloom 提供了布隆过滤器的接口定义和实现。
// 布隆过滤器是一种空间效率很高的概率型数据结构，用于判断一个元素是否在集合中。
package bloom

import (
	"context"
)

type (
	// Bloom 定义了布隆过滤器的核心接口。
	// 该接口提供了基本的元素判断和添加功能，以及分组操作的支持。
	Bloom interface {
		// Contain 用于判断指定元素是否可能存在于布隆过滤器中。
		// 返回值说明：
		// - false：元素肯定不存在。
		// - true：元素可能存在（存在误判可能）。
		// - error：操作过程中发生的错误。
		Contain(ctx context.Context, value string) (bool, error)

		// Put 将指定元素添加到布隆过滤器中。
		// 返回值说明：
		// - error：添加过程中发生的错误。
		Put(ctx context.Context, value string) error

		// GroupContain 用于判断指定分组中是否可能包含指定元素。
		// 返回值说明：
		// - false：元素在指定分组中肯定不存在。
		// - true：元素在指定分组中可能存在（存在误判可能）。
		// - error：操作过程中发生的错误。
		GroupContain(ctx context.Context, group string, value string) (bool, error)

		// GroupPut 将指定元素添加到指定分组的布隆过滤器中。
		// 返回值说明：
		// - error：添加过程中发生的错误。
		GroupPut(ctx context.Context, group string, value string) error
	}

	// Store 定义了布隆过滤器底层数据存储的接口。
	// 该接口负责实际的数据存储和查询操作。
	Store interface {
		// Exist 用于判断指定 key 对应的所有 hash 值是否都已存在。
		// 返回值说明：
		// - false：至少有一个 hash 值不存在。
		// - true：所有 hash 值都存在。
		// - error：查询过程中发生的错误。
		Exist(ctx context.Context, key string, hash []uint64) (bool, error)

		// Add 将一组 hash 值添加到指定 key 对应的存储中。
		// 返回值说明：
		// - error：添加过程中发生的错误。
		Add(ctx context.Context, key string, hash []uint64) error
	}
)
