// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package convert

import (
	"time"

	"github.com/gogf/gf/v2/util/gconv"
)

var (
	// converter 是包级 gconv 转换器实例，所有导出转换函数共享它执行实际转换。
	converter = gconv.NewConverter()
)

// ToInt 将 v 按 gconv 规则转换为 int。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式遵循 gconv.Converter.Int。
//
// 返回：
//   - int: 转换成功后的 int 值。
//   - error: v 无法转换为 int 时返回 gconv 产生的错误；成功时为 nil。
func ToInt(v any) (int, error) {
	return converter.Int(v)
}

// ToInt8 将 v 按 gconv 规则转换为 int8。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式遵循 gconv.Converter.Int8。
//
// 返回：
//   - int8: 转换成功后的 int8 值。
//   - error: v 无法转换为 int8 时返回 gconv 产生的错误；成功时为 nil。
func ToInt8(v any) (int8, error) {
	return converter.Int8(v)
}

// ToInt16 将 v 按 gconv 规则转换为 int16。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式遵循 gconv.Converter.Int16。
//
// 返回：
//   - int16: 转换成功后的 int16 值。
//   - error: v 无法转换为 int16 时返回 gconv 产生的错误；成功时为 nil。
func ToInt16(v any) (int16, error) {
	return converter.Int16(v)
}

// ToInt32 将 v 按 gconv 规则转换为 int32。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式遵循 gconv.Converter.Int32。
//
// 返回：
//   - int32: 转换成功后的 int32 值。
//   - error: v 无法转换为 int32 时返回 gconv 产生的错误；成功时为 nil。
func ToInt32(v any) (int32, error) {
	return converter.Int32(v)
}

// ToInt64 将 v 按 gconv 规则转换为 int64。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式遵循 gconv.Converter.Int64。
//
// 返回：
//   - int64: 转换成功后的 int64 值。
//   - error: v 无法转换为 int64 时返回 gconv 产生的错误；成功时为 nil。
func ToInt64(v any) (int64, error) {
	return converter.Int64(v)
}

// ToUint 将 v 转换为 uint，负数输入按 0 处理。
//
// 参数：
//   - v: 待转换的输入值；本函数会先按 gconv 规则解析为 int64。
//
// 返回：
//   - uint: 转换成功后的 uint 值；v 为负数或负数字符串时为 0，非负值按 Go 从 int64 到目标无符号类型的转换规则处理。
//   - error: v 无法解析为 int64 时返回 gconv 产生的错误；负数输入本身不产生错误。
//
// 本函数不额外做目标无符号类型的上界校验。
func ToUint(v any) (uint, error) {
	i, err := converter.Int64(v)
	if err != nil {
		return 0, err
	}
	if i < 0 {
		return 0, nil
	}
	return uint(i), nil
}

// ToUint8 将 v 转换为 uint8，负数输入按 0 处理。
//
// 参数：
//   - v: 待转换的输入值；本函数会先按 gconv 规则解析为 int64。
//
// 返回：
//   - uint8: 转换成功后的 uint8 值；v 为负数或负数字符串时为 0，非负值按 Go 从 int64 到目标无符号类型的转换规则处理。
//   - error: v 无法解析为 int64 时返回 gconv 产生的错误；负数输入本身不产生错误。
//
// 本函数不额外做目标无符号类型的上界校验。
func ToUint8(v any) (uint8, error) {
	i, err := converter.Int64(v)
	if err != nil {
		return 0, err
	}
	if i < 0 {
		return 0, nil
	}
	return uint8(i), nil
}

// ToUint16 将 v 转换为 uint16，负数输入按 0 处理。
//
// 参数：
//   - v: 待转换的输入值；本函数会先按 gconv 规则解析为 int64。
//
// 返回：
//   - uint16: 转换成功后的 uint16 值；v 为负数或负数字符串时为 0，非负值按 Go 从 int64 到目标无符号类型的转换规则处理。
//   - error: v 无法解析为 int64 时返回 gconv 产生的错误；负数输入本身不产生错误。
//
// 本函数不额外做目标无符号类型的上界校验。
func ToUint16(v any) (uint16, error) {
	i, err := converter.Int64(v)
	if err != nil {
		return 0, err
	}
	if i < 0 {
		return 0, nil
	}
	return uint16(i), nil
}

