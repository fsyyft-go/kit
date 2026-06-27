// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package aes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGCMInvalidNonceLengthPublicContract 验证公开 GCM API 对非法 nonce 长度的错误契约。
//
// 该测试通过表驱动用例覆盖 EncryptGCM 与 DecryptGCM 在空、短和超长 nonce 下的返回值，确保公开入口返回可诊断错误且不会触发 panic。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGCMInvalidNonceLengthPublicContract(t *testing.T) {
	validKey := []byte(testKeyBytes)
	validNonce := []byte("123456789012")
	validPlaintext := []byte(testPlainText)
	validCiphertext, err := EncryptGCM(validKey, validNonce, validPlaintext)
	require.NoError(t, err)
	require.Greater(t, len(validCiphertext), len(validNonce))
	validCiphertextBody := validCiphertext[len(validNonce):]

	tests := []struct {
		name        string
		description string
		giveNonce   []byte
		run         func([]byte) ([]byte, error)
	}{
		{
			name:        "error/encrypt-empty-nonce",
			description: "验证 EncryptGCM 在 nonce 为空时返回非法 nonce 长度错误而不是 panic。",
			giveNonce:   []byte{},
			run: func(nonce []byte) ([]byte, error) {
				return EncryptGCM(validKey, nonce, validPlaintext)
			},
		},
		{
			name:        "error/encrypt-short-nonce",
			description: "验证 EncryptGCM 在 nonce 少于 GCM 标准长度时返回非法 nonce 长度错误。",
			giveNonce:   []byte("12345678901"),
			run: func(nonce []byte) ([]byte, error) {
				return EncryptGCM(validKey, nonce, validPlaintext)
			},
		},
		{
			name:        "error/encrypt-long-nonce",
			description: "验证 EncryptGCM 在 nonce 多于 GCM 标准长度时返回非法 nonce 长度错误。",
			giveNonce:   []byte("1234567890123"),
			run: func(nonce []byte) ([]byte, error) {
				return EncryptGCM(validKey, nonce, validPlaintext)
			},
		},
		{
			name:        "error/decrypt-empty-nonce",
			description: "验证 DecryptGCM 在 nonce 为空时返回非法 nonce 长度错误而不是 panic。",
			giveNonce:   []byte{},
			run: func(nonce []byte) ([]byte, error) {
				return DecryptGCM(validKey, nonce, validCiphertextBody)
			},
		},
		{
			name:        "error/decrypt-short-nonce",
			description: "验证 DecryptGCM 在 nonce 少于 GCM 标准长度时返回非法 nonce 长度错误。",
			giveNonce:   []byte("12345678901"),
			run: func(nonce []byte) ([]byte, error) {
				return DecryptGCM(validKey, nonce, validCiphertextBody)
			},
		},
		{
			name:        "error/decrypt-long-nonce",
			description: "验证 DecryptGCM 在 nonce 多于 GCM 标准长度时返回非法 nonce 长度错误。",
			giveNonce:   []byte("1234567890123"),
			run: func(nonce []byte) ([]byte, error) {
				return DecryptGCM(validKey, nonce, validCiphertextBody)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			var got []byte
			var gotErr error
			require.NotPanics(t, func() {
				got, gotErr = tt.run(append([]byte(nil), tt.giveNonce...))
			})

			require.Error(t, gotErr)
			assert.Nil(t, got)
			assert.Contains(t, gotErr.Error(), "invalid nonce length")
			assert.Contains(t, gotErr.Error(), "want 12")
		})
	}
}
