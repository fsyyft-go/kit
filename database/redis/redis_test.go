// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//
// redis_test.go 设计说明：
//
// 本文件用于测试 database/redis 包的核心功能，覆盖 Redis 基础操作与扩展接口。
// 测试用例采用表格驱动，断言使用 stretchr/testify，便于批量验证多种场景。
// 所有测试均假定本地 Redis 服务可用，且使用默认配置。
//
// 使用方法：
//   go test -v -cover ./database/redis
//
// 主要覆盖内容：
//   - RedisExtension 基本命令（Get/Set/Del/Expire）
//   - Redis 接口的 Do/Pipelined/TxPipelined/Subscribe/PSubscribe
//   - Lua 脚本相关命令
//   - 错误处理与边界场景
//
// 注意：如 Redis 未启动，部分测试会自动跳过。

package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// testRedisConnected 测试 Redis 连接是否可用。
//
// 参数：
//   - redis：Redis 扩展实例
//
// 返回值：
//   - bool：如果连接成功返回 true，否则返回 false
func testRedisConnected(redis RedisExtension) bool {
	// 设置一个较短的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 尝试执行一个简单的 PING 命令
	_, err := redis.Do(ctx, "PING").Result()
	return err == nil
}

func TestRedisExtension_Basic(t *testing.T) {
	redis := NewRedis()
	redisExtension := NewRedisExtension(redis)
	if !testRedisConnected(redisExtension) {
		t.Skip("Redis 连接失败，跳过测试")
	}
	ctx := context.Background()
	t.Run("Set/Get/Del/Expire", func(t *testing.T) {
		testKey := "test:basic:key"
		// 表格驱动测试用例
		tests := []struct {
			name       string
			prepare    func()
			action     func() (interface{}, error)
			assertFunc func(t *testing.T, got interface{}, err error)
		}{
			{
				name:    "Set",
				prepare: func() {},
				action: func() (interface{}, error) {
					return redisExtension.Set(ctx, testKey, "v1", time.Second*5).Result()
				},
				assertFunc: func(t *testing.T, got interface{}, err error) {
					assert.NoError(t, err, "Set 应无错误")
					assert.Equal(t, "OK", got, "Set 返回应为 OK")
				},
			},
			{
				name: "Get",
				prepare: func() {
					redisExtension.Set(ctx, testKey, "v2", time.Second*5)
				},
				action: func() (interface{}, error) {
					return redisExtension.Get(ctx, testKey).Result()
				},
				assertFunc: func(t *testing.T, got interface{}, err error) {
					assert.NoError(t, err, "Get 应无错误")
					assert.Equal(t, "v2", got, "Get 返回应为 v2")
				},
			},
			{
				name: "Expire",
				prepare: func() {
					redisExtension.Set(ctx, testKey, "v3", time.Second*5)
				},
				action: func() (interface{}, error) {
					return redisExtension.Expire(ctx, testKey, 1*time.Second).Result()
				},
				assertFunc: func(t *testing.T, got interface{}, err error) {
					assert.NoError(t, err, "Expire 应无错误")
					assert.Equal(t, int64(1), got, "Expire 返回应为 1")
				},
			},
			{
				name: "Del",
				prepare: func() {
					redisExtension.Set(ctx, testKey, "v4", time.Second*5)
				},
				action: func() (interface{}, error) {
					return redisExtension.Del(ctx, testKey).Result()
				},
				assertFunc: func(t *testing.T, got interface{}, err error) {
					assert.NoError(t, err, "Del 应无错误")
					assert.Equal(t, int64(1), got, "Del 返回应为 1")
				},
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				if tc.prepare != nil {
					tc.prepare()
				}
				got, err := tc.action()
				tc.assertFunc(t, got, err)
			})
		}
	})
}

func TestRedis(t *testing.T) {
	redis := NewRedis()
	redisExtension := NewRedisExtension(redis)

	if !testRedisConnected(redisExtension) {
		t.Skip("Redis 连接失败，跳过测试")
	}

	redisExtension.Set(context.Background(), "key", "value", time.Second*10)
	val, err := redisExtension.Get(context.Background(), "key").Result()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(val)
}

