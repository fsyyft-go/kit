package config

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	kitgobuild "github.com/fsyyft-go/kit/go/build"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ kitgobuild.BuildingContext = (*fakeBuildingContext)(nil)

// fakeBuildingContext 提供稳定的构建上下文测试替身。
//
// 该辅助类型完整实现 kitgobuild.BuildingContext 接口，使 version 的代理方法、格式化和描述输出可在无外部依赖的条件下验证。
type fakeBuildingContext struct {
	version               string
	gitVersion            string
	gitShortVersion       string
	libGitVersion         string
	libGitShortVersion    string
	buildTimeString       string
	buildLibraryDirectory string
	buildWorkingDirectory string
	buildGopathDirectory  string
	buildGorootDirectory  string
	debug                 bool
}

// Version 返回测试替身中的软件版本号。
//
// 返回：
//   - string: 软件版本号。
func (c *fakeBuildingContext) Version() string {
	return c.version
}

// GitVersion 返回测试替身中的完整 Git 版本号。
//
// 返回：
//   - string: 完整 Git 版本号。
func (c *fakeBuildingContext) GitVersion() string {
	return c.gitVersion
}

// GitShortVersion 返回测试替身中的短格式 Git 版本号。
//
// 返回：
//   - string: 短格式 Git 版本号。
func (c *fakeBuildingContext) GitShortVersion() string {
	return c.gitShortVersion
}

// LibGitVersion 返回测试替身中的类库完整 Git 版本号。
//
// 返回：
//   - string: 类库完整 Git 版本号。
func (c *fakeBuildingContext) LibGitVersion() string {
	return c.libGitVersion
}

// LibGitShortVersion 返回测试替身中的类库短格式 Git 版本号。
//
// 返回：
//   - string: 类库短格式 Git 版本号。
func (c *fakeBuildingContext) LibGitShortVersion() string {
	return c.libGitShortVersion
}

// BuildTimeString 返回测试替身中的构建时间字符串。
//
// 返回：
//   - string: 构建时间字符串。
func (c *fakeBuildingContext) BuildTimeString() string {
	return c.buildTimeString
}

// BuildLibraryDirectory 返回测试替身中的类库目录。
//
// 返回：
//   - string: 类库目录。
func (c *fakeBuildingContext) BuildLibraryDirectory() string {
	return c.buildLibraryDirectory
}

// BuildWorkingDirectory 返回测试替身中的工作目录。
//
// 返回：
//   - string: 工作目录。
func (c *fakeBuildingContext) BuildWorkingDirectory() string {
	return c.buildWorkingDirectory
}

// BuildGopathDirectory 返回测试替身中的 GOPATH 目录。
//
// 返回：
//   - string: GOPATH 目录。
func (c *fakeBuildingContext) BuildGopathDirectory() string {
	return c.buildGopathDirectory
}

// BuildGorootDirectory 返回测试替身中的 GOROOT 目录。
//
// 返回：
//   - string: GOROOT 目录。
func (c *fakeBuildingContext) BuildGorootDirectory() string {
	return c.buildGorootDirectory
}

// Debug 返回测试替身中的调试状态。
//
// 返回：
//   - bool: 调试状态。
func (c *fakeBuildingContext) Debug() bool {
	return c.debug
}

// newTestVersion 构造带有稳定构建上下文的 version 测试对象。
//
// 该辅助函数为格式化和代理方法测试集中提供可预测的构建上下文，避免测试依赖真实构建环境。
//
// 返回：
//   - *version: 可直接用于断言的版本对象。
//   - *fakeBuildingContext: 版本对象持有的测试替身上下文。
func newTestVersion() (*version, *fakeBuildingContext) {
	ctx := &fakeBuildingContext{
		version:               "v1.2.3",
		gitVersion:            "0123456789abcdef0123456789abcdef01234567",
		gitShortVersion:       "01234567",
		libGitVersion:         "abcdef0123456789abcdef0123456789abcdef01",
		libGitShortVersion:    "abcdef01",
		buildTimeString:       "20250618123456000",
		buildLibraryDirectory: "/workspace/kit",
		buildWorkingDirectory: "/workspace/app",
		buildGopathDirectory:  "/workspace/go",
		buildGorootDirectory:  "/usr/local/go",
		debug:                 true,
	}

	return &version{buildingContext: ctx}, ctx
}

