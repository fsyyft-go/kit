// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package net 提供测试网络相关用例时使用的环境变量辅助函数。
//
// 本包当前用于读取 TEST_NETWORKS 环境变量，并据此决定是否启用依赖外部网络的测试。
// 它不封装通用网络客户端或协议实现，调用方通常只在测试代码中使用这些辅助函数。
package net