func TestRedis_EvalSha(t *testing.T) {
	// 创建 Redis 实例
	redis := NewRedis()

	redisExtension := NewRedisExtension(redis)

	if !testRedisConnected(redisExtension) {
		t.Skip("Redis 连接失败，跳过测试")
	}

	// 生成唯一的测试键名
	testKey := "test:counter:" + time.Now().Format("20060102150405")

	// 确保测试结束后清理环境
	defer func() {
		_, err := redis.Do(context.Background(), "DEL", testKey).Result()
		if err != nil {
			t.Logf("清理测试环境失败: %v", err)
		}
	}()

	// 定义一个简单的 Lua 脚本，实现计数器功能
	script := `
		local current = redis.call('GET', KEYS[1])
		if current == false then
			current = 0
		end
		local new = current + ARGV[1]
		redis.call('SET', KEYS[1], new)
		return new
	`

	// 加载脚本并获取 SHA1 值
	sha1, err := redis.ScriptLoad(context.Background(), script).Result()
	if err != nil {
		t.Fatalf("加载脚本失败: %v", err)
	}
	t.Logf("脚本加载成功，SHA1: %s", sha1)

	// 使用 EVALSHA 执行脚本
	// KEYS[1] = testKey, ARGV[1] = 1
	result, err := redis.EvalSha(context.Background(), sha1, []string{testKey}, 1).Result()
	if err != nil {
		t.Fatalf("执行脚本失败: %v", err)
	}

	// 将结果转换为数字
	resultNum, ok := result.(int64)
	if !ok {
		t.Fatalf("结果类型转换失败: %v", result)
	}

	// 验证第一次执行结果
	expected := int64(1)
	if resultNum != expected {
		t.Fatalf("第一次执行结果不符合预期: 期望 %v, 实际 %v", expected, resultNum)
	}
	t.Logf("第一次执行成功，计数器值: %v", resultNum)

	// 再次执行脚本，验证计数器递增
	result, err = redis.EvalSha(context.Background(), sha1, []string{testKey}, 1).Result()
	if err != nil {
		t.Fatalf("执行脚本失败: %v", err)
	}

	// 将结果转换为数字
	resultNum, ok = result.(int64)
	if !ok {
		t.Fatalf("结果类型转换失败: %v", result)
	}

	// 验证第二次执行结果
	expected = int64(2)
	if resultNum != expected {
		t.Fatalf("第二次执行结果不符合预期: 期望 %v, 实际 %v", expected, resultNum)
	}
	t.Logf("第二次执行成功，计数器值: %v", resultNum)
}

func TestRedis_InterfaceMethods(t *testing.T) {
	redis := NewRedis()
	redisExtension := NewRedisExtension(redis)
	if !testRedisConnected(redisExtension) {
		t.Skip("Redis 连接失败，跳过测试")
	}
	ctx := context.Background()

	t.Run("Do", func(t *testing.T) {
		// 测试 SET/GET/DEL 命令
		testKey := "test:do:key"
		_, err := redis.Do(ctx, "SET", testKey, "v1").Result()
		assert.NoError(t, err, "Do SET 应无错误")
		val, err := redis.Do(ctx, "GET", testKey).Result()
		assert.NoError(t, err, "Do GET 应无错误")
		assert.Equal(t, "v1", val, "Do GET 返回应为 v1")
		cnt, err := redis.Do(ctx, "DEL", testKey).Result()
		assert.NoError(t, err, "Do DEL 应无错误")
		assert.Equal(t, int64(1), cnt, "Do DEL 返回应为 1")
	})

	t.Run("Pipelined", func(t *testing.T) {
		testKey := "test:pipeline:key"
		_, err := redis.Pipelined(ctx, func(pipe Pipeliner) error {
			pipe.Do(ctx, "SET", testKey, "v2")
			pipe.Do(ctx, "GET", testKey)
			return nil
		})
		assert.NoError(t, err, "Pipelined 应无错误")
		val, err := redis.Do(ctx, "GET", testKey).Result()
		assert.NoError(t, err, "Pipelined GET 应无错误")
		assert.Equal(t, "v2", val, "Pipelined GET 返回应为 v2")
		redis.Do(ctx, "DEL", testKey)
	})

	t.Run("TxPipelined", func(t *testing.T) {
		testKey := "test:txpipeline:key"
		_, err := redis.TxPipelined(ctx, func(pipe Pipeliner) error {
			pipe.Do(ctx, "SET", testKey, "v3")
			pipe.Do(ctx, "GET", testKey)
			return nil
		})
		assert.NoError(t, err, "TxPipelined 应无错误")
		val, err := redis.Do(ctx, "GET", testKey).Result()
		assert.NoError(t, err, "TxPipelined GET 应无错误")
		assert.Equal(t, "v3", val, "TxPipelined GET 返回应为 v3")
		redis.Do(ctx, "DEL", testKey)
	})

	t.Run("Subscribe/PSubscribe", func(t *testing.T) {
		// 仅测试订阅对象创建，不做消息收发
		ps := redis.Subscribe(ctx, "test:channel")
		assert.NotNil(t, ps, "Subscribe 返回对象不应为 nil")
		pps := redis.PSubscribe(ctx, "test:chan*")
		assert.NotNil(t, pps, "PSubscribe 返回对象不应为 nil")
	})
}

