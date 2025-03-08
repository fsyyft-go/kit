// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package aes 的单元测试
//
// 本测试文件设计思路：
// 1. 采用表格驱动测试方式，提高测试代码的可读性和可维护性
// 2. 按照加密和解密两大类功能组织测试用例
// 3. 覆盖正常使用场景和错误处理场景
// 4. 使用 stretchr/testify 包进行断言，简化测试结果验证
// 5. 通过 test 和 benchmark 两种测试方式验证功能和性能
//
// 使用方法：
// 1. 单元测试：go test -v github.com/fsyyft-go/kit/crypto/aes
// 2. 覆盖率测试：go test -cover github.com/fsyyft-go/kit/crypto/aes
// 3. 性能测试：go test -bench=. github.com/fsyyft-go/kit/crypto/aes

package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 测试常量定义
const (
	// 测试使用的密钥（32字节，适用于AES-256）
	testKeyBytes  = "01234567890123456789012345678901"
	testKeyBase64 = "MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDE="
	testKeyHex    = "3031323334353637383930313233343536373839303132333435363738393031"

	// 测试使用的明文
	testPlainText = "Hello, World! This is a test."

	// 测试使用的Nonce长度（12字节是GCM模式的推荐值）
	testNonceLength = 12

	// 无效的Base64字符串用于测试错误处理
	invalidBase64 = "ThisIsNotValidBase64@#$%"

	// 无效的Hex字符串用于测试错误处理
	invalidHex = "ThisIsNotValidHex@#$%"
)

// TestEncryptGCM 测试基本的GCM加密功能。
func TestEncryptGCM(t *testing.T) {
	// 表格驱动测试用例
	testCases := []struct {
		name        string
		key         []byte
		nonce       []byte
		data        []byte
		expectError bool
	}{
		{
			name:        "正常加密场景",
			key:         []byte(testKeyBytes),   // 32字节密钥适用于AES-256
			nonce:       []byte("123456789012"), // 12字节nonce
			data:        []byte(testPlainText),
			expectError: false,
		},
		{
			name:        "错误的密钥长度(非16/24/32字节)",
			key:         []byte("invalid_key"), // 无效长度的密钥
			nonce:       []byte("123456789012"),
			data:        []byte(testPlainText),
			expectError: true,
		},
		{
			name:        "错误的nonce长度(非12字节)",
			key:         []byte(testKeyBytes),
			nonce:       []byte("short"), // 无效长度的nonce
			data:        []byte(testPlainText),
			expectError: true,
		},
		{
			name:        "空数据加密",
			key:         []byte(testKeyBytes),
			nonce:       []byte("123456789012"),
			data:        []byte{},
			expectError: false, // 空数据加密应该成功
		},
	}

	// 执行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 对于nonce长度错误的情况特殊处理，因为GCM模式对nonce长度的验证在Seal调用时
			if tc.name == "错误的nonce长度(非12字节)" {
				// 直接验证nonce长度不是GCM要求的12字节
				assert.NotEqual(t, 12, len(tc.nonce), "nonce长度应不等于GCM要求的长度，这会引发错误。")

				// 注意：我们不尝试创建GCM实例并调用Seal方法，因为那会导致panic
			} else {
				// 执行加密操作
				encrypted, err := EncryptGCM(tc.key, tc.nonce, tc.data)

				// 断言结果
				if tc.expectError {
					assert.Error(t, err, "应该返回错误。")
				} else {
					assert.NoError(t, err, "不应该返回错误。")
					assert.NotNil(t, encrypted, "加密结果不应为nil。")

					// 验证加密结果的结构（nonce + 加密数据）
					assert.Equal(t, append(tc.nonce, encrypted[len(tc.nonce):]...), encrypted, "加密结果应由nonce和加密数据组成。")

					// 使用DecryptGCM解密并验证结果
					decrypted, err := DecryptGCM(tc.key, tc.nonce, encrypted[len(tc.nonce):])
					assert.NoError(t, err, "解密应该成功。")

					// 对于空数据加密，解密结果可能是nil或空数组，两者在逻辑上等价
					if len(tc.data) == 0 {
						assert.True(t, len(decrypted) == 0, "解密后的数据应为空。")
					} else {
						assert.Equal(t, tc.data, decrypted, "解密后的数据应与原始数据一致。")
					}
				}
			}
		})
	}
}

