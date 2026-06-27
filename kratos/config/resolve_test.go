// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"errors"
	"os"
	"testing"

	kitcryptodes "github.com/fsyyft-go/kit/crypto/des"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewResolve 测试 newResolve 函数是否正确初始化解析器实例
func TestNewResolve(t *testing.T) {
	// 创建一个新的解析器实例。
	r := newResolve()
	// 验证解析器实例不为空。
	assert.NotNil(t, r, "newResolve() 返回了 nil")
	// 验证解析器的 resolvers 映射已初始化。
	assert.NotNil(t, r.resolvers, "newResolve() 未初始化 resolvers 映射")
	// 验证初始化的 resolvers 映射为空。
	assert.Empty(t, r.resolvers, "newResolve() 初始化的 resolvers 映射不为空")
}

// TestResolveRegister 测试 register 方法是否正确注册解析处理函数。
func TestResolveRegister(t *testing.T) {
	// 创建一个新的解析器实例。
	r := newResolve()

	// 创建一个测试用的解析处理函数。
	testResolver := func(target map[string]interface{}, key, val string) error {
		return nil
	}

	// 注册解析处理函数。
	r.register("test", testResolver)

	// 验证解析处理函数是否已正确注册。
	assert.Len(t, r.resolvers, 1, "register() 后 resolvers 映射应包含 1 个元素")

	// 验证是否可以通过键名获取到注册的解析处理函数。
	_, exists := r.resolvers["test"]
	assert.True(t, exists, "register() 未能正确注册解析处理函数")
}

// TestRegisterResolve 测试 RegisterResolve 函数是否正确向默认解析器注册解析处理函数。
func TestRegisterResolve(t *testing.T) {
	// 保存原始的默认解析器。
	originalDefaultResolve := defaultResolve
	// 创建一个新的默认解析器用于测试。
	defaultResolve = newResolve()
	// 测试完成后恢复原始的默认解析器。
	defer func() { defaultResolve = originalDefaultResolve }()

	// 创建一个测试用的解析处理函数。
	testResolver := func(target map[string]interface{}, key, val string) error {
		return nil
	}

	// 注册解析处理函数到默认解析器。
	RegisterResolve("test", testResolver)

	// 验证解析处理函数是否已正确注册到默认解析器。
	assert.Len(t, defaultResolve.resolvers, 1, "RegisterResolve() 后默认解析器的 resolvers 映射应包含 1 个元素")

	// 验证是否可以通过键名获取到注册的解析处理函数。
	_, exists := defaultResolve.resolvers["test"]
	assert.True(t, exists, "RegisterResolve() 未能正确向默认解析器注册解析处理函数")
}

// TestRegisterResolveBase64 测试 registerResolveBase64 函数是否正确处理 base64 编码的值。
func TestRegisterResolveBase64(t *testing.T) {
	// 定义测试用例。
	tests := []struct {
		name           string                 // 测试用例名称。
		description    string                 // 用例语义说明。
		target         map[string]interface{} // 目标映射。
		key            string                 // 键名。
		val            string                 // 值。
		expectedTarget map[string]interface{} // 期望的目标映射。
		expectedError  error                  // 期望的错误。
	}{
		{
			name:           "success/non-base64-suffix",
			description:    "验证非 base64 后缀的键不会被改写。",
			target:         map[string]interface{}{"key": "value"},
			key:            "key",
			val:            "value",
			expectedTarget: map[string]interface{}{"key": "value"},
			expectedError:  nil,
		},
		{
			name:           "success/valid-base64-value",
			description:    "验证有效 base64 编码会写入去后缀键并保留原键。",
			target:         map[string]interface{}{"key.b64": "SGVsbG8gV29ybGQ="},
			key:            "key.b64",
			val:            "SGVsbG8gV29ybGQ=",
			expectedTarget: map[string]interface{}{"key.b64": "SGVsbG8gV29ybGQ=", "key": "Hello World"},
			expectedError:  nil,
		},
		{
			name:           "error/invalid-base64-value",
			description:    "验证无效 base64 编码会返回解码错误并把错误字符串写入去后缀键。",
			target:         map[string]interface{}{"key.b64": "invalid-base64"},
			key:            "key.b64",
			val:            "invalid-base64",
			expectedTarget: map[string]interface{}{"key.b64": "invalid-base64", "key": "illegal base64 data at input byte 7"},
			expectedError:  errors.New("illegal base64 data at input byte 7"),
		},
	}

	// 遍历测试用例。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			// 调用 registerResolveBase64 函数。
			err := registerResolveBase64(tt.target, tt.key, tt.val)

			// 验证错误是否符合预期。
			if tt.expectedError != nil {
				assert.Error(t, err, "应该返回错误")
				assert.Equal(t, tt.expectedError.Error(), err.Error(), "错误消息不匹配")
			} else {
				assert.NoError(t, err, "不应该返回错误")
			}

			// 验证目标映射是否符合预期。
			for k, expectedV := range tt.expectedTarget {
				actualV, exists := tt.target[k]
				assert.True(t, exists, "目标映射中缺少键 %s", k)
				assert.Equal(t, expectedV, actualV, "键 %s 的值不匹配", k)
			}
		})
	}
}