func TestRedis_ScriptCommands(t *testing.T) {
	redis := NewRedis()
	redisExtension := NewRedisExtension(redis)
	if !testRedisConnected(redisExtension) {
		t.Skip("Redis 连接失败，跳过测试")
	}
	ctx := context.Background()
	script := `return ARGV[1]` // 简单返回参数

	t.Run("Eval", func(t *testing.T) {
		res, err := redis.Eval(ctx, script, []string{}, 123).Result()
		assert.NoError(t, err, "Eval 应无错误")
		assert.Equal(t, "123", res, "Eval 返回应为 123")
	})

	t.Run("EvalRO", func(t *testing.T) {
		res, err := redis.EvalRO(ctx, script, []string{}, 456).Result()
		assert.NoError(t, err, "EvalRO 应无错误")
		assert.Equal(t, "456", res, "EvalRO 返回应为 456")
	})

	t.Run("ScriptLoad/EvalSha/EvalShaRO/ScriptExists", func(t *testing.T) {
		sha, err := redis.ScriptLoad(ctx, script).Result()
		assert.NoError(t, err, "ScriptLoad 应无错误")
		assert.NotEmpty(t, sha, "ScriptLoad 返回 SHA1 不应为空")
		// EvalSha
		res, err := redis.EvalSha(ctx, sha, []string{}, 789).Result()
		assert.NoError(t, err, "EvalSha 应无错误")
		assert.Equal(t, "789", res, "EvalSha 返回应为 789")
		// EvalShaRO
		res, err = redis.EvalShaRO(ctx, sha, []string{}, 1011).Result()
		assert.NoError(t, err, "EvalShaRO 应无错误")
		assert.Equal(t, "1011", res, "EvalShaRO 返回应为 1011")
		// ScriptExists
		exists, err := redis.ScriptExists(ctx, sha).Result()
		assert.NoError(t, err, "ScriptExists 应无错误")
		assert.True(t, exists[0], "ScriptExists 应为 true")
	})

	t.Run("ScriptFlush", func(t *testing.T) {
		res, err := redisExtension.ScriptFlush(ctx).Result()
		assert.NoError(t, err, "ScriptFlush 应无错误")
		assert.Equal(t, "OK", res, "ScriptFlush 返回应为 OK")
	})

	t.Run("ScriptKill", func(t *testing.T) {
		// 一般情况下无正在运行脚本，ScriptKill 返回错误或 OK 都可接受
		_, _ = redisExtension.ScriptKill(ctx).Result()
	})
}

func TestRedis_Options(t *testing.T) {
	// 测试默认参数
	redis := NewRedis()
	assert.NotNil(t, redis, "NewRedis 返回对象不应为 nil")

	// 测试自定义地址和密码
	customAddr := "localhost:6380"
	customPwd := "testpwd"
	redis2 := NewRedis(WithAddr(customAddr), WithPassword(customPwd))
	assert.NotNil(t, redis2, "自定义参数 NewRedis 返回对象不应为 nil")

	// 通过类型断言检查参数是否生效
	if c, ok := redis2.(*redisClient); ok {
		assert.Equal(t, customAddr, c.addr, "WithAddr 应设置 addr")
		assert.Equal(t, customPwd, c.password, "WithPassword 应设置 password")
	}
}

