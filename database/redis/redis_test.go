// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package redis

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type respCommand struct {
	name string
	args []string
}

type respReply struct {
	kind  string
	value interface{}
}

type memoryRedisServer struct {
	mu      sync.Mutex
	kv      map[string]string
	scripts map[string]string
	records []respCommand
}

// TestNewRedis_OptionsAndCloseBehavior 验证 Redis 客户端创建、Option 覆盖和关闭后的错误行为。
//
// 该测试覆盖默认配置、自定义配置、多次 Option 覆盖，以及关闭客户端后继续执行命令返回 ErrClosed 的行为契约。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewRedis_OptionsAndCloseBehavior(t *testing.T) {
	tests := []struct {
		name         string
		description  string
		giveOptions  []Option
		wantAddr     string
		wantPassword string
	}{
		{
			name:         "success/default-options",
			description:  "验证 NewRedis 在未传入 Option 时使用包内默认地址和密码。",
			wantAddr:     addrDefault,
			wantPassword: passwordDefault,
		},
		{
			name:        "success/custom-address-and-password",
			description: "验证 WithAddr 与 WithPassword 会覆盖 NewRedis 的默认配置。",
			giveOptions: []Option{
				WithAddr("redis.internal:6380"),
				WithPassword("custom-password"),
			},
			wantAddr:     "redis.internal:6380",
			wantPassword: "custom-password",
		},
		{
			name:        "success/last-option-wins",
			description: "验证相同 Option 多次传入时最后一次配置生效。",
			giveOptions: []Option{
				WithAddr("redis-1.internal:6379"),
				WithAddr("redis-2.internal:6379"),
				WithPassword("first-password"),
				WithPassword("second-password"),
			},
			wantAddr:     "redis-2.internal:6379",
			wantPassword: "second-password",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			client := NewRedis(tt.giveOptions...)
			require.NotNil(t, client)
			t.Cleanup(func() {
				_ = client.Close()
			})

			redisClient, ok := client.(*redisClient)
			require.True(t, ok, "NewRedis 应返回 *redisClient 实现")
			assert.Equal(t, tt.wantAddr, redisClient.addr)
			assert.Equal(t, tt.wantPassword, redisClient.password)

			require.NoError(t, redisClient.Close())
			err := redisClient.Do(context.Background(), "PING").Err()
			assert.ErrorIs(t, err, ErrClosed)
		})
	}
}

// TestRedisClient_CommandDelegation 验证 redisClient 将基础命令委托给 go-redis 客户端。
//
// 该测试使用内存 RESP 服务覆盖 Do、Pipelined、TxPipelined、Subscribe、PSubscribe 与 Close，不依赖真实 Redis 服务。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRedisClient_CommandDelegation(t *testing.T) {
	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T, ctx context.Context, client *redisClient, server *memoryRedisServer)
	}{
		{
			name:        "success/do",
			description: "验证 Do 会将命令和参数发送给底层 go-redis 客户端并返回命令结果。",
			assert: func(t *testing.T, ctx context.Context, client *redisClient, server *memoryRedisServer) {
				got, err := client.Do(ctx, "SET", "client:do:key", "value").Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", got)
				assert.True(t, server.hasCommand("SET", "client:do:key", "value"))
			},
		},
		{
			name:        "success/pipelined",
			description: "验证 Pipelined 会将闭包中的多个命令作为管道命令执行并返回对应结果。",
			assert: func(t *testing.T, ctx context.Context, client *redisClient, server *memoryRedisServer) {
				cmds, err := client.Pipelined(ctx, func(pipe Pipeliner) error {
					pipe.Do(ctx, "SET", "client:pipeline:key", "pipeline-value")
					pipe.Do(ctx, "GET", "client:pipeline:key")
					return nil
				})
				require.NoError(t, err)
				require.Len(t, cmds, 2)
				got, err := cmds[1].(*Cmd).Result()
				require.NoError(t, err)
				assert.Equal(t, "pipeline-value", got)
				assert.True(t, server.hasCommand("SET", "client:pipeline:key", "pipeline-value"))
				assert.True(t, server.hasCommand("GET", "client:pipeline:key"))
			},
		},
		{
			name:        "success/tx-pipelined",
			description: "验证 TxPipelined 会使用 MULTI/EXEC 事务管道包装闭包中的命令并解析 EXEC 结果。",
			assert: func(t *testing.T, ctx context.Context, client *redisClient, server *memoryRedisServer) {
				cmds, err := client.TxPipelined(ctx, func(pipe Pipeliner) error {
					pipe.Do(ctx, "SET", "client:tx:key", "tx-value")
					pipe.Do(ctx, "GET", "client:tx:key")
					return nil
				})
				require.NoError(t, err)
				require.Len(t, cmds, 2)
				got, err := cmds[1].(*Cmd).Result()
				require.NoError(t, err)
				assert.Equal(t, "tx-value", got)
				assert.True(t, server.hasCommand("MULTI"))
				assert.True(t, server.hasCommand("EXEC"))
			},
		},
		{
			name:        "success/subscribe",
			description: "验证 Subscribe 创建发布订阅连接并接收服务端订阅确认消息。",
			assert: func(t *testing.T, ctx context.Context, client *redisClient, server *memoryRedisServer) {
				pubSub := client.Subscribe(ctx, "client:channel")
				require.NotNil(t, pubSub)
				t.Cleanup(func() {
					_ = pubSub.Close()
				})

				got, err := pubSub.ReceiveTimeout(ctx, time.Second)
				require.NoError(t, err)
				subscription, ok := got.(*Subscription)
				require.True(t, ok, "Subscribe 应收到订阅确认")
				assert.Equal(t, "subscribe", subscription.Kind)
				assert.Equal(t, "client:channel", subscription.Channel)
				assert.True(t, server.hasCommand("SUBSCRIBE", "client:channel"))
			},
		},
		{
			name:        "success/psubscribe",
			description: "验证 PSubscribe 创建模式订阅连接并接收服务端模式订阅确认消息。",
			assert: func(t *testing.T, ctx context.Context, client *redisClient, server *memoryRedisServer) {
				pubSub := client.PSubscribe(ctx, "client:*:channel")
				require.NotNil(t, pubSub)
				t.Cleanup(func() {
					_ = pubSub.Close()
				})

				got, err := pubSub.ReceiveTimeout(ctx, time.Second)
				require.NoError(t, err)
				subscription, ok := got.(*Subscription)
				require.True(t, ok, "PSubscribe 应收到模式订阅确认")
				assert.Equal(t, "psubscribe", subscription.Kind)
				assert.Equal(t, "client:*:channel", subscription.Channel)
				assert.True(t, server.hasCommand("PSUBSCRIBE", "client:*:channel"))
			},
		},
		{
			name:        "success/close",
			description: "验证 Close 会关闭底层 go-redis 客户端，并使后续命令返回 ErrClosed。",
			assert: func(t *testing.T, ctx context.Context, client *redisClient, _ *memoryRedisServer) {
				require.NoError(t, client.Close())
				err := client.Do(ctx, "PING").Err()
				assert.ErrorIs(t, err, ErrClosed)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			t.Cleanup(cancel)
			client, server := newMemoryRedisClient(t)

			tt.assert(t, ctx, client, server)
		})
	}
}

