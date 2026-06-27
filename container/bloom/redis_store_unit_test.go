// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package bloom

import (
	"context"
	"errors"
	"fmt"
	"testing"

	kitredis "github.com/fsyyft-go/kit/database/redis"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRedisEvalCall struct {
	sha  string
	keys []string
	args []any
}

type fakeRedisScriptLoadCall struct {
	script string
}

type fakeRedis struct {
	kitredis.Redis

	evalShaResults    []*kitredis.Cmd
	scriptLoadResults []*kitredis.StringCmd

	evalShaCalls    []fakeRedisEvalCall
	scriptLoadCalls []fakeRedisScriptLoadCall
}

// EvalSha 记录 EvalSha 调用并返回预设结果。
//
// 参数：
//   - ctx: 调用上下文。
//   - sha1: Redis 脚本 hash。
//   - keys: Redis Lua KEYS 参数。
//   - args: Redis Lua ARGV 参数。
//
// 返回：
//   - *kitredis.Cmd: 预设 Redis 命令结果。
func (f *fakeRedis) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *kitredis.Cmd {
	cmd := newRedisCmdResult(ctx, nil, errors.New("unexpected EvalSha call"))
	if len(f.evalShaResults) > len(f.evalShaCalls) {
		cmd = f.evalShaResults[len(f.evalShaCalls)]
	}

	copiedKeys := append([]string(nil), keys...)
	copiedArgs := append([]any(nil), args...)
	f.evalShaCalls = append(f.evalShaCalls, fakeRedisEvalCall{
		sha:  sha1,
		keys: copiedKeys,
		args: copiedArgs,
	})

	return cmd
}

// ScriptLoad 记录 ScriptLoad 调用并返回预设结果。
//
// 参数：
//   - ctx: 调用上下文。
//   - script: 待加载的 Lua 脚本。
//
// 返回：
//   - *kitredis.StringCmd: 预设 Redis 字符串命令结果。
func (f *fakeRedis) ScriptLoad(ctx context.Context, script string) *kitredis.StringCmd {
	cmd := newRedisStringCmdResult(ctx, "", errors.New("unexpected ScriptLoad call"))
	if len(f.scriptLoadResults) > len(f.scriptLoadCalls) {
		cmd = f.scriptLoadResults[len(f.scriptLoadCalls)]
	}

	f.scriptLoadCalls = append(f.scriptLoadCalls, fakeRedisScriptLoadCall{script: script})
	return cmd
}

// newRedisCmdResult 构造带预设值或错误的 Redis Cmd。
//
// 参数：
//   - ctx: 命令上下文。
//   - value: 命令返回值。
//   - err: 命令错误。
//
// 返回：
//   - *kitredis.Cmd: 可供被测代码调用 Result 的命令对象。
func newRedisCmdResult(ctx context.Context, value any, err error) *kitredis.Cmd {
	cmd := goredis.NewCmd(ctx)
	if err != nil {
		cmd.SetErr(err)
		return cmd
	}
	cmd.SetVal(value)
	return cmd
}

// newRedisStringCmdResult 构造带预设值或错误的 Redis StringCmd。
//
// 参数：
//   - ctx: 命令上下文。
//   - value: 命令返回值。
//   - err: 命令错误。
//
// 返回：
//   - *kitredis.StringCmd: 可供被测代码调用 Result 的字符串命令对象。
func newRedisStringCmdResult(ctx context.Context, value string, err error) *kitredis.StringCmd {
	cmd := goredis.NewStringCmd(ctx)
	if err != nil {
		cmd.SetErr(err)
		return cmd
	}
	cmd.SetVal(value)
	return cmd
}

// TestRedisStoreUnit_GenerateKeysAndArgs 验证 RedisStore 脚本参数生成逻辑。
//
// 该测试为纯单元测试，不访问 Redis，覆盖 Redis key 格式化和 hash 到 ARGV 的转换。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRedisStoreUnit_GenerateKeysAndArgs(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveName    string
		giveHash    []uint64
		wantKeys    []string
		wantArgs    []any
	}{
		{
			name:        "success/multiple-hashes",
			description: "验证多个 hash 被转换为 Redis Lua ARGV，且 key 使用统一前缀格式。",
			giveName:    "alpha",
			giveHash:    []uint64{0, 7, 8},
			wantKeys:    []string{fmt.Sprintf(redisKeyFormat, "alpha")},
			wantArgs:    []any{uint64(0), uint64(7), uint64(8)},
		},
		{
			name:        "boundary/empty-hash",
			description: "验证 hash 为空时仍生成 Redis key，ARGV 保持为空切片。",
			giveName:    "empty",
			giveHash:    []uint64{},
			wantKeys:    []string{fmt.Sprintf(redisKeyFormat, "empty")},
			wantArgs:    []any{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			store := &redisStore{}
			gotKeys, gotArgs := store.generateKeysAndArgs(tt.giveName, tt.giveHash)

			assert.Equal(t, tt.wantKeys, gotKeys)
			assert.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}

// TestRedisStoreUnit_Exist 验证 RedisStore Exist 对脚本返回值和错误分支的处理。
//
// 该测试使用 fake Redis 覆盖全量命中、部分未命中、返回类型错误和 EvalSha 失败语义，确保查询契约稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRedisStoreUnit_Exist(t *testing.T) {
	evalErr := errors.New("eval failed")

	tests := []struct {
		name        string
		description string
		giveResult  any
		giveErr     error
		wantExist   bool
		wantErrIs   error
	}{
		{
			name:        "success/all-bits-existed",
			description: "验证 Redis 脚本返回的所有位均为 1 时 Exist 返回 true。",
			giveResult:  []any{int64(1), int64(1), int64(1)},
			wantExist:   true,
		},
		{
			name:        "success/partial-bit-missing",
			description: "验证 Redis 脚本返回值存在非 1 位时 Exist 返回 false 且无错误。",
			giveResult:  []any{int64(1), int64(0), int64(1)},
			wantExist:   false,
		},
		{
			name:        "error/result-type-not-array",
			description: "验证 Redis 脚本返回非数组结果时 Exist 返回结果类型错误。",
			giveResult:  "not-array",
			wantErrIs:   ErrResultTypeNotArray,
		},
		{
			name:        "error/eval-sha-failed",
			description: "验证 EvalSha 返回非 NOSCRIPT 错误时 Exist 透传该错误。",
			giveErr:     evalErr,
			wantErrIs:   evalErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			fake := &fakeRedis{
				evalShaResults: []*kitredis.Cmd{
					newRedisCmdResult(t.Context(), tt.giveResult, tt.giveErr),
				},
			}
			store := &redisStore{
				redis:         fake,
				getScriptHash: "get-sha",
			}

			gotExist, err := store.Exist(t.Context(), "unit", []uint64{1, 2, 3})

			if tt.wantErrIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErrIs)
				assert.False(t, gotExist)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantExist, gotExist)
			require.Len(t, fake.evalShaCalls, 1)
			assert.Equal(t, "get-sha", fake.evalShaCalls[0].sha)
			assert.Equal(t, []string{fmt.Sprintf(redisKeyFormat, "unit")}, fake.evalShaCalls[0].keys)
			assert.Equal(t, []any{uint64(1), uint64(2), uint64(3)}, fake.evalShaCalls[0].args)
		})
	}
}

