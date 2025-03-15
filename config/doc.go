// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

/*
Package config 提供了应用程序的配置管理功能，特别是版本信息的管理和展示。

主要功能：

1. 版本信息管理：
  - 通过 version 结构体封装版本信息
  - 提供 CurrentVersion 全局实例
  - 支持获取软件版本和 Git 版本信息
  - 提供构建时间和环境路径信息

2. 信息展示格式：
  - String 方法：提供简短的版本信息，格式为 "version {git-short-hash}/{build-time} (build {go-version})"
  - Format 方法：支持通过 fmt 包的格式化功能展示不同详细程度的信息
  - Description 方法：提供多行的详细版本信息描述

3. 版本信息接口：
  - Version：获取软件版本号
  - GitVersion：获取完整的 Git 提交哈希值
  - GitShortVersion：获取短格式的 Git 提交哈希值（前 8 位）
  - LibGitVersion：获取类库的 Git 版本信息
  - BuildTimeString：获取构建时间
  - BuildLibraryDirectory：获取类库目录
  - BuildWorkingDirectory：获取工作目录
  - BuildGopathDirectory：获取 GOPATH 目录
  - BuildGorootDirectory：获取 GOROOT 目录
  - Debug：获取调试状态

基本用法：

	// 获取当前版本信息
	ver := config.CurrentVersion

	// 获取简短版本信息
	fmt.Println("版本信息:", ver)

	// 获取详细版本信息
	fmt.Printf("详细信息:\n%+v\n", ver)

	// 获取具体版本属性
	fmt.Printf("Git 版本: %s\n", ver.GitVersion())
	fmt.Printf("构建时间: %s\n", ver.BuildTimeString())

常量定义：

	verSimple = "version %[1]s/%[2]s (build %[3]s)"    // 简短版本信息的格式模板

注意事项：

1. 版本信息展示：
  - String 方法返回简短格式，适合日志输出
  - Format 方法支持 %+v 标记输出详细信息
  - Description 方法返回多行的完整信息

2. 调试状态：
  - Debug 方法用于检查是否处于调试模式
  - 调试状态会影响某些版本信息的展示
  - 目前存在已知的调试状态输出问题（见 TODO 注释）

3. 性能考虑：
  - Description 方法使用 bytes.Buffer 优化字符串构建
  - 版本信息通常只需要获取一次，可以缓存使用
  - 避免频繁调用 Description 方法
*/
package config
