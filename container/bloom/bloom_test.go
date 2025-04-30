// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package bloom 的测试文件
// 设计思路：
// 1. 使用表格驱动测试方式，覆盖所有主要功能点
// 2. 使用 mock 对象模拟存储层，确保测试的独立性
// 3. 测试用例包括正常和异常场景
// 4. 使用 testify 包进行断言，提高测试代码的可读性
//
// 使用方法：
// 1. 运行所有测试：go test -v
// 2. 运行特定测试：go test -v -run TestBloom_Contain
// 3. 查看测试覆盖率：go test -v -cover

package bloom

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	kitlog "github.com/fsyyft-go/kit/log"
)

// TestBloom_Contain 测试布隆过滤器的 Contain 方法
// 测试场景包括：
// 1. 元素存在的情况
// 2. 元素不存在的情况
// 3. 存储层发生错误的情况
func TestBloom_Contain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockStore(ctrl)
	bloom, _, err := NewBloom(
		WithName("test"),
		WithStore(mockStore),
		WithExpectedElements(1000),
		WithFalsePositiveRate(0.01),
	)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		value   string
		mock    func()
		want    bool
		wantErr bool
	}{
		{
			name:  "元素存在",
			value: "test",
			mock: func() {
				mockStore.EXPECT().
					Exist(gomock.Any(), "test", gomock.Any()).
					Return(true, nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:  "元素不存在",
			value: "test",
			mock: func() {
				mockStore.EXPECT().
					Exist(gomock.Any(), "test", gomock.Any()).
					Return(false, nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:  "存储错误",
			value: "test",
			mock: func() {
				mockStore.EXPECT().
					Exist(gomock.Any(), "test", gomock.Any()).
					Return(false, assert.AnError)
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := bloom.Contain(context.Background(), tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestBloom_Put 测试布隆过滤器的 Put 方法
// 测试场景包括：
// 1. 成功添加元素的情况
// 2. 存储层发生错误的情况
func TestBloom_Put(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockStore(ctrl)
	bloom, _, err := NewBloom(
		WithName("test"),
		WithStore(mockStore),
		WithExpectedElements(1000),
		WithFalsePositiveRate(0.01),
	)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		value   string
		mock    func()
		wantErr bool
	}{
		{
			name:  "添加成功",
			value: "test",
			mock: func() {
				mockStore.EXPECT().
					Add(gomock.Any(), "test", gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "添加失败",
			value: "test",
			mock: func() {
				mockStore.EXPECT().
					Add(gomock.Any(), "test", gomock.Any()).
					Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := bloom.Put(context.Background(), tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBloom_GroupContain 测试布隆过滤器的 GroupContain 方法
// 测试场景包括：
// 1. 分组中元素存在的情况
// 2. 分组中元素不存在的情况
// 3. 存储层发生错误的情况
func TestBloom_GroupContain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockStore(ctrl)
	bloom, _, err := NewBloom(
		WithName("test"),
		WithStore(mockStore),
		WithExpectedElements(1000),
		WithFalsePositiveRate(0.01),
	)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		group   string
		value   string
		mock    func()
		want    bool
		wantErr bool
	}{
		{
			name:  "元素存在",
			group: "group1",
			value: "test",
			mock: func() {
				mockStore.EXPECT().
					Exist(gomock.Any(), "test:group1", gomock.Any()).
					Return(true, nil)
			},
			want:    true,
			wantErr: false,
		},
		{
			name:  "元素不存在",
			group: "group1",
			value: "test",
			mock: func() {
				mockStore.EXPECT().
					Exist(gomock.Any(), "test:group1", gomock.Any()).
					Return(false, nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name:  "存储错误",
			group: "group1",
			value: "test",
			mock: func() {
				mockStore.EXPECT().
					Exist(gomock.Any(), "test:group1", gomock.Any()).
					Return(false, assert.AnError)
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := bloom.GroupContain(context.Background(), tt.group, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestBloom_GroupPut 测试布隆过滤器的 GroupPut 方法
// 测试场景包括：
// 1. 成功添加分组元素的情况
// 2. 存储层发生错误的情况
func TestBloom_GroupPut(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := NewMockStore(ctrl)
	bloom, _, err := NewBloom(
		WithName("test"),
		WithStore(mockStore),
		WithExpectedElements(1000),
		WithFalsePositiveRate(0.01),
	)
	assert.NoError(t, err)

	tests := []struct {
		name    string
		group   string
		value   string
		mock    func()
		wantErr bool
	}{
		{
			name:  "添加成功",
			group: "group1",
			value: "test",
			mock: func() {
				mockStore.EXPECT().
					Add(gomock.Any(), "test:group1", gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "添加失败",
			group: "group1",
			value: "test",
			mock: func() {
				mockStore.EXPECT().
					Add(gomock.Any(), "test:group1", gomock.Any()).
					Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := bloom.GroupPut(context.Background(), tt.group, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestNewBloom_Errors 测试 NewBloom 的所有错误分支和边界条件
func TestNewBloom_Errors(t *testing.T) {
	// 名称为空
	b, _, err := NewBloom(WithName("   "))
	assert.Nil(t, b)
	assert.Equal(t, ErrBloomNameEmpty, err)

	// 名称重复
	bloomNames["dup"] = "dup"
	b, _, err = NewBloom(WithName("dup"))
	assert.Nil(t, b)
	assert.Equal(t, ErrBloomNameRepeated, err)
	delete(bloomNames, "dup")

	// p > 1
	b, _, err = NewBloom(WithName("test2"), WithFalsePositiveRate(1.1))
	assert.Nil(t, b)
	assert.Equal(t, ErrBloomFalseProbabilityThanOne, err)

	// p < 0
	b, _, err = NewBloom(WithName("test3"), WithFalsePositiveRate(-0.1))
	assert.Nil(t, b)
	assert.Equal(t, ErrBloomFalseProbabilityNegative, err)
}

// TestBloom_WithLogger 测试 WithLogger Option
func TestBloom_WithLogger(t *testing.T) {
	logger, err := kitlog.NewLogger()
	assert.NoError(t, err)
	b, _, err := NewBloom(WithName("withlogger"), WithLogger(logger))
	assert.NoError(t, err)
	assert.NotNil(t, b)
	// 断言 logger 字段被正确设置
	bb, ok := b.(*bloom)
	assert.True(t, ok)
	assert.Equal(t, logger, bb.logger)
}

// TestMockBloom_AllMethods 形式化覆盖 MockBloom 所有方法
func TestMockBloom_AllMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockBloom(ctrl)
	mock.EXPECT().Contain(gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	mock.EXPECT().GroupContain(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()
	mock.EXPECT().GroupPut(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mock.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	_, _ = mock.Contain(context.Background(), "v")
	_, _ = mock.GroupContain(context.Background(), "g", "v")
	_ = mock.GroupPut(context.Background(), "g", "v")
	_ = mock.Put(context.Background(), "v")
}

// TestMockStore_AllMethods 形式化覆盖 MockStore 所有方法
func TestMockStore_AllMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockStore(ctrl)
	mock.EXPECT().Add(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mock.EXPECT().Exist(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil).AnyTimes()

	_ = mock.Add(context.Background(), "k", []uint64{1, 2})
	_, _ = mock.Exist(context.Background(), "k", []uint64{1, 2})
}
