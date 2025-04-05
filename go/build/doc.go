// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package build 提供了用于获取和管理 Go 程序构建环境信息的工具。

主要功能：

1. 构建信息管理：
  - 通过 BuildingContext 接口提供统一的信息访问方式
  - 支持获取软件版本和 Git 版本信息
  - 提供构建时间和环境路径信息
  - 自动检测调试环境状态

2. 版本信息：
  - Version：软件版本号
  - GitVersion：完整的 Git 提交哈希值
  - GitShortVersion：短格式的 Git 提交哈希值（前 8 位）
  - LibGitVersion：类库的 Git 版本信息

3. 构建环境：
  - BuildTimeString：构建时间，格式为 "20060102150405000"
  - BuildLibraryDirectory：类库所在目录
  - BuildWorkingDirectory：工作目录
  - BuildGopathDirectory：GOPATH 目录
  - BuildGorootDirectory：GOROOT 目录

基本用法：

	// 获取当前构建上下文
	ctx := build.CurrentBuildingContext

	// 获取版本信息
	version := ctx.Version()
	gitVersion := ctx.GitVersion()
	shortVersion := ctx.GitShortVersion()

	// 获取构建时间
	buildTime := ctx.BuildTimeString()

	// 获取构建环境信息
	workDir := ctx.BuildWorkingDirectory()
	gopath := ctx.BuildGopathDirectory()

	// 检查调试状态
	if ctx.Debug() {
	    // 处理调试模式
	}

构建时注入信息：

可以在构建时通过 -ldflags 参数注入信息，例如：

	go build -ldflags "-X github.com/fsyyft-go/kit/go/build.version=v1.0.0 \
	                   -X github.com/fsyyft-go/kit/go/build.gitVersion=$(git rev-parse HEAD) \
	                   -X github.com/fsyyft-go/kit/go/build.buildTimeString=$(date +%Y%m%d%H%M%S%3N)"

常量定义：

	TimeLayout = "20060102150405000"    // 时间格式化模板
	GoEnvNameRoot = "GOROOT"            // Go 根目录环境变量名
	GoEnvNamePath = "GOPATH"            // Go 路径环境变量名
	GoEnvNameTmpDir = "GOTMPDIR"        // Go 临时目录环境变量名

注意事项：

1. 版本信息：
  - 在非调试模式下，版本信息来自构建时注入
  - 在调试模式下，使用默认值或运行时环境信息

2. 环境路径：
  - 确保路径分隔符的正确使用
  - 注意环境变量的可用性

3. 调试状态：
  - 自动检测 go run 和 go test 环境
  - 根据调试状态调整行为
*/
package build