// ToUint32 将 v 转换为 uint32，负数输入按 0 处理。
//
// 参数：
//   - v: 待转换的输入值；本函数会先按 gconv 规则解析为 int64。
//
// 返回：
//   - uint32: 转换成功后的 uint32 值；v 为负数或负数字符串时为 0，非负值按 Go 从 int64 到目标无符号类型的转换规则处理。
//   - error: v 无法解析为 int64 时返回 gconv 产生的错误；负数输入本身不产生错误。
//
// 本函数不额外做目标无符号类型的上界校验。
func ToUint32(v any) (uint32, error) {
	i, err := converter.Int64(v)
	if err != nil {
		return 0, err
	}
	if i < 0 {
		return 0, nil
	}
	return uint32(i), nil
}

// ToUint64 将 v 转换为 uint64，负数输入按 0 处理。
//
// 参数：
//   - v: 待转换的输入值；本函数会先按 gconv 规则解析为 int64。
//
// 返回：
//   - uint64: 转换成功后的 uint64 值；v 为负数或负数字符串时为 0，非负值按 Go 从 int64 到目标无符号类型的转换规则处理。
//   - error: v 无法解析为 int64 时返回 gconv 产生的错误；负数输入本身不产生错误。
//
// 本函数不额外做目标无符号类型的上界校验。
func ToUint64(v any) (uint64, error) {
	i, err := converter.Int64(v)
	if err != nil {
		return 0, err
	}
	if i < 0 {
		return 0, nil
	}
	return uint64(i), nil
}

// ToFloat32 将 v 按 gconv 规则转换为 float32。
//
// 参数：
//   - v: 待转换的输入值；支持的数字、字符串等格式遵循 gconv.Converter.Float32。
//
// 返回：
//   - float32: 转换成功后的 float32 值。
//   - error: v 无法转换为 float32 时返回 gconv 产生的错误；成功时为 nil。
func ToFloat32(v any) (float32, error) {
	return converter.Float32(v)
}

// ToFloat64 将 v 按 gconv 规则转换为 float64。
//
// 参数：
//   - v: 待转换的输入值；支持的数字、字符串等格式遵循 gconv.Converter.Float64。
//
// 返回：
//   - float64: 转换成功后的 float64 值。
//   - error: v 无法转换为 float64 时返回 gconv 产生的错误；成功时为 nil。
func ToFloat64(v any) (float64, error) {
	return converter.Float64(v)
}

// ToBool 将 v 按 gconv 规则转换为 bool。
//
// 参数：
//   - v: 待转换的输入值；字符串、数字和布尔值等输入的解释规则遵循 gconv.Converter.Bool。
//
// 返回：
//   - bool: 转换成功后的 bool 值。
//   - error: v 无法转换为 bool 时返回 gconv 产生的错误；成功时为 nil。
func ToBool(v any) (bool, error) {
	return converter.Bool(v)
}

// ToString 将 v 按 gconv 规则转换为 string。
//
// 参数：
//   - v: 待转换的输入值；基础类型、结构体、map 等输入的字符串化规则遵循 gconv.Converter.String。
//
// 返回：
//   - string: 转换成功后的字符串。
//   - error: v 无法转换为 string 时返回 gconv 产生的错误；成功时为 nil。
func ToString(v any) (string, error) {
	return converter.String(v)
}

// ToTime 将 v 按 gconv 规则转换为 time.Time。
//
// 参数：
//   - v: 待转换的输入值；数值时间戳和可解析时间字符串等格式遵循 gconv.Converter.Time。
//
// 返回：
//   - time.Time: 转换成功后的时间值。
//   - error: v 无法解析为 time.Time 时返回 gconv 产生的错误；成功时为 nil。
func ToTime(v any) (time.Time, error) {
	return converter.Time(v)
}

