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

// Int/Int8/Int16/Int32/Int64 无 error 版本测试
func TestIntFamily_NoError(t *testing.T) {
	type testCase struct {
		name    string
		input   any
		expects map[string]any
	}

	tests := []testCase{
		{
			name:    "int string",
			input:   "123",
			expects: map[string]any{"int": 123, "int8": int8(123), "int16": int16(123), "int32": int32(123), "int64": int64(123)},
		},
		{
			name:    "float string",
			input:   "123.9",
			expects: map[string]any{"int": 123, "int8": int8(123), "int16": int16(123), "int32": int32(123), "int64": int64(123)},
		},
		{
			name:    "bool true",
			input:   true,
			expects: map[string]any{"int": 1, "int8": int8(1), "int16": int16(1), "int32": int32(1), "int64": int64(1)},
		},
		{
			name:    "invalid string",
			input:   "abc",
			expects: map[string]any{"int": 0, "int8": int8(0), "int16": int16(0), "int32": int32(0), "int64": int64(0)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expects["int"], Int(tc.input))
			assert.Equal(t, tc.expects["int8"], Int8(tc.input))
			assert.Equal(t, tc.expects["int16"], Int16(tc.input))
			assert.Equal(t, tc.expects["int32"], Int32(tc.input))
			assert.Equal(t, tc.expects["int64"], Int64(tc.input))
		})
	}
}

// Uint/Uint8/Uint16/Uint32/Uint64 无 error 版本测试
func TestUintFamily_NoError(t *testing.T) {
	type testCase struct {
		name    string
		input   any
		expects map[string]any
	}

	tests := []testCase{
		{
			name:    "uint string",
			input:   "123",
			expects: map[string]any{"uint": uint(123), "uint8": uint8(123), "uint16": uint16(123), "uint32": uint32(123), "uint64": uint64(123)},
		},
		{
			name:    "negative string",
			input:   "-1",
			expects: map[string]any{"uint": uint(0), "uint8": uint8(0), "uint16": uint16(0), "uint32": uint32(0), "uint64": uint64(0)},
		},
		{
			name:    "invalid string",
			input:   "abc",
			expects: map[string]any{"uint": uint(0), "uint8": uint8(0), "uint16": uint16(0), "uint32": uint32(0), "uint64": uint64(0)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expects["uint"], Uint(tc.input))
			assert.Equal(t, tc.expects["uint8"], Uint8(tc.input))
			assert.Equal(t, tc.expects["uint16"], Uint16(tc.input))
			assert.Equal(t, tc.expects["uint32"], Uint32(tc.input))
			assert.Equal(t, tc.expects["uint64"], Uint64(tc.input))
		})
	}
}

// Float32/Float64 无 error 版本测试
func TestFloatFamily_NoError(t *testing.T) {
	type testCase struct {
		name    string
		input   any
		expects map[string]any
	}

	tests := []testCase{
		{
			name:    "float string",
			input:   "123.456",
			expects: map[string]any{"float32": float32(123.456), "float64": float64(123.456)},
		},
		{
			name:    "int string",
			input:   "789",
			expects: map[string]any{"float32": float32(789), "float64": float64(789)},
		},
		{
			name:    "invalid string",
			input:   "abc",
			expects: map[string]any{"float32": float32(0), "float64": float64(0)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.InDelta(t, tc.expects["float32"], Float32(tc.input), 1e-5)
			assert.InDelta(t, tc.expects["float64"], Float64(tc.input), 1e-9)
		})
	}
}

// Bool 无 error 版本测试
func TestBool_NoError(t *testing.T) {
	type testCase struct {
		name   string
		input  any
		expect bool
	}

	tests := []testCase{
		{"true string", "true", true},
		{"false string", "false", false},
		{"1 string", "1", true},
		{"0 string", "0", false},
		{"invalid string", "abc", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, Bool(tc.input))
		})
	}
}

// String 无 error 版本测试
func TestString_NoError(t *testing.T) {
	type testCase struct {
		name   string
		input  any
		expect string
	}

	tests := []testCase{
		{"int", 123, "123"},
		{"float", 123.45, "123.45"},
		{"bool", true, "true"},
		{"string", "abc", "abc"},
		// 注意：struct{} 转换为字符串时，gconv 默认行为为 "{}"，不是空字符串。
		{"invalid", struct{}{}, "{}"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expect, String(tc.input))
		})
	}
}

