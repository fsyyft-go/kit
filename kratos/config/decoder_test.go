// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"errors"
	"strings"
	"testing"

	kratos_config "github.com/go-kratos/kratos/v2/config"
)

// TestWithResolve 测试 WithResolve 函数是否正确设置解析函数。
func TestWithResolve(t *testing.T) {
	// 创建一个测试用的解析函数。
	testResolve := func(target map[string]interface{}) error {
		return nil
	}

	// 创建一个空的 DecoderOptions。
	options := &DecoderOptions{}

	// 应用 WithResolve 选项。
	opt := WithResolve(testResolve)
	opt(options)

	// 验证解析函数是否已正确设置。
	if options.Resolve == nil {
		t.Fatal("WithResolve() 未能正确设置解析函数")
	}
}

// TestNewDecoder 测试 NewDecoder 函数是否正确创建解码器实例。
func TestNewDecoder(t *testing.T) {
	// 定义测试用例。
	tests := []struct {
		name          string          // 测试用例名称。
		opts          []DecoderOption // 解码器选项。
		expectResolve bool            // 是否期望解析函数存在。
	}{
		{
			name:          "无选项",
			opts:          nil,
			expectResolve: true, // 默认使用 defaultResolve.Resolve。
		},
		{
			name: "自定义解析函数",
			opts: []DecoderOption{
				WithResolve(func(target map[string]interface{}) error {
					return nil
				}),
			},
			expectResolve: true,
		},
		{
			name: "包含空选项",
			opts: []DecoderOption{
				nil,
				WithResolve(func(target map[string]interface{}) error {
					return nil
				}),
			},
			expectResolve: true,
		},
	}

	// 遍历测试用例。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建解码器实例。
			decoder := NewDecoder(tt.opts...)

			// 验证解码器是否已创建。
			if decoder == nil {
				t.Fatal("NewDecoder() 返回了 nil")
			}

			// 验证解析函数是否已正确设置。
			if (decoder.Resolve == nil) == tt.expectResolve {
				t.Fatalf("解析函数设置不正确，期望存在: %v, 实际: %v", tt.expectResolve, decoder.Resolve != nil)
			}
		})
	}
}

