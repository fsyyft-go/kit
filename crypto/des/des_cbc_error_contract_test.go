// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package des_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kitdes "github.com/fsyyft-go/kit/crypto/des"
)

// TestCBCPkCS7HexWrappers_PublicContracts 验证 CBC PKCS7 十六进制包装函数的公开行为。
//
// 该测试通过表驱动用例覆盖字符串密钥包装函数与十六进制密钥包装函数的一致加密、解密和大写密文契约。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestCBCPkCS7HexWrappers_PublicContracts(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveKey     string
		giveData    string
	}{
		{
			name:        "success/ascii-plaintext",
			description: "验证字符串密钥包装函数与十六进制密钥包装函数对 ASCII 明文返回一致的大写密文并可解密。",
			giveKey:     "12345678",
			giveData:    "Hello, World!",
		},
		{
			name:        "success/unicode-plaintext",
			description: "验证字符串密钥包装函数与十六进制密钥包装函数对 UTF-8 中文明文保持一致往返。",
			giveKey:     "go-kit-k",
			giveData:    "中文配置示例",
		},
		{
			name:        "success/empty-plaintext",
			description: "验证字符串密钥包装函数与十六进制密钥包装函数对空明文使用完整 PKCS7 填充块并可解密为空字符串。",
			giveKey:     "12345678",
			giveData:    "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			keyHex := hex.EncodeToString([]byte(tt.giveKey))

			stringCiphertext, err := kitdes.EncryptStringCBCPkCS7PaddingStringHex(tt.giveKey, tt.giveData)
			require.NoError(t, err)
			require.NotEmpty(t, stringCiphertext)
			directCiphertext, err := kitdes.EncryptStringCBCPkCS7PaddingHex(keyHex, tt.giveData)
			require.NoError(t, err)

			assert.Equal(t, directCiphertext, stringCiphertext, "字符串密钥包装函数应等价于传入相同字节密钥的十六进制包装函数。")
			assert.Equal(t, strings.ToUpper(stringCiphertext), stringCiphertext, "公开包装函数应返回大写十六进制密文。")

			stringPlaintext, err := kitdes.DecryptStringCBCPkCS7PaddingStringHex(tt.giveKey, stringCiphertext)
			require.NoError(t, err)
			directPlaintext, err := kitdes.DecryptStringCBCPkCS7PaddingHex(keyHex, directCiphertext)
			require.NoError(t, err)

			assert.Equal(t, tt.giveData, stringPlaintext)
			assert.Equal(t, tt.giveData, directPlaintext)
		})
	}
}

// TestCBCPkCS7DecryptErrorBoundaries 验证 CBC PKCS7 解密入口对非法块内容的错误返回。
//
// 该测试覆盖字符串包装、十六进制包装、默认 IV 和独立 IV 解密函数在块对齐但填充非法的密文下返回 error 且不产生明文。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestCBCPkCS7DecryptErrorBoundaries(t *testing.T) {
	invalidPaddingCipherHex := "0000000000000000"
	invalidPaddingCipherBytes, err := hex.DecodeString(invalidPaddingCipherHex)
	require.NoError(t, err)

	tests := []struct {
		name            string
		description     string
		run             func() ([]byte, error)
		wantErrContains string
	}{
		{
			name:        "error/string-key-wrapper-invalid-padding",
			description: "验证字符串密钥解密包装函数在块对齐但 PKCS7 填充非法时返回错误。",
			run: func() ([]byte, error) {
				plaintext, err := kitdes.DecryptStringCBCPkCS7PaddingStringHex("12345678", invalidPaddingCipherHex)
				return []byte(plaintext), err
			},
			wantErrContains: "invalid padding",
		},
		{
			name:        "error/hex-key-wrapper-invalid-padding",
			description: "验证十六进制密钥解密包装函数在块对齐但 PKCS7 填充非法时返回错误。",
			run: func() ([]byte, error) {
				plaintext, err := kitdes.DecryptStringCBCPkCS7PaddingHex("3132333435363738", invalidPaddingCipherHex)
				return []byte(plaintext), err
			},
			wantErrContains: "invalid padding",
		},
		{
			name:        "error/default-iv-invalid-padding",
			description: "验证默认 IV 解密函数在块对齐但 PKCS7 填充非法时返回错误且不产生明文。",
			run: func() ([]byte, error) {
				return kitdes.DecryptCBCPkCS7Padding([]byte("12345678"), invalidPaddingCipherBytes)
			},
			wantErrContains: "invalid padding",
		},
		{
			name:        "error/alone-iv-invalid-padding",
			description: "验证独立 IV 解密函数在块对齐但 PKCS7 填充非法时返回错误且不产生明文。",
			run: func() ([]byte, error) {
				return kitdes.DecryptCBCPkCS7PaddingAloneIV([]byte("12345678"), []byte("12345678"), invalidPaddingCipherBytes)
			},
			wantErrContains: "invalid padding",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			var got []byte
			var gotErr error
			require.NotPanics(t, func() {
				got, gotErr = tt.run()
			})

			require.Error(t, gotErr)
			assert.Empty(t, got)
			assert.Contains(t, gotErr.Error(), tt.wantErrContains)
		})
	}
}

