// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// convert/convert_test.go
//
// 设计思路：
// 本测试文件采用表格驱动法，针对 convert 包所有导出方法进行单元测试，涵盖常见类型、边界值、错误场景。
// 断言使用 stretchr/testify，确保类型转换的正确性和健壮性。
//
// 使用方法：
// go test ./convert -v -cover
//
// 依赖：
//   - github.com/stretchr/testify/assert
//
// 每个测试用例均有详细注释，便于理解和维护。

package convert

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 测试 ToInt, ToInt8, ToInt16, ToInt32, ToInt64
func TestToIntFamily(t *testing.T) {
	type testCase struct {
		name    string
		input   any
		expects map[string]any // 期望的各类型结果
		err     bool
	}

	tests := []testCase{
		{
			name:    "int string",
			input:   "123",
			expects: map[string]any{"int": 123, "int8": int8(123), "int16": int16(123), "int32": int32(123), "int64": int64(123)},
			err:     false,
		},
		{
			name:    "float string",
			input:   "123.9",
			expects: map[string]any{"int": 123, "int8": int8(123), "int16": int16(123), "int32": int32(123), "int64": int64(123)},
			err:     false,
		},
		{
			name:    "bool true",
			input:   true,
			expects: map[string]any{"int": 1, "int8": int8(1), "int16": int16(1), "int32": int32(1), "int64": int64(1)},
			err:     false,
		},
		{
			name:    "invalid string",
			input:   "abc",
			expects: nil,
			err:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v, err := ToInt(tc.input)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expects["int"], v)
			}

			v8, err := ToInt8(tc.input)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expects["int8"], v8)
			}

			v16, err := ToInt16(tc.input)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expects["int16"], v16)
			}

			v32, err := ToInt32(tc.input)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expects["int32"], v32)
			}

			v64, err := ToInt64(tc.input)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expects["int64"], v64)
			}
		})
	}
}

// 测试 ToUint, ToUint8, ToUint16, ToUint32, ToUint64
func TestToUintFamily(t *testing.T) {
	type testCase struct {
		name    string
		input   any
		expects map[string]any
		err     bool
	}

	tests := []testCase{
		{
			name:    "uint string",
			input:   "123",
			expects: map[string]any{"uint": uint(123), "uint8": uint8(123), "uint16": uint16(123), "uint32": uint32(123), "uint64": uint64(123)},
			err:     false,
		},
		{
			name:    "negative string",
			input:   "-1",
			expects: map[string]any{"uint": uint(0), "uint8": uint8(0), "uint16": uint16(0), "uint32": uint32(0), "uint64": uint64(0)},
			err:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v, err := ToUint(tc.input)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expects["uint"], v)
			}

			v8, err := ToUint8(tc.input)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expects["uint8"], v8)
			}

			v16, err := ToUint16(tc.input)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expects["uint16"], v16)
			}

			v32, err := ToUint32(tc.input)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expects["uint32"], v32)
			}

			v64, err := ToUint64(tc.input)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expects["uint64"], v64)
			}
		})
	}
}

// 测试 ToFloat32, ToFloat64
func TestToFloatFamily(t *testing.T) {
	type testCase struct {
		name    string
		input   any
		expects map[string]any
		err     bool
	}

	tests := []testCase{
		{
			name:    "float string",
			input:   "123.456",
			expects: map[string]any{"float32": float32(123.456), "float64": float64(123.456)},
			err:     false,
		},
		{
			name:    "int string",
			input:   "789",
			expects: map[string]any{"float32": float32(789), "float64": float64(789)},
			err:     false,
		},
		{
			name:    "invalid string",
			input:   "abc",
			expects: nil,
			err:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v32, err := ToFloat32(tc.input)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tc.expects["float32"], v32, 1e-5)
			}

			v64, err := ToFloat64(tc.input)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tc.expects["float64"], v64, 1e-9)
			}
		})
	}
}

// 测试 ToBool
func TestToBool(t *testing.T) {
	type testCase struct {
		name   string
		input  any
		expect bool
		err    bool
	}

	tests := []testCase{
		{"true string", "true", true, false},
		{"false string", "false", false, false},
		{"1 string", "1", true, false},
		{"0 string", "0", false, false},
		{"invalid string", "abc", true, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v, err := ToBool(tc.input)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expect, v)
			}
		})
	}
}

// 测试 ToString
func TestToString(t *testing.T) {
	type testCase struct {
		name   string
		input  any
		expect string
		err    bool
	}

	tests := []testCase{
		{"int", 123, "123", false},
		{"float", 123.45, "123.45", false},
		{"bool", true, "true", false},
		{"string", "abc", "abc", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v, err := ToString(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, v)
		})
	}
}

