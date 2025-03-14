# config

## 简介

config 包提供了应用程序的配置管理功能，特别是版本信息管理功能。通过这个包，开发者可以在编译时注入版本信息，并在运行时方便地获取和展示这些信息。

### 主要特性

- 提供统一的版本信息访问接口
- 支持在编译时注入版本号、Git 提交信息
- 支持获取构建时间信息
- 支持获取构建环境信息（工作目录、GOPATH、GOROOT 等）
- 提供标准化的字符串表示形式
- 支持调试模式标识

### 设计理念

config 包采用了以下设计理念：

- **接口分离**: 通过依赖 `BuildingContext` 接口而非具体实现，实现了关注点分离
- **全局访问点**: 提供 `CurrentVersion` 全局变量，便于在应用的任何位置访问版本信息
- **格式化输出**: 实现了 `fmt.Stringer` 和 `fmt.Formatter` 接口，支持灵活的版本信息展示
- **编译时注入**: 利用 Go 的链接标志（ldflags）机制，在编译时注入版本信息

## 安装

### 前置条件

- Go 版本要求：Go 1.16 或更高版本
- 依赖要求：

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/config
```

## 快速开始

### 基础用法

```go
package main

import (
    "fmt"

    "github.com/fsyyft-go/kit/config"
)

func main() {
    // 打印简短版本信息
    fmt.Println("版本信息:", config.CurrentVersion)

    // 打印详细版本信息
    fmt.Printf("详细信息:\n%+v\n", config.CurrentVersion)
}
```

### 配置选项

在编译应用程序时，可以通过 ldflags 注入版本信息：

```bash
go build -ldflags "
    -X github.com/fsyyft-go/kit/go/build.version=1.2.3
    -X github.com/fsyyft-go/kit/go/build.gitVersion=$(git rev-parse HEAD)
    -X github.com/fsyyft-go/kit/go/build.buildTimeString=$(date '+%Y%m%d%H%M%S000')
" -o myapp main.go
```

## 详细指南

### 核心概念

version 结构体是 config 包的核心，它封装了应用程序的版本信息并提供了访问这些信息的方法。这个结构体实现了 `fmt.Stringer` 和 `fmt.Formatter` 接口，使得版本信息可以方便地用于日志记录、调试输出等场景。

### 常见用例

#### 1. 在应用程序启动时显示版本信息

```go
package main

import (
    "fmt"
    "log"

    "github.com/fsyyft-go/kit/config"
)

func main() {
    log.Printf("启动应用 %s", config.CurrentVersion)

    // 应用程序代码...
}
```

#### 2. 在 API 响应中包含版本信息

```go
package main

import (
    "encoding/json"
    "net/http"

    "github.com/fsyyft-go/kit/config"
)