// TestRedis_BusinessCases
//
// 本测试函数补充有实际业务意义的边界场景和异常分支，采用表格驱动和详细注释，提升覆盖率。
// 覆盖内容：
//  1. 发布订阅消息收发与取消
//  2. 管道/事务中包含错误命令
//  3. EvalSha 脚本不存在
//  4. Get/Del 不存在 key
//  5. Set 过期后自动删除
//  6. Option 多次叠加
func TestRedis_BusinessCases(t *testing.T) {
	redis := NewRedis()
	redisExtension := NewRedisExtension(redis)
	if !testRedisConnected(redisExtension) {
		t.Skip("Redis 连接失败，跳过测试")
	}
	ctx := context.Background()

	t.Run("Publish/Subscribe 消息收发与取消", func(t *testing.T) {
		// 订阅频道
		channel := "test:pubsub:channel"
		pubsub := redis.Subscribe(ctx, channel)
		defer pubsub.Close()
		// 发布消息
		msg := "hello_pubsub"
		go func() {
			time.Sleep(100 * time.Millisecond)
			redis.Do(ctx, "PUBLISH", channel, msg)
		}()
		// 接收消息
		m, err := pubsub.ReceiveMessage(ctx)
		assert.NoError(t, err, "应能收到消息")
		assert.Equal(t, msg, m.Payload, "消息内容应一致")
		// 取消订阅
		err = pubsub.Unsubscribe(ctx, channel)
		assert.NoError(t, err, "取消订阅应无错误")
	})

	t.Run("Pipelined/TxPipelined 包含错误命令", func(t *testing.T) {
		testKey := "test:pipeline:error"
		// 管道中包含错误命令
		cmds, err := redis.Pipelined(ctx, func(pipe Pipeliner) error {
			pipe.Do(ctx, "SET", testKey, "v1")
			pipe.Do(ctx, "NOTACMD", testKey)
			return nil
		})
		assert.Error(t, err, "管道包含错误命令应返回错误")
		assert.Len(t, cmds, 2, "应有两个命令结果")
		// 事务中包含错误命令
		cmds, err = redis.TxPipelined(ctx, func(pipe Pipeliner) error {
			pipe.Do(ctx, "SET", testKey, "v2")
			pipe.Do(ctx, "NOTACMD", testKey)
			return nil
		})
		assert.Error(t, err, "事务包含错误命令应返回错误")
		assert.Len(t, cmds, 2, "应有两个命令结果")
	})

	t.Run("EvalSha 脚本不存在", func(t *testing.T) {
		sha := "ffffffffffffffffffffffffffffffffffffffff"
		_, err := redis.EvalSha(ctx, sha, []string{}, 1).Result()
		assert.Error(t, err, "EvalSha 脚本不存在应报错")
	})

	t.Run("Get/Del 不存在 key", func(t *testing.T) {
		nonExistKey := "test:nonexist:key"
		val, err := redisExtension.Get(ctx, nonExistKey).Result()
		assert.Error(t, err, "Get 不存在 key 应报错")
		assert.Nil(t, val, "Get 不存在 key 返回应为 nil")
		cnt, err := redisExtension.Del(ctx, nonExistKey).Result()
		assert.NoError(t, err, "Del 不存在 key 应无错误")
		assert.Equal(t, int64(0), cnt, "Del 不存在 key 返回应为 0")
	})

	t.Run("Set 过期后自动删除", func(t *testing.T) {
		testKey := "test:expire:key"
		redisExtension.Set(ctx, testKey, "v1", time.Millisecond*200)
		time.Sleep(300 * time.Millisecond)
		val, err := redisExtension.Get(ctx, testKey).Result()
		assert.Error(t, err, "过期后 Get 应报错")
		assert.Nil(t, val, "过期后 Get 返回应为 nil")
	})

	t.Run("Option 多次叠加", func(t *testing.T) {
		addr1 := "127.0.0.1:6379"
		addr2 := "127.0.0.1:6380"
		pwd1 := "pwd1"
		pwd2 := "pwd2"
		redis3 := NewRedis(WithAddr(addr1), WithAddr(addr2), WithPassword(pwd1), WithPassword(pwd2))
		if c, ok := redis3.(*redisClient); ok {
			assert.Equal(t, addr2, c.addr, "后设置的 addr 应生效")
			assert.Equal(t, pwd2, c.password, "后设置的 password 应生效")
		}
	})
}