// ToDuration 将 v 按 gconv 规则转换为 time.Duration。
//
// 参数：
//   - v: 待转换的输入值；时长字符串和数值等格式遵循 gconv.Converter.Duration。
//
// 返回：
//   - time.Duration: 转换成功后的时长值。
//   - error: v 无法解析为 time.Duration 时返回 gconv 产生的错误；成功时为 nil。
func ToDuration(v any) (time.Duration, error) {
	return converter.Duration(v)
}

// ToBytes 将 v 按 gconv 规则转换为字节切片。
//
// 参数：
//   - v: 待转换的输入值；字符串、字节切片和可编码值等输入的处理规则遵循 gconv.Converter.Bytes。
//
// 返回：
//   - []byte: 转换成功后的字节切片。
//   - error: v 无法转换为字节切片时返回 gconv 产生的错误；成功时为 nil。
func ToBytes(v any) ([]byte, error) {
	return converter.Bytes(v)
}

// ToRunes 将 v 按 gconv 规则转换为 rune 切片。
//
// 参数：
//   - v: 待转换的输入值；字符串和其他可转换输入的处理规则遵循 gconv.Converter.Runes。
//
// 返回：
//   - []rune: 转换成功后的 rune 切片。
//   - error: v 无法转换为 rune 切片时返回 gconv 产生的错误；成功时为 nil。
func ToRunes(v any) ([]rune, error) {
	return converter.Runes(v)
}

// ToSliceAny 将 v 按 gconv 规则转换为任意类型切片。
//
// 参数：
//   - v: 待转换的输入值；切片、数组、JSON 数组字符串等输入的处理规则遵循 gconv.Converter.SliceAny。
//
// 返回：
//   - []any: 转换成功后的任意类型切片。
//   - error: v 或其中元素无法按 SliceAny 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToSliceAny(v any) ([]any, error) {
	return converter.SliceAny(v)
}

// ToSlice ToSlice 是 ToSliceAny 的等价入口，将 v 按 gconv 规则转换为any 类型切片。
//
// 参数：
//   - v: 待转换的输入值；切片、数组、JSON 数组字符串等输入的处理规则遵循 gconv.Converter.SliceAny。
//
// 返回：
//   - []any: 转换成功后的any 类型切片。
//   - error: v 或其中元素无法按 SliceAny 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToSlice(v any) ([]any, error) {
	return converter.SliceAny(v)
}

// ToSliceInt 将 v 按 gconv 规则转换为int 类型切片。
//
// 参数：
//   - v: 待转换的输入值；切片、数组、JSON 数组字符串等输入的处理规则遵循 gconv.Converter.SliceInt。
//
// 返回：
//   - []int: 转换成功后的int 类型切片。
//   - error: v 或其中元素无法按 SliceInt 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToSliceInt(v any) ([]int, error) {
	return converter.SliceInt(v)
}

// ToSliceInt32 将 v 按 gconv 规则转换为int32 类型切片。
//
// 参数：
//   - v: 待转换的输入值；切片、数组、JSON 数组字符串等输入的处理规则遵循 gconv.Converter.SliceInt32。
//
// 返回：
//   - []int32: 转换成功后的int32 类型切片。
//   - error: v 或其中元素无法按 SliceInt32 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToSliceInt32(v any) ([]int32, error) {
	return converter.SliceInt32(v)
}

// ToSliceInt64 将 v 按 gconv 规则转换为int64 类型切片。
//
// 参数：
//   - v: 待转换的输入值；切片、数组、JSON 数组字符串等输入的处理规则遵循 gconv.Converter.SliceInt64。
//
// 返回：
//   - []int64: 转换成功后的int64 类型切片。
//   - error: v 或其中元素无法按 SliceInt64 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToSliceInt64(v any) ([]int64, error) {
	return converter.SliceInt64(v)
}

