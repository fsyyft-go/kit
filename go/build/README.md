# build

## 简介

build 包提供了一套用于获取和管理 Go 程序构建环境信息的工具。它允许应用程序检索构建时和运行时的各种上下文信息，如软件版本、Git 提交信息、构建时间以及 Go 环境路径等。

### 主要特性

- 提供统一的构建上下文接口 `BuildingContext`
- 支持获取软件版本和 Git 版本信息
- 支持获取构建时间信息
- 支持获取构建环境路径（工作目录、GOPATH、GOROOT等）
- 自动检测并标记调试环境（go run 或 go test）
- 提供全局可访问的当前构建上下文 `CurrentBuildingContext`

### 设计理念

build 包采用接口隔离原则，通过 `BuildingContext` 接口抽象构建信息的访问方式，实现与具体实现的解耦。包内部通过 `buildingContextValue` 结构体实现该接口，并在初始化时自动配置信息。设计上支持构建时注入（通过链接器标志）和运行时检测（调试模式下自动获取）两种信息获取方式，灵活满足不同场景需求。

## 安装

### 前置条件

- Go 版本要求：Go 1.13 或更高版本
- 依赖要求：标准库（无外部依赖）

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/go/build
```

## 快速开始

### 基础用法

```go
package main

import (
	"fmt"
	
	"github.com/fsyyft-go/kit/go/build"
)

func main() {
	// 获取当前构建上下文
	ctx := build.CurrentBuildingContext
	
	// 输出版本信息
	fmt.Printf("Version: %s\n", ctx.Version())
	fmt.Printf("Git Commit: %s\n", ctx.GitVersion())
	fmt.Printf("Short Git Commit: %s\n", ctx.GitShortVersion())
	
	// 输出构建时间
	fmt.Printf("Build Time: %s\n", ctx.BuildTimeString())
	
	// 检查是否为调试模式
	if ctx.Debug() {
		fmt.Println("Running in debug mode")
	} else {
		fmt.Println("Running in release mode")
	}
}
```

### 配置选项

```go
package main

import (
	"fmt"
	"log"
	"os"
	"time"
	
	"github.com/fsyyft-go/kit/go/build"
)

func main() {
	// 在构建时通过链接器标志注入版本信息
	// 使用命令:
	// go build -ldflags "-X github.com/fsyyft-go/kit/go/build.version=v1.0.0 -X github.com/fsyyft-go/kit/go/build.gitVersion=abcdef1234567890 -X github.com/fsyyft-go/kit/go/build.buildTimeString=20240317123456789"
	
	ctx := build.CurrentBuildingContext
	
	// 输出注入的信息
	fmt.Printf("Version: %s\n", ctx.Version())
	fmt.Printf("Git Version: %s\n", ctx.GitVersion())
	
	// 解析构建时间
	timeStr := ctx.BuildTimeString()
	buildTime, err := time.Parse(build.TimeLayout, timeStr)
	if err != nil {
		log.Fatalf("Failed to parse build time: %v", err)
	}
	
	fmt.Printf("Build Time: %s\n", buildTime.Format(time.RFC3339))
	
	// 输出构建环境信息
	fmt.Printf("Working Directory: %s\n", ctx.BuildWorkingDirectory())
	fmt.Printf("GOPATH: %s\n", ctx.BuildGopathDirectory())
	fmt.Printf("GOROOT: %s\n", ctx.BuildGorootDirectory())
}
```

## 详细指南

### 核心概念

build 包的核心是 `BuildingContext` 接口和 `CurrentBuildingContext` 全局变量。接口定义了获取构建信息的方法集，全局变量则提供了便捷访问方式。包的实现会根据运行环境自动判断是否处于调试模式，并相应调整信息获取方式。

在非调试模式下，版本信息通常来自构建时通过链接器标志（-ldflags）注入的值。而在调试模式下（如通过 `go run` 或 `go test` 运行），则更多地依赖运行时环境信息和默认值。

### 常见用例

#### 1. 在日志中添加版本信息

```go
package main

import (
	"log"
	
	"github.com/fsyyft-go/kit/go/build"
)

func main() {
	ctx := build.CurrentBuildingContext
	
	// 配置日志前缀，包含版本和提交信息
	prefix := "[" + ctx.Version() + " " + ctx.GitShortVersion() + "] "
	log.SetPrefix(prefix)
	
	log.Println("Application started")
	
	// 输出更详细的版本信息
	log.Printf("Version details: %s (%s)", ctx.Version(), ctx.GitVersion())
	log.Printf("Build time: %s", ctx.BuildTimeString())
}
```

#### 2. 在命令行工具中实现版本子命令

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	
	"github.com/fsyyft-go/kit/go/build"
)

func main() {
	// 定义版本标志
	versionFlag := flag.Bool("version", false, "Print version information and exit")
	flag.Parse()
	
	// 检查是否需要显示版本信息
	if *versionFlag {
		ctx := build.CurrentBuildingContext
		
		// 解析构建时间
		buildTime, _ := time.Parse(build.TimeLayout, ctx.BuildTimeString())
		timeStr := buildTime.Format(time.RFC3339)
		
		// 显示版本信息
		fmt.Printf("MyApp %s\n", ctx.Version())
		fmt.Printf("Git commit: %s\n", ctx.GitVersion())
		fmt.Printf("Built: %s\n", timeStr)
		fmt.Printf("Go path: %s\n", ctx.BuildGopathDirectory())
		os.Exit(0)
	}
	
	// 应用程序主逻辑
	fmt.Println("Application running...")
}
```

