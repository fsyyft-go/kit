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
	// verSimple 定义 String 返回的简短版本信息格式。
	verSimple = "version %[1]s/%[2]s (build %[3]s)"
)

var (
	// 编译期确认 version 实现版本格式化和构建上下文访问所需接口。
	_ fmt.Stringer               = (*version)(nil)
	_ fmt.Formatter              = (*version)(nil)
	_ kitgobuild.BuildingContext = (*version)(nil)
)

var (
	// CurrentVersion 是基于当前构建上下文初始化的默认版本信息实例。
	//
	// CurrentVersion 通过方法暴露应用版本、类库版本、Git 提交、构建时间和构建目录等只读元数据。
	// 由于其具体类型未导出，调用方通常直接调用方法，或通过 &CurrentVersion 传递给 fmt.Stringer、fmt.Formatter 和 kitgobuild.BuildingContext 接口。
	CurrentVersion = version{
		buildingContext: kitgobuild.CurrentBuildingContext,
	}
)

// version 保存构建上下文，并将版本访问方法委托给该上下文。
type version struct {
	// buildingContext 提供当前应用和类库的构建时元数据。
	buildingContext kitgobuild.BuildingContext
}

// Version 返回软件版本号。
//
// 参数：无。
//
// 返回：
//   - string: 底层构建上下文中的软件版本号。
func (v *version) Version() string {
	return v.buildingContext.Version()
}

// GitVersion 返回完整的应用 Git 版本号。
//
// 参数：无。
//
// 返回：
//   - string: 底层构建上下文中的完整应用 Git 版本号，通常为完整 Git 提交哈希。
func (v *version) GitVersion() string {
	return v.buildingContext.GitVersion()
}

// GitShortVersion 返回短格式的应用 Git 版本号。
//
// 参数：无。
//
// 返回：
//   - string: 底层构建上下文中的短格式应用 Git 版本号，通常为 Git 提交哈希前 8 位。
func (v *version) GitShortVersion() string {
	return v.buildingContext.GitShortVersion()
}

// LibGitVersion 返回类库的完整 Git 版本号。
//
// 参数：无。
//
// 返回：
//   - string: 底层构建上下文中的类库完整 Git 版本号，通常为完整 Git 提交哈希。
func (v *version) LibGitVersion() string {
	return v.buildingContext.LibGitVersion()
}

// LibGitShortVersion 返回类库的短格式 Git 版本号。
//
// 参数：无。
//
// 返回：
//   - string: 底层构建上下文中的类库短格式 Git 版本号，通常为 Git 提交哈希前 8 位。
func (v *version) LibGitShortVersion() string {
	return v.buildingContext.LibGitShortVersion()
}

// BuildTimeString 返回构建时间字符串。
//
// 参数：无。
//
// 返回：
//   - string: 底层构建上下文中的构建时间，格式为 "20060102150405000"。
func (v *version) BuildTimeString() string {
	return v.buildingContext.BuildTimeString()
}

// BuildLibraryDirectory 返回构建时的类库目录。
//
// 参数：无。
//
// 返回：
//   - string: 底层构建上下文中的类库绝对路径。
func (v *version) BuildLibraryDirectory() string {
	return v.buildingContext.BuildLibraryDirectory()
}

// BuildWorkingDirectory 返回构建时的工作目录。
//
// 参数：无。
//
// 返回：
//   - string: 底层构建上下文中的应用工作目录绝对路径。
func (v *version) BuildWorkingDirectory() string {
	return v.buildingContext.BuildWorkingDirectory()
}

// BuildGopathDirectory 返回构建时的 GOPATH 目录。
//
// 参数：无。
//
// 返回：
//   - string: 底层构建上下文中的 GOPATH 目录路径。
func (v *version) BuildGopathDirectory() string {
	return v.buildingContext.BuildGopathDirectory()
}

// BuildGorootDirectory 返回构建时的 GOROOT 目录。
//
// 参数：无。
//
// 返回：
//   - string: 底层构建上下文中的 GOROOT 目录路径。
func (v *version) BuildGorootDirectory() string {
	return v.buildingContext.BuildGorootDirectory()
}