// TestRedisClient_ScriptDelegation 验证 redisClient 将脚本相关命令委托给 go-redis 客户端。
//
// 该测试通过内存 RESP 服务覆盖 Eval、EvalRO、EvalSha、EvalShaRO、ScriptExists、ScriptLoad、ScriptFlush 与 ScriptKill 的结果解析。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRedisClient_ScriptDelegation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	client, server := newMemoryRedisClient(t)
	script := "return ARGV[1]"

	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "success/eval",
			description: "验证 Eval 会发送 EVAL 命令并返回脚本参数结果。",
			assert: func(t *testing.T) {
				got, err := client.Eval(ctx, script, []string{}, "eval-value").Result()
				require.NoError(t, err)
				assert.Equal(t, "eval-value", got)
				assert.True(t, server.hasCommand("EVAL", script, "0", "eval-value"))
			},
		},
		{
			name:        "success/eval-ro",
			description: "验证 EvalRO 会发送 EVAL_RO 命令并返回只读脚本结果。",
			assert: func(t *testing.T) {
				got, err := client.EvalRO(ctx, script, []string{}, "eval-ro-value").Result()
				require.NoError(t, err)
				assert.Equal(t, "eval-ro-value", got)
				assert.True(t, server.hasCommand("EVAL_RO", script, "0", "eval-ro-value"))
			},
		},
		{
			name:        "success/script-load-and-eval-sha",
			description: "验证 ScriptLoad 返回脚本 SHA，EvalSha 使用该 SHA 执行脚本并返回参数结果。",
			assert: func(t *testing.T) {
				sha, err := client.ScriptLoad(ctx, script).Result()
				require.NoError(t, err)
				assert.Equal(t, sha1Hex(script), sha)

				got, err := client.EvalSha(ctx, sha, []string{}, "eval-sha-value").Result()
				require.NoError(t, err)
				assert.Equal(t, "eval-sha-value", got)
				assert.True(t, server.hasCommand("SCRIPT", "load", script))
				assert.True(t, server.hasCommand("EVALSHA", sha, "0", "eval-sha-value"))
			},
		},
		{
			name:        "success/eval-sha-ro",
			description: "验证 EvalShaRO 使用脚本 SHA 发送只读脚本命令并返回参数结果。",
			assert: func(t *testing.T) {
				sha, err := client.ScriptLoad(ctx, script).Result()
				require.NoError(t, err)

				got, err := client.EvalShaRO(ctx, sha, []string{}, "eval-sha-ro-value").Result()
				require.NoError(t, err)
				assert.Equal(t, "eval-sha-ro-value", got)
				assert.True(t, server.hasCommand("EVALSHA_RO", sha, "0", "eval-sha-ro-value"))
			},
		},
		{
			name:        "success/script-exists",
			description: "验证 ScriptExists 会发送 SCRIPT EXISTS 并按脚本缓存状态返回布尔切片。",
			assert: func(t *testing.T) {
				sha, err := client.ScriptLoad(ctx, script).Result()
				require.NoError(t, err)

				got, err := client.ScriptExists(ctx, sha, "missing-sha").Result()
				require.NoError(t, err)
				assert.Equal(t, []bool{true, false}, got)
				assert.True(t, server.hasCommand("SCRIPT", "exists", sha, "missing-sha"))
			},
		},
		{
			name:        "success/script-flush",
			description: "验证 ScriptFlush 会发送 SCRIPT FLUSH 并返回 OK 状态。",
			assert: func(t *testing.T) {
				got, err := client.ScriptFlush(ctx).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", got)
				assert.True(t, server.hasCommand("SCRIPT", "flush"))
			},
		},
		{
			name:        "success/script-kill",
			description: "验证 ScriptKill 会发送 SCRIPT KILL 并返回 OK 状态。",
			assert: func(t *testing.T) {
				got, err := client.ScriptKill(ctx).Result()
				require.NoError(t, err)
				assert.Equal(t, "OK", got)
				assert.True(t, server.hasCommand("SCRIPT", "kill"))
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			tt.assert(t)
		})
	}
}

