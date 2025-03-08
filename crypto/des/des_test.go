// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package des 测试文件设计说明：

1. 测试目标：
   - 验证 PKCS7Padding 填充函数的正确性。
   - 验证 PKCS7UnPadding 去填充函数的正确性。
   - 验证填充和去填充函数的配对使用。

2. 测试策略：
   - 采用表格驱动测试方式，提高代码复用性。
   - 覆盖正常场景、边界场景和特殊场景。
   - 使用 testify/assert 包确保测试结果的准确性。

3. 测试执行方法：
   - 运行所有测试：go test ./crypto/des
   - 查看测试覆盖率：go test -cover ./crypto/des
   - 生成覆盖率报告：go test -coverprofile=coverage.out ./crypto/des
*/

package des

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPKCS7Padding 测试 PKCS7Padding 函数的各种场景。
func TestPKCS7Padding(t *testing.T) {
	// 定义测试用例集合，包含多个测试场景。
	tests := []struct {
		name      string // 测试用例的名称。
		input     []byte // 输入的原始数据。
		blockSize int    // 块大小（字节数）。
		want      []byte // 期望得到的填充后数据。
		wantErr   bool   // 是否期望发生错误。
	}{
		{
			name:      "空数据填充",
			input:     []byte{},
			blockSize: 8,
			want:      []byte{8, 8, 8, 8, 8, 8, 8, 8},
			wantErr:   false,
		},
		{
			name:      "数据长度小于块大小",
			input:     []byte{1, 2, 3},
			blockSize: 8,
			want:      []byte{1, 2, 3, 5, 5, 5, 5, 5},
			wantErr:   false,
		},
		{
			name:      "数据长度等于块大小",
			input:     []byte{1, 2, 3, 4, 5, 6, 7, 8},
			blockSize: 8,
			want:      []byte{1, 2, 3, 4, 5, 6, 7, 8, 8, 8, 8, 8, 8, 8, 8, 8},
			wantErr:   false,
		},
		{
			name:      "数据长度大于块大小",
			input:     []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
			blockSize: 8,
			want:      []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 7, 7, 7, 7, 7, 7, 7},
			wantErr:   false,
		},
	}

	// 遍历执行每个测试用例。
	for _, tt := range tests {
		// 使用子测试方式运行每个测试场景。
		t.Run(tt.name, func(t *testing.T) {
			// 执行填充操作并验证结果。
			got := PKCS7Padding(tt.input, tt.blockSize)
			assert.Equal(t, tt.want, got, "PKCS7Padding() = %v, want %v", got, tt.want)
		})
	}
}

// TestPKCS7UnPadding 测试 PKCS7UnPadding 函数的各种场景。
func TestPKCS7UnPadding(t *testing.T) {
	// 定义测试用例集合。
	tests := []struct {
		name    string // 测试用例的名称。
		input   []byte // 输入的已填充数据。
		want    []byte // 期望得到的去填充后数据。
		wantErr bool   // 是否期望发生错误。
	}{
		{
			name:    "正常去填充",
			input:   []byte{1, 2, 3, 4, 4, 4, 4, 4},
			want:    []byte{1, 2, 3, 4},
			wantErr: false,
		},
		{
			name:    "完整块去填充",
			input:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 8, 8, 8, 8, 8, 8, 8, 8},
			want:    []byte{1, 2, 3, 4, 5, 6, 7, 8},
			wantErr: false,
		},
	}

	// 遍历执行每个测试用例。
	for _, tt := range tests {
		// 使用子测试方式运行每个测试场景。
		t.Run(tt.name, func(t *testing.T) {
			// 执行去填充操作并验证结果。
			got := PKCS7UnPadding(tt.input)
			assert.Equal(t, tt.want, got, "PKCS7UnPadding() = %v, want %v", got, tt.want)
		})
	}
}

// TestPKCS7PaddingAndUnPadding 测试 PKCS7 填充和去填充函数的配对使用。
func TestPKCS7PaddingAndUnPadding(t *testing.T) {
	// 定义测试用例集合。
	tests := []struct {
		name      string // 测试用例的名称。
		input     []byte // 输入的原始数据。
		blockSize int    // 块大小（字节数）。
		wantErr   bool   // 是否期望发生错误。
	}{
		{
			name:      "填充和去填充配对测试-空数据",
			input:     []byte{},
			blockSize: 8,
			wantErr:   false,
		},
		{
			name:      "填充和去填充配对测试-短数据",
			input:     []byte{1, 2, 3},
			blockSize: 8,
			wantErr:   false,
		},
		{
			name:      "填充和去填充配对测试-块大小数据",
			input:     []byte{1, 2, 3, 4, 5, 6, 7, 8},
			blockSize: 8,
			wantErr:   false,
		},
		{
			name:      "填充和去填充配对测试-长数据",
			input:     []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
			blockSize: 8,
			wantErr:   false,
		},
	}

	// 遍历执行每个测试用例。
	for _, tt := range tests {
		// 使用子测试方式运行每个测试场景。
		t.Run(tt.name, func(t *testing.T) {
			// 首先执行填充操作。
			padded := PKCS7Padding(tt.input, tt.blockSize)
			// 然后执行去填充操作。
			unpadded := PKCS7UnPadding(padded)
			// 验证去填充后的结果是否与原始数据相同。
			assert.Equal(t, tt.input, unpadded, "PKCS7UnPadding(PKCS7Padding()) should return original data")
		})
	}
}
