// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
		target         map[string]interface{} // 目标映射。
		key            string                 // 键名。
		val            string                 // 值。
		expectedTarget map[string]interface{} // 期望的目标映射。
		expectedError  error                  // 期望的错误。
	}{
		{
			name:           "非 base64 后缀的键",
			target:         map[string]interface{}{"key": "value"},
			key:            "key",
			val:            "value",
			expectedTarget: map[string]interface{}{"key": "value"},
			expectedError:  nil,
		},
		{
			name:           "有效的 base64 编码值",
			target:         map[string]interface{}{"key.b64": "SGVsbG8gV29ybGQ="},
			key:            "key.b64",
			val:            "SGVsbG8gV29ybGQ=",
			expectedTarget: map[string]interface{}{"key.b64": "SGVsbG8gV29ybGQ=", "key": "Hello World"},
			expectedError:  nil,
		},
		{
			name:           "无效的 base64 编码值",
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

// TestResolve 测试 Resolve 方法是否正确处理各种类型的配置值。
func TestResolve(t *testing.T) {
	// 定义测试用例。
	tests := []struct {
		name           string                 // 测试用例名称。
		target         map[string]interface{} // 目标映射。
		resolvers      map[string]ResolveItem // 解析处理函数映射。
		expectedTarget map[string]interface{} // 期望的目标映射。
		expectedError  error                  // 期望的错误。
	}{
		{
			name:           "空映射",
			target:         map[string]interface{}{},
			resolvers:      nil,
			expectedTarget: map[string]interface{}{},
			expectedError:  nil,
		},
		{
			name: "无字符串值的映射",
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
			name: "嵌套映射",
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
			name: "包含数组的映射",
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
			name: "解析器返回错误",
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
		envKey         string                 // 环境变量名。
		envVal         string                 // 环境变量值。
		target         map[string]interface{} // 目标映射。
		key            string                 // 键名。
		val            string                 // 值。
		expectedTarget map[string]interface{} // 期望的目标映射。
		setEnv         bool                   // 是否设置环境变量。
	}{
		{
			name:   "环境变量存在",
			envKey: "TEST_ENV_KEY",
			envVal: "test_env_value",
			target: map[string]interface{}{"key.env": "TEST_ENV_KEY"},
			key:    "key.env",
			val:    "TEST_ENV_KEY",
			expectedTarget: map[string]interface{}{
				"key.env": "TEST_ENV_KEY",
				"key":     "test_env_value",
			},
			setEnv: true,
		},
		{
			name:   "环境变量不存在",
			envKey: "NOT_EXIST_ENV_KEY",
			envVal: "",
			target: map[string]interface{}{"key.env": "NOT_EXIST_ENV_KEY"},
			key:    "key.env",
			val:    "NOT_EXIST_ENV_KEY",
			expectedTarget: map[string]interface{}{
				"key.env": "NOT_EXIST_ENV_KEY",
			},
			setEnv: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				// 设置环境变量。
				err := os.Setenv(tt.envKey, tt.envVal)
				assert.NoError(t, err, "设置环境变量失败")
				defer func() {
					_ = os.Unsetenv(tt.envKey)
				}()
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
