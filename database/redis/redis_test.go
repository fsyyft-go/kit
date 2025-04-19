// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package redis

import (
	"context"
	"testing"
	"time"
)

func TestRedis(t *testing.T) {
	redis := NewRedis()

	redis.Do(context.Background(), "SET", "key", "value")
	val, err := redis.Do(context.Background(), "GET", "key").Result()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(val)
}

func TestRedis_EvalSha(t *testing.T) {
	// 创建 Redis 实例
	redis := NewRedis()

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
