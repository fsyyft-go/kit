// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package build

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 定义构建相关的常量。
const (
	// TimeLayout 定义时间格式化模板，格式为：年月日时分秒毫秒。
	TimeLayout = "20060102150405000"
	// GoEnvNameRoot 定义 Go 运行环境的根目录环境变量名。
	GoEnvNameRoot = "GOROOT"
	// GoEnvNamePath 定义 Go 运行环境的路径环境变量名。
	GoEnvNamePath = "GOPATH"
	// GoEnvNameTmpDir 定义 Go 运行环境的临时目录环境变量名。
	GoEnvNameTmpDir = "GOTMPDIR"
)

var (
	// 确保 buildingContextValue 实现了 BuildingContext 接口。
	_ BuildingContext = (*buildingContextValue)(nil)
)

type (
	// BuildingContext 定义构建上下文接口，提供访问构建时信息的方法。
	BuildingContext interface {
		// Version 获取软件版本。
		//
		// 返回值：
		//   - string：软件版本号。
		Version() string

		// GitVersion 获取完整的 Git 版本号。
		//
		// 返回值：
		//   - string：完整的 Git 提交哈希值。
		GitVersion() string

		// GitShortVersion 获取短格式的 Git 版本号。
		//
		// 返回值：
		//   - string：Git 提交哈希值的前 8 位字符。
		GitShortVersion() string

		// LibGitVersion 获取类库的完整 Git 版本号。
		//
		// 返回值：
		//   - string：类库的完整 Git 提交哈希值。
		LibGitVersion() string

		// LibGitShortVersion 获取类库的短格式 Git 版本号。
		//
		// 返回值：
		//   - string：类库的 Git 提交哈希值的前 8 位字符。
		LibGitShortVersion() string

		// BuildTimeString 获取构建时间字符串。
		//
		// 返回值：
		//   - string：构建时间，格式为 "20060102150405000"。
		BuildTimeString() string

		// BuildLibraryDirectory 获取构建时类库所在目录。
		//
		// 返回值：
		//   - string：类库的绝对路径。
		BuildLibraryDirectory() string

		// BuildWorkingDirectory 获取构建时的工作目录。
		//
		// 返回值：
		//   - string：工作目录的绝对路径。
		BuildWorkingDirectory() string

		// BuildGopathDirectory 获取构建时的 GOPATH 目录。
		//
		// 返回值：
		//   - string：GOPATH 环境变量指定的目录路径。
		BuildGopathDirectory() string

		// BuildGorootDirectory 获取构建时的 GOROOT 目录。
		//
		// 返回值：
		//   - string：GOROOT 环境变量指定的目录路径。
		BuildGorootDirectory() string

		// Debug 获取是否为调试状态。
		//
		// 返回值：
		//   - bool：true 表示处于调试状态，false 表示处于正常状态。
		Debug() bool
	}

	// buildingContextValue 实现 BuildingContext 接口，存储构建时的上下文信息。
	buildingContextValue struct {
		version               string // 软件版本。
		gitVersion            string // Git 版本。
		libGitVersion         string // 类库的 Git 版本。
		buildTimeString       string // 编译时间。
		buildLibraryDirectory string // 编译时的类库所在目录。
		buildWorkingDirectory string // 编译时的工作目录。
		buildGopathDirectory  string // 编译时的 GOPATH 目录。
		buildGorootDirectory  string // 编译时的 GOROOT 目录。
		debug                 bool   // 是否调试状态。
	}
)

// Version 获取软件版本。
//
// 返回值：
//   - string：软件版本号。
func (c *buildingContextValue) Version() string {
	return c.version
}

// GitVersion 获取完整的 Git 版本号。
//
// 返回值：
//   - string：完整的 Git 提交哈希值。
func (c *buildingContextValue) GitVersion() string {
	return c.gitVersion
}