// 测试 ToTime, ToDuration
func TestToTimeAndDuration(t *testing.T) {
	type testCase struct {
		name      string
		input     any
		expectDur time.Duration
		expectTm  time.Time
		err       bool
	}

	tm := time.Now()
	dur := 2 * time.Hour

	tests := []testCase{
		{"duration string", "2h", dur, time.Time{}, false},
		{"timestamp int", tm.Unix(), 0, time.Unix(tm.Unix(), 0), false},
		{"invalid", "abc", 0, time.Time{}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectDur != 0 {
				d, err := ToDuration(tc.input)
				if tc.err {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tc.expectDur, d)
				}
			}
			if !tc.expectTm.IsZero() {
				tm2, err := ToTime(tc.input)
				if tc.err {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tc.expectTm, tm2)
				}
			}
		})
	}
}

// 测试 ToBytes, ToRunes
func TestToBytesAndRunes(t *testing.T) {
	// 测试字符串转字节和 rune
	b, err := ToBytes("abc")
	assert.NoError(t, err)
	assert.Equal(t, []byte{'a', 'b', 'c'}, b)

	r, err := ToRunes("abc")
	assert.NoError(t, err)
	assert.Equal(t, []rune{'a', 'b', 'c'}, r)
}

// 测试 ToSliceAny, ToSlice, ToSliceInt, ToSliceInt32, ToSliceInt64, ToSliceUint, ToSliceUint32, ToSliceUint64, ToSliceFloat32, ToSliceFloat64, ToSliceStr
func TestToSliceFamily(t *testing.T) {
	input := []any{"1", 2, 3.0}

	// any 切片
	v, err := ToSliceAny(input)
	assert.NoError(t, err)
	assert.Equal(t, input, v)

	// int 切片
	vi, err := ToSliceInt([]any{"1", 2, 3})
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, vi)

	// int32 切片
	vi32, err := ToSliceInt32([]any{"1", 2, 3})
	assert.NoError(t, err)
	assert.Equal(t, []int32{1, 2, 3}, vi32)

	// int64 切片
	vi64, err := ToSliceInt64([]any{"1", 2, 3})
	assert.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, vi64)

	// uint 切片
	vu, err := ToSliceUint([]any{"1", 2, 3})
	assert.NoError(t, err)
	assert.Equal(t, []uint{1, 2, 3}, vu)

	// uint32 切片
	vu32, err := ToSliceUint32([]any{"1", 2, 3})
	assert.NoError(t, err)
	assert.Equal(t, []uint32{1, 2, 3}, vu32)

	// uint64 切片
	vu64, err := ToSliceUint64([]any{"1", 2, 3})
	assert.NoError(t, err)
	assert.Equal(t, []uint64{1, 2, 3}, vu64)

	// float32 切片
	vf32, err := ToSliceFloat32([]any{"1.1", 2, 3})
	assert.NoError(t, err)
	assert.InDeltaSlice(t, []float32{1.1, 2, 3}, vf32, 1e-5)

	// float64 切片
	vf64, err := ToSliceFloat64([]any{"1.1", 2, 3})
	assert.NoError(t, err)
	assert.InDeltaSlice(t, []float64{1.1, 2, 3}, vf64, 1e-9)

	// string 切片
	vs, err := ToSliceStr([]any{1, 2, 3})
	assert.NoError(t, err)
	assert.Equal(t, []string{"1", "2", "3"}, vs)
}

// 测试 ToSliceMap, ToSliceAnyMap
func TestToSliceMapFamily(t *testing.T) {
	input := []map[string]any{
		{"a": 1},
		{"b": 2},
	}
	v, err := ToSliceMap(input)
	assert.NoError(t, err)
	assert.Equal(t, input, v)

	v2, err := ToSliceAnyMap(input)
	assert.NoError(t, err)
	assert.Equal(t, input, v2)
}

// 测试 ToMap, ToMapStrAny, ToMapStrStr
func TestToMapFamily(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	user := User{Name: "John", Age: 28}

	m, err := ToMap(user)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"name": "John", "age": 28}, m)

	m2, err := ToMapStrAny(user)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"name": "John", "age": 28}, m2)

	m3, err := ToMapStrStr(map[string]any{"name": "John", "age": 28})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"name": "John", "age": "28"}, m3)
}

// 测试 ToStruct, ToStructs
func TestToStructFamily(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	m := map[string]any{"name": "Alice", "age": 30}
	var user User
	err := ToStruct(m, &user)
	assert.NoError(t, err)
	assert.Equal(t, User{Name: "Alice", Age: 30}, user)

	// 测试切片
	ms := []map[string]any{
		{"name": "Tom", "age": 30},
		{"name": "Jerry", "age": 25},
	}
	var users []User
	err = ToStructs(ms, &users)
	assert.NoError(t, err)
	assert.Equal(t, []User{{"Tom", 30}, {"Jerry", 25}}, users)
}