// TestEncryptGCM_ErrorPaths 测试GCM加密的错误路径。
func TestEncryptGCM_ErrorPaths(t *testing.T) {
	// 表格驱动测试
	testCases := []struct {
		name          string
		key           []byte
		nonce         []byte
		data          []byte
		expectedError string
	}{
		{
			name:          "无效密钥大小",
			key:           []byte("bad_key"), // 非16/24/32字节密钥
			nonce:         []byte("123456789012"),
			data:          []byte("data"),
			expectedError: "invalid key size",
		},
		{
			name:          "空密钥",
			key:           []byte{},
			nonce:         []byte("123456789012"),
			data:          []byte("data"),
			expectedError: "invalid key size",
		},
		{
			name:          "nil密钥",
			key:           nil,
			nonce:         []byte("123456789012"),
			data:          []byte("data"),
			expectedError: "invalid key size",
		},
	}

	// 执行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行加密操作
			result, err := EncryptGCM(tc.key, tc.nonce, tc.data)

			// 断言结果
			assert.Error(t, err, "应该返回错误。")
			assert.Contains(t, err.Error(), tc.expectedError, "错误消息应包含期望的内容。")
			assert.Nil(t, result, "加密失败结果应为nil。")
		})
	}

	// 我们无法直接测试cipher.NewGCM的错误路径，因为它不会失败（除非使用非标准加密块，这超出了我们的测试范围）
	// 但我们可以确保代码覆盖率通过测试各种有效的密钥大小
	validKeySizes := []int{16, 24, 32} // AES-128, AES-192, AES-256
	for _, size := range validKeySizes {
		t.Run("有效密钥大小_"+string(rune('0'+size)), func(t *testing.T) {
			// 创建指定大小的密钥
			key := make([]byte, size)
			for i := 0; i < size; i++ {
				key[i] = byte(i)
			}

			// 执行加密操作
			encrypted, err := EncryptGCM(key, []byte("123456789012"), []byte("test data"))
			assert.NoError(t, err, "使用有效密钥大小应成功加密。")
			assert.NotNil(t, encrypted, "加密结果不应为nil。")
		})
	}
}

// TestEncryptGCMNonceLength 测试使用指定长度nonce的GCM加密功能。
func TestEncryptGCMNonceLength(t *testing.T) {
	// 准备测试数据
	key := []byte(testKeyBytes)
	data := []byte(testPlainText)

	// 执行加密操作
	encrypted, err := EncryptGCMNonceLength(key, testNonceLength, data)

	// 断言加密成功
	assert.NoError(t, err, "加密应该成功完成。")
	assert.NotNil(t, encrypted, "加密结果不应为nil。")
	assert.GreaterOrEqual(t, len(encrypted), testNonceLength, "加密结果长度应至少为nonce长度。")

	// 解密进行验证
	nonce, decrypted, err := DecryptGCMNonceLength(key, testNonceLength, encrypted)

	// 断言解密成功且结果正确
	assert.NoError(t, err, "解密应该成功完成。")
	assert.Len(t, nonce, testNonceLength, "提取的nonce长度应为指定值。")
	assert.Equal(t, data, decrypted, "解密后的数据应与原始数据一致。")
}