### 最佳实践

- 使用链接器标志（-ldflags）在构建时注入版本信息
- 在 CI/CD 流程中自动化构建信息的注入
- 在日志、错误报告和诊断信息中包含版本和构建信息
- 结合构建时间戳实现版本控制和缓存失效策略
- 利用调试模式检测来调整应用行为（例如在开发环境使用更详细的日志）

## API 文档

### 主要类型

```go
// BuildingContext 定义构建上下文接口，提供访问构建时信息的方法。
type BuildingContext interface {
    // Version 获取软件版本。
    Version() string
    
    // GitVersion 获取完整的 Git 版本号。
    GitVersion() string
    
    // GitShortVersion 获取短格式的 Git 版本号。
    GitShortVersion() string
    
    // LibGitVersion 获取类库的完整 Git 版本号。
    LibGitVersion() string
    
    // LibGitShortVersion 获取类库的短格式 Git 版本号。
    LibGitShortVersion() string
    
    // BuildTimeString 获取构建时间字符串。
    BuildTimeString() string
    
    // BuildLibraryDirectory 获取构建时类库所在目录。
    BuildLibraryDirectory() string
    
    // BuildWorkingDirectory 获取构建时的工作目录。
    BuildWorkingDirectory() string
    
    // BuildGopathDirectory 获取构建时的 GOPATH 目录。
    BuildGopathDirectory() string
    
    // BuildGorootDirectory 获取构建时的 GOROOT 目录。
    BuildGorootDirectory() string
    
    // Debug 获取是否为调试状态。
    Debug() bool
}
```

### 关键函数

#### CurrentBuildingContext 全局变量

包提供了一个全局可访问的构建上下文变量 `CurrentBuildingContext`，它在包初始化时自动配置，包含当前运行环境的所有构建信息。

示例：
```go
// 获取版本信息
version := build.CurrentBuildingContext.Version()

// 检查是否处于调试模式
isDebug := build.CurrentBuildingContext.Debug()
```

### 错误处理

build 包中的方法不会返回错误，而是在初始化过程中处理错误并设置适当的默认值。例如，当无法获取某些信息（如执行路径）时，包会安全地失败并将相应的调试标志设置为默认值。

在没有构建信息注入的情况下，大多数方法会返回空字符串或默认值，应用程序应当检查返回值并相应处理（例如，检查版本字符串是否为空）。

## 性能指标

build 包的所有方法都是简单的数据访问，没有复杂计算或 I/O 操作，因此性能开销可以忽略不计。包的初始化过程涉及少量文件系统操作（如获取当前路径），但这些操作仅在程序启动时执行一次。

## 测试覆盖率

测试覆盖率信息待补充

## 调试指南

### 日志级别

build 包本身不生成日志，但在初始化时会自动检测应用程序是否在调试环境（go run 或 go test）中运行，并相应地设置 `Debug()` 标志。

### 常见问题排查

#### 版本信息缺失或不正确

如果 `Version()` 或 `GitVersion()` 返回空字符串，很可能是因为未在构建时正确注入这些值。确保使用以下命令进行构建：

```bash
go build -ldflags "-X github.com/fsyyft-go/kit/go/build.version=v1.0.0 -X github.com/fsyyft-go/kit/go/build.gitVersion=$(git rev-parse HEAD) -X github.com/fsyyft-go/kit/go/build.buildTimeString=$(date +%Y%m%d%H%M%S%3N)"
```

#### 构建时间格式化问题

如果需要将 `BuildTimeString()` 返回的时间戳转换为标准时间，请确保使用正确的格式模板：

```go
timeStr := build.CurrentBuildingContext.BuildTimeString()
buildTime, err := time.Parse(build.TimeLayout, timeStr)
if err != nil {
    // 处理错误
}
```

## 相关文档

- [Go 链接器标志文档](https://golang.org/cmd/link/)
- [Go 构建约束](https://golang.org/pkg/go/build/)

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考我们的[贡献指南](../../CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT 许可证。查看 [LICENSE](../../LICENSE) 文件了解更多信息。

## 补充说明

本文的大部分信息，由 AI 使用[模板](../../ai/templates/docs/package_readme_template.md)根据[提示词](../../ai/prompts/docs/package_readme_generator.md)自动生成，如有任何问题，请随时联系我。 