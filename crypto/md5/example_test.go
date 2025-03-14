// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package md5_test 提供了对 md5 包功能的示例。
package md5_test

import (
	"fmt"

	"github.com/fsyyft-go/kit/crypto/md5"
)

// Example 展示了一个基本的 MD5 计算示例。
func Example() {
	// 计算一个简单字符串的 MD5 哈希值，忽略错误。
	hash := md5.HashStringWithoutError("hello world")
	fmt.Println(hash)
	// Output: 5eb63bbbe01eeed093cb22bb8f5acdc3
}

// ExampleHashString 展示了如何使用 HashString 函数，包括错误处理。
func ExampleHashString() {
	// 计算字符串的 MD5 哈希值，并处理可能的错误。
	hash, err := md5.HashString("hello world")
	if err != nil {
		fmt.Printf("计算哈希值时发生错误: %v\n", err)
		return
	}
	fmt.Println(hash)
	// Output: 5eb63bbbe01eeed093cb22bb8f5acdc3
}

// ExampleHashStringWithoutError 展示了如何使用 HashStringWithoutError 函数。
func ExampleHashStringWithoutError() {
	// 计算一个简单字符串的 MD5 哈希值，忽略错误。
	hash := md5.HashStringWithoutError("hello world")
	fmt.Println(hash)
	// Output: 5eb63bbbe01eeed093cb22bb8f5acdc3
}

// ExampleHashStringWithoutError_empty 展示了对空字符串计算 MD5 哈希值的情况。
func ExampleHashStringWithoutError_empty() {
	// 计算空字符串的 MD5 哈希值，忽略错误。
	hash := md5.HashStringWithoutError("")
	fmt.Println(hash)
	// Output: d41d8cd98f00b204e9800998ecf8427e
}

// ExampleHashString_chinese 展示了对中文字符串计算 MD5 哈希值的示例。
func ExampleHashString_chinese() {
	// 计算中文字符串的 MD5 哈希值，并处理可能的错误。
	hash, err := md5.HashString("你好，世界")
	if err != nil {
		fmt.Printf("计算哈希值时发生错误: %v\n", err)
		return
	}
	fmt.Println(hash)
	// Output: dbefd3ada018615b35588a01e216ae6e
}
