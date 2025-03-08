// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// package runtime 提供了应用程序运行时管理的基础设施。
package runtime

import (
	"context"
)

// Runner 定义了可运行组件的接口。
// 实现此接口的组件可以被统一管理其生命周期。
type Runner interface {
	// Start 启动组件并开始处理。
	// ctx 提供生命周期控制和取消信号。
	// 返回：处理过程中可能发生的错误。
	Start(ctx context.Context) error

	// Stop 优雅地停止组件。
	// ctx 提供停止操作的截止时间。
	// 返回：停止过程中可能发生的错误。
	Stop(ctx context.Context) error
}