// GitShortVersion 获取短格式的 Git 版本号。
//
// 返回值：
//   - string：Git 提交哈希值的前 8 位字符。
func (c *buildingContextValue) GitShortVersion() string {
	shortVersion := c.gitVersion
	if len(shortVersion) > 8 {
		shortVersion = shortVersion[:8]
	}
	return shortVersion
}

// LibGitVersion 获取类库的完整 Git 版本号。
//
// 返回值：
//   - string：类库的完整 Git 提交哈希值。
func (c *buildingContextValue) LibGitVersion() string {
	return c.libGitVersion
}

// LibGitShortVersion 获取类库的短格式 Git 版本号。
//
// 返回值：
//   - string：类库的 Git 提交哈希值的前 8 位字符。
func (c *buildingContextValue) LibGitShortVersion() string {
	shortVersion := c.libGitVersion
	if len(shortVersion) > 8 {
		shortVersion = shortVersion[:8]
	}
	return shortVersion
}

// BuildTimeString 获取构建时间字符串。
//
// 返回值：
//   - string：构建时间，格式为 "20060102150405000"。
func (c *buildingContextValue) BuildTimeString() string {
	return c.buildTimeString
}

// BuildLibraryDirectory 获取构建时类库所在目录。
//
// 返回值：
//   - string：类库的绝对路径。
func (c *buildingContextValue) BuildLibraryDirectory() string {
	return c.buildLibraryDirectory
}

// BuildWorkingDirectory 获取构建时的工作目录。
//
// 返回值：
//   - string：工作目录的绝对路径。
func (c *buildingContextValue) BuildWorkingDirectory() string {
	return c.buildWorkingDirectory
}

// BuildGopathDirectory 获取构建时的 GOPATH 目录。
//
// 返回值：
//   - string：GOPATH 环境变量指定的目录路径。
func (c *buildingContextValue) BuildGopathDirectory() string {
	return c.buildGopathDirectory
}

// BuildGorootDirectory 获取构建时的 GOROOT 目录。
//
// 返回值：
//   - string：GOROOT 环境变量指定的目录路径。
func (c *buildingContextValue) BuildGorootDirectory() string {
	return c.buildGorootDirectory
}

// Debug 获取是否为调试状态。
//
// 返回值：
//   - bool：true 表示处于调试状态，false 表示处于正常状态。
func (c *buildingContextValue) Debug() bool {
	return c.debug
}

var (
	version               string // 软件版本。
	gitVersion            string // Git 版本。
	libGitVersion         string // 类库的 Git 版本。
	buildTimeString       string // 编译时间。
	buildLibraryDirectory string // 编译时的类库所在目录。
	buildWorkingDirectory string // 编译时的工作目录。
	// buildWorkingDirectory = "/development/Learning/Go"
	buildGopathDirectory string // 编译时的 GOPATH 目录。
	// buildGopathDirectory = "/data/var/lib/go/path"
	buildGorootDirectory string // 编译时的 GOROOT 目录。
	isDebug              bool   // 是否调试状态。
)

var (
	// CurrentBuildingContext 存储当前的构建上下文信息。
	CurrentBuildingContext *buildingContextValue
)

