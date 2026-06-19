// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package runtime

import (
	"context"
)

type (
	// Runner 定义具备显式 Start 和 Stop 生命周期的组件契约。
	//
	// Runner 仅约定启动与停止入口，不强制规定实现的并发安全、幂等性、阻塞
	// 方式或资源管理策略；调用方应结合具体实现文档决定调用时机与错误处理方式。
	Runner interface {
		// Start 启动组件并开始其运行流程。
		//
		// 参数：
		//   - ctx：用于传递启动过程的取消信号、截止时间和生命周期边界。
		//
		// 返回：
		//   - error：实现方无法完成启动，或 ctx 导致启动流程中止时返回错误。具体错误类型和调用方处理方式由实现方定义。
		Start(ctx context.Context) error

		// Stop 请求组件停止并执行必要的清理流程。
		//
		// 参数：
		//   - ctx：用于约束停止过程的取消信号、截止时间和清理边界。
		//
		// 返回：
		//   - error：实现方无法完成停止或清理，或 ctx 导致停止流程中止时返回错误。具体错误类型和调用方处理方式由实现方定义。
		Stop(ctx context.Context) error
	}
)
