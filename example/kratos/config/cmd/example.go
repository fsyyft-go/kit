// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"
	"strings"

	kit_kratos_config "github.com/fsyyft-go/kit/kratos/config"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
)

// Example 函数展示了配置的基本用法。
func Example() {
	// Config 结构体定义了配置文件的结构。
	type Config struct {
		// App 嵌套结构体包含应用程序相关配置。
		App struct {
			// Name 是应用程序名称。
			Name string `json:"name"`
			// Password 是应用程序密码，可能是从 base64 编码后解码得到的。
			Password string `json:"password"`
			// Addr 是应用程序地址，可能是从 base64 编码后解码得到的。
			Addr string `json:"addr"`
		} `json:"app"`
	}

	// 声明配置文件路径变量。
	var configPath string

	// 获取当前工作目录。
	pwd := os.Getenv("PWD")
	// 打印当前工作目录，用于调试。
	fmt.Println("pwd", pwd)

	// 根据当前工作目录确定配置文件路径。
	if strings.HasSuffix(pwd, "example/kratos/config") {
		// 如果在 example/kratos/config 目录下运行，使用相对路径。
		configPath = "config.yaml"
	} else {
		// 如果在项目根目录下运行，使用完整路径。
		configPath = "example/kratos/config/config.yaml"
	}

	// 创建配置管理器实例。
	c := config.New(
		// 设置配置源为文件源，指定配置文件路径。
		config.WithSource(
			file.NewSource(configPath),
		),
		// 设置自定义解码器，支持特殊格式处理（如 base64 解码）。
		config.WithDecoder(kit_kratos_config.NewDecoder().Decode),
	)
	// 加载配置，如果出错则触发 panic。
	if err := c.Load(); err != nil {
		panic(err)
	}

	// 声明配置结构体变量。
	var cfg Config
	// 将加载的配置扫描到结构体中，如果出错则触发 panic。
	if err := c.Scan(&cfg); err != nil {
		panic(err)
	}

	// 打印配置信息。
	fmt.Printf("%+v\n", cfg.App.Name)
	fmt.Printf("%+v\n", cfg.App.Password)
	fmt.Printf("%+v\n", cfg.App.Addr)
}