// TestRedisExtension_CommandArguments 验证 RedisExtension 基础扩展命令的参数组合。
//
// 该测试通过 fake Redis 捕获 Do 参数，覆盖 Get、Set、Del、Expire 以及过期时间边界语义。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRedisExtension_CommandArguments(t *testing.T) {
	tests := []struct {
		name        string
		description string
		act         func(ext RedisExtension, ctx context.Context)
		wantArgs    []interface{}
	}{
		{
			name:        "success/get",
			description: "验证 Get 使用 GET 命令和目标 key 进行委托。",
			act: func(ext RedisExtension, ctx context.Context) {
				ext.Get(ctx, "extension:get:key")
			},
			wantArgs: []interface{}{"GET", "extension:get:key"},
		},
		{
			name:        "success/set-without-expiration",
			description: "验证 Set 在过期时间为零时不发送 EX/PX 等过期参数。",
			act: func(ext RedisExtension, ctx context.Context) {
				ext.Set(ctx, "extension:set:key", "value", 0)
			},
			wantArgs: []interface{}{"SET", "extension:set:key", "value"},
		},
		{
			name:        "boundary/set-with-negative-expiration",
			description: "验证 Set 在负过期时间下不发送无效 TTL 参数。",
			act: func(ext RedisExtension, ctx context.Context) {
				ext.Set(ctx, "extension:set:key", "value", -time.Second)
			},
			wantArgs: []interface{}{"SET", "extension:set:key", "value"},
		},
		{
			name:        "success/set-with-integer-seconds",
			description: "验证 Set 对整秒过期时间使用 Redis EX 和整数秒参数。",
			act: func(ext RedisExtension, ctx context.Context) {
				ext.Set(ctx, "extension:set:key", "value", 5*time.Second)
			},
			wantArgs: []interface{}{"SET", "extension:set:key", "value", "EX", int64(5)},
		},
		{
			name:        "boundary/set-with-subsecond-expiration",
			description: "验证 Set 对非整秒过期时间使用 PX 和整数毫秒参数，避免发送浮点秒。",
			act: func(ext RedisExtension, ctx context.Context) {
				ext.Set(ctx, "extension:set:key", "value", 1500*time.Millisecond)
			},
			wantArgs: []interface{}{"SET", "extension:set:key", "value", "PX", int64(1500)},
		},
		{
			name:        "boundary/set-with-submillisecond-expiration",
			description: "验证 Set 对小于一毫秒的正过期时间向上取整为 1 毫秒。",
			act: func(ext RedisExtension, ctx context.Context) {
				ext.Set(ctx, "extension:set:key", "value", time.Nanosecond)
			},
			wantArgs: []interface{}{"SET", "extension:set:key", "value", "PX", int64(1)},
		},
		{
			name:        "success/del",
			description: "验证 Del 使用 DEL 命令和目标 key 进行委托。",
			act: func(ext RedisExtension, ctx context.Context) {
				ext.Del(ctx, "extension:del:key")
			},
			wantArgs: []interface{}{"DEL", "extension:del:key"},
		},
		{
			name:        "success/expire-with-integer-seconds",
			description: "验证 Expire 对整秒过期时间发送 EXPIRE 和整数秒参数。",
			act: func(ext RedisExtension, ctx context.Context) {
				ext.Expire(ctx, "extension:expire:key", 7*time.Second)
			},
			wantArgs: []interface{}{"EXPIRE", "extension:expire:key", int64(7)},
		},
		{
			name:        "boundary/expire-with-subsecond-expiration",
			description: "验证 Expire 对非整秒过期时间发送 PEXPIRE 和整数毫秒参数。",
			act: func(ext RedisExtension, ctx context.Context) {
				ext.Expire(ctx, "extension:expire:key", 2500*time.Millisecond)
			},
			wantArgs: []interface{}{"PEXPIRE", "extension:expire:key", int64(2500)},
		},
		{
			name:        "boundary/expire-with-zero-expiration",
			description: "验证 Expire 在零过期时间下发送 EXPIRE 0，保持 Redis 立即过期语义。",
			act: func(ext RedisExtension, ctx context.Context) {
				ext.Expire(ctx, "extension:expire:key", 0)
			},
			wantArgs: []interface{}{"EXPIRE", "extension:expire:key", int64(0)},
		},
		{
			name:        "boundary/expire-with-negative-expiration",
			description: "验证 Expire 在负过期时间下发送 EXPIRE 0，避免错误地移除 key 的 TTL。",
			act: func(ext RedisExtension, ctx context.Context) {
				ext.Expire(ctx, "extension:expire:key", -time.Second)
			},
			wantArgs: []interface{}{"EXPIRE", "extension:expire:key", int64(0)},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			ctx := context.Background()
			fake := newFakeRedis()
			ext := NewRedisExtension(fake)

			tt.act(ext, ctx)

			require.NotEmpty(t, fake.doCalls)
			assert.Equal(t, tt.wantArgs, fake.doCalls[len(fake.doCalls)-1])
		})
	}
}