// TestRedisStoreUnit_Add 验证 RedisStore Add 对脚本执行成功和失败的处理。
//
// 该测试使用 fake Redis 覆盖写入成功和 EvalSha 错误透传语义，确保添加契约稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRedisStoreUnit_Add(t *testing.T) {
	addErr := errors.New("add failed")

	tests := []struct {
		name        string
		description string
		giveErr     error
		wantErrIs   error
	}{
		{
			name:        "success/eval-sha-succeeded",
			description: "验证 Add 在 Redis 脚本执行成功时返回 nil。",
		},
		{
			name:        "error/eval-sha-failed",
			description: "验证 Add 在 EvalSha 返回错误时透传该错误。",
			giveErr:     addErr,
			wantErrIs:   addErr,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			fake := &fakeRedis{
				evalShaResults: []*kitredis.Cmd{
					newRedisCmdResult(t.Context(), []any{int64(0), int64(1)}, tt.giveErr),
				},
			}
			store := &redisStore{
				redis:         fake,
				setScriptHash: "set-sha",
			}

			err := store.Add(t.Context(), "unit", []uint64{8, 13})

			if tt.wantErrIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErrIs)
				return
			}

			require.NoError(t, err)
			require.Len(t, fake.evalShaCalls, 1)
			assert.Equal(t, "set-sha", fake.evalShaCalls[0].sha)
			assert.Equal(t, []string{fmt.Sprintf(redisKeyFormat, "unit")}, fake.evalShaCalls[0].keys)
			assert.Equal(t, []any{uint64(8), uint64(13)}, fake.evalShaCalls[0].args)
		})
	}
}