// TestRegisterResolveDES 验证 registerResolveDES 对 DES 后缀配置项的解析行为。
//
// 该测试通过表驱动用例覆盖非 .des 后缀不改写、有效 DES 密文成功解密以及无效 DES 密文返回错误并写入错误字符串。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRegisterResolveDES(t *testing.T) {
	validCiphertext, err := kitcryptodes.EncryptStringCBCPkCS7PaddingStringHex(defaultDESKey, "decrypted-secret")
	require.NoError(t, err)

	tests := []struct {
		name            string
		description     string
		giveTarget      map[string]interface{}
		giveKey         string
		giveValue       string
		wantTarget      map[string]interface{}
		wantErr         bool
		wantErrContains string
	}{
		{
			name:        "success/non-des-suffix",
			description: "验证非 .des 后缀键不会触发 DES 解密，也不会新增去后缀键。",
			giveTarget: map[string]interface{}{
				"password": "plain-text",
			},
			giveKey:   "password",
			giveValue: "plain-text",
			wantTarget: map[string]interface{}{
				"password": "plain-text",
			},
		},
		{
			name:        "success/valid-des-ciphertext",
			description: "验证有效 DES 密文会解密为明文并写入去后缀键，同时保留原始密文键。",
			giveTarget: map[string]interface{}{
				"password.des": validCiphertext,
			},
			giveKey:   "password.des",
			giveValue: validCiphertext,
			wantTarget: map[string]interface{}{
				"password.des": validCiphertext,
				"password":     "decrypted-secret",
			},
		},
		{
			name:        "error/invalid-des-ciphertext",
			description: "验证无效 DES 密文会返回解密错误，并把错误字符串写入去后缀键以便诊断。",
			giveTarget: map[string]interface{}{
				"password.des": "invalid-ciphertext",
			},
			giveKey:   "password.des",
			giveValue: "invalid-ciphertext",
			wantTarget: map[string]interface{}{
				"password.des": "invalid-ciphertext",
				"password":     "encoding/hex: invalid byte: U+0069 'i'",
			},
			wantErr:         true,
			wantErrContains: "invalid byte",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			err := registerResolveDES(tt.giveTarget, tt.giveKey, tt.giveValue)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantTarget, tt.giveTarget)
		})
	}
}

