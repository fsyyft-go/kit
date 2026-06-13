package rsa

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	stdrsa "crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"hash"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildAdditionalECDSAPublicKeyPEM 构造可解析但不是 RSA 算法的 PKIX 公钥 PEM。
//
// 该辅助函数用于覆盖 convertPublicKey 中“PKIX 解析成功但类型断言失败”的错误分支。
//
// 参数：
//   - t: 测试上下文，用于报告夹具构造失败并标记辅助函数调用栈。
//
// 返回：
//   - []byte: PEM 编码的 ECDSA 公钥。
func buildAdditionalECDSAPublicKeyPEM(t *testing.T) []byte {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err, "生成 ECDSA 测试密钥应该成功。")

	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err, "编码 ECDSA 公钥应该成功。")

	return pem.EncodeToMemory(&pem.Block{
		Type:  BlockTypePublicKey,
		Bytes: publicKeyDER,
	})
}

// TestAdditionalKeyParsingPEMBranches 验证 PEM 密钥转换函数的关键成功与错误分支。
//
// 该测试通过表驱动用例覆盖有效 PEM、错误 block type、非法 PEM、非法 DER 和非 RSA 公钥，确保密钥解析契约稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestAdditionalKeyParsingPEMBranches(t *testing.T) {
	privateKey, privateKeyBytes, publicKeyBytes := generateTestKeyPair(t, 2048)
	ecdsaPublicKeyPEM := buildAdditionalECDSAPublicKeyPEM(t)

	privateKeyCases := []struct {
		name        string
		description string
		givePEM     []byte
		wantErr     bool
		wantErrIs   error
	}{
		{
			name:        "success/valid-pkcs1-private-key",
			description: "验证 PKCS#1 RSA 私钥 PEM 能解析为原始 RSA 私钥对象。",
			givePEM:     privateKeyBytes,
		},
		{
			name:        "error/invalid-private-pem",
			description: "验证非 PEM 私钥输入返回 ErrDecodePrivateKey。",
			givePEM:     []byte("invalid private key"),
			wantErr:     true,
			wantErrIs:   ErrDecodePrivateKey,
		},
		{
			name:        "error/private-block-type-mismatch",
			description: "验证私钥 PEM block type 不是 RSA PRIVATE KEY 时返回 ErrDecodePrivateKey。",
			givePEM: pem.EncodeToMemory(&pem.Block{
				Type:  BlockTypePublicKey,
				Bytes: []byte("not a private key"),
			}),
			wantErr:   true,
			wantErrIs: ErrDecodePrivateKey,
		},
		{
			name:        "error/private-der-parse-failure",
			description: "验证私钥 PEM 类型正确但 DER 内容非法时返回底层解析错误。",
			givePEM: pem.EncodeToMemory(&pem.Block{
				Type:  BlockTypePrivateKey,
				Bytes: []byte("not valid DER private key"),
			}),
			wantErr: true,
		},
	}

	for _, tt := range privateKeyCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			gotKey, err := ConvertPrivateKey(tt.givePEM)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, gotKey, "解析失败时不应返回私钥对象。")
				if tt.wantErrIs != nil {
					assert.ErrorIs(t, err, tt.wantErrIs, "错误类型应保持稳定。")
				} else {
					assert.False(t, errors.Is(err, ErrDecodePrivateKey), "DER 解析错误不应被折叠成 ErrDecodePrivateKey。")
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, gotKey, "有效私钥 PEM 应返回私钥对象。")
			assert.Equal(t, privateKey.N, gotKey.N, "解析后的私钥应保留 RSA 模数。")
			assert.Equal(t, privateKey.E, gotKey.E, "解析后的私钥应保留 RSA 指数。")
		})
	}

	publicKeyCases := []struct {
		name        string
		description string
		givePEM     []byte
		wantErr     bool
		wantErrIs   error
	}{
		{
			name:        "success/valid-pkix-public-key",
			description: "验证 PKIX RSA 公钥 PEM 能解析为原始 RSA 公钥对象。",
			givePEM:     publicKeyBytes,
		},
		{
			name:        "error/invalid-public-pem",
			description: "验证非 PEM 公钥输入返回 ErrDecodePublicKey。",
			givePEM:     []byte("invalid public key"),
			wantErr:     true,
			wantErrIs:   ErrDecodePublicKey,
		},
		{
			name:        "error/public-block-type-mismatch",
			description: "验证公钥 PEM block type 不是 PUBLIC KEY 时返回 ErrDecodePublicKey。",
			givePEM: pem.EncodeToMemory(&pem.Block{
				Type:  BlockTypePrivateKey,
				Bytes: []byte("not a public key"),
			}),
			wantErr:   true,
			wantErrIs: ErrDecodePublicKey,
		},
		{
			name:        "error/public-der-parse-failure",
			description: "验证公钥 PEM 类型正确但 DER 内容非法时返回底层解析错误。",
			givePEM: pem.EncodeToMemory(&pem.Block{
				Type:  BlockTypePublicKey,
				Bytes: []byte("not valid DER public key"),
			}),
			wantErr: true,
		},
		{
			name:        "error/non-rsa-public-key",
			description: "验证可解析但算法不是 RSA 的公钥返回 ErrDecodePublicKey。",
			givePEM:     ecdsaPublicKeyPEM,
			wantErr:     true,
			wantErrIs:   ErrDecodePublicKey,
		},
	}

	for _, tt := range publicKeyCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			gotKey, err := convertPublicKey(tt.givePEM)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, gotKey, "解析失败时不应返回公钥对象。")
				if tt.wantErrIs != nil {
					assert.ErrorIs(t, err, tt.wantErrIs, "错误类型应保持稳定。")
				} else {
					assert.False(t, errors.Is(err, ErrDecodePublicKey), "DER 解析错误不应被折叠成 ErrDecodePublicKey。")
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, gotKey, "有效公钥 PEM 应返回公钥对象。")
			assert.Equal(t, privateKey.N, gotKey.N, "解析后的公钥应保留 RSA 模数。")
			assert.Equal(t, privateKey.E, gotKey.E, "解析后的公钥应保留 RSA 指数。")
		})
	}
}

