// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package runtime 提供了运行时工具函数和协程管理功能。

主要特性：

  - 协程管理和控制
  - 运行时状态监控
  - 性能优化工具
  - 安全的并发操作
  - 资源使用跟踪

基本功能：

1. 协程运行器：

	// 创建运行器
	runner := runtime.NewRunner()

	// 添加任务
	runner.Add(func() {
	    // 任务逻辑
	})

	// 启动所有任务
	runner.Start()

	// 等待任务完成
	runner.Wait()

	// 优雅关闭
	runner.Shutdown()

2. 并发控制：

	// 设置最大并发数
	runner := runtime.NewRunner(
	    runtime.WithMaxConcurrency(10),
	    runtime.WithQueueSize(100),
	)

	// 添加带超时的任务
	runner.AddWithTimeout(func() {
	    // 任务逻辑
	}, 5*time.Second)

3. 错误处理：

	// 设置错误处理器
	runner.OnError(func(err error) {
	    log.Printf("任务执行错误：%v", err)
	})

	// 添加可能失败的任务
	runner.Add(func() error {
	    // 任务逻辑
	    return nil
	})

4. 状态监控：

	// 获取运行状态
	stats := runner.Stats()
	fmt.Printf("运行中任务：%d\n", stats.Running)
	fmt.Printf("已完成任务：%d\n", stats.Completed)
	fmt.Printf("失败任务：%d\n", stats.Failed)

性能优化：

1. 任务调度：
  - 使用工作池模式
  - 动态调整并发度
  - 避免资源竞争

2. 内存管理：
  - 复用 goroutine
  - 控制内存使用
  - 及时释放资源

3. 错误处理：
  - 快速失败策略
  - 错误隔离机制
  - 优雅降级支持

使用建议：

1. 资源管理：
  - 合理设置并发数
  - 控制任务队列大小
  - 及时清理资源

2. 错误处理：
  - 设置错误回调
  - 实现熔断机制
  - 提供降级方案

3. 监控告警：
  - 监控运行状态
  - 设置性能指标
  - 及时处理异常

注意事项：

1. 并发安全：
  - 避免数据竞争
  - 正确使用同步原语
  - 注意死锁风险

2. 资源限制：
  - 控制内存使用
  - 限制 goroutine 数量
  - 避免资源耗尽

3. 性能影响：
  - 注意调度开销
  - 合理设置超时
  - 避免阻塞操作

更多示例和最佳实践请参考 example/runtime 目录。
*/
package runtime