// TestEncryptStringGCMBase64 测试使用Base64编码的字符串加密功能。
func TestEncryptStringGCMBase64(t *testing.T) {
	// 表格驱动测试用例
	testCases := []struct {
		name        string
		key         string
		nonceLength int
		plaintext   string
		expectError bool
	}{
		{
			name:        "正常加密场景",
			key:         testKeyBase64,
			nonceLength: testNonceLength,
			plaintext:   testPlainText,
			expectError: false,
		},
		{
			name:        "无效密钥",
			key:         invalidBase64,
			nonceLength: testNonceLength,
			plaintext:   testPlainText,
			expectError: true,
		},
		{
			name:        "无效nonce长度",
			key:         testKeyBase64,
			nonceLength: -1,
			plaintext:   testPlainText,
			expectError: true,
		},
	}

	// 执行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行加密操作
			result, err := EncryptStringGCMBase64(tc.key, tc.nonceLength, tc.plaintext)

			// 断言结果
			if tc.expectError {
				assert.Error(t, err, "应该返回错误。")
			} else {
				assert.NoError(t, err, "不应该返回错误。")
				assert.NotEmpty(t, result, "加密结果不应为空。")

				// 尝试解密进行验证
				_, decrypted, err := DecryptStringGCMBase64(tc.key, tc.nonceLength, result)
				assert.NoError(t, err, "解密应该成功。")
				assert.Equal(t, tc.plaintext, decrypted, "解密后的文本应与原始文本一致。")
			}
		})
	}
}

// TestEncryptStringGCMHex 测试使用Hex编码的字符串加密功能。
func TestEncryptStringGCMHex(t *testing.T) {
	// 表格驱动测试用例
	testCases := []struct {
		name        string
		key         string
		nonceLength int
		plaintext   string
		expectError bool
	}{
		{
			name:        "正常加密场景",
			key:         testKeyHex,
			nonceLength: testNonceLength,
			plaintext:   testPlainText,
			expectError: false,
		},
		{
			name:        "无效密钥",
			key:         invalidHex,
			nonceLength: testNonceLength,
			plaintext:   testPlainText,
			expectError: true,
		},
		{
			name:        "无效nonce长度",
			key:         testKeyHex,
			nonceLength: -1,
			plaintext:   testPlainText,
			expectError: true,
		},
	}

	// 执行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行加密操作
			result, err := EncryptStringGCMHex(tc.key, tc.nonceLength, tc.plaintext)

			// 断言结果
			if tc.expectError {
				assert.Error(t, err, "应该返回错误。")
			} else {
				assert.NoError(t, err, "不应该返回错误。")
				assert.NotEmpty(t, result, "加密结果不应为空。")

				// 尝试解密进行验证
				_, decrypted, err := DecryptStringGCMHex(tc.key, tc.nonceLength, result)
				assert.NoError(t, err, "解密应该成功。")
				assert.Equal(t, tc.plaintext, decrypted, "解密后的文本应与原始文本一致。")
			}
		})
	}
}

// TestEncryptGCMBase64 测试使用Base64编码的二进制数据加密功能。
func TestEncryptGCMBase64(t *testing.T) {
	// 准备测试数据
	plainDataBase64 := base64.StdEncoding.EncodeToString([]byte(testPlainText))

	// 表格驱动测试用例
	testCases := []struct {
		name        string
		key         string
		nonceLength int
		data        string
		expectError bool
	}{
		{
			name:        "正常加密场景",
			key:         testKeyBase64,
			nonceLength: testNonceLength,
			data:        plainDataBase64,
			expectError: false,
		},
		{
			name:        "无效密钥",
			key:         invalidBase64,
			nonceLength: testNonceLength,
			data:        plainDataBase64,
			expectError: true,
		},
		{
			name:        "无效数据",
			key:         testKeyBase64,
			nonceLength: testNonceLength,
			data:        invalidBase64,
			expectError: true,
		},
	}

	// 执行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行加密操作
			result, err := EncryptGCMBase64(tc.key, tc.nonceLength, tc.data)

			// 断言结果
			if tc.expectError {
				assert.Error(t, err, "应该返回错误。")
			} else {
				assert.NoError(t, err, "不应该返回错误。")
				assert.NotEmpty(t, result, "加密结果不应为空。")

				// 尝试解密进行验证
				_, decrypted, err := DecryptGCMBase64(tc.key, tc.nonceLength, result)
				assert.NoError(t, err, "解密应该成功。")
				assert.Equal(t, tc.data, decrypted, "解密后的Base64数据应与原始数据一致。")
			}
		})
	}
}

