// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package bloom 的内存存储测试文件
// 设计思路：
// 1. 使用表格驱动测试方式，覆盖所有主要功能点
// 2. 测试内存存储的基本操作：添加、查询、初始化
// 3. 测试用例包括正常和边界场景
// 4. 使用 testify 包进行断言，提高测试代码的可读性
//
// 使用方法：
// 1. 运行所有测试：go test -v
// 2. 运行特定测试：go test -v -run TestMemoryStore_Exist
// 3. 查看测试覆盖率：go test -v -cover

package bloom

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMemoryStore_Exist 测试内存存储的 Exist 方法
// 测试场景包括：
// 1. 查询不存在的元素
// 2. 查询已添加的元素
// 3. 查询部分匹配的元素
func TestMemoryStore_Exist(t *testing.T) {
	store := NewMemoryStore(1000)
	ctx := context.Background()

	tests := []struct {
		name    string
		key     string
		hashes  []uint64
		setup   func()
		want    bool
		wantErr bool
	}{
		{
			name:   "查询不存在的元素",
			key:    "test",
			hashes: []uint64{1, 2, 3},
			setup:  func() {},
			want:   false,
		},
		{
			name:   "查询已添加的元素",
			key:    "test",
			hashes: []uint64{1, 2, 3},
			setup: func() {
				_ = store.Add(ctx, "test", []uint64{1, 2, 3})
			},
			want: true,
		},
		{
			name:   "查询部分匹配的元素",
			key:    "test",
			hashes: []uint64{1, 2, 4},
			setup: func() {
				_ = store.Add(ctx, "test", []uint64{1, 2, 3})
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := store.Exist(ctx, tt.key, tt.hashes)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestMemoryStore_Add 测试内存存储的 Add 方法
// 测试场景包括：
// 1. 添加新元素
// 2. 重复添加相同元素
// 3. 添加不同 key 的元素
func TestMemoryStore_Add(t *testing.T) {
	store := NewMemoryStore(1000)
	ctx := context.Background()

	tests := []struct {
		name    string
		key     string
		hashes  []uint64
		wantErr bool
	}{
		{
			name:   "添加新元素",
			key:    "test",
			hashes: []uint64{1, 2, 3},
		},
		{
			name:   "重复添加相同元素",
			key:    "test",
			hashes: []uint64{1, 2, 3},
		},
		{
			name:   "添加不同 key 的元素",
			key:    "test2",
			hashes: []uint64{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Add(ctx, tt.key, tt.hashes)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMemoryStore_NewMemoryStore 测试内存存储的初始化
// 测试场景包括：
// 1. 使用默认大小初始化
// 2. 使用指定大小初始化
func TestMemoryStore_NewMemoryStore(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
	}{
		{
			name:     "使用默认大小初始化",
			capacity: 0,
		},
		{
			name:     "使用指定大小初始化",
			capacity: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemoryStore(tt.capacity)
			assert.NotNil(t, store)
		})
	}
}