// Time/Duration 无 error 版本测试
func TestTimeAndDuration_NoError(t *testing.T) {
	tm := time.Now()
	dur := 2 * time.Hour

	t.Run("duration string", func(t *testing.T) {
		assert.Equal(t, dur, Duration("2h"))
	})
	t.Run("timestamp int", func(t *testing.T) {
		assert.Equal(t, time.Unix(tm.Unix(), 0), Time(tm.Unix()))
	})
	// 错误输入应返回零值
	t.Run("invalid", func(t *testing.T) {
		assert.Equal(t, time.Time{}, Time("abc"))
		assert.Equal(t, time.Duration(0), Duration("abc"))
	})
}

// Bytes/Runes 无 error 版本测试
func TestBytesAndRunes_NoError(t *testing.T) {
	// 正常输入
	assert.Equal(t, []byte{'a', 'b', 'c'}, Bytes("abc"))
	assert.Equal(t, []rune{'a', 'b', 'c'}, Runes("abc"))
	// gconv.Bytes(123) 实际返回 []byte{'{'}, 即 123 的 ASCII 字符
	assert.Equal(t, []byte{'{'}, Bytes(123))
	// gconv.Runes(123) 实际返回 []rune{'1','2','3'}，即 "123" 的 rune 切片
	assert.Equal(t, []rune{'1', '2', '3'}, Runes(123))
}

// SliceAny/Slice/SliceInt/SliceInt32/SliceInt64/SliceUint/SliceUint32/SliceUint64/SliceFloat32/SliceFloat64/SliceStr 无 error 版本测试
func TestSliceFamily_NoError(t *testing.T) {
	input := []any{"1", 2, 3.0}
	assert.Equal(t, input, SliceAny(input))
	assert.Equal(t, input, Slice(input))
	assert.Equal(t, []int{1, 2, 3}, SliceInt([]any{"1", 2, 3}))
	assert.Equal(t, []int32{1, 2, 3}, SliceInt32([]any{"1", 2, 3}))
	assert.Equal(t, []int64{1, 2, 3}, SliceInt64([]any{"1", 2, 3}))
	assert.Equal(t, []uint{1, 2, 3}, SliceUint([]any{"1", 2, 3}))
	assert.Equal(t, []uint32{1, 2, 3}, SliceUint32([]any{"1", 2, 3}))
	assert.Equal(t, []uint64{1, 2, 3}, SliceUint64([]any{"1", 2, 3}))
	assert.InDeltaSlice(t, []float32{1.1, 2, 3}, SliceFloat32([]any{"1.1", 2, 3}), 1e-5)
	assert.InDeltaSlice(t, []float64{1.1, 2, 3}, SliceFloat64([]any{"1.1", 2, 3}), 1e-9)
	assert.Equal(t, []string{"1", "2", "3"}, SliceStr([]any{1, 2, 3}))
	// 错误输入
	assert.Nil(t, SliceInt([]any{"a", "b"}))
}

// SliceMap/SliceAnyMap 无 error 版本测试
func TestSliceMapFamily_NoError(t *testing.T) {
	input := []map[string]any{{"a": 1}, {"b": 2}}
	assert.Equal(t, input, SliceMap(input))
	assert.Equal(t, input, SliceAnyMap(input))
	// gconv.SliceMap([]any{"a", "b"}) 实际返回 [nil, nil]，不是 nil
	assert.Equal(t, []map[string]any{nil, nil}, SliceMap([]any{"a", "b"}))
}

// Map/MapStrAny/MapStrStr 无 error 版本测试
func TestMapFamily_NoError(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	user := User{Name: "John", Age: 28}
	assert.Equal(t, map[string]any{"name": "John", "age": 28}, Map(user))
	assert.Equal(t, map[string]any{"name": "John", "age": 28}, MapStrAny(user))
	assert.Equal(t, map[string]string{"name": "John", "age": "28"}, MapStrStr(map[string]any{"name": "John", "age": 28}))
	// gconv.Map(struct{ X int }{}) 实际返回 map[string]interface{}{"X":0}，不是 nil
	assert.Equal(t, map[string]any{"X": 0}, Map(struct{ X int }{}))
}
