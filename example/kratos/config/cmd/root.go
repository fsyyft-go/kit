// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd 表示在没有调用任何子命令时的基础命令。
var rootCmd = &cobra.Command{
	// 指定命令的名称。
	Use: "config",
	// 简短的命令描述。
	Short: "配置工具",
	// 详细的命令描述。
	Long: `一个提供各种配置处理实用工具的程序，
包含 DES 加密解密等功能。`,
	// RunE 函数定义了命令的执行逻辑。
	RunE: func(cmd *cobra.Command, args []string) error {
		// 当没有提供子命令时，运行示例函数。
		if len(args) == 0 {
			example()
			return nil
		}
		return nil
	},
}

// Execute 将所有子命令添加到根命令并适当设置标志。
// 这个函数由 main.main() 调用，只需要对 rootCmd 执行一次。
func Execute() {
	// 执行根命令，如果出现错误则打印错误信息并退出程序。
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// init 函数在包初始化时运行，用于设置全局标志。
func init() {
	// 这里可以添加全局标志。
}
