// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package redis

import (
	"context"
	"testing"
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
