// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package bloom

import (
	"context"
	"sync"
)

const (
	// defaultBlockSize 是未显式配置 size 时每个 key 使用的默认内存块大小，单位为 byte。
	// 设计考虑：
	// 1. 内存分配：
	//    - 每个 key 首次写入时分配固定大小的内存块，避免动态扩容。
	//    - 大小选择需要考虑实际应用场景和内存限制。
	// 2. 位图存储：
	//    - 每个 byte 包含 8 个 bit。
	//    - 128MB 可以提供约 10 亿个位图位置。
	// 3. 性能优化：
	//    - 内存对齐，提高访问效率。
	//    - 固定大小，避免内存碎片。
	defaultBlockSize = 128 * 1024 * 1024
)

// memoryStore 是基于内存位图的 Store 实现。
//
// memoryStore 使用 map[string][]byte 为每个 Store key 维护一份独立位图，避免最终 key
// 不同的位图互相污染。每个 key 首次写入时按固定 block size 分配位图，后续通过
// hash % (size*8) 将位置映射到该位图。它使用 RWMutex 保护 map 和位图访问，读操作可并发，
// 写操作串行化，并在 Exist 和 Add 中忽略 ctx。
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

// NewMemoryStore 创建一个新的内存 Store 实例。
//
// 参数：
//   - size: 每个 key 的内存块大小，单位为 byte；小于或等于 0 时使用 defaultBlockSize。
//
// 返回：
//   - *memoryStore: 初始化完成的内存 Store；不同 key 的位图相互隔离，并发访问由内部锁保护。
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
//   - data: 目标 key 对应的位图，调用方应保证 pos 位于该位图范围内。
//   - pos: 要设置的位位置。
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
//   - data: 目标 key 对应的位图，调用方应保证 pos 位于该位图范围内。
//   - pos: 要获取的位位置。
//
// 返回：
//   - bool: 位的值，true 表示 1，false 表示 0。
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

// Exist 实现 Store 接口，判断指定 key 下的 hash 位是否全部已设置。
//
// memoryStore 会忽略 ctx，并使用 key 选择独立位图；不同 key 之间互不影响。
//
// 参数：
//   - ctx: 调用上下文；当前实现忽略该参数。
//   - key: 位图命名空间标识；不同 key 对应独立内存位图。
//   - hash: 要判断的哈希值列表；每个值会按当前位图容量取模后定位到 bit。
//
// 返回：
//   - bool: 所有哈希值是否都已存在：
//   - false: key 不存在且 hash 非空，或至少有一个哈希位未设置。
//   - true: 所有哈希位均已设置；key 不存在但 hash 为空时也返回 true。
//   - error: 当前实现始终返回 nil。
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

// Add 实现 Store 接口，将指定 key 下的 hash 位全部设置为 1。
//
// memoryStore 会忽略 ctx，并按最终 Store key 选择独立内存命名空间；key 不同的位图互不影响。
//
// 参数：
//   - ctx: 调用上下文；当前实现忽略该参数。
//   - key: 位图命名空间标识；不同 key 对应独立内存位图。
//   - hash: 要写入的哈希值列表；每个值会按当前位图容量取模后定位到 bit。
//
// 返回：
//   - error: 当前实现始终返回 nil。
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
