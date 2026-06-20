// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package pinyin 提供对 github.com/mozillazg/go-pinyin 的简洁封装。
//
// NewPinyin 使用默认的 Normal 风格、非多音字模式和空格分隔符创建转换器；调用方可通
// 过 WithStyle、WithHeteronym 和 WithSeparator 调整输出行为。包级拼音风格常量直接
// 映射到底层 go-pinyin 的对应配置。
//
// Pinyin 接口将输入文本转换为 [][]string，以保留多音字的多个候选读音。非汉字字符的
// 处理行为遵循底层 go-pinyin 库。
package pinyin
