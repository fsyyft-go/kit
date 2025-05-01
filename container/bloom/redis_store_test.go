// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package bloom

import (
	"context"
	"testing"
	"time"

	kitbytes "github.com/fsyyft-go/kit/bytes"
	kitredis "github.com/fsyyft-go/kit/database/redis"
)

func generateRedisClient() (kitredis.Redis, error) {
	r := kitredis.NewRedis(kitredis.WithAddr("localhost:6379"), kitredis.WithPassword("redis*2025"))

	// 设置一个较短的超时时间。
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 尝试执行一个简单的 PING 命令。
	_, err := r.Do(ctx, "PING").Result()

	return r, err
}

func TestRedisStoreSimple(t *testing.T) {
	// 创建 Redis 客户端。
	redisClient, err := generateRedisClient()
	if err != nil {
		t.Skip("Redis 连接失败，跳过测试")
	}

	redisClient.Do(t.Context(), "DEL", "kit:bloom:test_bloom")

	bloom, cleanup, err := NewBloom(WithRedis(redisClient), WithName("test_bloom"))
	if err != nil {
		t.Skip("Bloom 创建失败，跳过测试")
	}
	defer cleanup()

	for range 1000 {
		// 生成随机字符串。
		random, err := kitbytes.GenerateNonce(10)
		if err != nil {
			t.Fatalf("Failed to generate random string: %v", err)
		}
		randomString := string(random)

		exists, err := bloom.Contain(t.Context(), randomString)
		if err != nil {
			t.Fatalf("Failed to check existence in RedisStore: %v", err)
		}
		if exists {
			t.Fatalf("Expected values to not exist in RedisStore, but they do")
		}

		err = bloom.Put(t.Context(), randomString)
		if err != nil {
			t.Fatalf("Failed to add values to RedisStore: %v", err)
		}

		exists, err = bloom.Contain(t.Context(), randomString)
		if err != nil {
			t.Fatalf("Failed to check existence in RedisStore: %v", err)
		}
		if !exists {
			t.Fatalf("Expected values to exist in RedisStore, but they do not")
		}
	}
}
