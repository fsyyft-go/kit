// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package rsa

import (
	"bytes"
	"crypto"
	"crypto/rand"
	stdrsa "crypto/rsa"
	"crypto/sha256"
	"encoding/pem"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBoundaryPEMWrapperKeyErrors 验证 PEM wrapper 对非法 DER 和非 RSA 公钥的错误契约。
//
// 该测试通过表驱动用例覆盖公开加解密 wrapper 在 PEM 类型正确但 DER 非法、PEM 类型不匹配和可解析但不是 RSA 公钥时的返回值。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestBoundaryPEMWrapperKeyErrors(t *testing.T) {
	_, privateKeyBytes, publicKeyBytes := generateTestKeyPair(t, 2048)
	plainText := []byte("pem wrapper key error contract")
	validCipherText, err := EncryptPubKey(publicKeyBytes, plainText)
	require.NoError(t, err)
	validSignature, err := EncryptPrivKey(privateKeyBytes, plainText)
	require.NoError(t, err)

	invalidPublicDER := pem.EncodeToMemory(&pem.Block{
		Type:  BlockTypePublicKey,
		Bytes: []byte("not valid DER public key"),
	})
	invalidPrivateDER := pem.EncodeToMemory(&pem.Block{
		Type:  BlockTypePrivateKey,
		Bytes: []byte("not valid DER private key"),
	})
	privateBlockAsPublic := pem.EncodeToMemory(&pem.Block{
		Type:  BlockTypePrivateKey,
		Bytes: []byte("not a public key"),
	})
	publicBlockAsPrivate := pem.EncodeToMemory(&pem.Block{
		Type:  BlockTypePublicKey,
		Bytes: []byte("not a private key"),
	})
	ecdsaPublicKey := buildAdditionalECDSAPublicKeyPEM(t)

	tests := []struct {
		name        string
		description string
		run         func() ([]byte, error)
		wantErrIs   error
	}{
		{
			name:        "error/encrypt-public-key-invalid-der",
			description: "验证 EncryptPubKey 在 PUBLIC KEY PEM 的 DER 内容非法时返回解析错误且不产生密文。",
			run: func() ([]byte, error) {
				return EncryptPubKey(invalidPublicDER, plainText)
			},
		},
		{
			name:        "error/encrypt-public-key-block-type-mismatch",
			description: "验证 EncryptPubKey 在 PEM block type 不是 PUBLIC KEY 时返回 ErrDecodePublicKey。",
			run: func() ([]byte, error) {
				return EncryptPubKey(privateBlockAsPublic, plainText)
			},
			wantErrIs: ErrDecodePublicKey,
		},
		{
			name:        "error/encrypt-public-key-non-rsa",
			description: "验证 EncryptPubKey 在 PKIX 公钥可解析但算法不是 RSA 时返回 ErrDecodePublicKey。",
			run: func() ([]byte, error) {
				return EncryptPubKey(ecdsaPublicKey, plainText)
			},
			wantErrIs: ErrDecodePublicKey,
		},
		{
			name:        "error/decrypt-public-key-non-rsa",
			description: "验证 DecryptPubKey 在 PKIX 公钥可解析但算法不是 RSA 时返回 ErrDecodePublicKey。",
			run: func() ([]byte, error) {
				return DecryptPubKey(ecdsaPublicKey, validSignature)
			},
			wantErrIs: ErrDecodePublicKey,
		},
		{
			name:        "error/decrypt-private-key-invalid-der",
			description: "验证 DecryptPrivKey 在 RSA PRIVATE KEY PEM 的 DER 内容非法时返回解析错误且不产生明文。",
			run: func() ([]byte, error) {
				return DecryptPrivKey(invalidPrivateDER, validCipherText)
			},
		},
		{
			name:        "error/encrypt-private-key-block-type-mismatch",
			description: "验证 EncryptPrivKey 在 PEM block type 不是 RSA PRIVATE KEY 时返回 ErrDecodePrivateKey。",
			run: func() ([]byte, error) {
				return EncryptPrivKey(publicBlockAsPrivate, plainText)
			},
			wantErrIs: ErrDecodePrivateKey,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := tt.run()

			require.Error(t, err)
			assert.Nil(t, got)
			if tt.wantErrIs != nil {
				assert.ErrorIs(t, err, tt.wantErrIs)
			} else {
				assert.False(t, errors.Is(err, ErrDecodePublicKey), "DER 解析错误不应被折叠成 ErrDecodePublicKey。")
				assert.False(t, errors.Is(err, ErrDecodePrivateKey), "DER 解析错误不应被折叠成 ErrDecodePrivateKey。")
			}
		})
	}
}

