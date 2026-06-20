// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// version 示例打印 config.CurrentVersion 暴露的版本与构建环境信息。
//
// 该命令适合配合 ldflags 注入的构建元数据运行，也可用于查看当前二进制内置的默认版本字段。
package main

import (
	"fmt"

	kitconfig "github.com/fsyyft-go/kit/config"
)

func main() {
	fmt.Printf("Version: %s\n", kitconfig.CurrentVersion.Version())
	fmt.Printf("Git Version: %s\n", kitconfig.CurrentVersion.GitVersion())
	fmt.Printf("Build Time: %s\n", kitconfig.CurrentVersion.BuildTimeString())
	fmt.Printf("Library Directory: %s\n", kitconfig.CurrentVersion.BuildLibraryDirectory())
	fmt.Printf("Working Directory: %s\n", kitconfig.CurrentVersion.BuildWorkingDirectory())
	fmt.Printf("GOPATH Directory: %s\n", kitconfig.CurrentVersion.BuildGopathDirectory())
	fmt.Printf("GOROOT Directory: %s\n", kitconfig.CurrentVersion.BuildGorootDirectory())
	fmt.Printf("Debug Mode: %v\n", kitconfig.CurrentVersion.Debug())
}
