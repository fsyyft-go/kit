// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package md5

import (
	"crypto/md5"
	"fmt"
	"io"
)

// writeString 默认指向 io.WriteString，测试可替换它以覆盖写入错误分支。
var writeString = io.WriteString

// HashStringWithoutError 返回 source 的 MD5 小写十六进制摘要，并忽略底层写入错误。
//
// 本函数适合调用方已接受“失败时返回空字符串”语义的场景；需要区分空字符串输入摘要和写入失败时，
// 应改用 HashString。
//
// 参数：
//   - source: 需要计算摘要的源字符串，按原始字节序列参与 MD5 计算。
//
// 返回：
//   - string: source 的 MD5 小写十六进制摘要；若底层写入失败则返回空字符串。
func HashStringWithoutError(source string) string {
	result, _ := HashString(source)
	return result
}

// HashString 计算 source 的 MD5 摘要。
//
// 参数：
//   - source: 需要计算摘要的源字符串，按原始字节序列参与 MD5 计算。
//
// 返回：
//   - string: source 的 MD5 小写十六进制摘要；若写入哈希状态失败则返回空字符串。
//   - error: 写入哈希状态失败时返回底层错误；成功时为 nil。
func HashString(source string) (string, error) {
	var result string
	var err error

	// 创建一个新的 MD5 哈希对象；每次调用使用独立状态，避免跨调用共享哈希缓冲。
	w := md5.New()
	// 将源字符串写入哈希对象，并检查是否发生错误；writeString 默认不会对内存哈希写入失败，测试会替换它以覆盖错误分支。
	if _, err = writeString(w, source); nil == err {
		// 计算哈希值并转换为十六进制字符串；Sum(nil) 生成当前摘要，%x 保持小写输出。
		result = fmt.Sprintf("%x", w.Sum(nil))
	} else {
		// 如果发生错误，将结果设置为空字符串，避免返回可能包含部分写入状态的不完整摘要。
		result = ""
	}

	// 返回计算结果和可能的错误，并将底层写入错误原样交给调用方处理。
	return result, err
}