// TestRedisExtension_Delegation 验证 RedisExtension 对基础 Redis 接口和脚本接口的委托。
//
// 该测试通过 fake Redis 覆盖 Do、Pipelined、TxPipelined、Subscribe、PSubscribe、Eval 系列、Script 系列和 Close 的委托行为。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRedisExtension_Delegation(t *testing.T) {
	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T, ctx context.Context, ext RedisExtension, fake *fakeRedis)
	}{
		{
			name:        "success/do",
			description: "验证 Do 会原样委托到底层 Redis 实现。",
			assert: func(t *testing.T, ctx context.Context, ext RedisExtension, fake *fakeRedis) {
				cmd := ext.Do(ctx, "PING")
				require.NotNil(t, cmd)
				assert.Equal(t, []interface{}{"PING"}, fake.doCalls[0])
			},
		},
		{
			name:        "success/pipelined",
			description: "验证 Pipelined 会把闭包委托到底层 Redis 实现并返回底层结果。",
			assert: func(t *testing.T, ctx context.Context, ext RedisExtension, fake *fakeRedis) {
				cmds, err := ext.Pipelined(ctx, func(Pipeliner) error { return nil })
				require.NoError(t, err)
				assert.Len(t, cmds, 1)
				assert.Equal(t, 1, fake.pipelinedCalls)
			},
		},
		{
			name:        "success/tx-pipelined",
			description: "验证 TxPipelined 会把事务管道闭包委托到底层 Redis 实现。",
			assert: func(t *testing.T, ctx context.Context, ext RedisExtension, fake *fakeRedis) {
				cmds, err := ext.TxPipelined(ctx, func(Pipeliner) error { return nil })
				require.NoError(t, err)
				assert.Len(t, cmds, 1)
				assert.Equal(t, 1, fake.txPipelinedCalls)
			},
		},
		{
			name:        "success/subscribe",
			description: "验证 Subscribe 会将频道列表原样委托到底层 Redis 实现。",
			assert: func(t *testing.T, ctx context.Context, ext RedisExtension, fake *fakeRedis) {
				assert.Same(t, fake.subscribeResult, ext.Subscribe(ctx, "channel-a", "channel-b"))
				assert.Equal(t, []string{"channel-a", "channel-b"}, fake.subscribeCalls[0])
			},
		},
		{
			name:        "success/psubscribe",
			description: "验证 PSubscribe 会将频道模式列表原样委托到底层 Redis 实现。",
			assert: func(t *testing.T, ctx context.Context, ext RedisExtension, fake *fakeRedis) {
				assert.Same(t, fake.psubscribeResult, ext.PSubscribe(ctx, "channel-*"))
				assert.Equal(t, []string{"channel-*"}, fake.psubscribeCalls[0])
			},
		},
		{
			name:        "success/eval",
			description: "验证 Eval 会将脚本、keys 和 args 原样委托到底层 Redis 实现。",
			assert: func(t *testing.T, ctx context.Context, ext RedisExtension, fake *fakeRedis) {
				ext.Eval(ctx, "return 1", []string{"k1"}, "a1")
				assert.Equal(t, scriptCall{method: "Eval", scriptOrSHA: "return 1", keys: []string{"k1"}, args: []interface{}{"a1"}}, fake.scriptCalls[0])
			},
		},
		{
			name:        "success/eval-ro",
			description: "验证 EvalRO 会将只读脚本参数原样委托到底层 Redis 实现。",
			assert: func(t *testing.T, ctx context.Context, ext RedisExtension, fake *fakeRedis) {
				ext.EvalRO(ctx, "return 2", []string{"k2"}, "a2")
				assert.Equal(t, scriptCall{method: "EvalRO", scriptOrSHA: "return 2", keys: []string{"k2"}, args: []interface{}{"a2"}}, fake.scriptCalls[0])
			},
		},
		{
			name:        "success/eval-sha",
			description: "验证 EvalSha 会将 SHA、keys 和 args 原样委托到底层 Redis 实现。",
			assert: func(t *testing.T, ctx context.Context, ext RedisExtension, fake *fakeRedis) {
				ext.EvalSha(ctx, "sha1", []string{"k3"}, "a3")
				assert.Equal(t, scriptCall{method: "EvalSha", scriptOrSHA: "sha1", keys: []string{"k3"}, args: []interface{}{"a3"}}, fake.scriptCalls[0])
			},
		},
		{
			name:        "success/eval-sha-ro",
			description: "验证 EvalShaRO 会将只读 SHA 脚本参数原样委托到底层 Redis 实现。",
			assert: func(t *testing.T, ctx context.Context, ext RedisExtension, fake *fakeRedis) {
				ext.EvalShaRO(ctx, "sha2", []string{"k4"}, "a4")
				assert.Equal(t, scriptCall{method: "EvalShaRO", scriptOrSHA: "sha2", keys: []string{"k4"}, args: []interface{}{"a4"}}, fake.scriptCalls[0])
			},
		},
		{
			name:        "success/script-exists",
			description: "验证 ScriptExists 会将 hash 列表原样委托到底层 Redis 实现。",
			assert: func(t *testing.T, ctx context.Context, ext RedisExtension, fake *fakeRedis) {
				ext.ScriptExists(ctx, "sha-a", "sha-b")
				assert.Equal(t, []string{"sha-a", "sha-b"}, fake.scriptExistsCalls[0])
			},
		},
		{
			name:        "success/script-load",
			description: "验证 ScriptLoad 会将脚本文本原样委托到底层 Redis 实现。",
			assert: func(t *testing.T, ctx context.Context, ext RedisExtension, fake *fakeRedis) {
				ext.ScriptLoad(ctx, "return 'load'")
				assert.Equal(t, []string{"return 'load'"}, fake.scriptLoadCalls)
			},
		},
		{
			name:        "success/script-flush-capability",
			description: "验证 ScriptFlush 通过能力接口委托给非 redisClient 的底层实现。",
			assert: func(t *testing.T, ctx context.Context, ext RedisExtension, fake *fakeRedis) {
				got := ext.ScriptFlush(ctx)
				require.NotNil(t, got)
				assert.Equal(t, 1, fake.scriptFlushCalls)
				assert.Equal(t, "OK", got.Val())
			},
		},
		{
			name:        "success/script-kill-capability",
			description: "验证 ScriptKill 通过能力接口委托给非 redisClient 的底层实现。",
			assert: func(t *testing.T, ctx context.Context, ext RedisExtension, fake *fakeRedis) {
				got := ext.ScriptKill(ctx)
				require.NotNil(t, got)
				assert.Equal(t, 1, fake.scriptKillCalls)
				assert.Equal(t, "OK", got.Val())
			},
		},
		{
			name:        "success/close",
			description: "验证 Close 会委托到底层 Redis 实现。",
			assert: func(t *testing.T, _ context.Context, ext RedisExtension, fake *fakeRedis) {
				require.NoError(t, ext.Close())
				assert.True(t, fake.closed)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			ctx := context.Background()
			fake := newFakeRedis()
			ext := NewRedisExtension(fake)

			tt.assert(t, ctx, ext, fake)
		})
	}
}

