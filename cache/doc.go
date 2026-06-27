// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package cache 提供统一的内存缓存接口、Ristretto 默认实现、类型安全包装器和包级默认缓存函数。
//
// 本包的 Cache 常规 Get、GetWithTTL、Set、SetWithTTL 和 Delete 操作遵循具体实现的并发能力；
// 内置 Ristretto 后端支持这些常规读写操作并发调用。Clear 非原子，调用方应避免将其与读写操作并发执行；
// Close 后不得继续使用缓存，且不应与其它操作并发调用。NewCache 创建独立缓存实例，未显式传入配置时使用包内默认
// Ristretto 参数；调用方在实例不再使用时应调用 Close 释放底层资源。SetWithTTL 的非正 ttl 表示永不过期，
// GetWithTTL 使用 -1 表示永不过期，使用 0 表示键不存在或已过期。
//
// AsTypedCache 在现有 Cache 上提供泛型类型断言包装，类型不匹配时按未命中处理。InitCache、Get、Set 等
// 包级函数操作进程内默认缓存；默认缓存由 sync.Once 控制只初始化一次，首次调用使用的 Option 会固定为后续
// 全局访问配置，首次初始化失败后也不会自动重试。
package cache