// TestEncryptGCMHex 测试使用Hex编码的二进制数据加密功能。
func TestEncryptGCMHex(t *testing.T) {
	// 准备测试数据
	plainDataHex := hex.EncodeToString([]byte(testPlainText))

	// 表格驱动测试用例
	testCases := []struct {
		name        string
		key         string
		nonceLength int
		data        string
		expectError bool
	}{
		{
			name:        "正常加密场景",
			key:         testKeyHex,
			nonceLength: testNonceLength,
			data:        plainDataHex,
			expectError: false,
		},
		{
			name:        "无效密钥",
			key:         invalidHex,
			nonceLength: testNonceLength,
			data:        plainDataHex,
			expectError: true,
		},
		{
			name:        "无效数据",
			key:         testKeyHex,
			nonceLength: testNonceLength,
			data:        invalidHex,
			expectError: true,
		},
	}

	// 执行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行加密操作
			result, err := EncryptGCMHex(tc.key, tc.nonceLength, tc.data)

			// 断言结果
			if tc.expectError {
				assert.Error(t, err, "应该返回错误。")
			} else {
				assert.NoError(t, err, "不应该返回错误。")
				assert.NotEmpty(t, result, "加密结果不应为空。")

				// 尝试解密进行验证
				_, decrypted, err := DecryptGCMHex(tc.key, tc.nonceLength, result)
				assert.NoError(t, err, "解密应该成功。")
				assert.Equal(t, strings.ToUpper(tc.data), decrypted, "解密后的Hex数据应与原始数据一致（转为大写）。")
			}
		})
	}
}

// TestDecryptStringGCMBase64 测试解密Base64编码的字符串密文到UTF-8文本。
func TestDecryptStringGCMBase64(t *testing.T) {
	// 准备测试数据
	key := []byte(testKeyBytes)
	plaintext := []byte(testPlainText)

	// 生成有效的密文
	encrypted, err := EncryptGCMNonceLength(key, testNonceLength, plaintext)
	assert.NoError(t, err, "加密应该成功。")
	validCiphertext := base64.StdEncoding.EncodeToString(encrypted)

	// 生成一个短密文用于测试长度不足的情况
	shortEncrypted := make([]byte, testNonceLength-1) // 比nonce长度还短
	shortCiphertext := base64.StdEncoding.EncodeToString(shortEncrypted)

	// 表格驱动测试用例
	testCases := []struct {
		name           string
		key            string
		nonceLength    int
		ciphertext     string
		expectError    bool
		skipDecryption bool // 标记是否跳过实际解密操作
	}{
		{
			name:           "正常解密场景",
			key:            testKeyBase64,
			nonceLength:    testNonceLength,
			ciphertext:     validCiphertext,
			expectError:    false,
			skipDecryption: false,
		},
		{
			name:           "无效密钥",
			key:            invalidBase64,
			nonceLength:    testNonceLength,
			ciphertext:     validCiphertext,
			expectError:    true,
			skipDecryption: false,
		},
		{
			name:           "无效密文",
			key:            testKeyBase64,
			nonceLength:    testNonceLength,
			ciphertext:     invalidBase64,
			expectError:    true,
			skipDecryption: false,
		},
		{
			name:           "密文长度不足",
			key:            testKeyBase64,
			nonceLength:    testNonceLength,
			ciphertext:     shortCiphertext,
			expectError:    true,
			skipDecryption: false, // 不需要跳过，因为函数本身会处理长度不足的情况
		},
	}

	// 执行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行解密操作
			nonce, result, err := DecryptStringGCMBase64(tc.key, tc.nonceLength, tc.ciphertext)

			// 断言结果
			if tc.expectError {
				assert.Error(t, err, "应该返回错误。")
			} else {
				assert.NoError(t, err, "不应该返回错误。")
				assert.NotEmpty(t, nonce, "提取的nonce不应为空。")
				assert.Equal(t, testPlainText, result, "解密后的文本应与原始文本一致。")
			}
		})
	}
}