// TestBoundaryPKCS1v15PlaintextLengthContracts 验证 PKCS#1 v1.5 公开 API 的明文长度边界。
//
// 该测试覆盖公钥加密和私钥签名式加密在最大可编码明文长度与超长明文下的成功、失败和往返契约。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestBoundaryPKCS1v15PlaintextLengthContracts(t *testing.T) {
	privateKey, privateKeyBytes, publicKeyBytes := generateTestKeyPair(t, 2048)
	publicKey := &privateKey.PublicKey
	keyBytes := (publicKey.N.BitLen() + 7) / 8
	maxPlaintextLen := keyBytes - 11
	maxPlaintext := bytes.Repeat([]byte{0x41}, maxPlaintextLen)
	tooLongPlaintext := bytes.Repeat([]byte{0x42}, maxPlaintextLen+1)

	tests := []struct {
		name        string
		description string
		run         func() ([]byte, error)
		verify      func(t *testing.T, cipherText []byte)
		wantErr     bool
	}{
		{
			name:        "success/public-pem-max-length",
			description: "验证 PEM 公钥加密在 PKCS#1 v1.5 最大明文长度下成功并可由私钥解密。",
			run: func() ([]byte, error) {
				return EncryptPubKey(publicKeyBytes, maxPlaintext)
			},
			verify: func(t *testing.T, cipherText []byte) {
				plainText, err := DecryptPrivKey(privateKeyBytes, cipherText)
				require.NoError(t, err)
				assert.Equal(t, maxPlaintext, plainText)
			},
		},
		{
			name:        "success/public-struct-max-length",
			description: "验证结构体公钥加密在 PKCS#1 v1.5 最大明文长度下成功并返回等于模数字节数的密文。",
			run: func() ([]byte, error) {
				return EncryptPublicKey(publicKey, maxPlaintext)
			},
			verify: func(t *testing.T, cipherText []byte) {
				assert.Len(t, cipherText, keyBytes)
				plainText, err := DecryptPrivateKey(privateKey, cipherText)
				require.NoError(t, err)
				assert.Equal(t, maxPlaintext, plainText)
			},
		},
		{
			name:        "error/public-pem-too-long",
			description: "验证 PEM 公钥加密在明文超过 PKCS#1 v1.5 最大长度时返回错误且不产生密文。",
			run: func() ([]byte, error) {
				return EncryptPubKey(publicKeyBytes, tooLongPlaintext)
			},
			wantErr: true,
		},
		{
			name:        "error/public-struct-too-long",
			description: "验证结构体公钥加密在明文超过 PKCS#1 v1.5 最大长度时返回错误且不产生密文。",
			run: func() ([]byte, error) {
				return EncryptPublicKey(publicKey, tooLongPlaintext)
			},
			wantErr: true,
		},
		{
			name:        "success/private-pem-max-length",
			description: "验证 PEM 私钥签名式加密在 PKCS#1 v1.5 最大明文长度下成功并可由公钥解包。",
			run: func() ([]byte, error) {
				return EncryptPrivKey(privateKeyBytes, maxPlaintext)
			},
			verify: func(t *testing.T, signature []byte) {
				plainText, err := DecryptPubKey(publicKeyBytes, signature)
				require.NoError(t, err)
				assert.Equal(t, maxPlaintext, plainText)
			},
		},
		{
			name:        "success/private-struct-max-length",
			description: "验证结构体私钥签名式加密在 PKCS#1 v1.5 最大明文长度下成功并返回等于模数字节数的签名。",
			run: func() ([]byte, error) {
				return EncryptPrivateKey(privateKey, maxPlaintext)
			},
			verify: func(t *testing.T, signature []byte) {
				assert.Len(t, signature, keyBytes)
				plainText, err := DecryptPublicKey(publicKey, signature)
				require.NoError(t, err)
				assert.Equal(t, maxPlaintext, plainText)
			},
		},
		{
			name:        "error/private-pem-too-long",
			description: "验证 PEM 私钥签名式加密在明文超过 PKCS#1 v1.5 最大长度时返回错误且不产生签名。",
			run: func() ([]byte, error) {
				return EncryptPrivKey(privateKeyBytes, tooLongPlaintext)
			},
			wantErr: true,
		},
		{
			name:        "error/private-struct-too-long",
			description: "验证结构体私钥签名式加密在明文超过 PKCS#1 v1.5 最大长度时返回错误且不产生签名。",
			run: func() ([]byte, error) {
				return EncryptPrivateKey(privateKey, tooLongPlaintext)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := tt.run()

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, got)
			tt.verify(t, got)
		})
	}
}