// TestAdditionalConvertPubKeyRoundTrip 验证 RSA 公钥导出 PEM 后可被重新解析。
//
// 该测试覆盖 ConvertPubKey 的有效转换路径和无效 RSA 公钥结构的错误路径，确保导出的 PEM 与原始公钥一致。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestAdditionalConvertPubKeyRoundTrip(t *testing.T) {
	privateKey, _, _ := generateTestKeyPair(t, 2048)

	testCases := []struct {
		name        string
		description string
		giveKey     *stdrsa.PublicKey
		wantErr     bool
	}{
		{
			name:        "success/rsa-public-key-round-trip",
			description: "验证有效 RSA 公钥转换为 PEM 后仍可解析为相同公钥。",
			giveKey:     &privateKey.PublicKey,
		},
		{
			name:        "error/nil-modulus-public-key",
			description: "验证缺少模数的 RSA 公钥结构无法编码为 PKIX PEM。",
			giveKey:     &stdrsa.PublicKey{N: nil, E: 65537},
			wantErr:     true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			pemBytes, err := ConvertPubKey(tt.giveKey)

			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, pemBytes, "编码失败时不应返回 PEM 内容。")
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, pemBytes, "有效 RSA 公钥应编码为 PEM。")
			parsedKey, err := convertPublicKey(pemBytes)
			require.NoError(t, err)
			assert.Equal(t, tt.giveKey.N, parsedKey.N, "PEM round-trip 应保留 RSA 模数。")
			assert.Equal(t, tt.giveKey.E, parsedKey.E, "PEM round-trip 应保留 RSA 指数。")
		})
	}
}

