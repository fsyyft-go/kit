// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package build

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildingContextValue_Getters 验证 buildingContextValue 的构建信息读取契约。
//
// 该测试通过表驱动用例覆盖版本、Git、构建时间、目录和调试状态字段，确保 BuildingContext getter 不改变已保存的构建上下文值。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestBuildingContextValue_Getters(t *testing.T) {
	giveContext := &buildingContextValue{
		version:               "v1.2.3",
		gitVersion:            "0123456789abcdef0123456789abcdef01234567",
		libGitVersion:         "abcdef0123456789abcdef0123456789abcdef01",
		buildTimeString:       "20250618123456000",
		buildLibraryDirectory: "/workspace/kit",
		buildWorkingDirectory: "/workspace/app",
		buildGopathDirectory:  "/workspace/go",
		buildGorootDirectory:  "/usr/local/go",
		debug:                 true,
	}

	tests := []struct {
		name        string
		description string
		give        func() interface{}
		want        interface{}
	}{
		{
			name:        "success/version",
			description: "验证 Version 返回构建上下文中保存的软件版本号。",
			give:        func() interface{} { return giveContext.Version() },
			want:        "v1.2.3",
		},
		{
			name:        "success/git-version",
			description: "验证 GitVersion 返回完整应用 Git 版本。",
			give:        func() interface{} { return giveContext.GitVersion() },
			want:        "0123456789abcdef0123456789abcdef01234567",
		},
		{
			name:        "success/lib-git-version",
			description: "验证 LibGitVersion 返回完整类库 Git 版本。",
			give:        func() interface{} { return giveContext.LibGitVersion() },
			want:        "abcdef0123456789abcdef0123456789abcdef01",
		},
		{
			name:        "success/build-time-string",
			description: "验证 BuildTimeString 返回构建时间字符串。",
			give:        func() interface{} { return giveContext.BuildTimeString() },
			want:        "20250618123456000",
		},
		{
			name:        "success/build-library-directory",
			description: "验证 BuildLibraryDirectory 返回类库目录。",
			give:        func() interface{} { return giveContext.BuildLibraryDirectory() },
			want:        "/workspace/kit",
		},
		{
			name:        "success/build-working-directory",
			description: "验证 BuildWorkingDirectory 返回应用工作目录。",
			give:        func() interface{} { return giveContext.BuildWorkingDirectory() },
			want:        "/workspace/app",
		},
		{
			name:        "success/build-gopath-directory",
			description: "验证 BuildGopathDirectory 返回 GOPATH 目录。",
			give:        func() interface{} { return giveContext.BuildGopathDirectory() },
			want:        "/workspace/go",
		},
		{
			name:        "success/build-goroot-directory",
			description: "验证 BuildGorootDirectory 返回 GOROOT 目录。",
			give:        func() interface{} { return giveContext.BuildGorootDirectory() },
			want:        "/usr/local/go",
		},
		{
			name:        "success/debug",
			description: "验证 Debug 返回构建上下文中保存的调试状态。",
			give:        func() interface{} { return giveContext.Debug() },
			want:        true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			assert.Equal(t, tt.want, tt.give())
		})
	}
}