// ToSliceUint 将 v 按 gconv 规则转换为uint 类型切片。
//
// 参数：
//   - v: 待转换的输入值；切片、数组、JSON 数组字符串等输入的处理规则遵循 gconv.Converter.SliceUint。
//
// 返回：
//   - []uint: 转换成功后的uint 类型切片。
//   - error: v 或其中元素无法按 SliceUint 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToSliceUint(v any) ([]uint, error) {
	return converter.SliceUint(v)
}

// ToSliceUint32 将 v 按 gconv 规则转换为uint32 类型切片。
//
// 参数：
//   - v: 待转换的输入值；切片、数组、JSON 数组字符串等输入的处理规则遵循 gconv.Converter.SliceUint32。
//
// 返回：
//   - []uint32: 转换成功后的uint32 类型切片。
//   - error: v 或其中元素无法按 SliceUint32 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToSliceUint32(v any) ([]uint32, error) {
	return converter.SliceUint32(v)
}

// ToSliceUint64 将 v 按 gconv 规则转换为uint64 类型切片。
//
// 参数：
//   - v: 待转换的输入值；切片、数组、JSON 数组字符串等输入的处理规则遵循 gconv.Converter.SliceUint64。
//
// 返回：
//   - []uint64: 转换成功后的uint64 类型切片。
//   - error: v 或其中元素无法按 SliceUint64 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToSliceUint64(v any) ([]uint64, error) {
	return converter.SliceUint64(v)
}

// ToSliceFloat32 将 v 按 gconv 规则转换为float32 类型切片。
//
// 参数：
//   - v: 待转换的输入值；切片、数组、JSON 数组字符串等输入的处理规则遵循 gconv.Converter.SliceFloat32。
//
// 返回：
//   - []float32: 转换成功后的float32 类型切片。
//   - error: v 或其中元素无法按 SliceFloat32 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToSliceFloat32(v any) ([]float32, error) {
	return converter.SliceFloat32(v)
}

// ToSliceFloat64 将 v 按 gconv 规则转换为float64 类型切片。
//
// 参数：
//   - v: 待转换的输入值；切片、数组、JSON 数组字符串等输入的处理规则遵循 gconv.Converter.SliceFloat64。
//
// 返回：
//   - []float64: 转换成功后的float64 类型切片。
//   - error: v 或其中元素无法按 SliceFloat64 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToSliceFloat64(v any) ([]float64, error) {
	return converter.SliceFloat64(v)
}

// ToSliceStr 将 v 按 gconv 规则转换为string 类型切片。
//
// 参数：
//   - v: 待转换的输入值；切片、数组、JSON 数组字符串等输入的处理规则遵循 gconv.Converter.SliceStr。
//
// 返回：
//   - []string: 转换成功后的string 类型切片。
//   - error: v 或其中元素无法按 SliceStr 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToSliceStr(v any) ([]string, error) {
	return converter.SliceStr(v)
}

// ToSliceMap 将 v 按 gconv 规则转换为map[string]any 类型切片。
//
// 参数：
//   - v: 待转换的输入值；切片、数组、JSON 数组字符串等输入的处理规则遵循 gconv.Converter.SliceMap。
//
// 返回：
//   - []map[string]any: 转换成功后的map[string]any 类型切片。
//   - error: v 或其中元素无法按 SliceMap 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToSliceMap(v any) ([]map[string]any, error) {
	return converter.SliceMap(v)
}

// ToSliceAnyMap 将 v 按 gconv 规则转换为map[string]any 类型切片。
//
// 参数：
//   - v: 待转换的输入值；切片、数组、JSON 数组字符串等输入的处理规则遵循 gconv.Converter.SliceMap。
//
// 返回：
//   - []map[string]any: 转换成功后的map[string]any 类型切片。
//   - error: v 或其中元素无法按 SliceMap 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToSliceAnyMap(v any) ([]map[string]any, error) {
	maps, err := converter.SliceMap(v)
	return maps, err
}