func versionHandler(w http.ResponseWriter, r *http.Request) {
    info := map[string]interface{}{
        "version":    config.CurrentVersion.Version(),
        "git_commit": config.CurrentVersion.GitVersion(),
        "build_time": config.CurrentVersion.BuildTimeString(),
        "debug_mode": config.CurrentVersion.Debug(),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(info)
}

func main() {
    http.HandleFunc("/version", versionHandler)
    http.ListenAndServe(":8080", nil)
}
```

### 最佳实践

- 在持续集成/持续部署 (CI/CD) 流程中自动注入版本信息
- 为每个发布版本使用[语义化版本号格式 (SemVer)](https://semver.org/lang/zh-CN/)
- 在日志输出中包含版本信息，便于问题追踪
- 在显示详细版本信息的场景中使用 `%+v` 格式化动词
- 在 Web 应用中提供版本信息 API 接口，便于版本确认和问题报告

## API 文档

### 主要类型

```go
// version 结构体封装了应用程序的版本信息。
type version struct {
    // buildingContext 包含了构建时的上下文信息。
    buildingContext kit_go_build.BuildingContext
}

// CurrentVersion 表示当前应用程序的版本信息实例。
var CurrentVersion = version{
    buildingContext: kit_go_build.CurrentBuildingContext,
}
```

### 关键函数

#### Version() string

返回软件版本号。

```go
func (v *version) Version() string
```

示例：

```go
ver := config.CurrentVersion.Version()
fmt.Println("软件版本:", ver)
```

#### GitVersion() string

返回完整的 Git 版本号（完整的哈希值）。

```go
func (v *version) GitVersion() string
```

示例：

```go
gitVer := config.CurrentVersion.GitVersion()
fmt.Println("Git 版本:", gitVer)
```

#### BuildTimeString() string

返回构建时间字符串。

```go
func (v *version) BuildTimeString() string
```

示例：

```go
buildTime := config.CurrentVersion.BuildTimeString()
fmt.Println("构建时间:", buildTime)
```

#### String() string

实现了 fmt.Stringer 接口，返回版本信息的简短字符串表示。

```go
func (v *version) String() string
```

示例：

```go
// 输出格式: version {git-short-hash}/{build-time} (build {go-version})
fmt.Println(config.CurrentVersion)
```

#### Format(s fmt.State, verb rune)

实现了 fmt.Formatter 接口，根据格式化标志返回不同详细程度的版本信息。

```go
func (v version) Format(s fmt.State, verb rune)
```

示例：

```go
// 输出简短信息
fmt.Printf("%v\n", config.CurrentVersion)

// 输出详细信息
fmt.Printf("%+v\n", config.CurrentVersion)
```

### 错误处理

config 包中的方法通常不会返回错误。如果某些版本信息在编译时未注入，相应的方法会返回空字符串或默认值。开发者应当确保在使用前检查这些返回值是否有效。

## 性能指标

config 包提供的方法主要是读取预注入的版本信息，这些操作通常是轻量级的，不会对应用程序的性能产生明显影响。

## 测试覆盖率

测试覆盖率信息待补充

## 调试指南

### 日志级别

config 包本身不产生日志输出。不过，开发者可以使用 `Debug()` 方法检查应用程序是否处于调试模式，并据此调整日志级别：

```go
if config.CurrentVersion.Debug() {
    // 设置为详细日志级别
} else {
    // 设置为普通日志级别
}
```

### 常见问题排查

#### 版本信息显示为空或默认值

**原因**: 编译时未正确注入版本信息。

**解决方案**: 确保在构建命令中使用正确的 ldflags 参数，例如：

```bash
go build -ldflags "
    -X github.com/fsyyft-go/kit/go/build.version=1.2.3
    -X github.com/fsyyft-go/kit/go/build.gitVersion=$(git rev-parse HEAD)
    -X github.com/fsyyft-go/kit/go/build.buildTimeString=$(date '+%Y%m%d%H%M%S000')
" -o myapp main.go
```

#### 版本格式不符合预期

**原因**: 字符串表示法可能与预期不一致。

**解决方案**: 使用特定的方法直接获取所需信息，而不是依赖自动格式化：

```go
// 不使用自动格式化
// fmt.Println(config.CurrentVersion)

// 而是直接获取所需信息
fmt.Printf("版本: %s, 提交: %s, 时间: %s\n",
    config.CurrentVersion.Version(),
    config.CurrentVersion.GitShortVersion(),
    config.CurrentVersion.BuildTimeString())
```

## 相关文档

- [版本信息管理示例](../example/config/version/README.md)
- [Go 构建工具文档](../go/build/README.md)
- [Go 编译链接标志文档](https://go.dev/cmd/link/)
- [语义化版本规范 (SemVer)](https://semver.org/lang/zh-CN/)
- [Go Modules 参考文档](https://go.dev/ref/mod)
- [Go 构建约束](https://pkg.go.dev/go/build#hdr-Build_Constraints)
- [GitHub Actions - 版本管理最佳实践](https://github.blog/2022-02-02-build-ci-cd-pipeline-github-actions-four-steps/)

## 贡献指南

我们欢迎任何形式的贡献，包括但不限于：

- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考我们的[贡献指南](../CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT 许可证。查看 [LICENSE](../LICENSE) 文件了解更多信息。

## 补充说明

本文的大部分信息，由 AI 使用[模板](../ai/templates/docs/package_readme_template.md)根据[提示词](../ai/prompts/docs/package_readme_generator.md)自动生成，如有任何问题，请随时联系我。
