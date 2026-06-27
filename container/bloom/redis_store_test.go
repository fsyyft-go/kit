// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package bloom

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	kitbytes "github.com/fsyyft-go/kit/bytes"
	kitredis "github.com/fsyyft-go/kit/database/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateRedisClient 构造 Redis integration 测试客户端。
//
// 该辅助函数仅在 KIT_TEST_INTEGRATION=1 的测试路径中使用，通过环境变量读取 Redis 地址和密码，避免默认单元测试访问外部服务。
//
// 返回：
//   - kitredis.Redis: 可用于 Redis integration 测试的客户端。
//   - error: Redis 地址缺失或 PING 失败时返回的错误。
func generateRedisClient() (kitredis.Redis, error) {
	addr := os.Getenv("KIT_TEST_REDIS_ADDR")
	if addr == "" {
		return nil, errors.New("KIT_TEST_REDIS_ADDR is required when KIT_TEST_INTEGRATION=1")
	}

	r := kitredis.NewRedis(
		kitredis.WithAddr(addr),
		kitredis.WithPassword(os.Getenv("KIT_TEST_REDIS_PASSWORD")),
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := r.Do(ctx, "PING").Result(); err != nil {
		_ = r.Close()
		return nil, err
	}

	return r, nil
}

// TestRedisStoreSimple 验证 RedisStore 在显式 integration 环境下可写入并查询已插入值。
//
// 该测试默认跳过，只有设置 KIT_TEST_INTEGRATION=1 和 KIT_TEST_REDIS_ADDR 后才访问 Redis；测试不强断言未插入值为 false。
//
// 参数：
//   - t: 测试上下文，用于运行测试、注册清理函数和报告断言失败。
func TestRedisStoreSimple(t *testing.T) {
	if os.Getenv("KIT_TEST_INTEGRATION") != "1" {
		t.Skipf("set KIT_TEST_INTEGRATION=1 and KIT_TEST_REDIS_ADDR to run Redis integration test")
	}

	redisClient, err := generateRedisClient()
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, redisClient.Close())
	})

	const bloomName = "test_bloom"
	redisKey := fmt.Sprintf(redisKeyFormat, bloomName)

	_, err = redisClient.Do(t.Context(), "DEL", redisKey).Result()
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = redisClient.Do(context.Background(), "DEL", redisKey).Result()
	})

	bloom, cleanup, err := NewBloom(WithRedis(redisClient), WithName(bloomName))
	require.NoError(t, err)
	require.NotNil(t, bloom)
	if cleanup != nil {
		t.Cleanup(cleanup)
	}

	for range 1000 {
		random, err := kitbytes.GenerateNonce(10)
		require.NoError(t, err)

		randomString := string(random)

		require.NoError(t, bloom.Put(t.Context(), randomString))

		exists, err := bloom.Contain(t.Context(), randomString)
		require.NoError(t, err)
		assert.True(t, exists)
	}
}