// TestVersion_DelegatesBuildingContext 验证 version 的代理方法返回底层构建上下文结果。
//
// 该测试通过表驱动用例覆盖所有 BuildingContext 代理方法，确保 version 不改变底层上下文的版本、路径、构建时间和调试状态语义。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestVersion_DelegatesBuildingContext(t *testing.T) {
	giveVersion, giveContext := newTestVersion()

	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "success/version",
			description: "验证 Version 代理返回底层构建上下文的软件版本号。",
			assert: func(t *testing.T) {
				assert.Equal(t, giveContext.version, giveVersion.Version())
			},
		},
		{
			name:        "success/git-version",
			description: "验证 GitVersion 代理返回底层构建上下文的完整应用 Git 版本。",
			assert: func(t *testing.T) {
				assert.Equal(t, giveContext.gitVersion, giveVersion.GitVersion())
			},
		},
		{
			name:        "success/git-short-version",
			description: "验证 GitShortVersion 代理返回底层构建上下文的短格式应用 Git 版本。",
			assert: func(t *testing.T) {
				assert.Equal(t, giveContext.gitShortVersion, giveVersion.GitShortVersion())
			},
		},
		{
			name:        "success/lib-git-version",
			description: "验证 LibGitVersion 代理返回底层构建上下文的完整类库 Git 版本。",
			assert: func(t *testing.T) {
				assert.Equal(t, giveContext.libGitVersion, giveVersion.LibGitVersion())
			},
		},
		{
			name:        "success/lib-git-short-version",
			description: "验证 LibGitShortVersion 代理返回底层构建上下文的短格式类库 Git 版本。",
			assert: func(t *testing.T) {
				assert.Equal(t, giveContext.libGitShortVersion, giveVersion.LibGitShortVersion())
			},
		},
		{
			name:        "success/build-time-string",
			description: "验证 BuildTimeString 代理返回底层构建上下文的构建时间字符串。",
			assert: func(t *testing.T) {
				assert.Equal(t, giveContext.buildTimeString, giveVersion.BuildTimeString())
			},
		},
		{
			name:        "success/build-library-directory",
			description: "验证 BuildLibraryDirectory 代理返回底层构建上下文的类库目录。",
			assert: func(t *testing.T) {
				assert.Equal(t, giveContext.buildLibraryDirectory, giveVersion.BuildLibraryDirectory())
			},
		},
		{
			name:        "success/build-working-directory",
			description: "验证 BuildWorkingDirectory 代理返回底层构建上下文的应用工作目录。",
			assert: func(t *testing.T) {
				assert.Equal(t, giveContext.buildWorkingDirectory, giveVersion.BuildWorkingDirectory())
			},
		},
		{
			name:        "success/build-gopath-directory",
			description: "验证 BuildGopathDirectory 代理返回底层构建上下文的 GOPATH 目录。",
			assert: func(t *testing.T) {
				assert.Equal(t, giveContext.buildGopathDirectory, giveVersion.BuildGopathDirectory())
			},
		},
		{
			name:        "success/build-goroot-directory",
			description: "验证 BuildGorootDirectory 代理返回底层构建上下文的 GOROOT 目录。",
			assert: func(t *testing.T) {
				assert.Equal(t, giveContext.buildGorootDirectory, giveVersion.BuildGorootDirectory())
			},
		},
		{
			name:        "success/debug",
			description: "验证 Debug 代理返回底层构建上下文的调试状态。",
			assert: func(t *testing.T) {
				assert.Equal(t, giveContext.debug, giveVersion.Debug())
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			require.NotNil(t, tt.assert)
			tt.assert(t)
		})
	}
}

