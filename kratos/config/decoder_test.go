// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"errors"
	"testing"

	kratosconfig "github.com/go-kratos/kratos/v2/config"
	_ "github.com/go-kratos/kratos/v2/encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWithResolve_Option 验证 WithResolve 能将自定义解析函数写入解码器选项。
//
// 该测试通过调用选项中保存的 Resolve 函数并观察 target 变化，确保选项保存的函数可被后续解码流程使用。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestWithResolve_Option(t *testing.T) {
	options := &DecoderOptions{}
	target := map[string]interface{}{}

	opt := WithResolve(func(target map[string]interface{}) error {
		target["resolved"] = true
		return nil
	})
	opt(options)

	require.NotNil(t, options.Resolve)
	require.NoError(t, options.Resolve(target))
	assert.Equal(t, true, target["resolved"])
}

// TestNewDecoder_Options 验证 NewDecoder 对默认配置、自定义 Resolve 和 nil option 的处理。
//
// 该测试通过表驱动用例覆盖解码器初始化时的默认解析函数、自定义解析函数以及 nil DecoderOption 跳过语义。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewDecoder_Options(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveOptions []DecoderOption
		assert      func(t *testing.T, decoder *Decoder)
	}{
		{
			name:        "success/default-resolve",
			description: "验证 NewDecoder 在未提供选项时使用默认 Resolve 函数。",
			assert: func(t *testing.T, decoder *Decoder) {
				t.Helper()
				assert.NotNil(t, decoder.Resolve)
			},
		},
		{
			name:        "success/custom-resolve",
			description: "验证 NewDecoder 能应用自定义 Resolve 函数并保留其行为。",
			giveOptions: []DecoderOption{
				WithResolve(func(target map[string]interface{}) error {
					target["source"] = "custom"
					return nil
				}),
			},
			assert: func(t *testing.T, decoder *Decoder) {
				t.Helper()
				target := map[string]interface{}{}
				require.NotNil(t, decoder.Resolve)
				require.NoError(t, decoder.Resolve(target))
				assert.Equal(t, "custom", target["source"])
			},
		},
		{
			name:        "success/nil-option-skipped",
			description: "验证 NewDecoder 遇到 nil DecoderOption 时跳过该选项并继续应用后续选项。",
			giveOptions: []DecoderOption{
				nil,
				WithResolve(func(target map[string]interface{}) error {
					target["source"] = "after-nil"
					return nil
				}),
			},
			assert: func(t *testing.T, decoder *Decoder) {
				t.Helper()
				target := map[string]interface{}{}
				require.NotNil(t, decoder.Resolve)
				require.NoError(t, decoder.Resolve(target))
				assert.Equal(t, "after-nil", target["source"])
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			decoder := NewDecoder(tt.giveOptions...)

			require.NotNil(t, decoder)
			if tt.assert != nil {
				tt.assert(t, decoder)
			}
		})
	}
}