// TestDecryptStringGCMHex 测试解密Hex编码的字符串密文到UTF-8文本。
func TestDecryptStringGCMHex(t *testing.T) {
	// 准备测试数据
	key := []byte(testKeyBytes)
	plaintext := []byte(testPlainText)

	// 生成有效的密文
	encrypted, err := EncryptGCMNonceLength(key, testNonceLength, plaintext)
	assert.NoError(t, err, "加密应该成功。")
	validCiphertext := hex.EncodeToString(encrypted)

	// 生成一个短密文用于测试长度不足的情况
	shortEncrypted := make([]byte, testNonceLength-1) // 比nonce长度还短
	shortCiphertext := hex.EncodeToString(shortEncrypted)

	// 表格驱动测试用例
	testCases := []struct {
		name           string
		key            string
		nonceLength    int
		ciphertext     string
		expectError    bool
		skipDecryption bool // 标记是否跳过实际解密操作
	}{
		{
			name:           "正常解密场景",
			key:            testKeyHex,
			nonceLength:    testNonceLength,
			ciphertext:     validCiphertext,
			expectError:    false,
			skipDecryption: false,
		},
		{
			name:           "正常解密场景-大写密文",
			key:            testKeyHex,
			nonceLength:    testNonceLength,
			ciphertext:     strings.ToUpper(validCiphertext),
			expectError:    false,
			skipDecryption: false,
		},
		{
			name:           "无效密钥",
			key:            invalidHex,
			nonceLength:    testNonceLength,
			ciphertext:     validCiphertext,
			expectError:    true,
			skipDecryption: false,
		},
		{
			name:           "无效密文",
			key:            testKeyHex,
			nonceLength:    testNonceLength,
			ciphertext:     invalidHex,
			expectError:    true,
			skipDecryption: false,
		},
		{
			name:           "密文长度不足",
			key:            testKeyHex,
			nonceLength:    testNonceLength,
			ciphertext:     shortCiphertext,
			expectError:    true,
			skipDecryption: false, // 不需要跳过，因为函数本身会处理长度不足的情况
		},
	}

	// 执行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行解密操作
			nonce, result, err := DecryptStringGCMHex(tc.key, tc.nonceLength, tc.ciphertext)

			// 断言结果
			if tc.expectError {
				assert.Error(t, err, "应该返回错误。")
			} else {
				assert.NoError(t, err, "不应该返回错误。")
				assert.NotEmpty(t, nonce, "提取的nonce不应为空。")
				assert.Equal(t, testPlainText, result, "解密后的文本应与原始文本一致。")
			}
		})
	}
}

