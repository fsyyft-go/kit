// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package convert

import (
	"time"

	"github.com/gogf/gf/v2/util/gconv"
)

var (
	// converter 是全局的类型转换器实例，用于统一处理各种类型的转换。
	converter = gconv.NewConverter()
)

// ToInt 将任意类型 v 转换为 int 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - int：转换后的 int 类型结果。
//   - error：转换过程中发生的错误。
func ToInt(v any) (int, error) {
	return converter.Int(v)
}

// ToInt8 将任意类型 v 转换为 int8 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - int8：转换后的 int8 类型结果。
//   - error：转换过程中发生的错误。
func ToInt8(v any) (int8, error) {
	return converter.Int8(v)
}

// ToInt16 将任意类型 v 转换为 int16 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - int16：转换后的 int16 类型结果。
//   - error：转换过程中发生的错误。
func ToInt16(v any) (int16, error) {
	return converter.Int16(v)
}

// ToInt32 将任意类型 v 转换为 int32 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - int32：转换后的 int32 类型结果。
//   - error：转换过程中发生的错误。
func ToInt32(v any) (int32, error) {
	return converter.Int32(v)
}

// ToInt64 将任意类型 v 转换为 int64 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - int64：转换后的 int64 类型结果。
//   - error：转换过程中发生的错误。
func ToInt64(v any) (int64, error) {
	return converter.Int64(v)
}

// ToUint 将任意类型 v 转换为 uint 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - uint：转换后的 uint 类型结果。
//   - error：转换过程中发生的错误。
func ToUint(v any) (uint, error) {
	return converter.Uint(v)
}

// ToUint8 将任意类型 v 转换为 uint8 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - uint8：转换后的 uint8 类型结果。
//   - error：转换过程中发生的错误。
func ToUint8(v any) (uint8, error) {
	return converter.Uint8(v)
}

// ToUint16 将任意类型 v 转换为 uint16 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - uint16：转换后的 uint16 类型结果。
//   - error：转换过程中发生的错误。
func ToUint16(v any) (uint16, error) {
	return converter.Uint16(v)
}

// ToUint32 将任意类型 v 转换为 uint32 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - uint32：转换后的 uint32 类型结果。
//   - error：转换过程中发生的错误。
func ToUint32(v any) (uint32, error) {
	return converter.Uint32(v)
}

// ToUint64 将任意类型 v 转换为 uint64 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - uint64：转换后的 uint64 类型结果。
//   - error：转换过程中发生的错误。
func ToUint64(v any) (uint64, error) {
	return converter.Uint64(v)
}

// ToFloat32 将任意类型 v 转换为 float32 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - float32：转换后的 float32 类型结果。
//   - error：转换过程中发生的错误。
func ToFloat32(v any) (float32, error) {
	return converter.Float32(v)
}

// ToFloat64 将任意类型 v 转换为 float64 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - float64：转换后的 float64 类型结果。
//   - error：转换过程中发生的错误。
func ToFloat64(v any) (float64, error) {
	return converter.Float64(v)
}

// ToBool 将任意类型 v 转换为 bool 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - bool：转换后的 bool 类型结果。
//   - error：转换过程中发生的错误。
func ToBool(v any) (bool, error) {
	return converter.Bool(v)
}

// ToString 将任意类型 v 转换为 string 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - string：转换后的 string 类型结果。
//   - error：转换过程中发生的错误。
func ToString(v any) (string, error) {
	return converter.String(v)
}

// ToTime 将任意类型 v 转换为 time.Time 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - time.Time：转换后的 time.Time 类型结果。
//   - error：转换过程中发生的错误。
func ToTime(v any) (time.Time, error) {
	return converter.Time(v)
}

// ToDuration 将任意类型 v 转换为 time.Duration 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - time.Duration：转换后的 time.Duration 类型结果。
//   - error：转换过程中发生的错误。
func ToDuration(v any) (time.Duration, error) {
	return converter.Duration(v)
}

// ToBytes 将任意类型 v 转换为字节切片。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - []byte：转换后的字节切片。
//   - error：转换过程中发生的错误。
func ToBytes(v any) ([]byte, error) {
	return converter.Bytes(v)
}

// ToRunes 将任意类型 v 转换为 rune 切片。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - []rune：转换后的 rune 切片。
//   - error：转换过程中发生的错误。
func ToRunes(v any) ([]rune, error) {
	return converter.Runes(v)
}

// ToSliceAny 将任意类型 v 转换为 any 类型切片。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - []any：转换后的 any 类型切片。
//   - error：转换过程中发生的错误。
func ToSliceAny(v any) ([]any, error) {
	return converter.SliceAny(v)
}

// ToSlice 将任意类型 v 转换为 interface{} 类型切片。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - []interface{}：转换后的 interface{} 类型切片。
//   - error：转换过程中发生的错误。
func ToSlice(v any) ([]any, error) {
	return converter.SliceAny(v)
}