// ToMap 将 v 按 gconv 规则转换为 map[string]any。
//
// 参数：
//   - v: 待转换的输入值；结构体、map 和 JSON 对象字符串等输入的处理规则遵循 gconv.Converter.Map。
//
// 返回：
//   - map[string]any: 转换成功后的 map[string]any。
//   - error: v 或其中键值无法按 Map 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToMap(v any) (map[string]any, error) {
	return converter.Map(v)
}

// ToMapStrAny ToMapStrAny 是 ToMap 的等价入口，将 v 按 gconv 规则转换为 map[string]any。
//
// 参数：
//   - v: 待转换的输入值；结构体、map 和 JSON 对象字符串等输入的处理规则遵循 gconv.Converter.Map。
//
// 返回：
//   - map[string]any: 转换成功后的 map[string]any。
//   - error: v 或其中键值无法按 Map 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToMapStrAny(v any) (map[string]any, error) {
	return converter.Map(v)
}

// ToMapStrStr 将 v 按 gconv 规则转换为 map[string]string。
//
// 参数：
//   - v: 待转换的输入值；结构体、map 和 JSON 对象字符串等输入的处理规则遵循 gconv.Converter.MapStrStr。
//
// 返回：
//   - map[string]string: 转换成功后的 map[string]string。
//   - error: v 或其中键值无法按 MapStrStr 规则转换时返回 gconv 产生的错误；成功时为 nil。
func ToMapStrStr(v any) (map[string]string, error) {
	return converter.MapStrStr(v)
}

// ToStruct 将 v 按 gconv 规则填充到 out 指向的结构体。
//
// 参数：
//   - v: 待转换的输入值；map、结构体和 JSON 对象字符串等输入的处理规则遵循 gconv.Converter.Struct。
//   - out: 接收转换结果的结构体指针，必须为非 nil 指针。
//
// 返回：
//   - error: out 不是有效目标指针、字段无法转换或 v 无法映射到目标结构体时返回 gconv 产生的错误；成功时为 nil。
func ToStruct(v any, out any) error {
	return converter.Struct(v, out)
}

// ToStructs 将 v 按 gconv 规则填充到 out 指向的结构体切片。
//
// 参数：
//   - v: 待转换的输入值；切片、数组和 JSON 数组字符串等输入的处理规则遵循 gconv.Converter.Structs。
//   - out: 接收转换结果的结构体切片指针，必须为非 nil 指针。
//
// 返回：
//   - error: out 不是有效目标指针、元素字段无法转换或 v 无法映射到目标切片时返回 gconv 产生的错误；成功时为 nil。
func ToStructs(v any, out any) error {
	return converter.Structs(v, out)
}

// Int 将 v 转换为 int 值，底层转换失败时返回 0。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToInt 相同。
//
// 返回：
//   - int: 转换成功后的int 值；ToInt 返回错误时为 0。
//
// 调用方需要区分真实 0与转换失败时，应使用 ToInt。
func Int(v any) int {
	val, err := ToInt(v)
	if err != nil {
		return 0
	}
	return val
}

// Int8 将 v 转换为 int8 值，底层转换失败时返回 0。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToInt8 相同。
//
// 返回：
//   - int8: 转换成功后的int8 值；ToInt8 返回错误时为 0。
//
// 调用方需要区分真实 0与转换失败时，应使用 ToInt8。
func Int8(v any) int8 {
	val, err := ToInt8(v)
	if err != nil {
		return 0
	}
	return val
}

// Int16 将 v 转换为 int16 值，底层转换失败时返回 0。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToInt16 相同。
//
// 返回：
//   - int16: 转换成功后的int16 值；ToInt16 返回错误时为 0。
//
// 调用方需要区分真实 0与转换失败时，应使用 ToInt16。
func Int16(v any) int16 {
	val, err := ToInt16(v)
	if err != nil {
		return 0
	}
	return val
}

