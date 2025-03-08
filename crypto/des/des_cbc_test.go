// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package des_test 实现了 DES 加密算法相关功能的测试。
//
// 测试文件设计思路：
// 1. 采用表格驱动测试方式，提高测试代码的可维护性和可读性；
// 2. 使用 testify 断言库，提供更直观和功能丰富的断言方式；
// 3. 测试用例覆盖：
//   - 正常用例：验证基本的加解密功能；
//   - 边界用例：空字符串、特殊字符等；
//   - 错误用例：非法密钥、非法数据等；
//   - 加解密配对：验证加密后能正确解密；
//
// 4. 每个函数的测试都包含详细的中文注释，说明测试目的和预期结果。
//
// 使用方法：
// 1. 运行所有测试：go test -v ./crypto/des
// 2. 运行覆盖率测试：go test -v -cover ./crypto/des
// 3. 生成覆盖率报告：go test -v -coverprofile=coverage.out ./crypto/des
// 4. 查看覆盖率报告：go tool cover -html=coverage.out
package des_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fsyyft-go/kit/crypto/des"
)

// TestEncryptStringCBCPkCS7PaddingStringHex 测试使用 UTF-8 编码的字符串密钥进行 DES CBC 加密。
func TestEncryptStringCBCPkCS7PaddingStringHex(t *testing.T) {
	tests := []struct {
		name    string // 测试用例名称
		key     string // 加密密钥
		data    string // 待加密数据
		want    string // 预期的加密结果（16 进制字符串）
		wantErr bool   // 是否期望错误
	}{
		{
			name: "正常加密中文字符串",
			key:  "12345678",
			data: "你好，世界！",
			want: "AA217B34D883AC1ECE1F8A6B8A45BC45B75815ADA93C87CD",
		},
		{
			name: "应用用例测试-英文",
			key:  "newbienb",
			data: "newbienb",
			want: "ED5C0836038E6E9739670D810E965521",
		},
		{
			name: "应用用例测试-中文",
			key:  "newbienb",
			data: "这是中文",
			want: "B6B93CE25531C9441EE463531E074876",
		},
		{
			name: "正常加密英文字符串",
			key:  "12345678",
			data: "Hello, World!",
			want: "738939092EEC608E2E2BE40CEB3A6EBE",
		},
		{
			name: "空字符串加密",
			key:  "12345678",
			data: "",
			want: "4431CC0267954866",
		},
		{
			name:    "密钥长度错误",
			key:     "123",
			data:    "test",
			wantErr: true,
		},
		{
			name: "特殊字符加密",
			key:  "12345678",
			data: "!@#$%^&*()_+",
			want: "E269AEFD9C8E95B9CD98B6F68156AA83",
		},
		{
			name: "长字符串加密",
			key:  "12345678",
			data: "这是一个很长的字符串，包含中文、English、数字123、特殊字符!@#$%^&*()_+",
			want: "07FF69BCE9F3E04FB146A407642F06FC8F039DA00549E4F193F27E10BD282FBB391C605CA1AE62FDFEAFC65BDD47869B16EA868DD53A388A81891EA135AE7C15A985C788F19068F631A55CE27D128C1BC734B6315497C093A20123E46508888D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 加密
			encrypted, err := des.EncryptStringCBCPkCS7PaddingStringHex(tt.key, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, encrypted)

			// 验证加密结果
			if !tt.wantErr {
				// 比对加密结果
				assert.Equal(t, tt.want, encrypted, "加密结果与预期不符")

				// 解密验证
				decrypted, err := des.DecryptStringCBCPkCS7PaddingStringHex(tt.key, encrypted)
				assert.NoError(t, err)
				assert.Equal(t, tt.data, decrypted, "解密结果与原始数据不符")
			}
		})
	}
}