// TestVersion_String 验证 version 的简短字符串格式。
//
// 该测试通过表驱动用例覆盖 String 的稳定输出契约，确保其使用短 Git 版本、构建时间和去除 go 前缀后的 runtime.Version 信息。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestVersion_String(t *testing.T) {
	giveVersion, giveContext := newTestVersion()

	tests := []struct {
		name        string
		description string
		giveVersion *version
		want        string
	}{
		{
			name:        "success/simple-format",
			description: "验证 String 输出包含短应用 Git 版本、构建时间和去除 go 前缀后的 Go 运行时版本。",
			giveVersion: giveVersion,
			want: fmt.Sprintf("version %s/%s (build %s)",
				giveContext.gitShortVersion,
				giveContext.buildTimeString,
				strings.ReplaceAll(runtime.Version(), "go", ""),
			),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			require.NotNil(t, tt.giveVersion)
			assert.Equal(t, tt.want, tt.giveVersion.String())
		})
	}
}

// TestVersion_Format 验证 version 的 fmt.Formatter 输出分支。
//
// 该测试通过表驱动用例覆盖 %+v 的详细描述输出，以及 %v、%s、%q 等非 %+v 格式回退到 String 输出的契约。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestVersion_Format(t *testing.T) {
	giveVersion, _ := newTestVersion()

	tests := []struct {
		name        string
		description string
		giveFormat  string
		want        string
	}{
		{
			name:        "success/plus-v-description",
			description: "验证 %+v 格式输出完整 Description 内容。",
			giveFormat:  "%+v",
			want:        giveVersion.Description(),
		},
		{
			name:        "success/v-string",
			description: "验证 %v 格式输出简短 String 内容。",
			giveFormat:  "%v",
			want:        giveVersion.String(),
		},
		{
			name:        "success/s-string",
			description: "验证 %s 格式输出简短 String 内容。",
			giveFormat:  "%s",
			want:        giveVersion.String(),
		},
		{
			name:        "success/q-string",
			description: "验证非 %+v 的 %q 格式仍输出简短 String 内容且不额外加引号。",
			giveFormat:  "%q",
			want:        giveVersion.String(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			assert.Equal(t, tt.want, fmt.Sprintf(tt.giveFormat, *giveVersion))
		})
	}
}

// TestVersion_Description 验证 version 的详细中文描述格式。
//
// 该测试通过表驱动用例覆盖 Description 的 8 行中文标签顺序与内容，确保详细版本信息输出可被稳定解析。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestVersion_Description(t *testing.T) {
	giveVersion, giveContext := newTestVersion()

	tests := []struct {
		name        string
		description string
		giveVersion *version
		wantLines   []string
	}{
		{
			name:        "success/chinese-label-order",
			description: "验证 Description 按固定顺序输出开发版本、编译时间、类库版本、应用版本、类库目录、应用目录、编译工具目录和编译环境目录。",
			giveVersion: giveVersion,
			wantLines: []string{
				"开发版本：" + runtime.Version(),
				"编译时间：" + giveContext.buildTimeString,
				"类库版本：" + giveContext.libGitVersion,
				"应用版本：" + giveContext.gitVersion,
				"类库目录：" + giveContext.buildLibraryDirectory,
				"应用目录：" + giveContext.buildWorkingDirectory,
				"编译工具目录：" + giveContext.buildGorootDirectory,
				"编译环境目录：" + giveContext.buildGopathDirectory,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			require.NotNil(t, tt.giveVersion)
			got := tt.giveVersion.Description()
			want := strings.Join(tt.wantLines, "\n")

			assert.Equal(t, want, got)
			assert.Len(t, strings.Split(got, "\n"), 8)
		})
	}
}

