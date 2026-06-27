// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package bloom

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMemoryStoreAdditional_Config 验证内存存储初始化时的容量与 key 位图映射。
//
// 该测试通过表驱动用例覆盖默认容量、负数容量和指定容量，确保初始化契约稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestMemoryStoreAdditional_Config(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveSize    int
		wantSize    int
	}{
		{
			name:        "success/default-size",
			description: "验证 size 为 0 时使用默认内存块大小且不预创建 key 位图。",
			giveSize:    0,
			wantSize:    defaultBlockSize,
		},
		{
			name:        "success/negative-size",
			description: "验证 size 为负数时使用默认内存块大小且不预创建 key 位图。",
			giveSize:    -1,
			wantSize:    defaultBlockSize,
		},
		{
			name:        "success/custom-size",
			description: "验证指定正数 size 时按该容量初始化存储。",
			giveSize:    16,
			wantSize:    16,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			store := NewMemoryStore(tt.giveSize)

			require.NotNil(t, store)
			assert.Equal(t, tt.wantSize, store.size)
			assert.NotNil(t, store.data)
			assert.Empty(t, store.data)
		})
	}
}

// TestMemoryStoreAdditional_AddExistBehavior 验证内存存储 Add 与 Exist 的核心行为。
//
// 该测试通过表驱动用例覆盖 key 缺失、部分命中、位边界、取模、空 hash 和 key 隔离语义。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestMemoryStoreAdditional_AddExistBehavior(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		description string
		giveSize    int
		setup       func(t *testing.T, store *memoryStore)
		giveKey     string
		giveHash    []uint64
		wantExist   bool
	}{
		{
			name:        "error/missing-key-non-empty-hash",
			description: "验证 key 不存在且 hash 非空时返回 false。",
			giveSize:    2,
			giveKey:     "missing",
			giveHash:    []uint64{1},
			wantExist:   false,
		},
		{
			name:        "success/add-and-exist",
			description: "验证 Add 写入指定 key 后，Exist 对相同 hash 返回 true。",
			giveSize:    2,
			setup: func(t *testing.T, store *memoryStore) {
				t.Helper()
				require.NoError(t, store.Add(ctx, "alpha", []uint64{1, 2, 3}))
			},
			giveKey:   "alpha",
			giveHash:  []uint64{1, 2, 3},
			wantExist: true,
		},
		{
			name:        "error/partial-hash-missing",
			description: "验证 key 已存在但任一 hash 位未设置时返回 false。",
			giveSize:    2,
			setup: func(t *testing.T, store *memoryStore) {
				t.Helper()
				require.NoError(t, store.Add(ctx, "alpha", []uint64{1, 2, 3}))
			},
			giveKey:   "alpha",
			giveHash:  []uint64{1, 2, 4},
			wantExist: false,
		},
		{
			name:        "boundary/bit-positions",
			description: "验证 0、7、8、size*8-1 等 byte 与 bit 边界位置可正确写入和读取。",
			giveSize:    2,
			setup: func(t *testing.T, store *memoryStore) {
				t.Helper()
				require.NoError(t, store.Add(ctx, "alpha", []uint64{0, 7, 8, 15}))
			},
			giveKey:   "alpha",
			giveHash:  []uint64{0, 7, 8, 15},
			wantExist: true,
		},
		{
			name:        "boundary/modulo-size-times-eight",
			description: "验证 hash 等于 size*8 时按位图容量取模并映射到 bit 0。",
			giveSize:    2,
			setup: func(t *testing.T, store *memoryStore) {
				t.Helper()
				require.NoError(t, store.Add(ctx, "alpha", []uint64{16}))
			},
			giveKey:   "alpha",
			giveHash:  []uint64{0},
			wantExist: true,
		},
		{
			name:        "boundary/empty-hash-missing-key",
			description: "验证 key 不存在但 hash 为空时保持全称判断语义并返回 true。",
			giveSize:    2,
			giveKey:     "missing",
			giveHash:    []uint64{},
			wantExist:   true,
		},
		{
			name:        "success/key-isolation",
			description: "验证相同 hash 写入 group-a 后不会污染 group-b 的位图。",
			giveSize:    2,
			setup: func(t *testing.T, store *memoryStore) {
				t.Helper()
				require.NoError(t, store.Add(ctx, "group-a", []uint64{5}))
			},
			giveKey:   "group-b",
			giveHash:  []uint64{5},
			wantExist: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			store := NewMemoryStore(tt.giveSize)
			if tt.setup != nil {
				tt.setup(t, store)
			}

			got, err := store.Exist(ctx, tt.giveKey, tt.giveHash)

			require.NoError(t, err)
			assert.Equal(t, tt.wantExist, got)
		})
	}
}

// TestMemoryStoreAdditional_Concurrent 验证内存存储在并发 Add 与 Exist 下保持数据一致性。
//
// 该测试启动多个 goroutine 对同一 key 写入和读取不同 hash，确保读写锁保护下不会产生 false negative。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestMemoryStoreAdditional_Concurrent(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore(16)

	const workerCount = 32

	var wg sync.WaitGroup
	errCh := make(chan error, workerCount)

	for i := 0; i < workerCount; i++ {
		i := i
		wg.Add(1)

		go func() {
			defer wg.Done()

			hash := []uint64{uint64(i), uint64(i + workerCount)}
			if err := store.Add(ctx, "shared", hash); err != nil {
				errCh <- fmt.Errorf("add worker %d: %w", i, err)
				return
			}

			got, err := store.Exist(ctx, "shared", hash)
			if err != nil {
				errCh <- fmt.Errorf("exist worker %d: %w", i, err)
				return
			}
			if !got {
				errCh <- fmt.Errorf("worker %d inserted hashes were not found", i)
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		assert.NoError(t, err)
	}

	for i := 0; i < workerCount; i++ {
		hash := []uint64{uint64(i), uint64(i + workerCount)}
		got, err := store.Exist(ctx, "shared", hash)
		require.NoError(t, err)
		assert.True(t, got)
	}
}
