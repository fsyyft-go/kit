// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package config

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"

	kitgobuild "github.com/fsyyft-go/kit/go/build"
)

const (
	// verSimple 定义版本信息的输出格式。
	verSimple = "version %[1]s/%[2]s (build %[3]s)"
)

var (
	// 确保 version 类型实现了这些接口。
	_ fmt.Stringer               = (*version)(nil)
	_ fmt.Formatter              = (*version)(nil)
	_ kitgobuild.BuildingContext = (*version)(nil)
)

var (
	// CurrentVersion 表示当前应用程序的版本信息实例。
	CurrentVersion = version{
		buildingContext: kitgobuild.CurrentBuildingContext,
	}
)

// version 结构体封装了应用程序的版本信息。
type version struct {
	// buildingContext 包含了构建时的上下文信息。
	buildingContext kitgobuild.BuildingContext
}

// Version 获取软件版本号。
//
// 返回值：
//   - string：软件版本号。
func (v *version) Version() string {
	return v.buildingContext.Version()
}

// GitVersion 获取完整的 Git 版本号。
//
// 返回值：
//   - string：完整的 Git 提交哈希值。
func (v *version) GitVersion() string {
	return v.buildingContext.GitVersion()
}

// GitShortVersion 获取短格式的 Git 版本号。
//
// 返回值：
//   - string：Git 提交哈希值的前 8 位字符。
func (v *version) GitShortVersion() string {
	return v.buildingContext.GitShortVersion()
}

// LibGitVersion 获取类库的完整 Git 版本号。
//
// 返回值：
//   - string：类库的完整 Git 提交哈希值。
func (v *version) LibGitVersion() string {
	return v.buildingContext.LibGitVersion()
}

// LibGitShortVersion 获取类库的短格式 Git 版本号。
//
// 返回值：
//   - string：类库的 Git 提交哈希值的前 8 位字符。
func (v *version) LibGitShortVersion() string {
	return v.buildingContext.LibGitShortVersion()
}

// BuildTimeString 获取构建时间字符串。
//
// 返回值：
//   - string：构建时间，格式为 "20060102150405000"。
func (v *version) BuildTimeString() string {
	return v.buildingContext.BuildTimeString()
}

// BuildLibraryDirectory 获取构建时类库所在目录。
//
// 返回值：
//   - string：类库的绝对路径。
func (v *version) BuildLibraryDirectory() string {
	return v.buildingContext.BuildLibraryDirectory()
}

// BuildWorkingDirectory 获取构建时的工作目录。
//
// 返回值：
//   - string：工作目录的绝对路径。
func (v *version) BuildWorkingDirectory() string {
	return v.buildingContext.BuildWorkingDirectory()
}

// BuildGopathDirectory 获取构建时的 GOPATH 目录。
//
// 返回值：
//   - string：GOPATH 环境变量指定的目录路径。
func (v *version) BuildGopathDirectory() string {
	return v.buildingContext.BuildGopathDirectory()
}

// BuildGorootDirectory 获取构建时的 GOROOT 目录。
//
// 返回值：
//   - string：GOROOT 环境变量指定的目录路径。
func (v *version) BuildGorootDirectory() string {
	return v.buildingContext.BuildGorootDirectory()
}

// Debug 获取是否为调试状态。
//
// 返回值：
//   - bool：true 表示处于调试状态，false 表示处于正常状态。
func (v *version) Debug() bool {
	return v.buildingContext.Debug()
}

// String 实现了 fmt.Stringer 接口，返回版本信息的简短字符串表示。
//
// 返回值：
//   - string：格式化后的版本信息字符串。
func (v *version) String() string {
	// TODO 调试状态有问题，先忽略输出。
	return fmt.Sprintf(verSimple,
		v.buildingContext.GitShortVersion(),
		v.buildingContext.BuildTimeString(),
		strings.ReplaceAll(runtime.Version(), "go", ""),
	)
}

// Format 实现了 fmt.Formatter 接口，根据格式化标志返回不同详细程度的版本信息。
//
// 参数：
//   - s：格式化状态。
//   - verb：格式化动词。
func (v version) Format(s fmt.State, verb rune) {
	// 当使用 %+v 格式化时，返回详细描述。
	isDesc := verb == 'v' && s.Flag('+')
	if isDesc {
		_, _ = s.Write([]byte(v.Description()))
	} else {
		_, _ = s.Write([]byte(v.String()))
	}
}

// Description 返回版本信息的详细描述。
//
// 返回值：
//   - string：包含完整版本信息的多行字符串。
func (v *version) Description() string {
	/**
	 * TODO 调试状态有问题，先忽略输出。
	 * 这个地方不经常修改，直接使用 buffer 代替 fmt 包，代码可读性差一些，但可以提升一定的性能。
	 */

	// 使用 bytes.Buffer 构建详细的版本信息字符串。
	buf := bytes.Buffer{}

	// 添加开发版本信息。
	buf.WriteString("开发版本：")
	buf.WriteString(runtime.Version())
	buf.WriteString("\n")

	// 添加编译时间信息。
	buf.WriteString("编译时间：")
	buf.WriteString(v.buildingContext.BuildTimeString())
	buf.WriteString("\n")

	// 添加类库版本信息。
	buf.WriteString("类库版本：")
	buf.WriteString(v.buildingContext.LibGitVersion())
	buf.WriteString("\n")

	// 添加应用版本信息。
	buf.WriteString("应用版本：")
	buf.WriteString(v.buildingContext.GitVersion())
	buf.WriteString("\n")

	// 添加类库目录信息。
	buf.WriteString("类库目录：")
	buf.WriteString(v.buildingContext.BuildLibraryDirectory())
	buf.WriteString("\n")

	// 添加应用目录信息。
	buf.WriteString("应用目录：")
	buf.WriteString(v.buildingContext.BuildWorkingDirectory())
	buf.WriteString("\n")

	// 添加编译工具目录信息。
	buf.WriteString("编译工具目录：")
	buf.WriteString(v.buildingContext.BuildGorootDirectory())
	buf.WriteString("\n")

	// 添加编译环境目录信息。
	buf.WriteString("编译环境目录：")
	buf.WriteString(v.buildingContext.BuildGopathDirectory())

	return buf.String()
}