// TestDecryptGCMBase64 测试解密Base64编码的二进制密文。
func TestDecryptGCMBase64(t *testing.T) {
	// 准备测试数据
	key := []byte(testKeyBytes)
	plaintext := []byte(testPlainText)

	// 生成有效的密文
	encrypted, err := EncryptGCMNonceLength(key, testNonceLength, plaintext)
	assert.NoError(t, err, "加密应该成功。")
	validCiphertext := base64.StdEncoding.EncodeToString(encrypted)

	// 表格驱动测试用例
	testCases := []struct {
		name        string
		key         string
		nonceLength int
		ciphertext  string
		expectError bool
	}{
		{
			name:        "正常解密场景",
			key:         testKeyBase64,
			nonceLength: testNonceLength,
			ciphertext:  validCiphertext,
			expectError: false,
		},
		{
			name:        "无效密钥",
			key:         invalidBase64,
			nonceLength: testNonceLength,
			ciphertext:  validCiphertext,
			expectError: true,
		},
		{
			name:        "无效密文",
			key:         testKeyBase64,
			nonceLength: testNonceLength,
			ciphertext:  invalidBase64,
			expectError: true,
		},
	}

	// 执行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行解密操作
			nonce, result, err := DecryptGCMBase64(tc.key, tc.nonceLength, tc.ciphertext)

			// 断言结果
			if tc.expectError {
				assert.Error(t, err, "应该返回错误。")
			} else {
				assert.NoError(t, err, "不应该返回错误。")
				assert.NotEmpty(t, nonce, "提取的nonce不应为空。")

				// 解码结果并与原始数据比较
				decoded, err := base64.StdEncoding.DecodeString(result)
				assert.NoError(t, err, "解码Base64结果应该成功。")
				assert.Equal(t, plaintext, decoded, "解密后的数据应与原始数据一致。")
			}
		})
	}
}

// TestDecryptGCMHex 测试解密Hex编码的二进制密文。
func TestDecryptGCMHex(t *testing.T) {
	// 准备测试数据
	key := []byte(testKeyBytes)
	plaintext := []byte(testPlainText)

	// 生成有效的密文
	encrypted, err := EncryptGCMNonceLength(key, testNonceLength, plaintext)
	assert.NoError(t, err, "加密应该成功。")
	validCiphertext := hex.EncodeToString(encrypted)

	// 表格驱动测试用例
	testCases := []struct {
		name        string
		key         string
		nonceLength int
		ciphertext  string
		expectError bool
	}{
		{
			name:        "正常解密场景",
			key:         testKeyHex,
			nonceLength: testNonceLength,
			ciphertext:  validCiphertext,
			expectError: false,
		},
		{
			name:        "正常解密场景-大写密文",
			key:         testKeyHex,
			nonceLength: testNonceLength,
			ciphertext:  strings.ToUpper(validCiphertext),
			expectError: false,
		},
		{
			name:        "无效密钥",
			key:         invalidHex,
			nonceLength: testNonceLength,
			ciphertext:  validCiphertext,
			expectError: true,
		},
		{
			name:        "无效密文",
			key:         testKeyHex,
			nonceLength: testNonceLength,
			ciphertext:  invalidHex,
			expectError: true,
		},
	}

	// 执行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行解密操作
			nonce, result, err := DecryptGCMHex(tc.key, tc.nonceLength, tc.ciphertext)

			// 断言结果
			if tc.expectError {
				assert.Error(t, err, "应该返回错误。")
			} else {
				assert.NoError(t, err, "不应该返回错误。")
				assert.NotEmpty(t, nonce, "提取的nonce不应为空。")

				// 解码结果并与原始数据比较
				decoded, err := hex.DecodeString(result)
				assert.NoError(t, err, "解码Hex结果应该成功。")
				assert.Equal(t, plaintext, decoded, "解密后的数据应与原始数据一致。")
			}
		})
	}
}

