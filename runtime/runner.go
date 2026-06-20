// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package runtime

import (
	"context"
)

type (
	// Runner 定义具备显式 Start/Stop 生命周期的组件接口。
	// 具体实现的幂等性、并发安全和阻塞语义由各实现自行约定。
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
