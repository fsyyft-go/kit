// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package runtime 定义具备显式 Start 和 Stop 生命周期的运行组件接口。
//
// 本包围绕 Runner 抽象约定组件的启动与停止流程。Start 和 Stop 均接收
// context.Context，用于传递取消信号、超时和清理边界。
// 本包当前只提供接口契约，不提供调度器、工作池或监督器实现；具体实现的
// 并发安全、幂等性、阻塞方式和错误语义由实现方文档约定。
package runtime