// Int32 将 v 转换为 int32 值，底层转换失败时返回 0。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToInt32 相同。
//
// 返回：
//   - int32: 转换成功后的int32 值；ToInt32 返回错误时为 0。
//
// 调用方需要区分真实 0与转换失败时，应使用 ToInt32。
func Int32(v any) int32 {
	val, err := ToInt32(v)
	if err != nil {
		return 0
	}
	return val
}

// Int64 将 v 转换为 int64 值，底层转换失败时返回 0。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToInt64 相同。
//
// 返回：
//   - int64: 转换成功后的int64 值；ToInt64 返回错误时为 0。
//
// 调用方需要区分真实 0与转换失败时，应使用 ToInt64。
func Int64(v any) int64 {
	val, err := ToInt64(v)
	if err != nil {
		return 0
	}
	return val
}

// Uint 将 v 转换为 uint 值，底层转换失败时返回 0。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToUint 相同。
//
// 返回：
//   - uint: 转换成功后的uint 值；ToUint 返回错误时为 0。
//
// 调用方需要区分真实 0与转换失败时，应使用 ToUint。
func Uint(v any) uint {
	val, err := ToUint(v)
	if err != nil {
		return 0
	}
	return val
}

// Uint8 将 v 转换为 uint8 值，底层转换失败时返回 0。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToUint8 相同。
//
// 返回：
//   - uint8: 转换成功后的uint8 值；ToUint8 返回错误时为 0。
//
// 调用方需要区分真实 0与转换失败时，应使用 ToUint8。
func Uint8(v any) uint8 {
	val, err := ToUint8(v)
	if err != nil {
		return 0
	}
	return val
}

// Uint16 将 v 转换为 uint16 值，底层转换失败时返回 0。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToUint16 相同。
//
// 返回：
//   - uint16: 转换成功后的uint16 值；ToUint16 返回错误时为 0。
//
// 调用方需要区分真实 0与转换失败时，应使用 ToUint16。
func Uint16(v any) uint16 {
	val, err := ToUint16(v)
	if err != nil {
		return 0
	}
	return val
}

// Uint32 将 v 转换为 uint32 值，底层转换失败时返回 0。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToUint32 相同。
//
// 返回：
//   - uint32: 转换成功后的uint32 值；ToUint32 返回错误时为 0。
//
// 调用方需要区分真实 0与转换失败时，应使用 ToUint32。
func Uint32(v any) uint32 {
	val, err := ToUint32(v)
	if err != nil {
		return 0
	}
	return val
}

// Uint64 将 v 转换为 uint64 值，底层转换失败时返回 0。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToUint64 相同。
//
// 返回：
//   - uint64: 转换成功后的uint64 值；ToUint64 返回错误时为 0。
//
// 调用方需要区分真实 0与转换失败时，应使用 ToUint64。
func Uint64(v any) uint64 {
	val, err := ToUint64(v)
	if err != nil {
		return 0
	}
	return val
}

// Float32 将 v 转换为 float32 值，底层转换失败时返回 0。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToFloat32 相同。
//
// 返回：
//   - float32: 转换成功后的float32 值；ToFloat32 返回错误时为 0。
//
// 调用方需要区分真实 0与转换失败时，应使用 ToFloat32。
func Float32(v any) float32 {
	val, err := ToFloat32(v)
	if err != nil {
		return 0
	}
	return val
}

// Float64 将 v 转换为 float64 值，底层转换失败时返回 0。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToFloat64 相同。
//
// 返回：
//   - float64: 转换成功后的float64 值；ToFloat64 返回错误时为 0。
//
// 调用方需要区分真实 0与转换失败时，应使用 ToFloat64。
func Float64(v any) float64 {
	val, err := ToFloat64(v)
	if err != nil {
		return 0
	}
	return val
}

// Bool 将 v 转换为 bool 值，底层转换失败时返回 false。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToBool 相同。
//
// 返回：
//   - bool: 转换成功后的bool 值；ToBool 返回错误时为 false。
//
// 调用方需要区分真实 false与转换失败时，应使用 ToBool。
func Bool(v any) bool {
	val, err := ToBool(v)
	if err != nil {
		return false
	}
	return val
}