// TestResolve 测试 Resolve 方法是否正确处理各种类型的配置值。
func TestResolve(t *testing.T) {
	// 定义测试用例。
	tests := []struct {
		name           string                 // 测试用例名称。
		description    string                 // 用例语义说明。
		target         map[string]interface{} // 目标映射。
		resolvers      map[string]ResolveItem // 解析处理函数映射。
		expectedTarget map[string]interface{} // 期望的目标映射。
		expectedError  error                  // 期望的错误。
	}{
		{
			name:           "空映射",
			description:    "验证空配置映射在无解析器时保持为空且不返回错误。",
			target:         map[string]interface{}{},
			resolvers:      nil,
			expectedTarget: map[string]interface{}{},
			expectedError:  nil,
		},
		{
			name:        "无字符串值的映射",
			description: "验证非字符串类型配置值不会触发解析器处理并保持原值。",
			target: map[string]interface{}{
				"int":  123,
				"bool": true,
			},
			resolvers: nil,
			expectedTarget: map[string]interface{}{
				"int":  123,
				"bool": true,
			},
			expectedError: nil,
		},
		{
			name:        "字符串值无解析器",
			description: "验证存在字符串值但未注册解析器时 Resolve 不改写目标映射且不返回错误。",
			target: map[string]interface{}{
				"plain": "value",
			},
			resolvers: nil,
			expectedTarget: map[string]interface{}{
				"plain": "value",
			},
			expectedError: nil,
		},
		{
			name:        "嵌套映射",
			description: "验证嵌套映射中的字符串配置项会递归应用注册解析器。",
			target: map[string]interface{}{
				"nested": map[string]interface{}{
					"key.b64": "SGVsbG8gV29ybGQ=",
				},
			},
			resolvers: map[string]ResolveItem{
				"b64": registerResolveBase64,
			},
			expectedTarget: map[string]interface{}{
				"nested": map[string]interface{}{
					"key.b64": "SGVsbG8gV29ybGQ=",
					"key":     "Hello World",
				},
			},
			expectedError: nil,
		},
		{
			name:        "嵌套映射解析器返回错误",
			description: "验证嵌套映射中的解析器错误会被外层 Resolve 透传。",
			target: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
			resolvers: map[string]ResolveItem{
				"error": func(target map[string]interface{}, key, val string) error {
					return errors.New("嵌套错误")
				},
			},
			expectedTarget: map[string]interface{}{
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
			expectedError: errors.New("嵌套错误"),
		},
		{
			name:        "包含数组的映射",
			description: "验证数组中的映射元素会递归解析，普通数组元素保持不变。",
			target: map[string]interface{}{
				"array": []interface{}{
					map[string]interface{}{
						"key.b64": "SGVsbG8gV29ybGQ=",
					},
					"string",
					123,
				},
			},
			resolvers: map[string]ResolveItem{
				"b64": registerResolveBase64,
			},
			expectedTarget: map[string]interface{}{
				"array": []interface{}{
					map[string]interface{}{
						"key.b64": "SGVsbG8gV29ybGQ=",
						"key":     "Hello World",
					},
					"string",
					123,
				},
			},
			expectedError: nil,
		},
		{
			name:        "数组映射解析器返回错误",
			description: "验证数组中映射元素的解析器错误会被 Resolve 立即透传。",
			target: map[string]interface{}{
				"array": []interface{}{
					map[string]interface{}{
						"key": "value",
					},
				},
			},
			resolvers: map[string]ResolveItem{
				"error": func(target map[string]interface{}, key, val string) error {
					return errors.New("数组错误")
				},
			},
			expectedTarget: map[string]interface{}{
				"array": []interface{}{
					map[string]interface{}{
						"key": "value",
					},
				},
			},
			expectedError: errors.New("数组错误"),
		},
		{
			name:        "解析器返回错误",
			description: "验证任一注册解析器返回错误时 Resolve 会立即透传该错误。",
			target: map[string]interface{}{
				"key": "value",
			},
			resolvers: map[string]ResolveItem{
				"error": func(target map[string]interface{}, key, val string) error {
					return errors.New("测试错误")
				},
			},
			expectedTarget: map[string]interface{}{
				"key": "value",
			},
			expectedError: errors.New("测试错误"),
		},
	}

	// 遍历测试用例。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			// 创建一个新的解析器实例。
			r := newResolve()

			// 注册解析处理函数。
			if tt.resolvers != nil {
				for k, resolver := range tt.resolvers {
					r.register(k, resolver)
				}
			}

			// 执行解析。
			err := r.Resolve(tt.target)

			// 验证错误是否符合预期。
			if tt.expectedError != nil {
				assert.Error(t, err, "应该返回错误")
				assert.Equal(t, tt.expectedError.Error(), err.Error(), "错误消息不匹配")
			} else {
				assert.NoError(t, err, "不应该返回错误")
			}

			// 验证嵌套映射是否符合预期。
			validateNestedMapWithAssert(t, tt.target, tt.expectedTarget)
		})
	}
}

// validateNestedMapWithAssert 递归验证嵌套映射是否匹配预期，使用 assert 包。
func validateNestedMapWithAssert(t *testing.T, actual, expected map[string]interface{}) {
	// 遍历期望的映射，验证实际映射中是否存在对应的键值对。
	for k, expectedV := range expected {
		actualV, exists := actual[k]
		assert.True(t, exists, "目标映射中缺少键 %s", k)

		// 根据值的类型进行不同的验证。
		switch expectedVTyped := expectedV.(type) {
		case map[string]interface{}:
			// 如果值是映射，递归验证。
			actualVTyped, ok := actualV.(map[string]interface{})
			assert.True(t, ok, "键 %s 的值类型不匹配，期望: map[string]interface{}, 实际: %T", k, actualV)
			if ok {
				validateNestedMapWithAssert(t, actualVTyped, expectedVTyped)
			}
		case []interface{}:
			// 如果值是数组，验证数组长度和内容。
			actualVTyped, ok := actualV.([]interface{})
			assert.True(t, ok, "键 %s 的值类型不匹配，期望: []interface{}, 实际: %T", k, actualV)
			if ok {
				assert.Len(t, actualVTyped, len(expectedVTyped), "键 %s 的数组长度不匹配", k)
				// 遍历数组，验证每个元素。
				for i, expectedItem := range expectedVTyped {
					if i < len(actualVTyped) { // 确保索引在范围内
						if expectedItemMap, ok := expectedItem.(map[string]interface{}); ok {
							// 如果元素是映射，递归验证。
							actualItemMap, ok := actualVTyped[i].(map[string]interface{})
							assert.True(t, ok, "键 %s 的数组项 %d 类型不匹配，期望: map[string]interface{}, 实际: %T", k, i, actualVTyped[i])
							if ok {
								validateNestedMapWithAssert(t, actualItemMap, expectedItemMap)
							}
						} else {
							// 如果元素是简单类型，直接比较。
							assert.Equal(t, expectedItem, actualVTyped[i], "键 %s 的数组项 %d 不匹配", k, i)
						}
					}
				}
			}
		default:
			// 对于简单类型，直接比较。
			assert.Equal(t, expectedV, actualV, "键 %s 的值不匹配", k)
		}
	}
}