// TestCurrentVersion_BasicInterfaces 验证 CurrentVersion 的基础接口行为。
//
// 该测试仅断言 CurrentVersion 与其底层构建上下文之间的代理一致性和格式化分支，不依赖具体构建环境中的版本号、目录或时间值。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestCurrentVersion_BasicInterfaces(t *testing.T) {
	require.NotNil(t, CurrentVersion.buildingContext)

	current := &CurrentVersion
	assert.Implements(t, (*fmt.Stringer)(nil), current)
	assert.Implements(t, (*fmt.Formatter)(nil), current)
	assert.Implements(t, (*kitgobuild.BuildingContext)(nil), current)

	tests := []struct {
		name        string
		description string
		assert      func(t *testing.T)
	}{
		{
			name:        "success/version-proxy",
			description: "验证 CurrentVersion 的 Version 与底层构建上下文保持一致。",
			assert: func(t *testing.T) {
				assert.Equal(t, CurrentVersion.buildingContext.Version(), current.Version())
			},
		},
		{
			name:        "success/git-version-proxy",
			description: "验证 CurrentVersion 的 GitVersion 与底层构建上下文保持一致。",
			assert: func(t *testing.T) {
				assert.Equal(t, CurrentVersion.buildingContext.GitVersion(), current.GitVersion())
			},
		},
		{
			name:        "success/git-short-version-proxy",
			description: "验证 CurrentVersion 的 GitShortVersion 与底层构建上下文保持一致。",
			assert: func(t *testing.T) {
				assert.Equal(t, CurrentVersion.buildingContext.GitShortVersion(), current.GitShortVersion())
			},
		},
		{
			name:        "success/lib-git-version-proxy",
			description: "验证 CurrentVersion 的 LibGitVersion 与底层构建上下文保持一致。",
			assert: func(t *testing.T) {
				assert.Equal(t, CurrentVersion.buildingContext.LibGitVersion(), current.LibGitVersion())
			},
		},
		{
			name:        "success/lib-git-short-version-proxy",
			description: "验证 CurrentVersion 的 LibGitShortVersion 与底层构建上下文保持一致。",
			assert: func(t *testing.T) {
				assert.Equal(t, CurrentVersion.buildingContext.LibGitShortVersion(), current.LibGitShortVersion())
			},
		},
		{
			name:        "success/build-time-string-proxy",
			description: "验证 CurrentVersion 的 BuildTimeString 与底层构建上下文保持一致。",
			assert: func(t *testing.T) {
				assert.Equal(t, CurrentVersion.buildingContext.BuildTimeString(), current.BuildTimeString())
			},
		},
		{
			name:        "success/build-library-directory-proxy",
			description: "验证 CurrentVersion 的 BuildLibraryDirectory 与底层构建上下文保持一致。",
			assert: func(t *testing.T) {
				assert.Equal(t, CurrentVersion.buildingContext.BuildLibraryDirectory(), current.BuildLibraryDirectory())
			},
		},
		{
			name:        "success/build-working-directory-proxy",
			description: "验证 CurrentVersion 的 BuildWorkingDirectory 与底层构建上下文保持一致。",
			assert: func(t *testing.T) {
				assert.Equal(t, CurrentVersion.buildingContext.BuildWorkingDirectory(), current.BuildWorkingDirectory())
			},
		},
		{
			name:        "success/build-gopath-directory-proxy",
			description: "验证 CurrentVersion 的 BuildGopathDirectory 与底层构建上下文保持一致。",
			assert: func(t *testing.T) {
				assert.Equal(t, CurrentVersion.buildingContext.BuildGopathDirectory(), current.BuildGopathDirectory())
			},
		},
		{
			name:        "success/build-goroot-directory-proxy",
			description: "验证 CurrentVersion 的 BuildGorootDirectory 与底层构建上下文保持一致。",
			assert: func(t *testing.T) {
				assert.Equal(t, CurrentVersion.buildingContext.BuildGorootDirectory(), current.BuildGorootDirectory())
			},
		},
		{
			name:        "success/debug-proxy",
			description: "验证 CurrentVersion 的 Debug 与底层构建上下文保持一致。",
			assert: func(t *testing.T) {
				assert.Equal(t, CurrentVersion.buildingContext.Debug(), current.Debug())
			},
		},
		{
			name:        "success/string-format",
			description: "验证 CurrentVersion 的默认格式化输出与 String 保持一致。",
			assert: func(t *testing.T) {
				assert.Equal(t, current.String(), fmt.Sprintf("%v", CurrentVersion))
			},
		},
		{
			name:        "success/description-format",
			description: "验证 CurrentVersion 的详细格式化输出与 Description 保持一致。",
			assert: func(t *testing.T) {
				assert.Equal(t, current.Description(), fmt.Sprintf("%+v", CurrentVersion))
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			require.NotNil(t, tt.assert)
			tt.assert(t)
		})
	}
}