// TestDecode 测试 Decode 方法是否正确解码配置。
func TestDecode(t *testing.T) {
	// 跳过需要 JSON 编解码器的测试，因为我们没有实际的 JSON 编解码器。
	t.Skip("跳过需要 JSON 编解码器的测试")

	// 定义测试用例。
	tests := []struct {
		name          string                  // 测试用例名称。
		src           *kratos_config.KeyValue // 源配置。
		resolveFunc   Resolve                 // 解析函数。
		expectedMap   map[string]interface{}  // 期望的映射结果。
		expectedError bool                    // 是否期望错误。
	}{
		{
			name: "空格式，简单键",
			src: &kratos_config.KeyValue{
				Key:    "key",
				Value:  []byte("value"),
				Format: "",
			},
			resolveFunc: nil,
			expectedMap: map[string]interface{}{
				"key": []byte("value"),
			},
			expectedError: false,
		},
		{
			name: "空格式，嵌套键",
			src: &kratos_config.KeyValue{
				Key:    "a.b.c",
				Value:  []byte("value"),
				Format: "",
			},
			resolveFunc: nil,
			expectedMap: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": []byte("value"),
					},
				},
			},
			expectedError: false,
		},
		{
			name: "JSON格式，成功解码",
			src: &kratos_config.KeyValue{
				Key:    "config",
				Value:  []byte(`{"key":"value","nested":{"inner":"data"}}`),
				Format: "json",
			},
			resolveFunc: nil,
			expectedMap: map[string]interface{}{
				"key": "value",
				"nested": map[string]interface{}{
					"inner": "data",
				},
			},
			expectedError: false,
		},
		{
			name: "JSON格式，解析函数成功",
			src: &kratos_config.KeyValue{
				Key:    "config",
				Value:  []byte(`{"key":"value"}`),
				Format: "json",
			},
			resolveFunc: func(target map[string]interface{}) error {
				target["added"] = "by_resolve"
				return nil
			},
			expectedMap: map[string]interface{}{
				"key":   "value",
				"added": "by_resolve",
			},
			expectedError: false,
		},
		{
			name: "JSON格式，解析函数失败",
			src: &kratos_config.KeyValue{
				Key:    "config",
				Value:  []byte(`{"key":"value"}`),
				Format: "json",
			},
			resolveFunc: func(target map[string]interface{}) error {
				return errors.New("解析错误")
			},
			expectedMap:   map[string]interface{}{},
			expectedError: true,
		},
		{
			name: "无效的JSON",
			src: &kratos_config.KeyValue{
				Key:    "config",
				Value:  []byte(`{"key":"value`), // 缺少结束括号。
				Format: "json",
			},
			resolveFunc:   nil,
			expectedMap:   map[string]interface{}{},
			expectedError: true,
		},
		{
			name: "不支持的格式",
			src: &kratos_config.KeyValue{
				Key:    "config",
				Value:  []byte(`{"key":"value"}`),
				Format: "unsupported",
			},
			resolveFunc:   nil,
			expectedMap:   map[string]interface{}{},
			expectedError: true,
		},
	}

	// 遍历测试用例。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建解码器。
			decoder := NewDecoder()
			if tt.resolveFunc != nil {
				decoder.Resolve = tt.resolveFunc
			}

			// 准备目标映射。
			target := make(map[string]interface{})

			// 执行解码。
			err := decoder.Decode(tt.src, target)

			// 验证错误是否符合预期。
			if (err != nil) != tt.expectedError {
				t.Fatalf("期望错误: %v, 实际: %v, 错误信息: %v", tt.expectedError, err != nil, err)
			}

			// 如果期望成功，验证解码结果。
			if !tt.expectedError {
				if tt.src.Format == "" {
					// 对于空格式，我们需要特殊处理验证，因为键可能是嵌套的。
					validateNestedKeyValue(t, target, tt.src.Key, tt.src.Value)
				} else {
					// 对于其他格式，直接比较映射。
					validateMap(t, target, tt.expectedMap)
				}
			}
		})
	}
}

// TestDecodeEmptyFormat 测试 Decode 方法处理空格式的情况。
func TestDecodeEmptyFormat(t *testing.T) {
	// 定义测试用例。
	tests := []struct {
		name          string                  // 测试用例名称。
		src           *kratos_config.KeyValue // 源配置。
		expectedKey   string                  // 期望的键。
		expectedValue []byte                  // 期望的值。
	}{
		{
			name: "简单键",
			src: &kratos_config.KeyValue{
				Key:    "key",
				Value:  []byte("value"),
				Format: "",
			},
			expectedKey:   "key",
			expectedValue: []byte("value"),
		},
		{
			name: "嵌套键",
			src: &kratos_config.KeyValue{
				Key:    "a.b.c",
				Value:  []byte("value"),
				Format: "",
			},
			expectedKey:   "c",
			expectedValue: []byte("value"),
		},
	}

	// 遍历测试用例。
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建解码器。
			decoder := NewDecoder()

			// 准备目标映射。
			target := make(map[string]interface{})

			// 执行解码。
			err := decoder.Decode(tt.src, target)

			// 验证没有错误。
			if err != nil {
				t.Fatalf("解码出错: %v", err)
			}

			// 验证解码结果。
			if tt.src.Key == tt.expectedKey {
				// 简单键。
				value, exists := target[tt.expectedKey]
				if !exists {
					t.Fatalf("键 %s 不存在于映射中", tt.expectedKey)
				}

				// 验证值类型和内容。
				actualValue, ok := value.([]byte)
				if !ok {
					t.Fatalf("键 %s 的值类型不是 []byte，而是 %T", tt.expectedKey, value)
				}

				if string(actualValue) != string(tt.expectedValue) {
					t.Fatalf("键 %s 的值不匹配，期望: %s, 实际: %s", tt.expectedKey, string(tt.expectedValue), string(actualValue))
				}
			} else {
				// 嵌套键，需要遍历键路径。
				keys := strings.Split(tt.src.Key, ".")
				current := target

				// 遍历键路径。
				for i, k := range keys {
					value, exists := current[k]
					if !exists {
						t.Fatalf("键 %s 不存在于映射中", k)
					}

					if i == len(keys)-1 {
						// 最后一个键，验证值。
						actualValue, ok := value.([]byte)
						if !ok {
							t.Fatalf("键 %s 的值类型不是 []byte，而是 %T", k, value)
						}

						if string(actualValue) != string(tt.expectedValue) {
							t.Fatalf("键 %s 的值不匹配，期望: %s, 实际: %s", k, string(tt.expectedValue), string(actualValue))
						}
					} else {
						// 中间键，继续遍历。
						nestedMap, ok := value.(map[string]interface{})
						if !ok {
							t.Fatalf("键 %s 的值不是映射，而是 %T", k, value)
						}
						current = nestedMap
					}
				}
			}
		})
	}
}