// TestRedisExtension_UnsupportedScriptFlushAndKill 验证扩展层对不支持脚本管理能力的兼容行为。
//
// 该测试确保底层 Redis 实现未提供 ScriptFlush 或 ScriptKill 能力接口时，扩展层保持历史兼容并返回 nil。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRedisExtension_UnsupportedScriptFlushAndKill(t *testing.T) {
	tests := []struct {
		name        string
		description string
		act         func(ext RedisExtension, ctx context.Context) *StatusCmd
	}{
		{
			name:        "success/script-flush-unsupported",
			description: "验证底层 Redis 不支持 ScriptFlush 能力接口时扩展层返回 nil。",
			act: func(ext RedisExtension, ctx context.Context) *StatusCmd {
				return ext.ScriptFlush(ctx)
			},
		},
		{
			name:        "success/script-kill-unsupported",
			description: "验证底层 Redis 不支持 ScriptKill 能力接口时扩展层返回 nil。",
			act: func(ext RedisExtension, ctx context.Context) *StatusCmd {
				return ext.ScriptKill(ctx)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			ext := NewRedisExtension(newBasicFakeRedis())
			assert.Nil(t, tt.act(ext, context.Background()))
		})
	}
}

// newMemoryRedisClient 构造使用内存 RESP 服务的 redisClient。
//
// 该辅助函数通过 go-redis 的 Dialer 注入 net.Pipe 连接，使客户端方法在单元测试中无需访问真实 Redis 服务。
//
// 参数：
//   - t: 测试上下文，用于注册清理逻辑并报告构造失败。
//
// 返回值：
//   - *redisClient: 可直接执行 Redis 命令的被测客户端。
//   - *memoryRedisServer: 捕获命令并返回稳定 RESP 响应的内存服务。
func newMemoryRedisClient(t *testing.T) (*redisClient, *memoryRedisServer) {
	t.Helper()

	server := &memoryRedisServer{
		kv:      make(map[string]string),
		scripts: make(map[string]string),
	}
	client := goredis.NewClient(&goredis.Options{
		Addr:     "memory.redis:6379",
		Protocol: 2,
		Dialer: func(_ context.Context, _, _ string) (net.Conn, error) {
			clientConn, serverConn := net.Pipe()
			go server.serve(serverConn)
			return clientConn, nil
		},
	})
	t.Cleanup(func() {
		_ = client.Close()
	})

	return &redisClient{client: client, addr: "memory.redis:6379"}, server
}

// serve 处理单条内存连接上的 RESP 命令。
//
// 该辅助方法循环读取 go-redis 发送的 RESP 数组命令，并写回 Redis 协议兼容的稳定响应。
//
// 参数：
//   - conn: net.Pipe 创建的服务端连接。
func (s *memoryRedisServer) serve(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	reader := bufio.NewReader(conn)
	var txActive bool
	var txReplies []respReply
	for {
		args, err := readRESPArray(reader)
		if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			return
		}
		if len(args) == 0 {
			writeRESP(conn, respReply{kind: "error", value: "ERR empty command"})
			continue
		}

		command := strings.ToUpper(args[0])
		s.record(args)

		switch command {
		case "MULTI":
			txActive = true
			txReplies = nil
			writeRESP(conn, respReply{kind: "simple", value: "OK"})
		case "EXEC":
			txActive = false
			writeRESP(conn, respReply{kind: "array", value: txReplies})
		case "DISCARD":
			txActive = false
			txReplies = nil
			writeRESP(conn, respReply{kind: "simple", value: "OK"})
		case "SUBSCRIBE":
			for i, channel := range args[1:] {
				writeRESP(conn, respReply{kind: "array", value: []respReply{
					{kind: "bulk", value: "subscribe"},
					{kind: "bulk", value: channel},
					{kind: "int", value: int64(i + 1)},
				}})
			}
		case "PSUBSCRIBE":
			for i, channel := range args[1:] {
				writeRESP(conn, respReply{kind: "array", value: []respReply{
					{kind: "bulk", value: "psubscribe"},
					{kind: "bulk", value: channel},
					{kind: "int", value: int64(i + 1)},
				}})
			}
		case "UNSUBSCRIBE", "PUNSUBSCRIBE":
			kind := strings.ToLower(command)
			for _, channel := range args[1:] {
				writeRESP(conn, respReply{kind: "array", value: []respReply{
					{kind: "bulk", value: kind},
					{kind: "bulk", value: channel},
					{kind: "int", value: int64(0)},
				}})
			}
		default:
			reply := s.handleCommand(args)
			if txActive {
				txReplies = append(txReplies, reply)
				writeRESP(conn, respReply{kind: "simple", value: "QUEUED"})
				continue
			}
			writeRESP(conn, reply)
		}
	}
}

// handleCommand 生成单条 Redis 命令的稳定响应。
//
// 该辅助方法实现测试所需的 GET、SET、脚本和生命周期命令，覆盖 go-redis 委托行为而不模拟完整 Redis。
//
// 参数：
//   - args: RESP 命令参数列表，首项为命令名。
//
// 返回值：
//   - respReply: 可序列化为 RESP 的命令响应。
func (s *memoryRedisServer) handleCommand(args []string) respReply {
	command := strings.ToUpper(args[0])

	s.mu.Lock()
	defer s.mu.Unlock()

	switch command {
	case "HELLO":
		return respReply{kind: "array", value: []respReply{
			{kind: "bulk", value: "server"},
			{kind: "bulk", value: "memory-redis"},
			{kind: "bulk", value: "proto"},
			{kind: "int", value: int64(2)},
		}}
	case "CLIENT", "AUTH", "SELECT", "PING":
		return respReply{kind: "simple", value: "OK"}
	case "SET":
		if len(args) >= 3 {
			s.kv[args[1]] = args[2]
		}
		return respReply{kind: "simple", value: "OK"}
	case "GET":
		if len(args) < 2 {
			return respReply{kind: "error", value: "ERR wrong number of arguments"}
		}
		value, ok := s.kv[args[1]]
		if !ok {
			return respReply{kind: "nil"}
		}
		return respReply{kind: "bulk", value: value}
	case "DEL":
		var deleted int64
		for _, key := range args[1:] {
			if _, ok := s.kv[key]; ok {
				delete(s.kv, key)
				deleted++
			}
		}
		return respReply{kind: "int", value: deleted}
	case "EXPIRE", "PEXPIRE", "PERSIST":
		return respReply{kind: "int", value: int64(1)}
	case "EVAL", "EVAL_RO", "EVALSHA", "EVALSHA_RO":
		return evalReply(args)
	case "SCRIPT":
		return s.handleScript(args)
	default:
		return respReply{kind: "error", value: fmt.Sprintf("ERR unknown command %s", command)}
	}
}