// TestAdditionalLegacyPKCS1v15RoundTrips 验证保留的 PKCS#1 v1.5 兼容 API 可完成往返。
//
// 该测试覆盖 PEM wrapper 和结构体 API 的公钥加密/私钥解密，以及历史私钥签名式加密/公钥解密场景。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestAdditionalLegacyPKCS1v15RoundTrips(t *testing.T) {
	privateKey, privateKeyBytes, publicKeyBytes := generateTestKeyPair(t, 2048)
	publicKey := &privateKey.PublicKey

	testCases := []struct {
		name        string
		description string
		givePlain   []byte
	}{
		{
			name:        "success/text-message",
			description: "验证文本明文可通过 legacy PKCS#1 v1.5 API 完成双向兼容往返。",
			givePlain:   []byte("legacy pkcs1v15 message"),
		},
		{
			name:        "success/empty-message",
			description: "验证空明文在 legacy PKCS#1 v1.5 API 中保持为空值语义。",
			givePlain:   []byte{},
		},
		{
			name:        "success/binary-message",
			description: "验证包含零值和高位字节的二进制明文可完成 legacy 往返。",
			givePlain:   []byte{0x00, 0x01, 0x7f, 0x80, 0xff},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			pemCipherText, err := EncryptPubKey(publicKeyBytes, tt.givePlain)
			require.NoError(t, err, "PEM 公钥加密应该成功。")
			require.NotEmpty(t, pemCipherText, "PKCS#1 v1.5 密文不应为空。")
			pemPlainText, err := DecryptPrivKey(privateKeyBytes, pemCipherText)
			require.NoError(t, err, "PEM 私钥解密应该成功。")
			assert.Equal(t, tt.givePlain, pemPlainText, "PEM 加解密结果应等于原始明文。")

			structCipherText, err := EncryptPublicKey(publicKey, tt.givePlain)
			require.NoError(t, err, "结构体公钥加密应该成功。")
			require.NotEmpty(t, structCipherText, "结构体 API 密文不应为空。")
			structPlainText, err := DecryptPrivateKey(privateKey, structCipherText)
			require.NoError(t, err, "结构体私钥解密应该成功。")
			assert.Equal(t, tt.givePlain, structPlainText, "结构体 API 加解密结果应等于原始明文。")

			pemSignature, err := EncryptPrivKey(privateKeyBytes, tt.givePlain)
			require.NoError(t, err, "PEM 私钥签名式加密应该成功。")
			require.NotEmpty(t, pemSignature, "签名式密文不应为空。")
			pemRecovered, err := DecryptPubKey(publicKeyBytes, pemSignature)
			require.NoError(t, err, "PEM 公钥解密签名式密文应该成功。")
			assert.Equal(t, tt.givePlain, pemRecovered, "PEM 签名式往返结果应等于原始明文。")

			structSignature, err := EncryptPrivateKey(privateKey, tt.givePlain)
			require.NoError(t, err, "结构体私钥签名式加密应该成功。")
			require.NotEmpty(t, structSignature, "结构体签名式密文不应为空。")
			structRecovered, err := DecryptPublicKey(publicKey, structSignature)
			require.NoError(t, err, "结构体公钥解密签名式密文应该成功。")
			assert.Equal(t, tt.givePlain, structRecovered, "结构体签名式往返结果应等于原始明文。")
		})
	}
}