// String 将 v 转换为 string 值，底层转换失败时返回空字符串。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToString 相同。
//
// 返回：
//   - string: 转换成功后的string 值；ToString 返回错误时为空字符串。
//
// 调用方需要区分真实空字符串与转换失败时，应使用 ToString。
func String(v any) string {
	val, err := ToString(v)
	if err != nil {
		return ""
	}
	return val
}

// Time 将 v 转换为 time.Time 值，底层转换失败时返回零值时间。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToTime 相同。
//
// 返回：
//   - time.Time: 转换成功后的time.Time 值；ToTime 返回错误时为零值时间。
//
// 调用方需要区分真实零值时间与转换失败时，应使用 ToTime。
func Time(v any) time.Time {
	val, err := ToTime(v)
	if err != nil {
		return time.Time{}
	}
	return val
}

// Duration 将 v 转换为 time.Duration 值，底层转换失败时返回 0。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToDuration 相同。
//
// 返回：
//   - time.Duration: 转换成功后的time.Duration 值；ToDuration 返回错误时为 0。
//
// 调用方需要区分真实 0与转换失败时，应使用 ToDuration。
func Duration(v any) time.Duration {
	val, err := ToDuration(v)
	if err != nil {
		return 0
	}
	return val
}

// Bytes 将 v 转换为 字节切片，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToBytes 相同。
//
// 返回：
//   - []byte: 转换成功后的字节切片；ToBytes 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToBytes。
func Bytes(v any) []byte {
	val, err := ToBytes(v)
	if err != nil {
		return nil
	}
	return val
}

// Runes 将 v 转换为 rune 切片，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToRunes 相同。
//
// 返回：
//   - []rune: 转换成功后的rune 切片；ToRunes 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToRunes。
func Runes(v any) []rune {
	val, err := ToRunes(v)
	if err != nil {
		return nil
	}
	return val
}

// SliceAny 将 v 转换为 any 类型切片，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToSliceAny 相同。
//
// 返回：
//   - []any: 转换成功后的any 类型切片；ToSliceAny 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToSliceAny。
func SliceAny(v any) []any {
	val, err := ToSliceAny(v)
	if err != nil {
		return nil
	}
	return val
}

// Slice 将 v 转换为 any 类型切片，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToSlice 相同。
//
// 返回：
//   - []any: 转换成功后的any 类型切片；ToSlice 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToSlice。
func Slice(v any) []any {
	val, err := ToSlice(v)
	if err != nil {
		return nil
	}
	return val
}

// SliceInt 将 v 转换为 int 类型切片，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToSliceInt 相同。
//
// 返回：
//   - []int: 转换成功后的int 类型切片；ToSliceInt 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToSliceInt。
func SliceInt(v any) []int {
	val, err := ToSliceInt(v)
	if err != nil {
		return nil
	}
	return val
}

// SliceInt32 将 v 转换为 int32 类型切片，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToSliceInt32 相同。
//
// 返回：
//   - []int32: 转换成功后的int32 类型切片；ToSliceInt32 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToSliceInt32。
func SliceInt32(v any) []int32 {
	val, err := ToSliceInt32(v)
	if err != nil {
		return nil
	}
	return val
}

// SliceInt64 将 v 转换为 int64 类型切片，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToSliceInt64 相同。
//
// 返回：
//   - []int64: 转换成功后的int64 类型切片；ToSliceInt64 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToSliceInt64。
func SliceInt64(v any) []int64 {
	val, err := ToSliceInt64(v)
	if err != nil {
		return nil
	}
	return val
}

// SliceUint 将 v 转换为 uint 类型切片，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToSliceUint 相同。
//
// 返回：
//   - []uint: 转换成功后的uint 类型切片；ToSliceUint 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToSliceUint。
func SliceUint(v any) []uint {
	val, err := ToSliceUint(v)
	if err != nil {
		return nil
	}
	return val
}