// TestBuildingContextValue_ShortVersions 验证 Git 短版本的截断与边界语义。
//
// 该测试通过表驱动用例覆盖空值、短值、刚好 8 位和超过 8 位场景，确保短版本函数只在长度超过 8 时截断。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestBuildingContextValue_ShortVersions(t *testing.T) {
	tests := []struct {
		name              string
		description       string
		giveGitVersion    string
		giveLibGitVersion string
		wantGitShort      string
		wantLibGitShort   string
	}{
		{
			name:        "boundary/empty",
			description: "验证空 Git 版本返回空短版本。",
		},
		{
			name:              "boundary/shorter-than-eight",
			description:       "验证少于 8 位的 Git 版本保持原值。",
			giveGitVersion:    "1234567",
			giveLibGitVersion: "abcdefg",
			wantGitShort:      "1234567",
			wantLibGitShort:   "abcdefg",
		},
		{
			name:              "boundary/eight-characters",
			description:       "验证刚好 8 位的 Git 版本保持原值。",
			giveGitVersion:    "12345678",
			giveLibGitVersion: "abcdefgh",
			wantGitShort:      "12345678",
			wantLibGitShort:   "abcdefgh",
		},
		{
			name:              "success/truncate-long-version",
			description:       "验证超过 8 位的 Git 版本截断为前 8 位。",
			giveGitVersion:    "1234567890abcdef",
			giveLibGitVersion: "abcdefghijklmnop",
			wantGitShort:      "12345678",
			wantLibGitShort:   "abcdefgh",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			giveContext := &buildingContextValue{
				gitVersion:    tt.giveGitVersion,
				libGitVersion: tt.giveLibGitVersion,
			}

			assert.Equal(t, tt.wantGitShort, giveContext.GitShortVersion())
			assert.Equal(t, tt.wantLibGitShort, giveContext.LibGitShortVersion())
		})
	}
}

// TestNewBuildingContext_DebugBranches 验证构建上下文初始化的调试与非调试分支。
//
// 该测试通过表驱动用例覆盖调试模式补齐构建时间和运行目录，以及非调试模式保持未注入字段为空的契约。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewBuildingContext_DebugBranches(t *testing.T) {
	t.Setenv(GoEnvNamePath, "/tmp/test-gopath")
	t.Setenv(GoEnvNameRoot, "/tmp/test-goroot")

	tests := []struct {
		name          string
		description   string
		giveDebug     bool
		wantDebug     bool
		wantGopath    string
		wantGoroot    string
		wantWorkdir   bool
		wantBuildTime bool
	}{
		{
			name:          "success/debug-fills-runtime-fields",
			description:   "验证调试模式会补齐构建时间、工作目录、GOPATH 和 GOROOT。",
			giveDebug:     true,
			wantDebug:     true,
			wantGopath:    "/tmp/test-gopath",
			wantGoroot:    "/tmp/test-goroot",
			wantWorkdir:   true,
			wantBuildTime: true,
		},
		{
			name:          "success/non-debug-keeps-empty-runtime-fields",
			description:   "验证非调试模式在未通过 ldflags 注入时不补齐运行时目录和构建时间。",
			giveDebug:     false,
			wantDebug:     false,
			wantBuildTime: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := newBuildingContext(tt.giveDebug)

			require.NotNil(t, got)
			assert.Equal(t, tt.wantDebug, got.Debug())
			assert.Equal(t, version, got.Version())
			assert.Equal(t, gitVersion, got.GitVersion())
			assert.Equal(t, libGitVersion, got.LibGitVersion())
			assert.Equal(t, tt.wantGopath, got.BuildGopathDirectory())
			assert.Equal(t, tt.wantGoroot, got.BuildGorootDirectory())
			if tt.wantWorkdir {
				assert.NotEmpty(t, got.BuildWorkingDirectory())
			} else {
				assert.Empty(t, got.BuildWorkingDirectory())
			}
			if tt.wantBuildTime {
				_, err := time.Parse(TimeLayout, got.BuildTimeString())
				require.NoError(t, err)
			} else {
				assert.Empty(t, got.BuildTimeString())
			}
		})
	}
}

// TestDefaultCheckDebug_TestProcess 验证默认调试检测能识别 go test 进程。
//
// 该测试断言当前测试二进制会被 defaultCheckDebug 判定为调试上下文，以保护包初始化时的 go test 回归语义。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestDefaultCheckDebug_TestProcess(t *testing.T) {
	// 该用例验证 go test 进程名和测试标志足以让默认检测逻辑返回 true。
	assert.True(t, defaultCheckDebug())
}

