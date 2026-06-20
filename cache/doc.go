// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package cache 提供统一的内存缓存接口、基于 Ristretto 的默认实现，以及全局默认缓存访问函数。
//
// Cache 定义 Get、GetWithTTL、Set、SetWithTTL、Delete、Clear 和 Close 等基础操作。
// NewCache 使用当前唯一的内置 Ristretto 后端创建独立缓存实例，并支持通过
// WithNumCounters、WithMaxCost 和 WithBufferItems 调整容量与并发参数；创建的独立实例在不再使用时应调用
// Close 释放底层资源。
// AsTypedCache 提供类型安全包装；InitCache、Get、Set 等函数操作包级默认缓存。
// 默认缓存通过 sync.Once 控制为单次初始化，首次调用采用的 Option 会固定为进程级配置，首次初始化失败后也不会
// 自动重试。
package cache
