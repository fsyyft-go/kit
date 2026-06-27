// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package bloom

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kitlog "github.com/fsyyft-go/kit/log"
)

type recordingStoreCall struct {
	method string
	key    string
	hash   []uint64
}

type recordingStore struct {
	addErr      error
	existResult bool
	existErr    error
	calls       []recordingStoreCall
}

// Add 记录 Add 调用并返回预设错误。
//
// 参数：
//   - ctx: 调用上下文，本 helper 不依赖该参数。
//   - key: 被测 Bloom 传递给 Store 的 key。
//   - hash: 被测 Bloom 传递给 Store 的 hash 列表。
//
// 返回：
//   - error: 预设的 Add 错误。
func (s *recordingStore) Add(_ context.Context, key string, hash []uint64) error {
	s.calls = append(s.calls, recordingStoreCall{
		method: "add",
		key:    key,
		hash:   cloneHash(hash),
	})

	return s.addErr
}

// Exist 记录 Exist 调用并返回预设结果。
//
// 参数：
//   - ctx: 调用上下文，本 helper 不依赖该参数。
//   - key: 被测 Bloom 传递给 Store 的 key。
//   - hash: 被测 Bloom 传递给 Store 的 hash 列表。
//
// 返回：
//   - bool: 预设的存在性结果。
//   - error: 预设的 Exist 错误。
func (s *recordingStore) Exist(_ context.Context, key string, hash []uint64) (bool, error) {
	s.calls = append(s.calls, recordingStoreCall{
		method: "exist",
		key:    key,
		hash:   cloneHash(hash),
	})

	return s.existResult, s.existErr
}

// cloneHash 复制 hash 切片以避免后续修改影响调用记录。
//
// 参数：
//   - hash: 待复制的 hash 切片。
//
// 返回：
//   - []uint64: 复制后的 hash 切片。
func cloneHash(hash []uint64) []uint64 {
	if hash == nil {
		return nil
	}

	cloned := make([]uint64, len(hash))
	copy(cloned, hash)

	return cloned
}

// resetBloomNames 临时重置 bloom 名称注册表以隔离全局状态。
//
// 参数：
//   - t: 测试上下文，用于注册清理函数并标记辅助函数调用栈。
func resetBloomNames(t *testing.T) {
	t.Helper()

	previous := bloomNames
	bloomNames = make(map[string]string)

	t.Cleanup(func() {
		bloomNames = previous
	})
}

// TestNewBloomAdditional_Success 验证 NewBloom 的成功路径和自定义配置。
//
// 该测试覆盖默认配置和自定义 name、store、logger、n、p，确保创建后的实现字段符合配置契约。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewBloomAdditional_Success(t *testing.T) {
	logger, err := kitlog.NewLogger()
	require.NoError(t, err)

	customStore := NewMemoryStore(16)

	tests := []struct {
		name        string
		description string
		giveOpts    []Option
		assert      func(t *testing.T, got *bloom)
	}{
		{
			name:        "success/default-config",
			description: "验证 NewBloom 在默认参数下创建可用实例并计算 m/k。",
			giveOpts: []Option{
				WithName("new-bloom-default"),
			},
			assert: func(t *testing.T, got *bloom) {
				t.Helper()

				assert.Equal(t, "new-bloom-default", got.name)
				assert.True(t, got.store == storeDefault)
				assert.Equal(t, expectedElementsDefault, got.n)
				assert.Equal(t, falsePositiveRateDefault, got.p)
			},
		},
		{
			name:        "success/custom-config",
			description: "验证 NewBloom 正确应用自定义 name、store、logger、预计元素数和误判率。",
			giveOpts: []Option{
				WithName("new-bloom-custom"),
				WithStore(customStore),
				WithLogger(logger),
				WithExpectedElements(128),
				WithFalsePositiveRate(0.001),
			},
			assert: func(t *testing.T, got *bloom) {
				t.Helper()

				assert.Equal(t, "new-bloom-custom", got.name)
				assert.True(t, got.store == customStore)
				assert.Equal(t, logger, got.logger)
				assert.Equal(t, uint64(128), got.n)
				assert.Equal(t, 0.001, got.p)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			resetBloomNames(t)

			got, cleanup, err := NewBloom(tt.giveOpts...)

			require.NoError(t, err)
			require.NotNil(t, cleanup)
			t.Cleanup(cleanup)
			require.NotNil(t, got)

			impl, ok := got.(*bloom)
			require.True(t, ok)
			assert.NotZero(t, impl.m)
			assert.NotZero(t, impl.k)

			tt.assert(t, impl)
		})
	}
}

