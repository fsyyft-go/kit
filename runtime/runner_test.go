// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package runtime

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// compileTimeRunnerInterfaceAssertion 在编译期验证 compileTimeRunner 满足 Runner 接口。
var _ Runner = (*compileTimeRunner)(nil)

// compileTimeRunner 是用于验证 Runner 方法集的最小测试实现。
//
// 该类型不承载运行时行为，仅为编译期接口断言提供 Start 和 Stop 方法。
type compileTimeRunner struct{}

// Start 满足 Runner 接口的启动方法签名。
//
// 参数：
//   - ctx: 启动阶段传入的上下文，本测试实现仅用于签名验证。
//
// 返回：
//   - error: 固定返回 nil，表示测试实现不模拟启动错误。
func (*compileTimeRunner) Start(ctx context.Context) error { return nil }

// Stop 满足 Runner 接口的停止方法签名。
//
// 参数：
//   - ctx: 停止阶段传入的上下文，本测试实现仅用于签名验证。
//
// 返回：
//   - error: 固定返回 nil，表示测试实现不模拟停止错误。
func (*compileTimeRunner) Stop(ctx context.Context) error { return nil }

// TestRunner_MethodSet 验证 Runner 接口仅暴露 Start 与 Stop 生命周期方法签名。
//
// 该测试检查接口方法集本身；具体 Runner 实现的启动、停止和错误处理语义应由实现类型的测试覆盖。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRunner_MethodSet(t *testing.T) {
	runnerType := reflect.TypeOf((*Runner)(nil)).Elem()
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType := reflect.TypeOf((*error)(nil)).Elem()

	assert.Equal(t, 2, runnerType.NumMethod())

	tests := []struct {
		name           string
		description    string
		giveMethodName string
	}{
		{
			name:           "method-set/start-context-error",
			description:    "验证 Runner.Start 接收 context.Context 参数并返回 error。",
			giveMethodName: "Start",
		},
		{
			name:           "method-set/stop-context-error",
			description:    "验证 Runner.Stop 接收 context.Context 参数并返回 error。",
			giveMethodName: "Stop",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			method, ok := runnerType.MethodByName(tt.giveMethodName)
			require.True(t, ok)
			require.Equal(t, 1, method.Type.NumIn())
			require.Equal(t, contextType, method.Type.In(0))
			require.Equal(t, 1, method.Type.NumOut())
			assert.Equal(t, errorType, method.Type.Out(0))
		})
	}
}