// TestRedisStoreUnit_EvalScriptNoScript 验证 evalScript 对 NOSCRIPT 场景的脚本重载与重试行为。
//
// 该测试使用 fake Redis 覆盖重载成功、加载失败和重试失败语义，确保脚本缓存失效时处理稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRedisStoreUnit_EvalScriptNoScript(t *testing.T) {
	loadErr := errors.New("script load failed")
	retryErr := errors.New("retry failed")

	tests := []struct {
		name              string
		description       string
		giveLoadErr       error
		giveRetryErr      error
		wantResult        any
		wantErrIs         error
		wantEvalCallCount int
	}{
		{
			name:              "success/reload-and-retry-succeeded",
			description:       "验证 EvalSha 返回 NOSCRIPT 后重新加载脚本并重试成功。",
			wantResult:        []any{int64(1)},
			wantEvalCallCount: 2,
		},
		{
			name:              "error/script-load-failed",
			description:       "验证 EvalSha 返回 NOSCRIPT 后脚本加载失败时返回加载错误且不再重试。",
			giveLoadErr:       loadErr,
			wantErrIs:         loadErr,
			wantEvalCallCount: 1,
		},
		{
			name:              "error/retry-failed",
			description:       "验证 EvalSha 返回 NOSCRIPT 且脚本加载成功后重试失败时返回重试错误。",
			giveRetryErr:      retryErr,
			wantErrIs:         retryErr,
			wantEvalCallCount: 2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			fake := &fakeRedis{
				evalShaResults: []*kitredis.Cmd{
					newRedisCmdResult(t.Context(), nil, errors.New("NOSCRIPT No matching script")),
					newRedisCmdResult(t.Context(), tt.wantResult, tt.giveRetryErr),
				},
				scriptLoadResults: []*kitredis.StringCmd{
					newRedisStringCmdResult(t.Context(), "loaded-sha", tt.giveLoadErr),
				},
			}
			store := &redisStore{redis: fake}

			gotResult, err := store.evalScript(t.Context(), "sha", "return 1", []string{"key"}, []any{uint64(1)})

			if tt.wantErrIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErrIs)
				assert.False(t, gotResult.(bool))
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, gotResult)
			}

			assert.Len(t, fake.evalShaCalls, tt.wantEvalCallCount)
			require.Len(t, fake.scriptLoadCalls, 1)
			assert.Equal(t, "return 1", fake.scriptLoadCalls[0].script)
		})
	}
}

// TestRedisStoreUnit_NewRedisStore 验证 NewRedisStore 对 Redis 脚本预加载结果的处理。
//
// 该测试使用 fake Redis 覆盖 set/get 脚本加载成功和加载失败语义，确保初始化契约稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRedisStoreUnit_NewRedisStore(t *testing.T) {
	loadSetErr := errors.New("load set script failed")
	loadGetErr := errors.New("load get script failed")

	tests := []struct {
		name                string
		description         string
		giveScriptLoadErrs  []error
		wantErrIs           error
		wantScriptLoadCount int
	}{
		{
			name:                "success/load-all-scripts",
			description:         "验证 set 与 get 脚本均加载成功时返回包含脚本哈希的 RedisStore。",
			giveScriptLoadErrs:  []error{nil, nil},
			wantScriptLoadCount: 2,
		},
		{
			name:                "error/load-set-script-failed",
			description:         "验证 set 脚本加载失败时 NewRedisStore 立即返回错误。",
			giveScriptLoadErrs:  []error{loadSetErr},
			wantErrIs:           loadSetErr,
			wantScriptLoadCount: 1,
		},
		{
			name:                "error/load-get-script-failed",
			description:         "验证 get 脚本加载失败时 NewRedisStore 返回该错误。",
			giveScriptLoadErrs:  []error{nil, loadGetErr},
			wantErrIs:           loadGetErr,
			wantScriptLoadCount: 2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			scriptLoadResults := make([]*kitredis.StringCmd, 0, len(tt.giveScriptLoadErrs))
			for i, err := range tt.giveScriptLoadErrs {
				scriptLoadResults = append(scriptLoadResults, newRedisStringCmdResult(t.Context(), fmt.Sprintf("sha-%d", i), err))
			}
			fake := &fakeRedis{scriptLoadResults: scriptLoadResults}

			gotStore, err := NewRedisStore(fake)

			if tt.wantErrIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErrIs)
				assert.Nil(t, gotStore)
			} else {
				require.NoError(t, err)
				require.NotNil(t, gotStore)
				assert.Equal(t, "sha-0", gotStore.setScriptHash)
				assert.Equal(t, "sha-1", gotStore.getScriptHash)
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