// TestAdditionalOAEPRoundTripAndErrors 验证 OAEP 默认参数、可配置参数和错误路径。
//
// 该测试覆盖默认 SHA-256/nil label、WithHash+label、nil hash、label/hash 不匹配、非法 key 和非法 ciphertext 等行为。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestAdditionalOAEPRoundTripAndErrors(t *testing.T) {
	privateKey, privateKeyBytes, publicKeyBytes := generateTestKeyPair(t, 2048)
	publicKey := &privateKey.PublicKey
	plainText := []byte("oaep additional behavior")
	label := []byte("kit:rsa:oaep:additional")

	roundTripCases := []struct {
		name        string
		description string
		encrypt     func() ([]byte, error)
		decrypt     func([]byte) ([]byte, error)
	}{
		{
			name:        "success/default-sha256-pem",
			description: "验证 PEM OAEP 默认 API 使用 SHA-256 和 nil label 完成往返。",
			encrypt: func() ([]byte, error) {
				return EncryptPubKeyOAEP(publicKeyBytes, plainText)
			},
			decrypt: func(cipherText []byte) ([]byte, error) {
				return DecryptPrivKeyOAEP(privateKeyBytes, cipherText)
			},
		},
		{
			name:        "success/default-sha256-struct",
			description: "验证结构体 OAEP 默认 API 使用 SHA-256 和 nil label 完成往返。",
			encrypt: func() ([]byte, error) {
				return EncryptPublicKeyOAEP(publicKey, plainText)
			},
			decrypt: func(cipherText []byte) ([]byte, error) {
				return DecryptPrivateKeyOAEP(privateKey, cipherText)
			},
		},
		{
			name:        "success/sha256-label-pem",
			description: "验证 PEM OAEP WithHash API 在 SHA-256 和相同 label 下完成往返。",
			encrypt: func() ([]byte, error) {
				return EncryptPubKeyOAEPWithHash(publicKeyBytes, plainText, sha256.New(), label)
			},
			decrypt: func(cipherText []byte) ([]byte, error) {
				return DecryptPrivKeyOAEPWithHash(privateKeyBytes, cipherText, sha256.New(), label)
			},
		},
		{
			name:        "success/sha1-label-struct",
			description: "验证结构体 OAEP WithHash API 在 SHA-1 和相同 label 下完成往返。",
			encrypt: func() ([]byte, error) {
				return EncryptPublicKeyOAEPWithHash(publicKey, plainText, sha1.New(), label)
			},
			decrypt: func(cipherText []byte) ([]byte, error) {
				return DecryptPrivateKeyOAEPWithHash(privateKey, cipherText, sha1.New(), label)
			},
		},
	}

	for _, tt := range roundTripCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			cipherText, err := tt.encrypt()
			require.NoError(t, err)
			require.NotEmpty(t, cipherText, "OAEP 密文不应为空。")

			gotPlainText, err := tt.decrypt(cipherText)
			require.NoError(t, err)
			assert.Equal(t, plainText, gotPlainText, "OAEP 解密结果应等于原始明文。")
		})
	}

	nilHashCases := []struct {
		name        string
		description string
		run         func() error
	}{
		{
			name:        "error/nil-hash-encrypt-pem",
			description: "验证 PEM OAEP 加密在 hash 为 nil 时返回 ErrNilHash。",
			run: func() error {
				_, err := EncryptPubKeyOAEPWithHash(publicKeyBytes, plainText, nil, nil)
				return err
			},
		},
		{
			name:        "error/nil-hash-encrypt-struct",
			description: "验证结构体 OAEP 加密在 hash 为 nil 时返回 ErrNilHash。",
			run: func() error {
				_, err := EncryptPublicKeyOAEPWithHash(publicKey, plainText, nil, nil)
				return err
			},
		},
		{
			name:        "error/nil-hash-decrypt-pem",
			description: "验证 PEM OAEP 解密在 hash 为 nil 时返回 ErrNilHash。",
			run: func() error {
				_, err := DecryptPrivKeyOAEPWithHash(privateKeyBytes, []byte("cipher"), nil, nil)
				return err
			},
		},
		{
			name:        "error/nil-hash-decrypt-struct",
			description: "验证结构体 OAEP 解密在 hash 为 nil 时返回 ErrNilHash。",
			run: func() error {
				_, err := DecryptPrivateKeyOAEPWithHash(privateKey, []byte("cipher"), nil, nil)
				return err
			},
		},
	}

	for _, tt := range nilHashCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			err := tt.run()
			assert.ErrorIs(t, err, ErrNilHash, "nil hash 应返回 ErrNilHash。")
		})
	}

	t.Run("error/label-mismatch", func(t *testing.T) {
		cipherText, err := EncryptPublicKeyOAEPWithHash(publicKey, plainText, sha256.New(), []byte("label-a"))
		require.NoError(t, err, "准备带 label 的 OAEP 密文应该成功。")

		gotPlainText, err := DecryptPrivateKeyOAEPWithHash(privateKey, cipherText, sha256.New(), []byte("label-b"))

		require.Error(t, err, "label 不一致应该导致 OAEP 解密失败。")
		assert.Nil(t, gotPlainText, "解密失败时不应返回明文。")
	})

	t.Run("error/hash-mismatch", func(t *testing.T) {
		cipherText, err := EncryptPublicKeyOAEPWithHash(publicKey, plainText, sha256.New(), nil)
		require.NoError(t, err, "准备 SHA-256 OAEP 密文应该成功。")

		gotPlainText, err := DecryptPrivateKeyOAEPWithHash(privateKey, cipherText, sha1.New(), nil)

		require.Error(t, err, "hash 不一致应该导致 OAEP 解密失败。")
		assert.Nil(t, gotPlainText, "解密失败时不应返回明文。")
	})

	t.Run("error/invalid-public-key", func(t *testing.T) {
		cipherText, err := EncryptPubKeyOAEP([]byte("invalid public key"), plainText)

		require.Error(t, err, "无效 PEM 公钥应该导致 OAEP 加密失败。")
		assert.Nil(t, cipherText, "加密失败时不应返回密文。")
		assert.ErrorIs(t, err, ErrDecodePublicKey, "无效公钥错误类型应保持稳定。")
	})

	t.Run("error/invalid-private-key", func(t *testing.T) {
		cipherText, err := EncryptPubKeyOAEP(publicKeyBytes, plainText)
		require.NoError(t, err, "准备 OAEP 密文应该成功。")

		gotPlainText, err := DecryptPrivKeyOAEP([]byte("invalid private key"), cipherText)

		require.Error(t, err, "无效 PEM 私钥应该导致 OAEP 解密失败。")
		assert.Nil(t, gotPlainText, "解密失败时不应返回明文。")
		assert.ErrorIs(t, err, ErrDecodePrivateKey, "无效私钥错误类型应保持稳定。")
	})

	t.Run("error/invalid-ciphertext", func(t *testing.T) {
		invalidCipherText := []byte("invalid OAEP ciphertext")

		gotPlainText, err := DecryptPrivateKeyOAEP(privateKey, invalidCipherText)
		require.Error(t, err, "结构体 API 解密非法 OAEP 密文应该失败。")
		assert.Nil(t, gotPlainText, "结构体 API 解密失败时不应返回明文。")

		gotPlainText, err = DecryptPrivKeyOAEP(privateKeyBytes, invalidCipherText)
		require.Error(t, err, "PEM API 解密非法 OAEP 密文应该失败。")
		assert.Nil(t, gotPlainText, "PEM API 解密失败时不应返回明文。")
	})

	t.Run("error/too-long-message", func(t *testing.T) {
		keyBytes := (publicKey.N.BitLen() + 7) / 8
		maxPlaintextLen := keyBytes - 2*sha256.Size - 2
		tooLongPlainText := make([]byte, maxPlaintextLen+1)

		cipherText, err := EncryptPublicKeyOAEP(publicKey, tooLongPlainText)

		require.Error(t, err, "超过 SHA-256 OAEP 最大明文长度的输入应该加密失败。")
		assert.Nil(t, cipherText, "加密失败时不应返回密文。")
	})
}

