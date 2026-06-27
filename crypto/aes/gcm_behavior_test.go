// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package aes

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGCMRawBehavior 验证原始 GCM 加解密函数的核心行为。
//
// 该测试通过表驱动用例覆盖成功往返、认证失败、无效密钥、无效 nonce 和密文长度不足语义，确保底层 GCM 契约稳定且错误路径不会 panic。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGCMRawBehavior(t *testing.T) {
	validKey := []byte(testKeyBytes)
	validNonce := []byte("123456789012")
	validPlaintext := []byte(testPlainText)
	validCiphertext, err := EncryptGCM(validKey, validNonce, validPlaintext)
	require.NoError(t, err)
	require.Greater(t, len(validCiphertext), len(validNonce))

	tests := []struct {
		name           string
		description    string
		giveKey        []byte
		giveNonce      []byte
		givePlaintext  []byte
		giveCiphertext []byte
		assert         func(t *testing.T, tt gcmRawBehaviorCase)
	}{
		{
			name:          "success/encrypt-prefixes-nonce-and-decrypts",
			description:   "验证 EncryptGCM 返回 nonce 前缀和可由 DecryptGCM 还原的认证密文。",
			giveKey:       validKey,
			giveNonce:     validNonce,
			givePlaintext: validPlaintext,
			assert: func(t *testing.T, tt gcmRawBehaviorCase) {
				encrypted, err := EncryptGCM(tt.giveKey, tt.giveNonce, tt.givePlaintext)
				require.NoError(t, err)
				require.Greater(t, len(encrypted), len(tt.giveNonce))
				assert.Equal(t, tt.giveNonce, encrypted[:len(tt.giveNonce)])

				decrypted, err := DecryptGCM(tt.giveKey, tt.giveNonce, encrypted[len(tt.giveNonce):])
				require.NoError(t, err)
				assert.Equal(t, tt.givePlaintext, decrypted)
			},
		},
		{
			name:           "success/decrypt-known-ciphertext",
			description:    "验证 DecryptGCM 可解开由相同密钥和 nonce 生成的认证密文片段。",
			giveKey:        validKey,
			giveNonce:      validNonce,
			giveCiphertext: validCiphertext[len(validNonce):],
			assert: func(t *testing.T, tt gcmRawBehaviorCase) {
				decrypted, err := DecryptGCM(tt.giveKey, tt.giveNonce, tt.giveCiphertext)
				require.NoError(t, err)
				assert.Equal(t, validPlaintext, decrypted)
			},
		},
		{
			name:          "error/encrypt-invalid-key",
			description:   "验证 EncryptGCM 在密钥长度非法时返回错误且不产生密文。",
			giveKey:       []byte("invalid-key"),
			giveNonce:     validNonce,
			givePlaintext: validPlaintext,
			assert: func(t *testing.T, tt gcmRawBehaviorCase) {
				encrypted, err := EncryptGCM(tt.giveKey, tt.giveNonce, tt.givePlaintext)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid key size")
				assert.Nil(t, encrypted)
			},
		},
		{
			name:          "error/encrypt-invalid-nonce-without-panic",
			description:   "验证 EncryptGCM 在 nonce 长度非法时返回可诊断错误而不是触发 panic。",
			giveKey:       validKey,
			giveNonce:     []byte("short"),
			givePlaintext: validPlaintext,
			assert: func(t *testing.T, tt gcmRawBehaviorCase) {
				require.NotPanics(t, func() {
					encrypted, err := EncryptGCM(tt.giveKey, tt.giveNonce, tt.givePlaintext)
					require.Error(t, err)
					assert.Contains(t, err.Error(), "invalid nonce length")
					assert.Nil(t, encrypted)
				})
			},
		},
		{
			name:           "error/decrypt-invalid-nonce-without-panic",
			description:    "验证 DecryptGCM 在 nonce 长度非法时返回可诊断错误而不是触发 panic。",
			giveKey:        validKey,
			giveNonce:      []byte("short"),
			giveCiphertext: validCiphertext[len(validNonce):],
			assert: func(t *testing.T, tt gcmRawBehaviorCase) {
				require.NotPanics(t, func() {
					decrypted, err := DecryptGCM(tt.giveKey, tt.giveNonce, tt.giveCiphertext)
					require.Error(t, err)
					assert.Contains(t, err.Error(), "invalid nonce length")
					assert.Nil(t, decrypted)
				})
			},
		},
		{
			name:           "error/decrypt-authentication-failure",
			description:    "验证 DecryptGCM 在认证标签或密文被篡改时返回认证失败错误。",
			giveKey:        validKey,
			giveNonce:      validNonce,
			giveCiphertext: append([]byte(nil), validCiphertext[len(validNonce):]...),
			assert: func(t *testing.T, tt gcmRawBehaviorCase) {
				tampered := append([]byte(nil), tt.giveCiphertext...)
				tampered[len(tampered)-1] ^= 0x01

				decrypted, err := DecryptGCM(tt.giveKey, tt.giveNonce, tampered)
				require.Error(t, err)
				assert.Nil(t, decrypted)
			},
		},
		{
			name:           "error/decrypt-nonce-length-negative-without-panic",
			description:    "验证 DecryptGCMNonceLength 在公开入参 nonceLength 为负数且数据非空时返回错误而不是触发 panic。",
			giveKey:        validKey,
			giveCiphertext: validCiphertext,
			assert: func(t *testing.T, tt gcmRawBehaviorCase) {
				var nonce []byte
				var decrypted []byte
				var err error

				require.NotPanics(t, func() {
					nonce, decrypted, err = DecryptGCMNonceLength(tt.giveKey, -1, tt.giveCiphertext)
				})
				require.Error(t, err)
				assert.Contains(t, err.Error(), "nonce")
				assert.Nil(t, nonce)
				assert.Nil(t, decrypted)
			},
		},
		{
			name:           "error/decrypt-nonce-length-data-too-short",
			description:    "验证 DecryptGCMNonceLength 在密文长度不足以提取 nonce 时返回错误。",
			giveKey:        validKey,
			giveCiphertext: []byte("too-short"),
			assert: func(t *testing.T, tt gcmRawBehaviorCase) {
				nonce, decrypted, err := DecryptGCMNonceLength(tt.giveKey, testNonceLength, tt.giveCiphertext)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "数据长度不足")
				assert.Nil(t, nonce)
				assert.Nil(t, decrypted)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			tt.assert(t, gcmRawBehaviorCase{
				giveKey:        append([]byte(nil), tt.giveKey...),
				giveNonce:      append([]byte(nil), tt.giveNonce...),
				givePlaintext:  append([]byte(nil), tt.givePlaintext...),
				giveCiphertext: append([]byte(nil), tt.giveCiphertext...),
			})
		})
	}
}

