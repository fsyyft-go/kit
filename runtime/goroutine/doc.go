// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package goroutine 提供了用于获取和操作 Go 协程相关信息的工具函数。

主要特性：

  - 获取当前 goroutine 的 ID
  - 支持多个 Go 版本（1.5-1.24）
  - 提供不同 CPU 架构的实现（amd64、arm64）
  - 支持性能优化的实现方式
  - 提供回退实现方案
  - 线程安全

基本功能：

1. 获取 Goroutine ID：

	// 获取当前 goroutine 的 ID
	goid := goroutine.GetGoID()
	fmt.Printf("当前 goroutine ID: %d\n", goid)

	// 在并发环境中使用
	go func() {
	    goid := goroutine.GetGoID()
	    fmt.Printf("新 goroutine ID: %d\n", goid)
	}()

2. 性能优化版本：

	// 使用优化的方式获取 goroutine ID
	goid := goroutine.GetGoIDFast()

	// 如果优化版本不可用，会自动回退到标准实现
	if goid == -1 {
	    goid = goroutine.GetGoID()
	}

实现说明：

1. 版本支持：
  - Go 1.5：基础实现
  - Go 1.6-1.8：改进的运行时支持
  - Go 1.9-1.22：优化的实现
  - Go 1.23+：最新的优化实现

2. 架构支持：
  - AMD64：完整优化支持
  - ARM64：完整优化支持
  - 其他架构：基础实现支持

3. 实现方式：
  - 直接访问运行时数据结构
  - 使用汇编优化关键路径
  - 提供回退实现保证可用性

性能考虑：

1. 快速路径：
  - 优化的汇编实现
  - 最小化指令数
  - 避免运行时调用

2. 回退路径：
  - 稳定可靠的实现
  - 适用所有平台
  - 性能略低但更通用

3. 缓存策略：
  - 合理使用 CPU 缓存
  - 避免不必要的内存访问
  - 减少锁竞争

使用建议：

1. 版本选择：
  - 优先使用 GetGoIDFast
  - 必要时回退到 GetGoID
  - 注意版本兼容性

2. 错误处理：
  - 处理无效返回值（-1）
  - 提供合理的回退方案
  - 记录异常情况

3. 性能优化：
  - 避免频繁调用
  - 合理缓存结果
  - 注意调用开销

限制和注意事项：

1. 兼容性：
  - 不同 Go 版本行为可能不同
  - 某些平台可能不支持优化实现
  - 需要适当的回退策略

2. 安全性：
  - 直接访问运行时数据结构
  - 可能受到运行时更改影响
  - 建议在开发环境充分测试

3. 调试：
  - 提供详细的错误信息
  - 支持运行时诊断
  - 便于问题排查

最佳实践：

1. 开发环境：
  - 使用最新的 Go 版本
  - 进行充分的测试
  - 验证所有目标平台

2. 生产环境：
  - 进行性能基准测试
  - 监控异常情况
  - 提供降级方案

3. 代码维护：
  - 保持实现的简洁性
  - 提供清晰的文档
  - 及时更新兼容性信息

更多示例和最佳实践请参考 example/runtime/goroutine 目录。
*/
package goroutine
