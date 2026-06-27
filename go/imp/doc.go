// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package imp 提供 Go import 声明的分组、排序和别名规范检查。
//
// Check 会递归遍历目录下的 Go 文件，解析 import 声明，并按内置包、第三方包、
// github.com/fsyyft-go/kit 包以及项目内包分组检查字母序。对于带别名的导入，本包还会
// 校验 kit 和 app 前缀约定，以及别名仅由小写字母和数字组成。
//
// 检查结果以问题字符串切片返回；当文件遍历或语法解析失败时，Check 返回 error 并中止
// 本次检查。
package imp