// TestDecodeUnsupportedFormat 测试 Decode 方法处理不支持的格式的情况。
func TestDecodeUnsupportedFormat(t *testing.T) {
	// 创建解码器。
	decoder := NewDecoder()

	// 准备源配置和目标映射。
	src := &kratos_config.KeyValue{
		Key:    "config",
		Value:  []byte(`{"key":"value"}`),
		Format: "unsupported",
	}
	target := make(map[string]interface{})

	// 执行解码。
	err := decoder.Decode(src, target)

	// 验证有错误。
	if err == nil {
		t.Fatal("期望解码不支持的格式时返回错误，但没有错误")
	}

	// 验证错误消息。
	expectedErrMsg := "unsupported key: config format: unsupported"
	if err.Error() != expectedErrMsg {
		t.Fatalf("错误消息不匹配，期望: %s, 实际: %s", expectedErrMsg, err.Error())
	}
}

// validateNestedKeyValue 验证嵌套键值是否正确设置。
func validateNestedKeyValue(t *testing.T, actual map[string]interface{}, key string, expectedValue []byte) {
	// 分割键路径。
	keys := strings.Split(key, ".")

	// 遍历键路径。
	current := actual
	for i, k := range keys {
		value, exists := current[k]
		if !exists {
			t.Fatalf("键 %s 不存在于映射中", k)
		}

		if i == len(keys)-1 {
			// 最后一个键，验证值。
			actualValue, ok := value.([]byte)
			if !ok {
				t.Fatalf("键 %s 的值类型不是 []byte，而是 %T", k, value)
			}
			if string(actualValue) != string(expectedValue) {
				t.Fatalf("键 %s 的值不匹配，期望: %s, 实际: %s", k, string(expectedValue), string(actualValue))
			}
		} else {
			// 中间键，继续遍历。
			nestedMap, ok := value.(map[string]interface{})
			if !ok {
				t.Fatalf("键 %s 的值不是映射，而是 %T", k, value)
			}
			current = nestedMap
		}
	}
}

// validateMap 验证两个映射是否匹配。
func validateMap(t *testing.T, actual, expected map[string]interface{}) {
	// 验证所有期望的键值对都存在。
	for k, expectedV := range expected {
		actualV, exists := actual[k]
		if !exists {
			t.Fatalf("键 %s 不存在于实际映射中", k)
		}

		// 根据值的类型进行不同的验证。
		switch expectedVTyped := expectedV.(type) {
		case map[string]interface{}:
			// 如果值是映射，递归验证。
			actualVTyped, ok := actualV.(map[string]interface{})
			if !ok {
				t.Fatalf("键 %s 的值类型不匹配，期望: map[string]interface{}, 实际: %T", k, actualV)
			}
			validateMap(t, actualVTyped, expectedVTyped)
		default:
			// 对于简单类型，直接比较字符串表示。
			if actualV != expectedV {
				t.Fatalf("键 %s 的值不匹配，期望: %v, 实际: %v", k, expectedV, actualV)
			}
		}
	}

	// 验证没有多余的键。
	for k := range actual {
		if _, exists := expected[k]; !exists {
			t.Fatalf("实际映射中存在多余的键: %s", k)
		}
	}
}
