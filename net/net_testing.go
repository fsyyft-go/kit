// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package net

import (
	"os"
)

const (
	// EnvTestNetwork 是控制网络相关测试是否启用的环境变量名。
	//
	// 当 TEST_NETWORKS 为任意非空字符串时，[TestNetwork] 返回 true。
	EnvTestNetwork = "TEST_NETWORKS"
)

// TestNetwork 判断是否启用依赖外部网络的测试。
//
// 当 [TestNetworkValue] 返回非空字符串时，TestNetwork 返回 true。
//
// 参数：无。
//
// 返回：
//   - bool: true 表示启用网络相关测试；false 表示跳过这类测试。
func TestNetwork() bool {
	needTest := len(TestNetworkValue()) > 0
	return needTest
}

// TestNetworkValue 返回 TEST_NETWORKS 环境变量的原始值。
//
// 未设置该环境变量时返回空字符串。
//
// 参数：无。
//
// 返回：
//   - string: TEST_NETWORKS 的原始环境变量值。
func TestNetworkValue() string {
	return os.Getenv(EnvTestNetwork)
}