// init 初始化构建上下文，设置默认值和运行时状态。
func init() {
	funcCheckDebug := func() bool {
		// 获取临时构建目录。
		goTmpDir := os.Getenv(GoEnvNameTmpDir)
		if goTmpDir == "" {
			// 如果未设置 GOTMPDIR，则使用系统临时目录。
			goTmpDir = os.TempDir()
		}

		// 确保路径以分隔符结尾。
		if !strings.HasSuffix(goTmpDir, string(os.PathSeparator)) {
			goTmpDir = goTmpDir + string(os.PathSeparator)
		}

		// 构建 go-build 临时目录路径。
		goBuildDir := filepath.Join(goTmpDir, "go-build")

		// 获取当前执行文件的路径。
		exePath, err := os.Executable()
		if err != nil {
			return false
		}

		// 规范化路径，处理符号链接。
		exePath, err = filepath.EvalSymlinks(exePath)
		if err != nil {
			return false
		}

		// 获取程序名称。
		exeName := filepath.Base(os.Args[0])

		// 检查是否为测试模式。
		isTestMode := strings.HasSuffix(exeName, ".test") ||
			strings.Contains(exeName, ".test.") ||
			strings.HasPrefix(exeName, "_testmain")

		// 检查是否通过 go run 运行。
		isGoRun := strings.HasPrefix(
			strings.ToLower(exePath),
			strings.ToLower(goBuildDir),
		)

		// 检查环境变量，判断是否在 go 命令执行环境中。
		goCommand := os.Getenv("GOEXE") != "" || // go 命令设置的环境变量
			os.Getenv("GOPATH") != "" ||
			os.Getenv("GOROOT") != ""

		// 检查命令行参数是否包含测试相关标志。
		hasTestFlags := false
		for _, arg := range os.Args[1:] {
			if strings.HasPrefix(arg, "-test.") {
				hasTestFlags = true
				break
			}
		}

		return isTestMode || isGoRun || (goCommand && hasTestFlags)
	}

	// 初始化当前构建上下文对象。
	CurrentBuildingContext = &buildingContextValue{
		version:               version,
		gitVersion:            gitVersion,
		libGitVersion:         libGitVersion,
		buildTimeString:       buildTimeString,
		buildLibraryDirectory: buildLibraryDirectory,
		buildWorkingDirectory: buildWorkingDirectory,
		buildGopathDirectory:  buildGopathDirectory,
		buildGorootDirectory:  buildGorootDirectory,
		debug:                 isDebug,
	}

	// 获取当前运行环境的临时目录路径。
	// goTmpBuild := os.Getenv(GoEnvNameTmpDir)
	// 检查路径是否以分隔符结尾，如果没有则添加。
	// noPathSeparator := len(goTmpBuild) > 0 && goTmpBuild[len(goTmpBuild)-1] != os.PathSeparator
	// if noPathSeparator {
	// 	goTmpBuild = filepath.Join(goTmpBuild, string(os.PathSeparator))
	// }
	// 拼接 go-build 目录路径。
	// goTmpBuild = filepath.Join(goTmpBuild, "go-build")

	// 通过检查执行文件路径判断是否为调试模式（go run 或 go test）。
	// exeIndex := strings.Index(strings.ToLower(os.Args[0]), strings.ToLower(goTmpBuild))
	// CurrentBuildingContext.debug = exeIndex >= 0 && exeIndex <= 7
	CurrentBuildingContext.debug = funcCheckDebug()

	// 设置构建时间(如果未设置则使用当前时间)。
	if len(CurrentBuildingContext.buildTimeString) == 0 {
		if CurrentBuildingContext.debug {
			// 调试模式使用当前时间。
			CurrentBuildingContext.buildTimeString = time.Now().Format(TimeLayout)
		} else {
			// 非调试模式使用默认时间格式。
			CurrentBuildingContext.buildTimeString = buildTimeString
		}
	}

	// 在调试模式下设置额外的环境信息。
	if CurrentBuildingContext.debug {
		// 设置工作目录。
		if len(CurrentBuildingContext.buildWorkingDirectory) == 0 {
			if pwd, err := os.Getwd(); nil == err {
				CurrentBuildingContext.buildWorkingDirectory = pwd
			}
		}

		// 设置 GOPATH 目录。
		if len(CurrentBuildingContext.buildGopathDirectory) == 0 {
			CurrentBuildingContext.buildGopathDirectory = os.Getenv(GoEnvNamePath)
		}

		// 设置 GOROOT 目录。
		if len(CurrentBuildingContext.buildGorootDirectory) == 0 {
			CurrentBuildingContext.buildGorootDirectory = os.Getenv(GoEnvNameRoot)
		}
	}
}