// TestDecode_Behavior 验证 Decoder.Decode 对空格式、已注册 codec、Resolve 和错误路径的处理。
//
// 该测试通过表驱动用例覆盖简单键、嵌套键、JSON 解码、Resolve 修改 target、Resolve 错误透传、codec.Unmarshal 错误以及 unsupported format 错误。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestDecode_Behavior(t *testing.T) {
	errResolve := errors.New("resolve failed")

	tests := []struct {
		name              string
		description       string
		giveSrc           *kratosconfig.KeyValue
		giveNilOption     bool
		giveResolve       Resolve
		wantTarget        map[string]interface{}
		wantErr           bool
		wantErrIs         error
		wantErrEqual      string
		wantErrContains   string
		wantResolveCalled bool
	}{
		{
			name:        "success/empty-format-simple-key",
			description: "验证空格式配置会把简单键直接写入 target，并保留原始字节值。",
			giveSrc: &kratosconfig.KeyValue{
				Key:    "plain",
				Value:  []byte("value"),
				Format: "",
			},
			wantTarget: map[string]interface{}{
				"plain": []byte("value"),
			},
		},
		{
			name:        "success/empty-format-nested-key",
			description: "验证空格式配置会把点分隔键展开为嵌套 map，并在叶子节点写入原始字节值。",
			giveSrc: &kratosconfig.KeyValue{
				Key:    "database.password",
				Value:  []byte("secret"),
				Format: "",
			},
			wantTarget: map[string]interface{}{
				"database": map[string]interface{}{
					"password": []byte("secret"),
				},
			},
		},
		{
			name:        "success/json-registered-codec",
			description: "验证已注册 JSON codec 能将 JSON 配置成功解码到 target。",
			giveSrc: &kratosconfig.KeyValue{
				Key:    "config",
				Value:  []byte(`{"name":"kit","nested":{"enabled":true}}`),
				Format: "json",
			},
			wantTarget: map[string]interface{}{
				"name": "kit",
				"nested": map[string]interface{}{
					"enabled": true,
				},
			},
		},
		{
			name:          "success/json-resolve-mutates-target-after-nil-option",
			description:   "验证 JSON 解码成功后会调用 Resolve，且 nil DecoderOption 不会阻止后续 Resolve 修改 target。",
			giveNilOption: true,
			giveSrc: &kratosconfig.KeyValue{
				Key:    "config",
				Value:  []byte(`{"name":"kit"}`),
				Format: "json",
			},
			giveResolve: func(target map[string]interface{}) error {
				target["name"] = "resolved"
				target["added"] = "by-resolve"
				return nil
			},
			wantTarget: map[string]interface{}{
				"name":  "resolved",
				"added": "by-resolve",
			},
			wantResolveCalled: true,
		},
		{
			name:        "error/json-resolve-error",
			description: "验证 Resolve 返回错误时 Decode 会透传该错误并停止返回成功结果。",
			giveSrc: &kratosconfig.KeyValue{
				Key:    "config",
				Value:  []byte(`{"name":"kit"}`),
				Format: "json",
			},
			giveResolve: func(target map[string]interface{}) error {
				return errResolve
			},
			wantErr:           true,
			wantErrIs:         errResolve,
			wantResolveCalled: true,
		},
		{
			name:        "error/json-unmarshal",
			description: "验证已注册 codec 反序列化失败时 Decode 返回 codec.Unmarshal 的错误。",
			giveSrc: &kratosconfig.KeyValue{
				Key:    "config",
				Value:  []byte(`{"name":`),
				Format: "json",
			},
			wantErr:         true,
			wantErrContains: "unexpected end of JSON input",
		},
		{
			name:        "error/unsupported-format",
			description: "验证未注册格式会返回包含 key 与 format 的 unsupported format 错误。",
			giveSrc: &kratosconfig.KeyValue{
				Key:    "config",
				Value:  []byte(`{"name":"kit"}`),
				Format: "unsupported",
			},
			wantErr:      true,
			wantErrEqual: "unsupported key: config format: unsupported",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			resolveCalled := false
			options := make([]DecoderOption, 0, 2)
			if tt.giveNilOption {
				options = append(options, nil)
			}
			if tt.giveResolve != nil {
				options = append(options, WithResolve(func(target map[string]interface{}) error {
					resolveCalled = true
					return tt.giveResolve(target)
				}))
			}
			decoder := NewDecoder(options...)
			target := make(map[string]interface{})

			err := decoder.Decode(tt.giveSrc, target)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrIs != nil {
					assert.ErrorIs(t, err, tt.wantErrIs)
				}
				if tt.wantErrEqual != "" {
					assert.EqualError(t, err, tt.wantErrEqual)
				}
				if tt.wantErrContains != "" {
					assert.Contains(t, err.Error(), tt.wantErrContains)
				}
				assert.Equal(t, tt.wantResolveCalled, resolveCalled)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantTarget, target)
			assert.Equal(t, tt.wantResolveCalled, resolveCalled)
		})
	}
}
