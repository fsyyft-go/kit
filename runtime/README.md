# runtime

## 简介

runtime 包提供了应用程序运行时管理的基础设施，主要用于统一管理各种组件的生命周期。该包定义了运行时组件的标准接口，使得应用程序可以以一致的方式启动和停止各种后台服务和处理程序。

### 主要特性

- 标准化的组件生命周期管理接口
- 上下文感知的启动和停止机制
- 支持优雅关闭
- 与 Go 上下文（context）包无缝集成

### 设计理念

runtime 包的设计遵循了"组合优于继承"的 Go 语言哲学，通过接口定义统一的行为规范，使得不同类型的运行时组件可以被统一管理。包的设计注重简洁性和可扩展性，为构建健壮的后台服务和长时间运行的应用程序提供基础支持。

## 安装

### 前置条件

- Go 版本要求：Go 1.24 或更高版本
- 依赖要求：
  - 无外部依赖

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/runtime
```

## 快速开始

### 基础用法

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/fsyyft-go/kit/runtime"
)

// MyService 实现了 runtime.Runner 接口
type MyService struct {
	// 服务所需字段
}

func (s *MyService) Start(ctx context.Context) error {
	// 初始化服务并启动
	log.Println("服务已启动")
	return nil
}

func (s *MyService) Stop(ctx context.Context) error {
	// 优雅关闭服务
	log.Println("服务已停止")
	return nil
}

func main() {
	// 创建服务
	service := &MyService{}
	
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// 启动服务
	if err := service.Start(ctx); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
	
	// 应用程序运行逻辑...
	
	// 停止服务
	if err := service.Stop(ctx); err != nil {
		log.Fatalf("服务停止失败: %v", err)
	}
}
```

## 详细指南

### 核心概念

runtime 包的核心是 `Runner` 接口，它定义了组件的生命周期管理方法。任何实现了 `Runner` 接口的组件都可以被统一管理，这种方式使得应用程序可以轻松地集成多种服务组件，并以一致的方式管理它们的启动和停止过程。

### 最佳实践

- 在 `Start` 方法中实现对 ctx.Done() 的监听，以支持取消操作
- 在 `Stop` 方法中尊重上下文的截止时间，确保在超时前完成关闭
- 将复杂组件拆分为多个 Runner 实现，通过组合的方式构建完整系统
- 在应用程序退出前，始终调用 Stop 方法以确保资源被正确释放

## API 文档

### 主要类型

```go
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
```

## 子包

runtime 包包含以下子包：

- [goroutine](./goroutine/README.md) - 提供与 goroutine 相关的功能，如获取 goroutine ID 等

## 相关文档

- [示例代码](../example/runtime/goroutine/README.md)
- [Github 仓库](https://github.com/fsyyft-go/kit)

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考我们的[贡献指南](../CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT 许可证。查看 [LICENSE](../LICENSE) 文件了解更多信息。