// TestCurrentBuildingContext_InitContract 验证包初始化后的当前构建上下文契约。
//
// 该测试只断言 go test 环境下稳定成立的初始化结果，包括接口实现、调试模式、构建时间格式和调试目录填充，避免依赖具体机器路径。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestCurrentBuildingContext_InitContract(t *testing.T) {
	// 该用例验证包 init 会创建可用的当前构建上下文，并在 go test 环境中识别为调试模式。
	require.NotNil(t, CurrentBuildingContext)
	assert.Implements(t, (*BuildingContext)(nil), CurrentBuildingContext)
	assert.True(t, CurrentBuildingContext.Debug())
	assert.NotEmpty(t, CurrentBuildingContext.BuildTimeString())

	_, err := time.Parse(TimeLayout, CurrentBuildingContext.BuildTimeString())
	require.NoError(t, err)

	assert.NotEmpty(t, CurrentBuildingContext.BuildWorkingDirectory())
	assert.Equal(t, version, CurrentBuildingContext.Version())
	assert.Equal(t, gitVersion, CurrentBuildingContext.GitVersion())
	assert.Equal(t, libGitVersion, CurrentBuildingContext.LibGitVersion())
}

// TestNewBuildingContext_PreservesInjectedFields 验证构建上下文初始化保留编译期注入字段。
//
// 该测试临时设置包级构建变量并在用例结束后恢复，确保 newBuildingContext 不覆盖已注入的版本、时间和目录信息。
//
// 参数：
//   - t: 测试上下文，用于注册恢复逻辑和报告断言失败。
func TestNewBuildingContext_PreservesInjectedFields(t *testing.T) {
	// 该用例验证调试模式也不会覆盖已经通过包级变量注入的构建字段。
	oldVersion := version
	oldGitVersion := gitVersion
	oldLibGitVersion := libGitVersion
	oldBuildTimeString := buildTimeString
	oldBuildLibraryDirectory := buildLibraryDirectory
	oldBuildWorkingDirectory := buildWorkingDirectory
	oldBuildGopathDirectory := buildGopathDirectory
	oldBuildGorootDirectory := buildGorootDirectory
	oldIsDebug := isDebug
	t.Cleanup(func() {
		version = oldVersion
		gitVersion = oldGitVersion
		libGitVersion = oldLibGitVersion
		buildTimeString = oldBuildTimeString
		buildLibraryDirectory = oldBuildLibraryDirectory
		buildWorkingDirectory = oldBuildWorkingDirectory
		buildGopathDirectory = oldBuildGopathDirectory
		buildGorootDirectory = oldBuildGorootDirectory
		isDebug = oldIsDebug
	})

	version = "v9.9.9"
	gitVersion = "1234567890abcdef"
	libGitVersion = "abcdef1234567890"
	buildTimeString = "20250102030405006"
	buildLibraryDirectory = "/injected/lib"
	buildWorkingDirectory = "/injected/work"
	buildGopathDirectory = "/injected/gopath"
	buildGorootDirectory = "/injected/goroot"
	isDebug = true

	got := newBuildingContext(true)

	require.NotNil(t, got)
	assert.Equal(t, "v9.9.9", got.Version())
	assert.Equal(t, "1234567890abcdef", got.GitVersion())
	assert.Equal(t, "abcdef1234567890", got.LibGitVersion())
	assert.Equal(t, "20250102030405006", got.BuildTimeString())
	assert.Equal(t, "/injected/lib", got.BuildLibraryDirectory())
	assert.Equal(t, "/injected/work", got.BuildWorkingDirectory())
	assert.Equal(t, "/injected/gopath", got.BuildGopathDirectory())
	assert.Equal(t, "/injected/goroot", got.BuildGorootDirectory())
}

// TestDefaultCheckDebug_GoBuildTempDirNormalization 验证 GOTMPDIR 规范化不会破坏调试检测。
//
// 该测试将 GOTMPDIR 设置为不带路径分隔符的值，确保 defaultCheckDebug 在测试进程下仍可稳定识别调试上下文。
//
// 参数：
//   - t: 测试上下文，用于设置环境变量和报告断言失败。
func TestDefaultCheckDebug_GoBuildTempDirNormalization(t *testing.T) {
	// 该用例覆盖 GOTMPDIR 不以路径分隔符结尾时的路径补齐分支。
	t.Setenv(GoEnvNameTmpDir, t.TempDir())

	assert.True(t, defaultCheckDebug())
}
