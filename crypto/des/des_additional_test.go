// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package des_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kitdes "github.com/fsyyft-go/kit/crypto/des"
)

// TestGetDefaultDESKey_ReturnsStableKey 验证默认 DES 密钥的公开兼容值。
//
// 该测试断言默认密钥内容和 DES 所需的 8 字节长度保持稳定，避免依赖默认值的调用方发生兼容性回归。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestGetDefaultDESKey_ReturnsStableKey(t *testing.T) {
	// 用例语义：验证默认 DES 密钥保持历史公开值且满足 DES 密钥长度要求。
	got := kitdes.GetDefaultDESKey()

	assert.Equal(t, "go-kit-k", got)
	assert.Len(t, got, 8)
}

// TestEncryptStringCBCPkCS7PaddingHex_DirectAPI 验证十六进制密钥字符串加密入口的成功和错误行为。
//
// 该测试通过表驱动用例覆盖直接传入十六进制密钥时的标准加密结果、非法十六进制密钥和解码后密钥长度错误。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestEncryptStringCBCPkCS7PaddingHex_DirectAPI(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		giveKeyHex      string
		giveData        string
		want            string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:        "success/uppercase-ciphertext",
			description: "验证直接传入合法十六进制 DES 密钥时返回稳定的大写十六进制密文。",
			giveKeyHex:  "3132333435363738",
			giveData:    "Hello, World!",
			want:        "738939092EEC608E2E2BE40CEB3A6EBE",
		},
		{
			name:            "error/non-hex-key",
			description:     "验证密钥不是合法十六进制字符串时返回解码错误且不产生密文。",
			giveKeyHex:      "zz",
			giveData:        "payload",
			wantErr:         true,
			wantErrContains: "invalid byte",
		},
		{
			name:            "error/decoded-key-invalid-size",
			description:     "验证十六进制密钥可解码但长度不满足 DES 要求时返回密钥长度错误。",
			giveKeyHex:      "73686f7274",
			giveData:        "payload",
			wantErr:         true,
			wantErrContains: "invalid key size",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := kitdes.EncryptStringCBCPkCS7PaddingHex(tt.giveKeyHex, tt.giveData)

			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestDecryptStringCBCPkCS7PaddingHex_DirectAPI 验证十六进制密钥字符串解密入口的成功和错误行为。
//
// 该测试通过表驱动用例覆盖直接传入十六进制密钥时的标准解密结果、非法密钥、非法密文和解码后密钥长度错误。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestDecryptStringCBCPkCS7PaddingHex_DirectAPI(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		giveKeyHex      string
		giveDataHex     string
		want            string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:        "success/plaintext",
			description: "验证直接传入合法十六进制 DES 密钥和密文时返回原始明文。",
			giveKeyHex:  "3132333435363738",
			giveDataHex: "738939092EEC608E2E2BE40CEB3A6EBE",
			want:        "Hello, World!",
		},
		{
			name:            "error/non-hex-key",
			description:     "验证密钥不是合法十六进制字符串时返回解码错误且不产生明文。",
			giveKeyHex:      "zz",
			giveDataHex:     "738939092EEC608E2E2BE40CEB3A6EBE",
			wantErr:         true,
			wantErrContains: "invalid byte",
		},
		{
			name:            "error/non-hex-ciphertext",
			description:     "验证密文不是合法十六进制字符串时返回解码错误且不产生明文。",
			giveKeyHex:      "3132333435363738",
			giveDataHex:     "zz",
			wantErr:         true,
			wantErrContains: "invalid byte",
		},
		{
			name:            "error/decoded-key-invalid-size",
			description:     "验证十六进制密钥可解码但长度不满足 DES 要求时返回密钥长度错误。",
			giveKeyHex:      "73686f7274",
			giveDataHex:     "4431CC0267954866",
			wantErr:         true,
			wantErrContains: "invalid key size",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := kitdes.DecryptStringCBCPkCS7PaddingHex(tt.giveKeyHex, tt.giveDataHex)

			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestDecryptCBCPkCS7PaddingAloneIV_ErrorBranches 验证独立 IV 解密入口的参数校验错误行为。
//
// 该测试通过表驱动用例覆盖 DES 解密前的密钥长度校验和 IV 长度校验，确保参数错误以 error 返回而不是产生明文。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestDecryptCBCPkCS7PaddingAloneIV_ErrorBranches(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		giveKey         []byte
		giveIV          []byte
		giveData        []byte
		wantErrContains string
	}{
		{
			name:            "error/invalid-key-size",
			description:     "验证 DES 密钥长度错误时解密函数返回标准库密钥长度错误。",
			giveKey:         []byte("short"),
			giveIV:          []byte("12345678"),
			giveData:        []byte("12345678"),
			wantErrContains: "invalid key size",
		},
		{
			name:            "error/invalid-iv-length",
			description:     "验证 DES 密钥合法但 IV 长度不等于块大小时解密函数返回 IV 长度错误。",
			giveKey:         []byte("12345678"),
			giveIV:          []byte("short"),
			giveData:        []byte("12345678"),
			wantErrContains: "IV length must equal block size",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := kitdes.DecryptCBCPkCS7PaddingAloneIV(tt.giveKey, tt.giveIV, tt.giveData)

			require.Error(t, err)
			assert.Nil(t, got)
			assert.Contains(t, err.Error(), tt.wantErrContains)
		})
	}
}