// TestBoundaryOAEPExactPlaintextLength 验证 OAEP 公开 API 在最大明文长度上的成功契约。
//
// 该测试覆盖 PEM 和结构体 OAEP API 在 SHA-256 最大明文长度下成功加密，并补充超长边界的明确错误断言。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestBoundaryOAEPExactPlaintextLength(t *testing.T) {
	privateKey, privateKeyBytes, publicKeyBytes := generateTestKeyPair(t, 2048)
	publicKey := &privateKey.PublicKey
	keyBytes := (publicKey.N.BitLen() + 7) / 8
	maxPlaintextLen := keyBytes - 2*sha256.Size - 2
	maxPlaintext := bytes.Repeat([]byte{0x61}, maxPlaintextLen)
	tooLongPlaintext := bytes.Repeat([]byte{0x62}, maxPlaintextLen+1)

	tests := []struct {
		name        string
		description string
		run         func() ([]byte, error)
		verify      func(t *testing.T, cipherText []byte)
		wantErr     bool
	}{
		{
			name:        "success/pem-max-length",
			description: "验证 PEM OAEP 默认 API 在 SHA-256 最大明文长度下成功并可解密。",
			run: func() ([]byte, error) {
				return EncryptPubKeyOAEP(publicKeyBytes, maxPlaintext)
			},
			verify: func(t *testing.T, cipherText []byte) {
				plainText, err := DecryptPrivKeyOAEP(privateKeyBytes, cipherText)
				require.NoError(t, err)
				assert.Equal(t, maxPlaintext, plainText)
			},
		},
		{
			name:        "success/struct-max-length",
			description: "验证结构体 OAEP 默认 API 在 SHA-256 最大明文长度下成功并返回等于模数字节数的密文。",
			run: func() ([]byte, error) {
				return EncryptPublicKeyOAEP(publicKey, maxPlaintext)
			},
			verify: func(t *testing.T, cipherText []byte) {
				assert.Len(t, cipherText, keyBytes)
				plainText, err := DecryptPrivateKeyOAEP(privateKey, cipherText)
				require.NoError(t, err)
				assert.Equal(t, maxPlaintext, plainText)
			},
		},
		{
			name:        "error/pem-too-long",
			description: "验证 PEM OAEP 默认 API 在明文超过 SHA-256 最大长度时返回错误且不产生密文。",
			run: func() ([]byte, error) {
				return EncryptPubKeyOAEP(publicKeyBytes, tooLongPlaintext)
			},
			wantErr: true,
		},
		{
			name:        "error/struct-too-long",
			description: "验证结构体 OAEP 默认 API 在明文超过 SHA-256 最大长度时返回错误且不产生密文。",
			run: func() ([]byte, error) {
				return EncryptPublicKeyOAEP(publicKey, tooLongPlaintext)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := tt.run()

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, got)
			tt.verify(t, got)
		})
	}
}