// handleScript 生成 SCRIPT 子命令的稳定响应。
//
// 该辅助方法覆盖测试涉及的 LOAD、EXISTS、FLUSH 和 KILL 子命令。
//
// 参数：
//   - args: SCRIPT 命令参数列表。
//
// 返回值：
//   - respReply: 可序列化为 RESP 的脚本命令响应。
func (s *memoryRedisServer) handleScript(args []string) respReply {
	if len(args) < 2 {
		return respReply{kind: "error", value: "ERR missing SCRIPT subcommand"}
	}

	switch strings.ToUpper(args[1]) {
	case "LOAD":
		if len(args) < 3 {
			return respReply{kind: "error", value: "ERR missing script"}
		}
		sha := sha1Hex(args[2])
		s.scripts[sha] = args[2]
		return respReply{kind: "bulk", value: sha}
	case "EXISTS":
		replies := make([]respReply, 0, len(args)-2)
		for _, hash := range args[2:] {
			var exists int64
			if _, ok := s.scripts[hash]; ok {
				exists = 1
			}
			replies = append(replies, respReply{kind: "int", value: exists})
		}
		return respReply{kind: "array", value: replies}
	case "FLUSH":
		s.scripts = make(map[string]string)
		return respReply{kind: "simple", value: "OK"}
	case "KILL":
		return respReply{kind: "simple", value: "OK"}
	default:
		return respReply{kind: "error", value: "ERR unsupported SCRIPT subcommand"}
	}
}

// record 保存收到的命令参数。
//
// 该辅助方法用于测试断言客户端是否将命令委托到内存 RESP 服务。
//
// 参数：
//   - args: RESP 命令参数列表。
func (s *memoryRedisServer) record(args []string) {
	record := respCommand{name: strings.ToUpper(args[0]), args: append([]string(nil), args[1:]...)}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, record)
}

// hasCommand 判断内存 RESP 服务是否收到指定命令。
//
// 该辅助方法按命令名和参数进行精确匹配，用于验证委托命令的稳定性。
//
// 参数：
//   - name: 期望的 Redis 命令名。
//   - args: 期望的 Redis 命令参数。
//
// 返回值：
//   - bool: 收到完全匹配的命令时返回 true。
func (s *memoryRedisServer) hasCommand(name string, args ...string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	name = strings.ToUpper(name)
	for _, record := range s.records {
		if record.name == name && equalStrings(record.args, args) {
			return true
		}
	}
	return false
}

// readRESPArray 读取一条 RESP 数组命令。
//
// 该辅助函数仅实现 go-redis 测试命令所需的 RESP 数组、Bulk String、Simple String 和 Integer 解析。
//
// 参数：
//   - reader: RESP 输入流。
//
// 返回值：
//   - []string: 解析后的命令参数。
//   - error: 读取或协议解析失败时返回错误。
func readRESPArray(reader *bufio.Reader) ([]string, error) {
	line, err := readRESPLine(reader)
	if err != nil {
		return nil, err
	}
	if len(line) == 0 || line[0] != '*' {
		return nil, fmt.Errorf("expected RESP array, got %q", line)
	}

	count, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, err
	}
	values := make([]string, 0, count)
	for i := 0; i < count; i++ {
		value, err := readRESPValue(reader)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

// readRESPValue 读取 RESP 数组内的一个标量值。
//
// 该辅助函数支持 Bulk String、Simple String 与 Integer，满足 go-redis 命令编码的测试需求。
//
// 参数：
//   - reader: RESP 输入流。
//
// 返回值：
//   - string: 解析后的标量值。
//   - error: 读取或协议解析失败时返回错误。
func readRESPValue(reader *bufio.Reader) (string, error) {
	line, err := readRESPLine(reader)
	if err != nil {
		return "", err
	}
	if len(line) == 0 {
		return "", fmt.Errorf("empty RESP value header")
	}

	switch line[0] {
	case '$':
		length, err := strconv.Atoi(line[1:])
		if err != nil {
			return "", err
		}
		if length < 0 {
			return "", nil
		}
		buf := make([]byte, length+2)
		if _, err := io.ReadFull(reader, buf); err != nil {
			return "", err
		}
		return string(buf[:length]), nil
	case '+', ':':
		return line[1:], nil
	default:
		return "", fmt.Errorf("unsupported RESP value %q", line)
	}
}

// readRESPLine 读取去除 CRLF 的 RESP 行。
//
// 该辅助函数集中处理 RESP 行结尾，避免协议解析逻辑重复。
//
// 参数：
//   - reader: RESP 输入流。
//
// 返回值：
//   - string: 去除 CRLF 后的行内容。
//   - error: 读取失败时返回错误。
func readRESPLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r"), nil
}

// writeRESP 将响应写入 RESP 输出流。
//
// 该辅助函数支持测试所需的 Simple String、Bulk String、Integer、Array、Nil 和 Error 响应类型。
//
// 参数：
//   - writer: RESP 输出流。
//   - reply: 待写入的响应对象。
func writeRESP(writer io.Writer, reply respReply) {
	switch reply.kind {
	case "simple":
		_, _ = fmt.Fprintf(writer, "+%s\r\n", reply.value)
	case "bulk":
		value := fmt.Sprint(reply.value)
		_, _ = fmt.Fprintf(writer, "$%d\r\n%s\r\n", len(value), value)
	case "int":
		_, _ = fmt.Fprintf(writer, ":%d\r\n", reply.value)
	case "array":
		replies := reply.value.([]respReply)
		_, _ = fmt.Fprintf(writer, "*%d\r\n", len(replies))
		for _, item := range replies {
			writeRESP(writer, item)
		}
	case "nil":
		_, _ = io.WriteString(writer, "$-1\r\n")
	case "error":
		_, _ = fmt.Fprintf(writer, "-%s\r\n", reply.value)
	}
}

