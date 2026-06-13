// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package bloom 提供了布隆过滤器的接口定义和实现。
package bloom

import (
	"context"
	"sync"
)

const (
	// defaultBlockSize 默认内存块大小，128MB。
	// 设计考虑：
	// 1. 内存分配：
	//    - 预分配固定大小的内存块，避免动态扩容。
	//    - 大小选择需要考虑实际应用场景和内存限制。
	// 2. 位图存储：
	//    - 每个 byte 包含 8 个 bit。
	//    - 128MB 可以存储约 10 亿个元素（128 * 1024 * 1024 * 8）。
	// 3. 性能优化：
	//    - 内存对齐，提高访问效率。
	//    - 固定大小，避免内存碎片。
	defaultBlockSize = 128 * 1024 * 1024
)

// memoryStore 是基于内存块的布隆过滤器存储实现。
// 该实现为每个 key 维护独立的固定大小内存块，每个 bit 表示一个 hash 值是否存在。
// 设计原理：
// 1. 数据结构：
//   - 使用 map[string][]byte 按 key 保存底层位图。
//   - 每个 key 的 []byte 是独立位图，避免不同 Bloom name 或 group 互相污染。
//   - 使用读写锁保证并发安全。
//
// 2. 内存管理：
//   - 每个 key 首次写入时分配固定大小的内存块。
//   - 使用取模运算将 hash 值映射到对应 key 的内存块。
//   - 支持自定义内存块大小。
//
// 3. 并发安全：
//   - 使用读写锁（RWMutex）保护 map 和位图访问。
//   - 读操作使用读锁，允许多个读操作并发。
//   - 写操作使用写锁，保证写操作的原子性。
type memoryStore struct {
	// mu 用于保护 data 的并发访问。
	// 设计考虑：
	// 1. 锁粒度：
	//    - 使用读写锁而不是互斥锁，提高并发性能。
	//    - 读操作可以并发执行，提高查询效率。
	// 2. 性能影响：
	//    - 读操作使用读锁，对性能影响较小。
	//    - 写操作使用写锁，可能成为性能瓶颈。
	mu sync.RWMutex

	// data 按 key 存储实际的位数据。
	// 设计考虑：
	// 1. 存储结构：
	//    - 每个 key 对应一个 []byte 位图。
	//    - 每个 byte 包含 8 个 bit。
	//    - 每个 bit 表示一个 hash 值是否存在。
	// 2. 内存访问：
	//    - 使用位操作提高访问效率。
	//    - 同一 key 的内存连续，提高缓存命中率。
	data map[string][]byte

	// size 内存块的大小，以 byte 为单位。
	// 设计考虑：
	// 1. 大小选择：
	//    - 需要根据实际应用场景选择合适的大小。
	//    - 太小会导致误判率增加。
	//    - 太大会浪费内存。
	size int
}

// NewMemoryStore 创建一个新的内存存储实例。
//
// 参数：
//   - size：每个 key 的内存块大小，以 byte 为单位。如果为 0，则使用默认大小。
//
// 返回值：
//   - *memoryStore：新的内存存储实例
func NewMemoryStore(size int) *memoryStore {
	// 设计考虑：
	// 1. 内存分配：
	//   - 初始化 key 到位图的映射。
	//   - 每个 key 的位图在首次写入时按固定大小分配。
	//
	// 2. 初始化：
	//   - 尚未写入任何 key 时不预创建位图。
	//   - 设置每个 key 对应位图的内存块大小。
	if size <= 0 {
		size = defaultBlockSize
	}
	return &memoryStore{
		data: make(map[string][]byte),
		size: size,
	}
}

// setBit 设置指定位图位置的位为 1。
//
// 参数：
//   - data：目标 key 对应的位图
//   - pos：要设置的位的位置
func (s *memoryStore) setBit(data []byte, pos uint64) {
	// 计算 byte 位置和 bit 位置。
	// 设计原理：
	// 1. 位操作：
	//   - 使用位运算提高性能。
	//   - 使用或运算设置位。
	//
	// 2. 位置计算：
	//   - 使用除法和取模计算 byte 和 bit 位置。
	//   - 保证位置在有效范围内。

	bytePos := pos / 8
	bitPos := pos % 8

	// 设置对应的位。
	data[bytePos] |= 1 << bitPos
}

// getBit 获取指定位图位置的位的值。
//
// 参数：
//   - data：目标 key 对应的位图
//   - pos：要获取的位的位置
//
// 返回值：
//   - bool：位的值，true 表示 1，false 表示 0
func (s *memoryStore) getBit(data []byte, pos uint64) bool {
	// 计算 byte 位置和 bit 位置。
	// 设计原理：
	// 1. 位操作：
	//   - 使用位运算提高性能。
	//   - 使用与运算获取位值。
	//
	// 2. 位置计算：
	//   - 使用除法和取模计算 byte 和 bit 位置。
	//   - 保证位置在有效范围内。
	bytePos := pos / 8
	bitPos := pos % 8

	// 获取对应的位。
	return (data[bytePos] & (1 << bitPos)) != 0
}

// Exist 实现了 Store 接口的 Exist 方法，判断指定 key 对应的所有 hash 值是否都已存在。
//
// 参数：
//   - ctx：上下文对象，用于控制请求的生命周期
//   - key：存储键名
//   - hash：要判断的哈希值列表
//
// 返回值：
//   - bool：所有哈希值是否都已存在
//   - false：至少有一个哈希值不存在
//   - true：所有哈希值都存在
//   - error：查询过程中发生的错误
func (s *memoryStore) Exist(_ context.Context, key string, hash []uint64) (bool, error) {
	// 设计原理：
	// 1. 查询流程：
	//   - 根据 key 找到对应位图。
	//   - 遍历所有 hash 值。
	//   - 检查每个 hash 值对应的位。
	//   - 如果任一位置为 0，返回 false。
	//
	// 2. 并发安全：
	//   - 使用读锁保护数据访问。
	//   - 允许多个读操作并发执行。
	//
	// 3. 性能优化：
	//   - 使用位操作提高查询效率。
	//   - 提前返回，减少不必要的检查。
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.data[key]
	if !ok {
		return len(hash) == 0, nil
	}

	// 检查所有 hash 值是否都存在。
	for _, h := range hash {
		// 计算位的位置。
		pos := h % (uint64(s.size) * 8)
		if !s.getBit(data, pos) {
			return false, nil
		}
	}

	return true, nil
}

// Add 实现了 Store 接口的 Add 方法，将一组 hash 值添加到存储中。
//
// 参数：
//   - ctx：上下文对象，用于控制请求的生命周期
//   - key：存储键名
//   - hash：要添加的哈希值列表
//
// 返回值：
//   - error：添加过程中发生的错误
func (s *memoryStore) Add(_ context.Context, key string, hash []uint64) error {
	// 设计原理：
	// 1. 添加流程：
	//   - 根据 key 获取或创建独立位图。
	//   - 遍历所有 hash 值。
	//   - 设置每个 hash 值对应的位。
	//
	// 2. 并发安全：
	//   - 使用写锁保护数据访问。
	//   - 保证写操作的原子性。
	//
	// 3. 性能优化：
	//   - 使用位操作提高设置效率。
	//   - 批量设置，减少锁的获取和释放。
	s.mu.Lock()
	defer s.mu.Unlock()

	data, ok := s.data[key]
	if !ok {
		data = make([]byte, s.size)
		s.data[key] = data
	}

	// 设置所有 hash 值对应的位。
	for _, h := range hash {
		// 计算位的位置。
		pos := h % (uint64(s.size) * 8)
		s.setBit(data, pos)
	}

	return nil
}
