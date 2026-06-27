// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package convert

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type convertAdditionalUser struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// TestToUnsignedFamily_InvalidInputs 验证无符号转换函数会拒绝非法输入并返回零值。
//
// 该测试通过表驱动用例覆盖所有带 error 的无符号转换入口，确保非法字符串不会被静默转换为有效数值。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestToUnsignedFamily_InvalidInputs(t *testing.T) {
	tests := []struct {
		name        string
		description string
		convert     func(any) (any, error)
		want        any
	}{
		{
			name:        "error/uint-invalid-string",
			description: "验证 ToUint 对非数字字符串返回错误和 uint 零值。",
			convert: func(v any) (any, error) {
				return ToUint(v)
			},
			want: uint(0),
		},
		{
			name:        "error/uint8-invalid-string",
			description: "验证 ToUint8 对非数字字符串返回错误和 uint8 零值。",
			convert: func(v any) (any, error) {
				return ToUint8(v)
			},
			want: uint8(0),
		},
		{
			name:        "error/uint16-invalid-string",
			description: "验证 ToUint16 对非数字字符串返回错误和 uint16 零值。",
			convert: func(v any) (any, error) {
				return ToUint16(v)
			},
			want: uint16(0),
		},
		{
			name:        "error/uint32-invalid-string",
			description: "验证 ToUint32 对非数字字符串返回错误和 uint32 零值。",
			convert: func(v any) (any, error) {
				return ToUint32(v)
			},
			want: uint32(0),
		},
		{
			name:        "error/uint64-invalid-string",
			description: "验证 ToUint64 对非数字字符串返回错误和 uint64 零值。",
			convert: func(v any) (any, error) {
				return ToUint64(v)
			},
			want: uint64(0),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := tt.convert("not-a-number")

			require.Error(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestToTimeAndDuration_ErrorPaths 验证时间与时长转换函数对非法输入返回错误。
//
// 该测试显式断言 ToTime 与 ToDuration 的错误分支，避免非法输入用例只声明但未执行转换。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestToTimeAndDuration_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		description string
		convert     func(any) (any, error)
		want        any
	}{
		{
			name:        "error/time-invalid-string",
			description: "验证 ToTime 对无法解析的时间字符串返回错误和零值时间。",
			convert: func(v any) (any, error) {
				return ToTime(v)
			},
			want: time.Time{},
		},
		{
			name:        "error/duration-invalid-string",
			description: "验证 ToDuration 对无法解析的时长字符串返回错误和零值时长。",
			convert: func(v any) (any, error) {
				return ToDuration(v)
			},
			want: time.Duration(0),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := tt.convert("not-a-time-value")

			require.Error(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestStructConversions_ErrorPaths 验证结构体转换函数会报告无效目标和字段转换错误。
//
// 该测试覆盖 ToStruct 与 ToStructs 的错误输入，确保调用方能诊断 nil 目标、非指针目标和字段类型不兼容问题。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestStructConversions_ErrorPaths(t *testing.T) {
	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "error/struct-nil-destination",
			description: "验证 ToStruct 在目标对象为 nil 时返回参数错误。",
			assert: func(t *testing.T) {
				err := ToStruct(map[string]any{"name": "Alice"}, nil)

				require.Error(t, err)
				assert.Contains(t, err.Error(), "pointer")
			},
		},
		{
			name:        "error/struct-non-pointer-destination",
			description: "验证 ToStruct 在目标对象不是指针时返回参数错误。",
			assert: func(t *testing.T) {
				var got convertAdditionalUser

				err := ToStruct(map[string]any{"name": "Alice"}, got)

				require.Error(t, err)
				assert.Contains(t, err.Error(), "destination pointer")
				assert.Equal(t, convertAdditionalUser{}, got)
			},
		},
		{
			name:        "error/struct-field-incompatible-value",
			description: "验证 ToStruct 在字段值无法转换为目标字段类型时返回错误并保留字段零值。",
			assert: func(t *testing.T) {
				var got convertAdditionalUser

				err := ToStruct(map[string]any{"age": make(chan int)}, &got)

				require.Error(t, err)
				assert.Contains(t, err.Error(), "chan int")
				assert.Equal(t, convertAdditionalUser{}, got)
			},
		},
		{
			name:        "error/structs-nil-destination",
			description: "验证 ToStructs 在切片目标为 nil 时返回参数错误。",
			assert: func(t *testing.T) {
				err := ToStructs([]map[string]any{{"name": "Alice"}}, nil)

				require.Error(t, err)
				assert.Contains(t, err.Error(), "pointer")
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

// TestNoErrorWrappers_ErrorFallbacks 验证无 error 包装函数在底层转换失败时返回安全零值。
//
// 该测试通过表驱动用例覆盖布尔、字符串、字节、切片和映射包装函数的错误兜底分支，确保错误不会泄漏为部分转换结果。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNoErrorWrappers_ErrorFallbacks(t *testing.T) {
	unsupportedChannel := make(chan int)
	unsupportedMapValue := map[string]any{"bad": func() {}}
	invalidArrayJSON := `[{]`
	invalidObjectJSON := `{"bad":}`

	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "error/bool-unsupported-channel",
			description: "验证 Bool 在底层 ToBool 无法转换 channel 时返回 false。",
			assert: func(t *testing.T) {
				assert.False(t, Bool(unsupportedChannel))
			},
		},
		{
			name:        "error/string-unsupported-function",
			description: "验证 String 在底层 ToString 无法转换函数时返回空字符串。",
			assert: func(t *testing.T) {
				assert.Empty(t, String(func() {}))
			},
		},
		{
			name:        "error/bytes-unmarshalable-map",
			description: "验证 Bytes 在 map 值无法 JSON 编码时返回 nil。",
			assert: func(t *testing.T) {
				assert.Nil(t, Bytes(unsupportedMapValue))
			},
		},
		{
			name:        "error/runes-unsupported-channel",
			description: "验证 Runes 在底层 ToRunes 无法转换 channel 时返回 nil。",
			assert: func(t *testing.T) {
				assert.Nil(t, Runes(unsupportedChannel))
			},
		},
		{
			name:        "error/slice-int32-unsupported-channel",
			description: "验证 SliceInt32 在元素无法转换为 int32 时返回 nil。",
			assert: func(t *testing.T) {
				assert.Nil(t, SliceInt32(unsupportedChannel))
			},
		},
		{
			name:        "error/slice-int64-unsupported-channel",
			description: "验证 SliceInt64 在元素无法转换为 int64 时返回 nil。",
			assert: func(t *testing.T) {
				assert.Nil(t, SliceInt64(unsupportedChannel))
			},
		},
		{
			name:        "error/slice-uint-unsupported-channel",
			description: "验证 SliceUint 在元素无法转换为 uint 时返回 nil。",
			assert: func(t *testing.T) {
				assert.Nil(t, SliceUint(unsupportedChannel))
			},
		},
		{
			name:        "error/slice-uint32-unsupported-channel",
			description: "验证 SliceUint32 在元素无法转换为 uint32 时返回 nil。",
			assert: func(t *testing.T) {
				assert.Nil(t, SliceUint32(unsupportedChannel))
			},
		},
		{
			name:        "error/slice-uint64-unsupported-channel",
			description: "验证 SliceUint64 在元素无法转换为 uint64 时返回 nil。",
			assert: func(t *testing.T) {
				assert.Nil(t, SliceUint64(unsupportedChannel))
			},
		},
		{
			name:        "error/slice-float32-unsupported-channel",
			description: "验证 SliceFloat32 在元素无法转换为 float32 时返回 nil。",
			assert: func(t *testing.T) {
				assert.Nil(t, SliceFloat32(unsupportedChannel))
			},
		},
		{
			name:        "error/slice-float64-unsupported-channel",
			description: "验证 SliceFloat64 在元素无法转换为 float64 时返回 nil。",
			assert: func(t *testing.T) {
				assert.Nil(t, SliceFloat64(unsupportedChannel))
			},
		},
		{
			name:        "error/slice-str-unsupported-channel",
			description: "验证 SliceStr 在元素无法转换为 string 时返回 nil。",
			assert: func(t *testing.T) {
				assert.Nil(t, SliceStr(unsupportedChannel))
			},
		},
		{
			name:        "error/slice-map-invalid-json",
			description: "验证 SliceMap 在 JSON 数组语法非法时返回 nil。",
			assert: func(t *testing.T) {
				assert.Nil(t, SliceMap(invalidArrayJSON))
			},
		},
		{
			name:        "error/slice-any-map-invalid-json",
			description: "验证 SliceAnyMap 在 JSON 数组语法非法时返回 nil。",
			assert: func(t *testing.T) {
				assert.Nil(t, SliceAnyMap(invalidArrayJSON))
			},
		},
		{
			name:        "error/map-invalid-json",
			description: "验证 Map 在 JSON 对象语法非法时返回 nil。",
			assert: func(t *testing.T) {
				assert.Nil(t, Map(invalidObjectJSON))
			},
		},
		{
			name:        "error/map-str-any-invalid-json",
			description: "验证 MapStrAny 在 JSON 对象语法非法时返回 nil。",
			assert: func(t *testing.T) {
				assert.Nil(t, MapStrAny(invalidObjectJSON))
			},
		},
		{
			name:        "error/map-str-str-unstringable-value",
			description: "验证 MapStrStr 在 map 值无法转换为 string 时返回 nil。",
			assert: func(t *testing.T) {
				assert.Nil(t, MapStrStr(unsupportedMapValue))
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