// evalReply 生成 Eval 系列命令的响应。
//
// 该辅助函数返回第一个 ARGV 参数，模拟测试脚本 return ARGV[1] 的稳定行为。
//
// 参数：
//   - args: Eval 系列命令参数。
//
// 返回值：
//   - respReply: Eval 系列命令的响应。
func evalReply(args []string) respReply {
	if len(args) < 3 {
		return respReply{kind: "error", value: "ERR wrong number of arguments"}
	}

	keyCount, err := strconv.Atoi(args[2])
	if err != nil {
		return respReply{kind: "error", value: "ERR invalid key count"}
	}
	argIndex := 3 + keyCount
	if len(args) <= argIndex {
		return respReply{kind: "bulk", value: ""}
	}
	return respReply{kind: "bulk", value: args[argIndex]}
}

// sha1Hex 计算脚本文本的 SHA1 十六进制摘要。
//
// 该辅助函数用于模拟 Redis SCRIPT LOAD 返回的脚本摘要。
//
// 参数：
//   - script: Lua 脚本文本。
//
// 返回值：
//   - string: SHA1 十六进制摘要。
func sha1Hex(script string) string {
	sum := sha1.Sum([]byte(script))
	return hex.EncodeToString(sum[:])
}

// equalStrings 判断两个字符串切片是否完全一致。
//
// 该辅助函数用于命令记录断言，避免引入与业务无关的切片比较细节。
//
// 参数：
//   - left: 左侧字符串切片。
//   - right: 右侧字符串切片。
//
// 返回值：
//   - bool: 两个切片长度和元素均一致时返回 true。
func equalStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

type scriptCall struct {
	method      string
	scriptOrSHA string
	keys        []string
	args        []interface{}
}

type basicFakeRedis struct {
	doCalls           [][]interface{}
	pipelinedCalls    int
	txPipelinedCalls  int
	subscribeCalls    [][]string
	psubscribeCalls   [][]string
	scriptCalls       []scriptCall
	scriptExistsCalls [][]string
	scriptLoadCalls   []string
	closed            bool
	subscribeResult   *PubSub
	psubscribeResult  *PubSub
}

type fakeRedis struct {
	basicFakeRedis
	scriptFlushCalls int
	scriptKillCalls  int
}

// newBasicFakeRedis 构造不支持 ScriptFlush 和 ScriptKill 能力接口的 fake Redis。
//
// 该辅助函数用于验证 RedisExtension 对底层能力缺失的兼容行为。
//
// 返回值：
//   - *basicFakeRedis: 可记录基础 Redis 委托调用的 fake 实现。
func newBasicFakeRedis() *basicFakeRedis {
	return &basicFakeRedis{
		subscribeResult:  &PubSub{},
		psubscribeResult: &PubSub{},
	}
}

// newFakeRedis 构造支持脚本管理能力接口的 fake Redis。
//
// 该辅助函数用于验证 RedisExtension 对 ScriptFlush 和 ScriptKill 的能力接口委托。
//
// 返回值：
//   - *fakeRedis: 可记录完整 Redis 委托调用的 fake 实现。
func newFakeRedis() *fakeRedis {
	return &fakeRedis{basicFakeRedis: *newBasicFakeRedis()}
}

// Do 记录任意 Redis 命令调用。
//
// 参数：
//   - ctx: 上下文对象，用于保持签名与 Redis 接口一致。
//   - args: Redis 命令参数列表。
//
// 返回值：
//   - *Cmd: 已设置 OK 结果的命令对象。
func (f *basicFakeRedis) Do(ctx context.Context, args ...interface{}) *Cmd {
	f.doCalls = append(f.doCalls, append([]interface{}(nil), args...))
	cmd := goredis.NewCmd(ctx, args...)
	cmd.SetVal("OK")
	return cmd
}

// Pipelined 记录管道委托调用。
//
// 参数：
//   - ctx: 上下文对象，用于保持签名与 Redis 接口一致。
//   - fn: 调用方提供的管道闭包。
//
// 返回值：
//   - []Cmder: fake 管道命令结果。
//   - error: 闭包返回的错误。
func (f *basicFakeRedis) Pipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error) {
	f.pipelinedCalls++
	if err := fn(nil); err != nil {
		return nil, err
	}
	return []Cmder{goredis.NewCmd(ctx, "PIPELINED")}, nil
}

// TxPipelined 记录事务管道委托调用。
//
// 参数：
//   - ctx: 上下文对象，用于保持签名与 Redis 接口一致。
//   - fn: 调用方提供的事务管道闭包。
//
// 返回值：
//   - []Cmder: fake 事务管道命令结果。
//   - error: 闭包返回的错误。
func (f *basicFakeRedis) TxPipelined(ctx context.Context, fn func(Pipeliner) error) ([]Cmder, error) {
	f.txPipelinedCalls++
	if err := fn(nil); err != nil {
		return nil, err
	}
	return []Cmder{goredis.NewCmd(ctx, "TXPIPELINED")}, nil
}

// Subscribe 记录频道订阅委托调用。
//
// 参数：
//   - ctx: 上下文对象，用于保持签名与 Redis 接口一致。
//   - channels: 待订阅频道列表。
//
// 返回值：
//   - *PubSub: fake 订阅对象。
func (f *basicFakeRedis) Subscribe(_ context.Context, channels ...string) *PubSub {
	f.subscribeCalls = append(f.subscribeCalls, append([]string(nil), channels...))
	return f.subscribeResult
}

