// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package bloom 提供可分组的 Bloom Filter 及其存储抽象。
//
// Bloom Filter 只提供 Put 和 Contain 一类追加式操作：Contain 返回 false 时表示元素一定不存在，
// 返回 true 时表示元素可能存在且存在误判，不支持删除后再精确恢复“不存在”语义。
// 本包通过 Bloom name 与 group 派生独立 key，隔离不同位图命名空间。默认内存存储按 key
// 分区并使用 RWMutex 保护并发访问，但会忽略传入的 ctx；Redis 存储会透传 ctx，将派生 key
// 转换为 Redis key，并通过 Lua 脚本批量执行位图读写。
package bloom
