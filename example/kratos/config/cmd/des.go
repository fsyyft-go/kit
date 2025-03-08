// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package cmd 实现了配置工具的命令行功能。
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	kit_crypto_des "github.com/fsyyft-go/kit/crypto/des"
)

// 定义命令行参数变量。
var (
	// key 用于存储用户提供的加密密钥。
	key string
	// data 用于存储待处理的数据。
	data string
	// encrypt 用于指示是执行加密还是解密操作。
	encrypt bool
)

// desCmd 代表 des 加密解密命令。
var desCmd = &cobra.Command{
	// 指定命令的名称。
	Use: "des",
	// 简短的命令描述。
	Short: "DES 加密解密工具",
	// 详细的命令描述和使用示例。
	Long: `DES 加密解密的命令行工具。
使用示例：
  # 使用默认密钥加密数据
  des --data 待加密数据 --encrypt
  # 使用自定义密钥加密数据
  des --key 你的密钥 --data 待加密数据 --encrypt
  # 使用自定义密钥解密数据
  des --key 你的密钥 --data 已加密数据 --encrypt=false`,
	// RunE 函数定义了命令的执行逻辑。
	RunE: func(cmd *cobra.Command, args []string) error {
		// 如果未提供密钥，使用默认密钥。
		if key == "" {
			key = kit_crypto_des.GetDefaultDESKey()
		}
		// 验证数据参数不能为空。
		if data == "" {
			return fmt.Errorf("数据不能为空")
		}

		var result string
		var err error

		// 根据 encrypt 标志决定执行加密或解密操作。
		if encrypt {
			result, err = kit_crypto_des.EncryptStringCBCPkCS7PaddingStringHex(key, data)
		} else {
			result, err = kit_crypto_des.EncryptStringCBCPkCS7PaddingStringHex(key, data)
		}

		// 处理操作结果。
		if nil != err {
			fmt.Printf("发生错误: %v\n", err)
		} else {
			fmt.Println("原始数据:", data)
			fmt.Println("操作结果:", result)
		}

		return nil
	},
}

// init 函数在包初始化时运行，用于设置命令行参数。
func init() {
	// 将 des 命令添加到根命令。
	rootCmd.AddCommand(desCmd)

	// 定义命令行标志。
	desCmd.PersistentFlags().StringVar(&key, "key", "", "加密/解密使用的密钥（可选，默认使用内置密钥）")
	desCmd.PersistentFlags().StringVar(&data, "data", "", "需要处理的数据（必填）")
	desCmd.PersistentFlags().BoolVar(&encrypt, "encrypt", true, "true表示加密，false表示解密")

	// 将 data 参数标记为必填项。
	_ = desCmd.MarkPersistentFlagRequired("data")
}