// TestBloomAdditional_PutContainStoreContract 验证 Put 与 Contain 向 Store 传递正确 key 和 hash。
//
// 该测试覆盖成功、未命中和错误传播场景，并断言同一 value 的 hash 与 impl.multiHash(value) 一致。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestBloomAdditional_PutContainStoreContract(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		description     string
		action          string
		giveValue       string
		giveExistResult bool
		giveErr         error
		wantMethod      string
		wantContain     bool
		wantErr         bool
	}{
		{
			name:        "success/put-passes-key-and-hash",
			description: "验证 Put 使用 Bloom name 作为 Store key，并传递 value 对应的 multiHash。",
			action:      "put",
			giveValue:   "value-a",
			wantMethod:  "add",
		},
		{
			name:        "error/put-propagates-store-error",
			description: "验证 Put 在 Store Add 返回错误时原样传播错误。",
			action:      "put",
			giveValue:   "value-b",
			giveErr:     assert.AnError,
			wantMethod:  "add",
			wantErr:     true,
		},
		{
			name:            "success/contain-existing-passes-key-and-hash",
			description:     "验证 Contain 使用 Bloom name 作为 Store key，并返回 Store 的命中结果。",
			action:          "contain",
			giveValue:       "value-c",
			giveExistResult: true,
			wantMethod:      "exist",
			wantContain:     true,
		},
		{
			name:            "success/contain-missing-passes-key-and-hash",
			description:     "验证 Contain 在 Store 返回未命中时返回 false。",
			action:          "contain",
			giveValue:       "value-d",
			giveExistResult: false,
			wantMethod:      "exist",
			wantContain:     false,
		},
		{
			name:        "error/contain-propagates-store-error",
			description: "验证 Contain 在 Store Exist 返回错误时原样传播错误。",
			action:      "contain",
			giveValue:   "value-e",
			giveErr:     assert.AnError,
			wantMethod:  "exist",
			wantErr:     true,
		},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			resetBloomNames(t)

			store := &recordingStore{existResult: tt.giveExistResult}
			if tt.action == "put" {
				store.addErr = tt.giveErr
			} else {
				store.existErr = tt.giveErr
			}

			bloomName := fmt.Sprintf("store-contract-%d", i)
			gotBloom, cleanup, err := NewBloom(
				WithName(bloomName),
				WithStore(store),
				WithExpectedElements(1000),
				WithFalsePositiveRate(0.01),
			)
			require.NoError(t, err)
			require.NotNil(t, cleanup)
			t.Cleanup(cleanup)

			impl, ok := gotBloom.(*bloom)
			require.True(t, ok)

			wantHash := impl.multiHash(tt.giveValue)
			assert.Equal(t, wantHash, impl.multiHash(tt.giveValue))

			switch tt.action {
			case "put":
				err = gotBloom.Put(ctx, tt.giveValue)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			case "contain":
				got, err := gotBloom.Contain(ctx, tt.giveValue)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				assert.Equal(t, tt.wantContain, got)
			default:
				require.Failf(t, "unsupported action", "action=%s", tt.action)
			}

			require.Len(t, store.calls, 1)
			call := store.calls[0]
			assert.Equal(t, tt.wantMethod, call.method)
			assert.Equal(t, bloomName, call.key)
			assert.Equal(t, wantHash, call.hash)
		})
	}
}

