// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package net

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTestNetwork_EnvironmentContract 验证网络测试开关环境变量的判定契约。
//
// 该测试通过表驱动用例覆盖未设置、空值和非空值场景，确保 TestNetworkValue 返回原始环境变量值，
// TestNetwork 仅依据 TEST_NETWORKS 是否为非空字符串启用网络测试。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestTestNetwork_EnvironmentContract(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		giveEnvSet      bool
		giveEnvValue    string
		wantValue       string
		wantTestNetwork bool
	}{
		{
			name:            "boundary/unset-env",
			description:     "验证 TEST_NETWORKS 未设置时返回空字符串并关闭网络测试开关。",
			giveEnvSet:      false,
			wantValue:       "",
			wantTestNetwork: false,
		},
		{
			name:            "boundary/empty-env",
			description:     "验证 TEST_NETWORKS 设置为空字符串时仍关闭网络测试开关。",
			giveEnvSet:      true,
			giveEnvValue:    "",
			wantValue:       "",
			wantTestNetwork: false,
		},
		{
			name:            "success/enabled-flag",
			description:     "验证 TEST_NETWORKS 设置为常用启用标记时返回原始值并开启网络测试开关。",
			giveEnvSet:      true,
			giveEnvValue:    "1",
			wantValue:       "1",
			wantTestNetwork: true,
		},
		{
			name:            "success/non-empty-literal",
			description:     "验证 TEST_NETWORKS 设置为任意非空字符串时按历史契约开启网络测试开关。",
			giveEnvSet:      true,
			giveEnvValue:    "false",
			wantValue:       "false",
			wantTestNetwork: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			setTestNetworkEnv(t, tt.giveEnvSet, tt.giveEnvValue)

			assert.Equal(t, tt.wantValue, TestNetworkValue())
			assert.Equal(t, tt.wantTestNetwork, TestNetwork())
		})
	}
}

// setTestNetworkEnv 为单个测试用例设置 TEST_NETWORKS 环境变量并在用例结束后恢复。
//
// 该辅助函数集中处理环境变量隔离，确保测试不会污染调用者已有的 TEST_NETWORKS 状态。
//
// 参数：
//   - t: 测试上下文，用于注册清理逻辑并标记辅助函数调用栈。
//   - giveSet: 是否将 TEST_NETWORKS 设置为指定值；为 false 时清除该环境变量。
//   - giveValue: 当 giveSet 为 true 时写入 TEST_NETWORKS 的环境变量值。
func setTestNetworkEnv(t *testing.T, giveSet bool, giveValue string) {
	t.Helper()

	oldValue, oldSet := os.LookupEnv(EnvTestNetwork)
	t.Cleanup(func() {
		if oldSet {
			require.NoError(t, os.Setenv(EnvTestNetwork, oldValue))
			return
		}
		require.NoError(t, os.Unsetenv(EnvTestNetwork))
	})

	if giveSet {
		require.NoError(t, os.Setenv(EnvTestNetwork, giveValue))
		return
	}
	require.NoError(t, os.Unsetenv(EnvTestNetwork))
}