// TestDecryptStringCBCPkCS7PaddingStringHex 测试使用 UTF-8 编码的字符串密钥进行 DES CBC 解密。
func TestDecryptStringCBCPkCS7PaddingStringHex(t *testing.T) {
	tests := []struct {
		name    string // 测试用例名称
		key     string // 解密密钥
		data    string // 待解密数据（16 进制字符串）
		wantErr bool   // 是否期望错误
	}{
		{
			name:    "密钥长度错误",
			key:     "123",
			data:    "42DC76E9479E0E37",
			wantErr: true,
		},
		{
			name:    "非法 16 进制字符串",
			key:     "12345678",
			data:    "ZZZZ",
			wantErr: true,
		},
		{
			name:    "空 16 进制字符串",
			key:     "12345678",
			data:    "",
			wantErr: true,
		},
		{
			name:    "奇数长度 16 进制字符串",
			key:     "12345678",
			data:    "123",
			wantErr: true,
		},
		{
			name:    "错误的填充数据",
			key:     "12345678",
			data:    "0000000000000000",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := des.DecryptStringCBCPkCS7PaddingStringHex(tt.key, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

// TestEncryptDecryptCBCPkCS7Padding 测试 DES CBC 加解密配对功能。
func TestEncryptDecryptCBCPkCS7Padding(t *testing.T) {
	tests := []struct {
		name    string // 测试用例名称
		key     []byte // 加解密密钥
		data    []byte // 原始数据
		wantErr bool   // 是否期望错误
	}{
		{
			name: "加解密中文字符串",
			key:  []byte("12345678"),
			data: []byte("你好，世界！"),
		},
		{
			name: "加解密英文字符串",
			key:  []byte("12345678"),
			data: []byte("Hello, World!"),
		},
		{
			name: "加解密空字符串",
			key:  []byte("12345678"),
			data: []byte(""),
		},
		{
			name: "加解密特殊字符",
			key:  []byte("12345678"),
			data: []byte("!@#$%^&*()_+"),
		},
		{
			name: "加解密长字符串",
			key:  []byte("12345678"),
			data: []byte("这是一个很长的字符串，包含中文、English、数字123、特殊字符!@#$%^&*()_+"),
		},
		{
			name:    "密钥长度错误",
			key:     []byte("123"),
			data:    []byte("test"),
			wantErr: true,
		},
		{
			name:    "空密钥",
			key:     []byte{},
			data:    []byte("test"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 先加密
			encrypted, err := des.EncryptCBCPkCS7Padding(tt.key, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// 再解密
			decrypted, err := des.DecryptCBCPkCS7Padding(tt.key, encrypted)
			assert.NoError(t, err)
			assert.Equal(t, tt.data, decrypted)
		})
	}
}

// TestEncryptDecryptCBCPkCS7PaddingAloneIV 测试使用独立 IV 的 DES CBC 加解密配对功能。
func TestEncryptDecryptCBCPkCS7PaddingAloneIV(t *testing.T) {
	tests := []struct {
		name    string // 测试用例名称
		key     []byte // 加解密密钥
		iv      []byte // 初始化向量
		data    []byte // 原始数据
		wantErr bool   // 是否期望错误
	}{
		{
			name: "使用不同 IV 加解密中文",
			key:  []byte("12345678"),
			iv:   []byte("87654321"),
			data: []byte("你好，世界！"),
		},
		{
			name: "使用不同 IV 加解密英文",
			key:  []byte("12345678"),
			iv:   []byte("87654321"),
			data: []byte("Hello, World!"),
		},
		{
			name: "使用相同 IV 加解密",
			key:  []byte("12345678"),
			iv:   []byte("12345678"),
			data: []byte("测试数据"),
		},
		{
			name: "使用全零 IV",
			key:  []byte("12345678"),
			iv:   make([]byte, 8),
			data: []byte("test data"),
		},
		{
			name:    "IV 长度错误",
			key:     []byte("12345678"),
			iv:      []byte("123"),
			data:    []byte("test"),
			wantErr: true,
		},
		{
			name:    "空 IV",
			key:     []byte("12345678"),
			iv:      []byte{},
			data:    []byte("test"),
			wantErr: true,
		},
		{
			name:    "密钥长度错误",
			key:     []byte("123"),
			iv:      []byte("12345678"),
			data:    []byte("test"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 先加密
			encrypted, err := des.EncryptCBCPkCS7PaddingAloneIV(tt.key, tt.iv, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// 再解密
			decrypted, err := des.DecryptCBCPkCS7PaddingAloneIV(tt.key, tt.iv, encrypted)
			assert.NoError(t, err)
			assert.Equal(t, tt.data, decrypted)
		})
	}
}
