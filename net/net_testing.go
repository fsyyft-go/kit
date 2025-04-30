// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// 声明 net 包，提供与网络相关的工具函数。
package net

import (
	"os"
)

const (
	// EnvTestNetwork 是环境变量 TEST_NETWORKS 的常量名。
	// 用于控制是否执行网络相关的单元测试。
	EnvTestNetwork = "TEST_NETWORKS"
)

// TestNetwork 判断是否需要进行网络相关的单元测试。
// 通过调用 TestNetworkValue 获取测试网络相关的环境变量值，
// 若环境变量值非空，则返回 true，表示需要进行网络相关的单元测试。
// 返回值：
//   - bool 类型，true 表示需要进行网络相关的单元测试，false 表示不需要。
func TestNetwork() bool {
	// needTest 变量用于存储是否需要进行网络相关单元测试的判断结果。
	needTest := len(TestNetworkValue()) > 0
	return needTest
}

// TestNetworkValue 获取测试网络相关的环境变量值。
// 该函数通过 os.Getenv 获取名为 EnvTestNetwork 的环境变量值，
// 若未设置该环境变量，则返回空字符串。
// 返回值：
//   - string 类型，表示 TEST_NETWORKS 环境变量的值。
func TestNetworkValue() string {
	return os.Getenv(EnvTestNetwork)
}
