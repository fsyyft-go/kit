// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package runtime 定义具备显式 Start 和 Stop 生命周期的公共接口。
//
// Runner 约定组件通过 Start(ctx) 启动、通过 Stop(ctx) 停止；ctx 用于传递取消信号、
// 超时和清理边界。
// 本包当前不提供调度器、工作池或监督器实现，具体的并发安全、幂等性和阻塞语义由实现方文档约定。
package runtime
