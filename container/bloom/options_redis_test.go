// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package bloom

import (
	"errors"
	"testing"

	kitredis "github.com/fsyyft-go/kit/database/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWithRedis_OptionBehavior 验证 WithRedis 使用 Redis 客户端构造 Store 的配置行为。
//
// 该测试使用 fake Redis 覆盖脚本加载成功、set 脚本加载失败和 get 脚本加载失败语义，确保 Option 不访问真实 Redis 且错误路径通过 panic 暴露。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestWithRedis_OptionBehavior(t *testing.T) {
	loadSetErr := errors.New("load set script failed")
	loadGetErr := errors.New("load get script failed")

	tests := []struct {
		name                string
		description         string
		giveScriptLoadErrs  []error
		wantPanicErr        error
		wantScriptLoadCount int
		wantSetScriptHash   string
		wantGetScriptHash   string
	}{
		{
			name:                "success/load-scripts-and-set-redis-store",
			description:         "验证 WithRedis 成功加载 Redis Lua 脚本后将 Bloom store 替换为 redisStore。",
			giveScriptLoadErrs:  []error{nil, nil},
			wantScriptLoadCount: 2,
			wantSetScriptHash:   "with-redis-sha-0",
			wantGetScriptHash:   "with-redis-sha-1",
		},
		{
			name:                "panic/load-set-script-failed",
			description:         "验证 WithRedis 在 set 脚本加载失败时 panic 并保持原 store 未被替换。",
			giveScriptLoadErrs:  []error{loadSetErr},
			wantPanicErr:        loadSetErr,
			wantScriptLoadCount: 1,
		},
		{
			name:                "panic/load-get-script-failed",
			description:         "验证 WithRedis 在 get 脚本加载失败时 panic 并保持原 store 未被替换。",
			giveScriptLoadErrs:  []error{nil, loadGetErr},
			wantPanicErr:        loadGetErr,
			wantScriptLoadCount: 2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			scriptLoadResults := make([]*kitredis.StringCmd, 0, len(tt.giveScriptLoadErrs))
			for i, err := range tt.giveScriptLoadErrs {
				scriptLoadResults = append(scriptLoadResults, newRedisStringCmdResult(t.Context(), "with-redis-sha-"+string(rune('0'+i)), err))
			}
			fake := &fakeRedis{scriptLoadResults: scriptLoadResults}
			initialStore := NewMemoryStore(1)
			target := &bloom{store: initialStore}
			apply := func() {
				WithRedis(fake)(target)
			}

			if tt.wantPanicErr != nil {
				require.PanicsWithError(t, tt.wantPanicErr.Error(), apply)
				assert.Same(t, initialStore, target.store)
			} else {
				require.NotPanics(t, apply)

				gotStore, ok := target.store.(*redisStore)
				require.True(t, ok)
				gotRedis, ok := gotStore.redis.(*fakeRedis)
				require.True(t, ok)
				assert.Same(t, fake, gotRedis)
				assert.Equal(t, tt.wantSetScriptHash, gotStore.setScriptHash)
				assert.Equal(t, tt.wantGetScriptHash, gotStore.getScriptHash)
			}

			assert.Len(t, fake.scriptLoadCalls, tt.wantScriptLoadCount)
			if tt.wantScriptLoadCount >= 1 {
				assert.Equal(t, bloomSetScript, fake.scriptLoadCalls[0].script)
			}
			if tt.wantScriptLoadCount >= 2 {
				assert.Equal(t, bloomGetScript, fake.scriptLoadCalls[1].script)
			}
		})
	}
}
