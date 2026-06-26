// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package context 提供对标准库 context 的轻量补充。
//
// 本包提供 WithoutCancel，用于在保留父 context 值的同时断开取消、截止时间和错误
// 状态。返回的 context 不继承父 context 的 Deadline、Done 和 Err，只通过 Value 委托
// 读取父 context 链上的值。该行为与标准库 context.WithoutCancel 一致，适合将请求级
// 元数据传递给不应受上游取消影响的后台流程。
package context