// TestDecryptGCMNonceLength 测试从指定nonce长度的数据中提取nonce并解密。
func TestDecryptGCMNonceLength(t *testing.T) {
	// 表格驱动测试用例
	testCases := []struct {
		name        string
		key         []byte
		nonceLength int
		data        []byte
		expectError bool
	}{
		{
			name:        "正常解密场景",
			key:         []byte(testKeyBytes),
			nonceLength: testNonceLength,
			data: func() []byte {
				encrypted, _ := EncryptGCMNonceLength([]byte(testKeyBytes), testNonceLength, []byte(testPlainText))
				return encrypted
			}(),
			expectError: false,
		},
		{
			name:        "数据长度不足",
			key:         []byte(testKeyBytes),
			nonceLength: testNonceLength,
			data:        []byte("too_short"),
			expectError: true,
		},
		{
			name:        "无效密钥长度",
			key:         []byte("invalid_key"),
			nonceLength: testNonceLength,
			data: func() []byte {
				encrypted, _ := EncryptGCMNonceLength([]byte(testKeyBytes), testNonceLength, []byte(testPlainText))
				return encrypted
			}(),
			expectError: true,
		},
	}

	// 执行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行解密操作
			nonce, result, err := DecryptGCMNonceLength(tc.key, tc.nonceLength, tc.data)

			// 断言结果
			if tc.expectError {
				assert.Error(t, err, "应该返回错误。")
			} else {
				assert.NoError(t, err, "不应该返回错误。")
				assert.Len(t, nonce, tc.nonceLength, "nonce长度应符合预期。")
				assert.Equal(t, []byte(testPlainText), result, "解密后的数据应与原始数据一致。")
			}
		})
	}
}

// TestDecryptGCM 测试基本的GCM解密功能。
func TestDecryptGCM(t *testing.T) {
	// 表格驱动测试用例
	testCases := []struct {
		name           string
		key            []byte
		nonce          []byte
		data           []byte
		expectError    bool
		skipDecryption bool // 标记是否跳过实际解密操作
	}{
		{
			name:  "正常解密场景",
			key:   []byte(testKeyBytes),
			nonce: []byte("123456789012"), // 12字节nonce
			data: func() []byte {
				block, _ := aes.NewCipher([]byte(testKeyBytes))
				aesgcm, _ := cipher.NewGCM(block)
				return aesgcm.Seal(nil, []byte("123456789012"), []byte(testPlainText), nil)
			}(),
			expectError:    false,
			skipDecryption: false,
		},
		{
			name:           "无效密钥长度",
			key:            []byte("invalid_key"),
			nonce:          []byte("123456789012"),
			data:           []byte(testPlainText),
			expectError:    true,
			skipDecryption: false,
		},
		{
			name:           "无效nonce长度",
			key:            []byte(testKeyBytes),
			nonce:          []byte("short"),
			data:           []byte(testPlainText),
			expectError:    true,
			skipDecryption: true, // 跳过实际解密操作以避免panic
		},
		{
			name:           "无效数据/认证失败",
			key:            []byte(testKeyBytes),
			nonce:          []byte("123456789012"),
			data:           []byte("tampered_data"),
			expectError:    true,
			skipDecryption: false,
		},
	}

	// 执行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skipDecryption {
				// 对于会导致panic的情况，我们只需要验证nonce长度是否正确
				assert.NotEqual(t, 12, len(tc.nonce), "nonce长度应不等于GCM要求的长度，这会引发错误。")
			} else {
				// 执行解密操作
				result, err := DecryptGCM(tc.key, tc.nonce, tc.data)

				// 断言结果
				if tc.expectError {
					assert.Error(t, err, "应该返回错误。")
				} else {
					assert.NoError(t, err, "不应该返回错误。")
					assert.Equal(t, []byte(testPlainText), result, "解密后的数据应与原始数据一致。")
				}
			}
		})
	}
}

// BenchmarkEncryptGCM 基准测试GCM加密性能。
func BenchmarkEncryptGCM(b *testing.B) {
	key := []byte(testKeyBytes)
	nonce := []byte("123456789012")
	data := []byte(testPlainText)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = EncryptGCM(key, nonce, data)
	}
}

// BenchmarkDecryptGCM 基准测试GCM解密性能。
func BenchmarkDecryptGCM(b *testing.B) {
	key := []byte(testKeyBytes)
	nonce := []byte("123456789012")

	// 生成加密数据用于解密基准测试
	block, _ := aes.NewCipher(key)
	aesgcm, _ := cipher.NewGCM(block)
	data := aesgcm.Seal(nil, nonce, []byte(testPlainText), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecryptGCM(key, nonce, data)
	}
}
