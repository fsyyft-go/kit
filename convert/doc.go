// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package convert 提供基于 github.com/gogf/gf/v2/util/gconv 的常用类型转换封装。
//
// 本包覆盖数值、布尔、字符串、time、slice、map 以及 struct 填充等常见转换场景。
// ToXxx 系列函数返回转换结果和 error，调用方可以据此区分真实零值和转换失败；Xxx
// 简写函数会吞掉错误并返回对应零值，适合输入来源可控或允许使用兜底值的路径。
//
// 具体输入格式、结构体标签处理和错误信息由 gconv 当前实现决定。本包额外约定无符号整数
// 转换会先按 int64 解析，负数或负数字符串返回 0 且不产生错误。
package convert