// TestAdditionalPublicDecryptAndHelperBoundaries 验证 publicDecrypt 与内部 helper 的边界行为。
//
// 该测试覆盖 publicDecrypt 的密钥长度错误、无哈希签名解包、pkcs1v15HashInfo 错误分支，以及 leftPad/unLeftPad 的边界语义。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestAdditionalPublicDecryptAndHelperBoundaries(t *testing.T) {
	privateKey, _, _ := generateTestKeyPair(t, 2048)
	plainText := []byte("public decrypt helper boundary")

	t.Run("error/public-decrypt-key-too-short", func(t *testing.T) {
		shortPublicKey := &stdrsa.PublicKey{N: big.NewInt(3), E: 3}

		gotPlainText, err := publicDecrypt(shortPublicKey, crypto.Hash(0), plainText, []byte{0x01})

		require.Error(t, err, "过小 RSA 公钥无法容纳 PKCS#1 v1.5 填充块时应返回错误。")
		assert.Nil(t, gotPlainText, "错误分支不应返回解密结果。")
		assert.Contains(t, err.Error(), "length illegal", "错误信息应表明长度不合法。")
	})

	t.Run("success/public-decrypt-no-hash-signature", func(t *testing.T) {
		signature, err := stdrsa.SignPKCS1v15(nil, privateKey, crypto.Hash(0), plainText)
		require.NoError(t, err, "准备无哈希 PKCS#1 v1.5 签名应该成功。")

		gotPlainText, err := publicDecrypt(&privateKey.PublicKey, crypto.Hash(0), nil, signature)

		require.NoError(t, err, "无哈希模式应能解包标准库生成的签名。")
		assert.Equal(t, plainText, gotPlainText, "解包结果应等于原始无哈希消息。")
	})

	hashInfoCases := []struct {
		name        string
		description string
		giveHash    crypto.Hash
		giveLen     int
		wantLen     int
		wantPrefix  []byte
		wantErr     bool
	}{
		{
			name:        "success/no-hash",
			description: "验证 hash 为 0 时 pkcs1v15HashInfo 使用输入长度且不返回 ASN.1 前缀。",
			giveHash:    crypto.Hash(0),
			giveLen:     len(plainText),
			wantLen:     len(plainText),
		},
		{
			name:        "success/sha256-prefix",
			description: "验证 SHA-256 返回标准哈希长度和预定义 ASN.1 DER 前缀。",
			giveHash:    crypto.SHA256,
			giveLen:     crypto.SHA256.Size(),
			wantLen:     crypto.SHA256.Size(),
			wantPrefix:  hashPrefixes[crypto.SHA256],
		},
		{
			name:        "error/hash-length-mismatch",
			description: "验证哈希输入长度不匹配时返回已哈希消息错误。",
			giveHash:    crypto.SHA256,
			giveLen:     crypto.SHA256.Size() - 1,
			wantErr:     true,
		},
		{
			name:        "error/unsupported-hash",
			description: "验证未注册 ASN.1 前缀的哈希算法返回 unsupported hash 错误。",
			giveHash:    crypto.BLAKE2b_256,
			giveLen:     crypto.BLAKE2b_256.Size(),
			wantErr:     true,
		},
	}

	for _, tt := range hashInfoCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			gotLen, gotPrefix, err := pkcs1v15HashInfo(tt.giveHash, tt.giveLen)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Zero(t, gotLen, "错误分支不应返回哈希长度。")
				assert.Nil(t, gotPrefix, "错误分支不应返回前缀。")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantLen, gotLen, "哈希长度应符合算法定义。")
			assert.Equal(t, tt.wantPrefix, gotPrefix, "哈希前缀应符合预定义映射。")
		})
	}

	paddingCases := []struct {
		name        string
		description string
		run         func(t *testing.T)
	}{
		{
			name:        "boundary/left-pad-prefixes-zero-bytes",
			description: "验证 leftPad 在目标长度大于输入长度时向左补零并保留输入后缀。",
			run: func(t *testing.T) {
				got := leftPad([]byte{0x01, 0x02}, 5)
				assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x01, 0x02}, got, "leftPad 应在左侧补零。")
			},
		},
		{
			name:        "boundary/left-pad-truncates-to-leading-bytes",
			description: "验证 leftPad 在输入长度超过目标长度时按当前实现保留输入前缀字节。",
			run: func(t *testing.T) {
				got := leftPad([]byte{0x01, 0x02, 0x03, 0x04}, 2)
				assert.Equal(t, []byte{0x01, 0x02}, got, "leftPad 截断时应保留当前实现定义的前缀字节。")
			},
		},
		{
			name:        "success/un-left-pad-ff-separator",
			description: "验证 unLeftPad 能从 0x00 0x01 0xff...0x00 结构中提取原始数据。",
			run: func(t *testing.T) {
				got := unLeftPad([]byte{0x00, 0x01, 0xff, 0xff, 0x00, 0x61, 0x62})
				assert.Equal(t, []byte("ab"), got, "标准 PKCS#1 v1.5 风格填充应还原 payload。")
			},
		},
		{
			name:        "success/un-left-pad-length-byte-separator",
			description: "验证 unLeftPad 兼容第二字节参与累计偏移的历史分隔模式。",
			run: func(t *testing.T) {
				got := unLeftPad([]byte{0x42, 0x01, 0x42, 0x61, 0x62})
				assert.Equal(t, []byte("ab"), got, "历史长度字节分隔模式应还原 payload。")
			},
		},
	}

	for _, tt := range paddingCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			tt.run(t)
		})
	}
}

