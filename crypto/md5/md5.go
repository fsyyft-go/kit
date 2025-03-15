// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package md5

import (
	"crypto/md5"
	"fmt"
	"io"
)

// writeString 是对 io.WriteString 的封装，便于测试时替换。
var writeString = io.WriteString

// HashStringWithoutError 计算字符串的 MD5 哈希值，忽略可能发生的错误。
// 该函数是 HashString 的简化版本，适用于确定不会发生错误的场景。
//
// 参数：
//   - source：需要计算哈希值的源字符串。
//
// 返回值：
//   - string：计算得到的 MD5 哈希值的十六进制字符串表示。
func HashStringWithoutError(source string) string {
	result, _ := HashString(source)
	return result
}

// HashString 计算字符串的 MD5 哈希值，并返回可能发生的错误。
//
// 参数：
//   - source：需要计算哈希值的源字符串。
//
// 返回值：
//   - string：计算得到的 MD5 哈希值的十六进制字符串表示。
//   - error：操作过程中可能发生的错误。
func HashString(source string) (string, error) {
	var result string
	var err error

	// 创建一个新的 MD5 哈希对象。
	w := md5.New()
	// 将源字符串写入哈希对象，并检查是否发生错误。
	if _, err = writeString(w, source); nil == err {
		// 计算哈希值并转换为十六进制字符串。
		result = fmt.Sprintf("%x", w.Sum(nil))
	} else {
		// 如果发生错误，将结果设置为空字符串。
		result = ""
	}

	// 返回计算结果和可能的错误。
	return result, err
}