// TestBloomAdditional_GroupOperationsStoreContract 验证 GroupPut 与 GroupContain 使用正确分组 key 和 hash。
//
// 该测试覆盖不同 case 的 group/value，确保 Store key 始终为 name:group 且 hash 来源于对应 value。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestBloomAdditional_GroupOperationsStoreContract(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		description     string
		action          string
		giveGroup       string
		giveValue       string
		giveExistResult bool
		giveErr         error
		wantMethod      string
		wantContain     bool
		wantErr         bool
	}{
		{
			name:        "success/group-put-uses-case-group-and-value",
			description: "验证 GroupPut 使用当前 case 的 group 构造 name:group，并传递当前 case 的 value hash。",
			action:      "group-put",
			giveGroup:   "group-a",
			giveValue:   "value-a",
			wantMethod:  "add",
		},
		{
			name:        "error/group-put-propagates-store-error",
			description: "验证 GroupPut 在 Store Add 返回错误时原样传播错误。",
			action:      "group-put",
			giveGroup:   "group-error",
			giveValue:   "value-b",
			giveErr:     assert.AnError,
			wantMethod:  "add",
			wantErr:     true,
		},
		{
			name:            "success/group-contain-existing-uses-case-group-and-value",
			description:     "验证 GroupContain 使用当前 case 的 group/value，并返回 Store 命中结果。",
			action:          "group-contain",
			giveGroup:       "group-b",
			giveValue:       "value-c",
			giveExistResult: true,
			wantMethod:      "exist",
			wantContain:     true,
		},
		{
			name:            "success/group-contain-missing-uses-case-group-and-value",
			description:     "验证 GroupContain 在 Store 返回未命中时返回 false。",
			action:          "group-contain",
			giveGroup:       "group-c",
			giveValue:       "value-d",
			giveExistResult: false,
			wantMethod:      "exist",
			wantContain:     false,
		},
		{
			name:        "error/group-contain-propagates-store-error",
			description: "验证 GroupContain 在 Store Exist 返回错误时原样传播错误。",
			action:      "group-contain",
			giveGroup:   "group-d",
			giveValue:   "value-e",
			giveErr:     assert.AnError,
			wantMethod:  "exist",
			wantErr:     true,
		},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			resetBloomNames(t)

			store := &recordingStore{existResult: tt.giveExistResult}
			if tt.action == "group-put" {
				store.addErr = tt.giveErr
			} else {
				store.existErr = tt.giveErr
			}

			bloomName := fmt.Sprintf("group-contract-%d", i)
			gotBloom, cleanup, err := NewBloom(
				WithName(bloomName),
				WithStore(store),
				WithExpectedElements(1000),
				WithFalsePositiveRate(0.01),
			)
			require.NoError(t, err)
			require.NotNil(t, cleanup)
			t.Cleanup(cleanup)

			impl, ok := gotBloom.(*bloom)
			require.True(t, ok)

			wantHash := impl.multiHash(tt.giveValue)
			wantKey := impl.buildGroupKey(tt.giveGroup)
			assert.Equal(t, fmt.Sprintf("%s:%s", bloomName, tt.giveGroup), wantKey)

			switch tt.action {
			case "group-put":
				err = gotBloom.GroupPut(ctx, tt.giveGroup, tt.giveValue)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			case "group-contain":
				got, err := gotBloom.GroupContain(ctx, tt.giveGroup, tt.giveValue)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				assert.Equal(t, tt.wantContain, got)
			default:
				require.Failf(t, "unsupported action", "action=%s", tt.action)
			}

			require.Len(t, store.calls, 1)
			call := store.calls[0]
			assert.Equal(t, tt.wantMethod, call.method)
			assert.Equal(t, wantKey, call.key)
			assert.Equal(t, wantHash, call.hash)
		})
	}
}

// TestBloomAdditional_MemoryStoreBehavior 验证 Bloom 结合 memoryStore 时不存在 false negative 且分组隔离。
//
// 该测试使用真实 memoryStore 覆盖多值写入查询和 group-a/group-b 隔离，防止内存位图跨 key 污染。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestBloomAdditional_MemoryStoreBehavior(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T, bloom Bloom)
	}{
		{
			name:        "success/multiple-values-no-false-negative",
			description: "验证多个已 Put 的 value 后续 Contain 均返回 true。",
			assert: func(t *testing.T, bloom Bloom) {
				t.Helper()

				values := []string{"alpha", "beta", "gamma", "delta"}
				for _, value := range values {
					require.NoError(t, bloom.Put(ctx, value))
				}
				for _, value := range values {
					got, err := bloom.Contain(ctx, value)
					require.NoError(t, err)
					assert.True(t, got)
				}
			},
		},
		{
			name:        "success/group-a-and-group-b-isolated",
			description: "验证同一 value 写入 group-a 后不会在 group-b 中命中。",
			assert: func(t *testing.T, bloom Bloom) {
				t.Helper()

				require.NoError(t, bloom.GroupPut(ctx, "group-a", "shared-value"))

				gotA, err := bloom.GroupContain(ctx, "group-a", "shared-value")
				require.NoError(t, err)
				assert.True(t, gotA)

				gotB, err := bloom.GroupContain(ctx, "group-b", "shared-value")
				require.NoError(t, err)
				assert.False(t, gotB)
			},
		},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			resetBloomNames(t)

			gotBloom, cleanup, err := NewBloom(
				WithName(fmt.Sprintf("memory-store-%d", i)),
				WithStore(NewMemoryStore(64)),
				WithExpectedElements(100),
				WithFalsePositiveRate(0.01),
			)
			require.NoError(t, err)
			require.NotNil(t, cleanup)
			t.Cleanup(cleanup)

			tt.assert(t, gotBloom)
		})
	}
}