// TestBoundaryPKCS1v15SignaturePaddingContract 验证 PKCS#1 v1.5 签名填充块的结构断言。
//
// 该测试使用标准库签名构造确定性填充块，断言 0x00 0x01、0xff 填充、分隔符和 SHA-256 DigestInfo 均符合预期。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestBoundaryPKCS1v15SignaturePaddingContract(t *testing.T) {
	privateKey, _, _ := generateTestKeyPair(t, 2048)
	publicKey := &privateKey.PublicKey
	messageDigest := sha256.Sum256([]byte("signature padding contract"))

	signature, err := stdrsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, messageDigest[:])
	require.NoError(t, err)
	keyBytes := (publicKey.N.BitLen() + 7) / 8
	require.Len(t, signature, keyBytes)

	signatureInt := new(big.Int).SetBytes(signature)
	encodedInt := encrypt(new(big.Int), publicKey, signatureInt)
	encodedBlock := leftPad(encodedInt.Bytes(), keyBytes)
	require.Len(t, encodedBlock, keyBytes)
	separatorIndex := bytes.IndexByte(encodedBlock[2:], 0x00)
	require.GreaterOrEqual(t, separatorIndex, 8, "PKCS#1 v1.5 签名填充至少需要 8 字节 0xff。")
	separatorIndex += 2
	wantDigestInfo := append(append([]byte(nil), hashPrefixes[crypto.SHA256]...), messageDigest[:]...)

	assert.Equal(t, byte(0x00), encodedBlock[0], "签名编码块第一个字节应为 0x00。")
	assert.Equal(t, byte(0x01), encodedBlock[1], "签名编码块第二个字节应为 0x01。")
	assert.Equal(t, bytes.Repeat([]byte{0xff}, separatorIndex-2), encodedBlock[2:separatorIndex], "签名填充区应全部为 0xff。")
	assert.Equal(t, byte(0x00), encodedBlock[separatorIndex], "签名填充区后应使用 0x00 分隔 DigestInfo。")
	assert.Equal(t, wantDigestInfo, encodedBlock[separatorIndex+1:], "签名 DigestInfo 应包含 SHA-256 ASN.1 前缀和消息摘要。")

	recoveredDigestInfo, err := publicDecrypt(publicKey, crypto.SHA256, messageDigest[:], signature)
	require.NoError(t, err)
	assert.Equal(t, wantDigestInfo, recoveredDigestInfo, "publicDecrypt 应返回完整 DigestInfo 内容。")
}

// TestBoundaryInvalidSignatureDoesNotRecoverMessage 验证无效签名不会被当前公钥解包流程当作原始消息接受。
//
// 该测试使用全零签名覆盖确定性的无效签名路径，断言结构体和 PEM wrapper 均不会恢复出原始消息内容。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestBoundaryInvalidSignatureDoesNotRecoverMessage(t *testing.T) {
	privateKey, _, publicKeyBytes := generateTestKeyPair(t, 2048)
	publicKey := &privateKey.PublicKey
	message := []byte("valid signature message")
	keyBytes := (publicKey.N.BitLen() + 7) / 8
	invalidSignature := make([]byte, keyBytes)

	tests := []struct {
		name        string
		description string
		run         func() ([]byte, error)
	}{
		{
			name:        "error/struct-invalid-signature-does-not-recover-message",
			description: "验证结构体公钥解包全零无效签名时不会返回原始消息内容。",
			run: func() ([]byte, error) {
				return DecryptPublicKey(publicKey, invalidSignature)
			},
		},
		{
			name:        "error/pem-invalid-signature-does-not-recover-message",
			description: "验证 PEM 公钥解包全零无效签名时不会返回原始消息内容。",
			run: func() ([]byte, error) {
				return DecryptPubKey(publicKeyBytes, invalidSignature)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := tt.run()

			require.NoError(t, err)
			assert.NotEqual(t, message, got)
			assert.NotEmpty(t, got)
		})
	}
}
