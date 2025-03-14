// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package runtime

import (
	"context"
)

type (
	// Runner 定义了可运行组件的接口。
	// 实现此接口的组件可以被统一管理其生命周期。
	Runner interface {
		// Start 启动组件并开始处理。
		//
		// 参数：
		//   - ctx：提供生命周期控制和取消信号。
		//
		// 返回值：
		//   - error：返回处理过程中可能发生的错误。
		Start(ctx context.Context) error

		// Stop 优雅地停止组件。
		//
		// 参数：
		//   - ctx：提供停止操作的截止时间。
		//
		// 返回值：
		//   - error：返回停止过程中可能发生的错误。
		Stop(ctx context.Context) error
	}
)
