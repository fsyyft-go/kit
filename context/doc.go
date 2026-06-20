// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package context 提供对标准库 context 的轻量补充。
//
// 本包当前实现 WithoutCancel，用于在保留父 context 值的同时断开取消、截止时间和错误
// 状态。该行为与 Go 1.21 引入的 context.WithoutCancel 一致，适合将请求级元数据继续传
// 递给不应受上游取消影响的后台流程。
package context
