// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// config 运行 Kratos 配置示例命令，并委托 cmd 包加载示例配置或执行 DES 子命令。
package main

import (
	kitkratosconfigcmd "github.com/fsyyft-go/kit/example/kratos/config/cmd"
)

// Config 结构体定义了应用程序的配置结构。

// main 函数是程序入口点。
func main() {
	kitkratosconfigcmd.Execute()
}