// SliceUint32 将 v 转换为 uint32 类型切片，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToSliceUint32 相同。
//
// 返回：
//   - []uint32: 转换成功后的uint32 类型切片；ToSliceUint32 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToSliceUint32。
func SliceUint32(v any) []uint32 {
	val, err := ToSliceUint32(v)
	if err != nil {
		return nil
	}
	return val
}

// SliceUint64 将 v 转换为 uint64 类型切片，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToSliceUint64 相同。
//
// 返回：
//   - []uint64: 转换成功后的uint64 类型切片；ToSliceUint64 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToSliceUint64。
func SliceUint64(v any) []uint64 {
	val, err := ToSliceUint64(v)
	if err != nil {
		return nil
	}
	return val
}

// SliceFloat32 将 v 转换为 float32 类型切片，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToSliceFloat32 相同。
//
// 返回：
//   - []float32: 转换成功后的float32 类型切片；ToSliceFloat32 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToSliceFloat32。
func SliceFloat32(v any) []float32 {
	val, err := ToSliceFloat32(v)
	if err != nil {
		return nil
	}
	return val
}

// SliceFloat64 将 v 转换为 float64 类型切片，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToSliceFloat64 相同。
//
// 返回：
//   - []float64: 转换成功后的float64 类型切片；ToSliceFloat64 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToSliceFloat64。
func SliceFloat64(v any) []float64 {
	val, err := ToSliceFloat64(v)
	if err != nil {
		return nil
	}
	return val
}

// SliceStr 将 v 转换为 string 类型切片，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToSliceStr 相同。
//
// 返回：
//   - []string: 转换成功后的string 类型切片；ToSliceStr 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToSliceStr。
func SliceStr(v any) []string {
	val, err := ToSliceStr(v)
	if err != nil {
		return nil
	}
	return val
}

// SliceMap 将 v 转换为 map[string]any 类型切片，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToSliceMap 相同。
//
// 返回：
//   - []map[string]any: 转换成功后的map[string]any 类型切片；ToSliceMap 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToSliceMap。
func SliceMap(v any) []map[string]any {
	val, err := ToSliceMap(v)
	if err != nil {
		return nil
	}
	return val
}

// SliceAnyMap 将 v 转换为 map[string]any 类型切片，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToSliceAnyMap 相同。
//
// 返回：
//   - []map[string]any: 转换成功后的map[string]any 类型切片；ToSliceAnyMap 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToSliceAnyMap。
func SliceAnyMap(v any) []map[string]any {
	val, err := ToSliceAnyMap(v)
	if err != nil {
		return nil
	}
	return val
}

// Map 将 v 转换为 map[string]any，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToMap 相同。
//
// 返回：
//   - map[string]any: 转换成功后的map[string]any；ToMap 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToMap。
func Map(v any) map[string]any {
	val, err := ToMap(v)
	if err != nil {
		return nil
	}
	return val
}

// MapStrAny 将 v 转换为 map[string]any，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToMapStrAny 相同。
//
// 返回：
//   - map[string]any: 转换成功后的map[string]any；ToMapStrAny 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToMapStrAny。
func MapStrAny(v any) map[string]any {
	val, err := ToMapStrAny(v)
	if err != nil {
		return nil
	}
	return val
}

// MapStrStr 将 v 转换为 map[string]string，底层转换失败时返回 nil。
//
// 参数：
//   - v: 待转换的输入值；支持的源类型和格式与 ToMapStrStr 相同。
//
// 返回：
//   - map[string]string: 转换成功后的map[string]string；ToMapStrStr 返回错误时为 nil。
//
// 调用方需要区分真实 nil与转换失败时，应使用 ToMapStrStr。
func MapStrStr(v any) map[string]string {
	val, err := ToMapStrStr(v)
	if err != nil {
		return nil
	}
	return val
}