// TestCBCPkCS7DecryptNonBlockAlignedCiphertextErrors 验证 CBC PKCS7 解密入口对非块对齐密文返回诊断性错误。
//
// 该测试覆盖公开解密入口在密文字节数不是 DES 块大小倍数时不触发 panic，并通过 error 返回可诊断的输入边界问题。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestCBCPkCS7DecryptNonBlockAlignedCiphertextErrors(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		run             func() ([]byte, error)
		wantErrContains string
	}{
		{
			name:        "error/string-key-wrapper-non-block-aligned-ciphertext",
			description: "验证字符串密钥解密包装函数在十六进制密文可解码但字节数不是 DES 块大小倍数时返回诊断性错误。",
			run: func() ([]byte, error) {
				plaintext, err := kitdes.DecryptStringCBCPkCS7PaddingStringHex("12345678", "0000")
				return []byte(plaintext), err
			},
			wantErrContains: "ciphertext length must be a multiple of block size",
		},
		{
			name:        "error/hex-key-wrapper-non-block-aligned-ciphertext",
			description: "验证十六进制密钥解密包装函数在密文字节数不是 DES 块大小倍数时返回诊断性错误。",
			run: func() ([]byte, error) {
				plaintext, err := kitdes.DecryptStringCBCPkCS7PaddingHex("3132333435363738", "0000")
				return []byte(plaintext), err
			},
			wantErrContains: "ciphertext length must be a multiple of block size",
		},
		{
			name:        "error/default-iv-non-block-aligned-ciphertext",
			description: "验证默认 IV 解密函数在密文字节数不是 DES 块大小倍数时返回诊断性错误且不产生明文。",
			run: func() ([]byte, error) {
				return kitdes.DecryptCBCPkCS7Padding([]byte("12345678"), []byte{0x00, 0x00})
			},
			wantErrContains: "ciphertext length must be a multiple of block size",
		},
		{
			name:        "error/alone-iv-non-block-aligned-ciphertext",
			description: "验证独立 IV 解密函数在密文字节数不是 DES 块大小倍数时返回诊断性错误且不产生明文。",
			run: func() ([]byte, error) {
				return kitdes.DecryptCBCPkCS7PaddingAloneIV([]byte("12345678"), []byte("12345678"), []byte{0x00, 0x00})
			},
			wantErrContains: "ciphertext length must be a multiple of block size",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			var got []byte
			var gotErr error
			require.NotPanics(t, func() {
				got, gotErr = tt.run()
			})

			require.Error(t, gotErr)
			assert.Empty(t, got)
			assert.Contains(t, gotErr.Error(), tt.wantErrContains)
		})
	}
}