// TestInit 测试 init 函数是否正确初始化默认解析器并注册 base64 解析处理器。
func TestInit(t *testing.T) {
	// 验证默认解析器是否已初始化。
	assert.NotNil(t, defaultResolve, "init() 未初始化默认解析器")

	// 验证 base64 解析处理器是否已注册。
	_, exists := defaultResolve.resolvers[suffixBase64]
	assert.True(t, exists, "init() 未注册 base64 解析处理器")

	// 测试默认解析器的 base64 解析功能。
	target := map[string]interface{}{
		"key.b64": "SGVsbG8gV29ybGQ=",
	}

	// 使用默认解析器解析目标映射。
	err := defaultResolve.Resolve(target)
	assert.NoError(t, err, "默认解析器解析 base64 值时出错")

	// 验证 base64 解析结果。
	val, exists := target["key"]
	assert.True(t, exists, "默认解析器未能解析 base64 值")
	assert.Equal(t, "Hello World", val, "默认解析器解析 base64 值不正确")
}

// TestRegisterResolveEnv 测试 registerResolveEnv 函数是否正确处理环境变量解析。
func TestRegisterResolveEnv(t *testing.T) {
	// 定义测试用例。
	tests := []struct {
		name           string                 // 测试用例名称。
		description    string                 // 用例语义说明。
		envKey         string                 // 环境变量名。
		envVal         string                 // 环境变量值。
		target         map[string]interface{} // 目标映射。
		key            string                 // 键名。
		val            string                 // 值。
		expectedTarget map[string]interface{} // 期望的目标映射。
		setEnv         bool                   // 是否设置环境变量。
	}{
		{
			name:        "环境变量存在",
			description: "验证 .env 后缀键在环境变量存在时写入去后缀键对应的环境变量值。",
			envKey:      "TEST_ENV_KEY",
			envVal:      "test_env_value",
			target:      map[string]interface{}{"key.env": "TEST_ENV_KEY"},
			key:         "key.env",
			val:         "TEST_ENV_KEY",
			expectedTarget: map[string]interface{}{
				"key.env": "TEST_ENV_KEY",
				"key":     "test_env_value",
			},
			setEnv: true,
		},
		{
			name:        "环境变量不存在",
			description: "验证 .env 后缀键在环境变量不存在时不会新增去后缀键。",
			envKey:      "NOT_EXIST_ENV_KEY",
			envVal:      "",
			target:      map[string]interface{}{"key.env": "NOT_EXIST_ENV_KEY"},
			key:         "key.env",
			val:         "NOT_EXIST_ENV_KEY",
			expectedTarget: map[string]interface{}{
				"key.env": "NOT_EXIST_ENV_KEY",
			},
			setEnv: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			originalValue, originalExists := os.LookupEnv(tt.envKey)
			t.Cleanup(func() {
				if originalExists {
					_ = os.Setenv(tt.envKey, originalValue)
				} else {
					_ = os.Unsetenv(tt.envKey)
				}
			})

			if tt.setEnv {
				// 设置环境变量。
				err := os.Setenv(tt.envKey, tt.envVal)
				assert.NoError(t, err, "设置环境变量失败")
			} else {
				_ = os.Unsetenv(tt.envKey)
			}

			// 调用 registerResolveEnv 函数。
			_ = registerResolveEnv(tt.target, tt.key, tt.val)

			// 验证目标映射是否符合预期。
			for k, expectedV := range tt.expectedTarget {
				actualV, exists := tt.target[k]
				assert.True(t, exists, "目标映射中缺少键 %s", k)
				assert.Equal(t, expectedV, actualV, "键 %s 的值不匹配", k)
			}
		})
	}
}