// TestAdditionalOAEPNilHashPrecedence 验证 nil hash 在结构体 API 中优先于底层 RSA 操作返回。
//
// 该测试使用自定义 hash 工厂字段表达 nil hash 场景，避免把 nil hash 错误路径与密钥或密文错误混淆。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestAdditionalOAEPNilHashPrecedence(t *testing.T) {
	privateKey, _, _ := generateTestKeyPair(t, 2048)
	publicKey := &privateKey.PublicKey

	testCases := []struct {
		name        string
		description string
		giveHash    func() hash.Hash
		run         func(hash.Hash) error
	}{
		{
			name:        "error/encrypt-nil-hash-with-valid-key",
			description: "验证有效公钥加密时 nil hash 直接返回 ErrNilHash。",
			giveHash:    func() hash.Hash { return nil },
			run: func(h hash.Hash) error {
				_, err := EncryptPublicKeyOAEPWithHash(publicKey, []byte("plain"), h, nil)
				return err
			},
		},
		{
			name:        "error/decrypt-nil-hash-with-valid-key",
			description: "验证有效私钥解密时 nil hash 直接返回 ErrNilHash。",
			giveHash:    func() hash.Hash { return nil },
			run: func(h hash.Hash) error {
				_, err := DecryptPrivateKeyOAEPWithHash(privateKey, []byte("cipher"), h, nil)
				return err
			},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			err := tt.run(tt.giveHash())
			assert.ErrorIs(t, err, ErrNilHash, "nil hash 应稳定返回 ErrNilHash。")
		})
	}
}
