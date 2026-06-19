// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package pinyin 提供基于 github.com/mozillazg/go-pinyin 的中文转拼音封装。
//
// 本包保留底层 go-pinyin 的主要风格常量，并通过 NewPinyin 创建轻量转换器。转换器默认
// 使用 Normal 风格、关闭多音字模式，并将底层 Separator 初始化为空格；调用方可通过 Option
// 覆盖 style、heteronym 和 separator 配置。当前公开的 Pinyin 方法返回按可转换字符组织的
// [][]string，style 和 heteronym 会影响候选拼音，separator 对该二维切片结果没有可观察的
// 拼接效果。
//
// 转换结果、非汉字字符的处理规则和多音字候选来源均遵循底层 go-pinyin 的词典与实现。本包不
// 提供错误返回；空输入或不含可转换字符的输入按底层转换行为返回空结果。
package pinyin