// Debug 返回是否处于调试状态。
//
// 参数：无。
//
// 返回：
//   - bool: true 表示处于调试状态，false 表示处于正常状态。
func (v *version) Debug() bool {
	return v.buildingContext.Debug()
}

// String 实现 fmt.Stringer，返回版本信息的简短字符串表示。
//
// 参数：无。
//
// 返回：
//   - string: 使用短应用 Git 版本、构建时间和 Go 运行时版本组成的字符串，格式为 "version <git-short>/<build-time> (build <go-version>)"。
func (v *version) String() string {
	// TODO(fsyyft-go): 调试状态输出与当前格式化契约不一致，确认语义后再恢复展示。
	return fmt.Sprintf(verSimple,
		v.buildingContext.GitShortVersion(),
		v.buildingContext.BuildTimeString(),
		strings.ReplaceAll(runtime.Version(), "go", ""),
	)
}

// Format 实现 fmt.Formatter，根据格式化动词和标志写入版本信息。
//
// 参数：
//   - s: 格式化状态，接收输出内容；写入错误会被忽略以满足 fmt.Formatter 接口约定。
//   - verb: 格式化动词；当 verb 为 'v' 且包含 '+' 标志时输出 Description，否则输出 String。
func (v version) Format(s fmt.State, verb rune) {
	// 当使用 %+v 格式化时输出详细描述，其它格式保持 String 的简短输出。
	isDesc := verb == 'v' && s.Flag('+')
	if isDesc {
		_, _ = s.Write([]byte(v.Description()))
	} else {
		_, _ = s.Write([]byte(v.String()))
	}
}

// Description 返回版本信息的详细多行描述。
//
// 参数：无。
//
// 返回：
//   - string: 由开发版本、编译时间、类库版本、应用版本、类库目录、应用目录、编译工具目录和编译环境目录组成的多行字符串。
func (v *version) Description() string {
	// TODO(fsyyft-go): 调试状态输出与当前详细描述契约不一致，确认语义后再恢复展示。
	// 本函数很少调整，沿用 bytes.Buffer 直接拼接固定字段，以减少 fmt 格式化开销。

	// 预分阶段写入 bytes.Buffer，保持下方固定输出顺序清晰。
	buf := bytes.Buffer{}

	// 第 1 行固定展示 Go 运行时版本。
	buf.WriteString("开发版本：")
	buf.WriteString(runtime.Version())
	buf.WriteString("\n")

	// 第 2 行固定展示构建时间字符串。
	buf.WriteString("编译时间：")
	buf.WriteString(v.buildingContext.BuildTimeString())
	buf.WriteString("\n")

	// 第 3 行固定展示类库完整 Git 版本。
	buf.WriteString("类库版本：")
	buf.WriteString(v.buildingContext.LibGitVersion())
	buf.WriteString("\n")

	// 第 4 行固定展示应用完整 Git 版本。
	buf.WriteString("应用版本：")
	buf.WriteString(v.buildingContext.GitVersion())
	buf.WriteString("\n")

	// 第 5 行固定展示类库构建目录。
	buf.WriteString("类库目录：")
	buf.WriteString(v.buildingContext.BuildLibraryDirectory())
	buf.WriteString("\n")

	// 第 6 行固定展示应用工作目录。
	buf.WriteString("应用目录：")
	buf.WriteString(v.buildingContext.BuildWorkingDirectory())
	buf.WriteString("\n")

	// 第 7 行固定展示 GOROOT 编译工具目录。
	buf.WriteString("编译工具目录：")
	buf.WriteString(v.buildingContext.BuildGorootDirectory())
	buf.WriteString("\n")

	// 第 8 行固定展示 GOPATH 编译环境目录。
	buf.WriteString("编译环境目录：")
	buf.WriteString(v.buildingContext.BuildGopathDirectory())

	return buf.String()
}