// PSubscribe 记录模式订阅委托调用。
//
// 参数：
//   - ctx: 上下文对象，用于保持签名与 Redis 接口一致。
//   - channels: 待订阅频道模式列表。
//
// 返回值：
//   - *PubSub: fake 模式订阅对象。
func (f *basicFakeRedis) PSubscribe(_ context.Context, channels ...string) *PubSub {
	f.psubscribeCalls = append(f.psubscribeCalls, append([]string(nil), channels...))
	return f.psubscribeResult
}

// Eval 记录 Lua 脚本执行委托调用。
//
// 参数：
//   - ctx: 上下文对象，用于保持签名与 Redis 接口一致。
//   - script: Lua 脚本文本。
//   - keys: 脚本 key 列表。
//   - args: 脚本参数列表。
//
// 返回值：
//   - *Cmd: 已设置 OK 结果的命令对象。
func (f *basicFakeRedis) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *Cmd {
	f.scriptCalls = append(f.scriptCalls, scriptCall{method: "Eval", scriptOrSHA: script, keys: append([]string(nil), keys...), args: append([]interface{}(nil), args...)})
	cmd := goredis.NewCmd(ctx, "EVAL")
	cmd.SetVal("OK")
	return cmd
}

// EvalRO 记录只读 Lua 脚本执行委托调用。
//
// 参数：
//   - ctx: 上下文对象，用于保持签名与 Redis 接口一致。
//   - script: Lua 脚本文本。
//   - keys: 脚本 key 列表。
//   - args: 脚本参数列表。
//
// 返回值：
//   - *Cmd: 已设置 OK 结果的命令对象。
func (f *basicFakeRedis) EvalRO(ctx context.Context, script string, keys []string, args ...interface{}) *Cmd {
	f.scriptCalls = append(f.scriptCalls, scriptCall{method: "EvalRO", scriptOrSHA: script, keys: append([]string(nil), keys...), args: append([]interface{}(nil), args...)})
	cmd := goredis.NewCmd(ctx, "EVAL_RO")
	cmd.SetVal("OK")
	return cmd
}

// EvalSha 记录 SHA 脚本执行委托调用。
//
// 参数：
//   - ctx: 上下文对象，用于保持签名与 Redis 接口一致。
//   - sha1: Lua 脚本 SHA1 摘要。
//   - keys: 脚本 key 列表。
//   - args: 脚本参数列表。
//
// 返回值：
//   - *Cmd: 已设置 OK 结果的命令对象。
func (f *basicFakeRedis) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *Cmd {
	f.scriptCalls = append(f.scriptCalls, scriptCall{method: "EvalSha", scriptOrSHA: sha1, keys: append([]string(nil), keys...), args: append([]interface{}(nil), args...)})
	cmd := goredis.NewCmd(ctx, "EVALSHA")
	cmd.SetVal("OK")
	return cmd
}

// EvalShaRO 记录只读 SHA 脚本执行委托调用。
//
// 参数：
//   - ctx: 上下文对象，用于保持签名与 Redis 接口一致。
//   - sha1: Lua 脚本 SHA1 摘要。
//   - keys: 脚本 key 列表。
//   - args: 脚本参数列表。
//
// 返回值：
//   - *Cmd: 已设置 OK 结果的命令对象。
func (f *basicFakeRedis) EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...interface{}) *Cmd {
	f.scriptCalls = append(f.scriptCalls, scriptCall{method: "EvalShaRO", scriptOrSHA: sha1, keys: append([]string(nil), keys...), args: append([]interface{}(nil), args...)})
	cmd := goredis.NewCmd(ctx, "EVALSHA_RO")
	cmd.SetVal("OK")
	return cmd
}

// ScriptExists 记录脚本存在性检查委托调用。
//
// 参数：
//   - ctx: 上下文对象，用于保持签名与 Redis 接口一致。
//   - hashes: 脚本 SHA1 摘要列表。
//
// 返回值：
//   - *BoolSliceCmd: 已设置 true 结果的命令对象。
func (f *basicFakeRedis) ScriptExists(ctx context.Context, hashes ...string) *BoolSliceCmd {
	f.scriptExistsCalls = append(f.scriptExistsCalls, append([]string(nil), hashes...))
	cmd := goredis.NewBoolSliceCmd(ctx, "SCRIPT", "EXISTS")
	cmd.SetVal([]bool{true})
	return cmd
}

// ScriptLoad 记录脚本加载委托调用。
//
// 参数：
//   - ctx: 上下文对象，用于保持签名与 Redis 接口一致。
//   - script: Lua 脚本文本。
//
// 返回值：
//   - *StringCmd: 已设置 SHA 结果的命令对象。
func (f *basicFakeRedis) ScriptLoad(ctx context.Context, script string) *StringCmd {
	f.scriptLoadCalls = append(f.scriptLoadCalls, script)
	cmd := goredis.NewStringCmd(ctx, "SCRIPT", "LOAD")
	cmd.SetVal(sha1Hex(script))
	return cmd
}

// Close 记录客户端关闭委托调用。
//
// 返回值：
//   - error: fake 实现始终返回 nil。
func (f *basicFakeRedis) Close() error {
	f.closed = true
	return nil
}

// ScriptFlush 记录脚本缓存清理能力调用。
//
// 参数：
//   - ctx: 上下文对象，用于保持签名与能力接口一致。
//
// 返回值：
//   - *StatusCmd: 已设置 OK 结果的状态命令对象。
func (f *fakeRedis) ScriptFlush(ctx context.Context) *StatusCmd {
	f.scriptFlushCalls++
	cmd := goredis.NewStatusCmd(ctx, "SCRIPT", "FLUSH")
	cmd.SetVal("OK")
	return cmd
}

// ScriptKill 记录脚本终止能力调用。
//
// 参数：
//   - ctx: 上下文对象，用于保持签名与能力接口一致。
//
// 返回值：
//   - *StatusCmd: 已设置 OK 结果的状态命令对象。
func (f *fakeRedis) ScriptKill(ctx context.Context) *StatusCmd {
	f.scriptKillCalls++
	cmd := goredis.NewStatusCmd(ctx, "SCRIPT", "KILL")
	cmd.SetVal("OK")
	return cmd
}