// TestGCMEncodedBehavior 验证 GCM 编码包装函数的核心行为。
//
// 该测试覆盖字符串和二进制包装函数在 Base64、Hex 编码下的成功往返与解码错误语义，确保外层 API 的编码契约稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGCMEncodedBehavior(t *testing.T) {
	plainBase64 := base64.StdEncoding.EncodeToString([]byte(testPlainText))
	plainHex := hex.EncodeToString([]byte(testPlainText))

	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "success/string-base64-round-trip",
			description: "验证字符串 Base64 包装函数可完成加密并解密回原始 UTF-8 文本。",
			assert: func(t *testing.T) {
				ciphertext, err := EncryptStringGCMBase64(testKeyBase64, testNonceLength, testPlainText)
				require.NoError(t, err)
				require.NotEmpty(t, ciphertext)

				nonce, plaintext, err := DecryptStringGCMBase64(testKeyBase64, testNonceLength, ciphertext)
				require.NoError(t, err)
				assert.NotEmpty(t, nonce)
				assert.Equal(t, testPlainText, plaintext)
			},
		},
		{
			name:        "success/string-hex-round-trip",
			description: "验证字符串 Hex 包装函数可完成加密并解密回原始 UTF-8 文本。",
			assert: func(t *testing.T) {
				ciphertext, err := EncryptStringGCMHex(testKeyHex, testNonceLength, testPlainText)
				require.NoError(t, err)
				require.NotEmpty(t, ciphertext)

				nonce, plaintext, err := DecryptStringGCMHex(testKeyHex, testNonceLength, ciphertext)
				require.NoError(t, err)
				assert.NotEmpty(t, nonce)
				assert.Equal(t, testPlainText, plaintext)
			},
		},
		{
			name:        "success/binary-base64-round-trip",
			description: "验证二进制 Base64 包装函数返回 Base64 编码的 nonce 和明文。",
			assert: func(t *testing.T) {
				ciphertext, err := EncryptGCMBase64(testKeyBase64, testNonceLength, plainBase64)
				require.NoError(t, err)
				require.NotEmpty(t, ciphertext)

				nonce, plaintext, err := DecryptGCMBase64(testKeyBase64, testNonceLength, ciphertext)
				require.NoError(t, err)
				assert.NotEmpty(t, nonce)
				assert.Equal(t, plainBase64, plaintext)
			},
		},
		{
			name:        "success/binary-hex-round-trip-uppercases-output",
			description: "验证二进制 Hex 包装函数返回大写 Hex 编码的 nonce 和明文。",
			assert: func(t *testing.T) {
				ciphertext, err := EncryptGCMHex(testKeyHex, testNonceLength, plainHex)
				require.NoError(t, err)
				require.NotEmpty(t, ciphertext)

				nonce, plaintext, err := DecryptGCMHex(testKeyHex, testNonceLength, ciphertext)
				require.NoError(t, err)
				assert.NotEmpty(t, nonce)
				assert.Equal(t, strings.ToUpper(plainHex), plaintext)
			},
		},
		{
			name:        "error/base64-key-decode",
			description: "验证 Base64 包装函数在密钥不是合法 Base64 时返回解码错误。",
			assert: func(t *testing.T) {
				ciphertext, err := EncryptStringGCMBase64(invalidBase64, testNonceLength, testPlainText)
				require.Error(t, err)
				assert.Empty(t, ciphertext)
			},
		},
		{
			name:        "error/hex-data-decode",
			description: "验证 Hex 包装函数在明文不是合法 Hex 时返回解码错误。",
			assert: func(t *testing.T) {
				ciphertext, err := EncryptGCMHex(testKeyHex, testNonceLength, invalidHex)
				require.Error(t, err)
				assert.Empty(t, ciphertext)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			tt.assert(t)
		})
	}
}

// gcmRawBehaviorCase 描述原始 GCM 行为测试的单个用例输入。
//
// 该辅助结构体集中承载可变字节切片输入，便于子测试在断言前复制隔离数据。
type gcmRawBehaviorCase struct {
	giveKey        []byte
	giveNonce      []byte
	givePlaintext  []byte
	giveCiphertext []byte
}
