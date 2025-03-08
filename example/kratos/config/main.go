// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"

	kit_kratos_config "github.com/fsyyft-go/kit/kratos/config"
)

type Config struct {
	App struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	} `json:"app"`
}

func main() {
	var configPath string

	pwd := os.Getenv("PWD")
	fmt.Println("pwd", pwd)

	if strings.HasSuffix(pwd, "example/kratos/config") {
		configPath = "config.yaml"
	} else {
		configPath = pwd + "/example/kratos/config/config.yaml"
	}

	c := config.New(
		config.WithSource(
			file.NewSource(configPath),
		),
		config.WithDecoder(kit_kratos_config.NewDecoder().Decode),
	)
	if err := c.Load(); err != nil {
		panic(err)
	}

	var cfg Config
	if err := c.Scan(&cfg); err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", cfg.App.Name)
	fmt.Printf("%+v\n", cfg.App.Password)
}
