// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package convert 提供基于 github.com/gogf/gf/v2/util/gconv 的常用类型转换封装。
//
// ToXxx 系列函数返回转换结果和 error，覆盖数值、布尔、字符串、time、slice、map 以及
// struct 填充等场景；Xxx 简写函数在转换失败时返回对应零值，便于在已知输入可控的路径
// 上使用。
//
// 对无符号整数转换，本包会先按有符号整数解析；当输入为负数或负数字符串时返回 0，而
// 不进行符号溢出转换。
package convert