// ToSliceInt 将任意类型 v 转换为 int 类型切片。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - []int：转换后的 int 类型切片。
//   - error：转换过程中发生的错误。
func ToSliceInt(v any) ([]int, error) {
	return converter.SliceInt(v)
}

// ToSliceInt32 将任意类型 v 转换为 int32 类型切片。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - []int32：转换后的 int32 类型切片。
//   - error：转换过程中发生的错误。
func ToSliceInt32(v any) ([]int32, error) {
	return converter.SliceInt32(v)
}

// ToSliceInt64 将任意类型 v 转换为 int64 类型切片。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - []int64：转换后的 int64 类型切片。
//   - error：转换过程中发生的错误。
func ToSliceInt64(v any) ([]int64, error) {
	return converter.SliceInt64(v)
}

// ToSliceUint 将任意类型 v 转换为 uint 类型切片。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - []uint：转换后的 uint 类型切片。
//   - error：转换过程中发生的错误。
func ToSliceUint(v any) ([]uint, error) {
	return converter.SliceUint(v)
}

// ToSliceUint32 将任意类型 v 转换为 uint32 类型切片。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - []uint32：转换后的 uint32 类型切片。
//   - error：转换过程中发生的错误。
func ToSliceUint32(v any) ([]uint32, error) {
	return converter.SliceUint32(v)
}

// ToSliceUint64 将任意类型 v 转换为 uint64 类型切片。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - []uint64：转换后的 uint64 类型切片。
//   - error：转换过程中发生的错误。
func ToSliceUint64(v any) ([]uint64, error) {
	return converter.SliceUint64(v)
}

// ToSliceFloat32 将任意类型 v 转换为 float32 类型切片。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - []float32：转换后的 float32 类型切片。
//   - error：转换过程中发生的错误。
func ToSliceFloat32(v any) ([]float32, error) {
	return converter.SliceFloat32(v)
}

// ToSliceFloat64 将任意类型 v 转换为 float64 类型切片。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - []float64：转换后的 float64 类型切片。
//   - error：转换过程中发生的错误。
func ToSliceFloat64(v any) ([]float64, error) {
	return converter.SliceFloat64(v)
}

// ToSliceStr 将任意类型 v 转换为 string 类型切片。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - []string：转换后的 string 类型切片。
//   - error：转换过程中发生的错误。
func ToSliceStr(v any) ([]string, error) {
	return converter.SliceStr(v)
}

// ToSliceMap 将任意类型 v 转换为 map[string]interface{} 类型切片。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - []map[string]any：转换后的 map[string]any 类型切片。
//   - error：转换过程中发生的错误。
func ToSliceMap(v any) ([]map[string]any, error) {
	return converter.SliceMap(v)
}

// ToSliceAnyMap 将任意类型 v 转换为 map[string]any 类型切片。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - []map[string]any：转换后的 map[string]any 类型切片。
//   - error：转换过程中发生的错误。
func ToSliceAnyMap(v any) ([]map[string]any, error) {
	maps, err := converter.SliceMap(v)
	return maps, err
}

// ToMap 将任意类型 v 转换为 map[string]interface{} 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - map[string]any：转换后的 map[string]any 类型。
//   - error：转换过程中发生的错误。
func ToMap(v any) (map[string]any, error) {
	return converter.Map(v)
}

// ToMapStrAny 将任意类型 v 转换为 map[string]interface{} 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - map[string]any：转换后的 map[string]any 类型。
//   - error：转换过程中发生的错误。
func ToMapStrAny(v any) (map[string]any, error) {
	return converter.Map(v)
}

// ToMapStrStr 将任意类型 v 转换为 map[string]string 类型。
//
// 参数：
//   - v：待转换的任意类型。
//
// 返回值：
//   - map[string]string：转换后的 map[string]string 类型。
//   - error：转换过程中发生的错误。
func ToMapStrStr(v any) (map[string]string, error) {
	return converter.MapStrStr(v)
}

// ToStruct 将任意类型 v 转换为结构体，结果存储到 out 指针指向的结构体中。
//
// 参数：
//   - v：待转换的任意类型。
//   - out：结构体指针，转换结果存储于此。
//
// 返回值：
//   - error：转换过程中发生的错误。
func ToStruct(v any, out any) error {
	return converter.Struct(v, out)
}

// ToStructs 将任意类型 v 转换为结构体切片，结果存储到 out 指针指向的切片中。
//
// 参数：
//   - v：待转换的任意类型。
//   - out：结构体切片指针，转换结果存储于此。
//
// 返回值：
//   - error：转换过程中发生的错误。
func ToStructs(v any, out any) error {
	return converter.Structs(v, out)
}
