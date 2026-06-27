// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package testing 提供用于在测试中写入带统一前缀日志的辅助函数。
//
// 本包封装 fmt.Println 和 fmt.Printf 的标准输出行为，在每次调用的正文前
// 添加固定日志前缀，便于从测试输出中识别辅助日志。输出格式、换行规则和
// 格式化诊断遵循 fmt 包语义；本包不接管 *testing.T 日志，也不保证并发调用时
// 单条日志记录以原子方式写入。
package testing
