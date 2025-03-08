// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// 内部测试包，可以访问 md5 包的内部实现。
// 该文件使用与被测试包相同的包名，可以直接访问未导出的函数和变量。
package md5

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 替换 io.WriteString 的模拟函数，始终返回错误。
func mockWriteStringWithError(w io.Writer, s string) (int, error) {
	return 0, errors.New("模拟的写入错误")
}

// 模拟写入部分数据后返回错误的函数。
func mockWriteStringWithPartialWrite(w io.Writer, s string) (int, error) {
	if len(s) > 0 {
		return 1, errors.New("模拟的部分写入错误")
	}
	return 0, nil
}

// 测试 HashString 函数在写入错误时的行为。
func TestHashStringWriteError(t *testing.T) {
	// 保存原始的 io.WriteString 函数，以便测试后恢复。
	originalWriteString := writeString

	// 替换 writeString 为我们的模拟函数。
	writeString = mockWriteStringWithError

	// 测试结束后恢复原始函数。
	defer func() {
		writeString = originalWriteString
	}()

	// 调用 HashString 函数，期望返回错误。
	result, err := HashString("test")

	// 断言结果为空字符串，且错误不为 nil。
	assert.Equal(t, "", result, "当写入失败时，应返回空字符串。")
	assert.Error(t, err, "当写入失败时，应返回错误。")
	assert.Contains(t, err.Error(), "模拟的写入错误", "错误消息应包含原始错误信息。")
}

// 测试 HashString 函数在部分写入后发生错误的情况。
func TestHashStringPartialWriteError(t *testing.T) {
	// 保存原始的 io.WriteString 函数，以便测试后恢复。
	originalWriteString := writeString

	// 替换 writeString 为我们的模拟函数。
	writeString = mockWriteStringWithPartialWrite

	// 测试结束后恢复原始函数。
	defer func() {
		writeString = originalWriteString
	}()

	// 调用 HashString 函数，期望返回错误。
	result, err := HashString("test")

	// 断言结果为空字符串，且错误不为 nil。
	assert.Equal(t, "", result, "当部分写入失败时，应返回空字符串。")
	assert.Error(t, err, "当部分写入失败时，应返回错误。")
	assert.Contains(t, err.Error(), "模拟的部分写入错误", "错误消息应包含原始错误信息。")

	// 测试空字符串情况下不会出错。
	result, err = HashString("")
	assert.NoError(t, err, "对于空字符串，不应返回错误。")
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", result, "空字符串的MD5哈希值不匹配。")
}

// 测试 HashStringWithoutError 在内部错误时的行为。
func TestHashStringWithoutErrorHandlesError(t *testing.T) {
	// 保存原始的 io.WriteString 函数，以便测试后恢复。
	originalWriteString := writeString

	// 替换 writeString 为我们的模拟函数。
	writeString = mockWriteStringWithError

	// 测试结束后恢复原始函数。
	defer func() {
		writeString = originalWriteString
	}()

	// 调用 HashStringWithoutError 函数。
	result := HashStringWithoutError("test")

	// 断言结果为空字符串，因为内部的 HashString 返回了错误。
	assert.Equal(t, "", result, "当内部写入失败时，应返回空字符串。")
}

// 测试 HashStringWithoutError 在部分写入错误的情况下的行为。
func TestHashStringWithoutErrorHandlesPartialWriteError(t *testing.T) {
	// 保存原始的 io.WriteString 函数，以便测试后恢复。
	originalWriteString := writeString

	// 替换 writeString 为我们的模拟函数。
	writeString = mockWriteStringWithPartialWrite

	// 测试结束后恢复原始函数。
	defer func() {
		writeString = originalWriteString
	}()

	// 调用 HashStringWithoutError 函数，非空字符串应该触发错误。
	result := HashStringWithoutError("test")
	assert.Equal(t, "", result, "当内部部分写入失败时，应返回空字符串。")

	// 对于空字符串，不应触发错误。
	result = HashStringWithoutError("")
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", result, "空字符串的MD5哈希值不匹配。")
}

// 测试 HashString 和 HashStringWithoutError 的一致性。
func TestHashFunctionConsistency(t *testing.T) {
	// 保存原始的 io.WriteString 函数，以便测试后恢复。
	originalWriteString := writeString

	// 测试结束后恢复原始函数。
	defer func() {
		writeString = originalWriteString
	}()

	// 测试正常情况下两个函数返回一致的结果。
	tests := []string{
		"",
		"hello world",
		"12345",
		"你好，世界",
		"!@#$%^&*()_+",
	}

	for _, test := range tests {
		resultWithError, err := HashString(test)
		assert.NoError(t, err, "HashString不应返回错误: %v", test)

		resultWithoutError := HashStringWithoutError(test)
		assert.Equal(t, resultWithError, resultWithoutError, "两个函数返回的结果应该一致: %v", test)
	}
}
