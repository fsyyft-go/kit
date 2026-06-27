// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	kitcryptodes "github.com/fsyyft-go/kit/crypto/des"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDesCommand_Behavior 验证 DES 子命令的加密、解密和错误处理契约。
//
// 该测试通过表驱动用例覆盖默认密钥加密、自定义密钥解密、缺少数据、非法密钥和非法密文，确保命令行为无需外部依赖即可回归。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestDesCommand_Behavior(t *testing.T) {
	validCiphertext := mustEncryptDESForCommand(t, "go-kit-k", "敏感配置")
	tests := []struct {
		name           string
		description    string
		giveArgs       []string
		wantErr        bool
		wantErrContain string
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:        "success/default-key-encrypt",
			description: "验证 des 命令省略 key 时使用默认 DES 密钥加密 data 并输出原始数据与密文。",
			giveArgs:    []string{"des", "--data", "敏感配置", "--encrypt=true"},
			wantContains: []string{
				"原始数据: 敏感配置",
				"操作结果: " + validCiphertext,
			},
		},
		{
			name:        "success/custom-key-decrypt",
			description: "验证 des 命令在 encrypt=false 时执行解密而不是再次加密。",
			giveArgs: []string{
				"des",
				"--key", "12345678",
				"--data", mustEncryptDESForCommand(t, "12345678", "plain text"),
				"--encrypt=false",
			},
			wantContains: []string{
				"操作结果: plain text",
			},
			wantNotContain: []string{
				"发生错误:",
			},
		},
		{
			name:           "error/missing-required-data",
			description:    "验证缺少必填 data 标志时由 Cobra 返回参数校验错误且不执行命令体。",
			giveArgs:       []string{"des"},
			wantErr:        true,
			wantErrContain: "required flag(s) \"data\" not set",
		},
		{
			name:           "error/empty-data-after-flag-validation",
			description:    "验证 data 标志显式设置为空字符串时由命令体返回业务校验错误。",
			giveArgs:       []string{"des", "--data", ""},
			wantErr:        true,
			wantErrContain: "数据不能为空",
		},
		{
			name:        "error/invalid-key-is-reported",
			description: "验证非法 DES 密钥长度不会使命令失败退出，但会在输出中报告加密错误。",
			giveArgs:    []string{"des", "--key", "bad", "--data", "payload", "--encrypt=true"},
			wantContains: []string{
				"发生错误:",
				"invalid key size 3",
			},
			wantNotContain: []string{
				"操作结果:",
			},
		},
		{
			name:        "error/invalid-ciphertext-is-reported",
			description: "验证解密非法 16 进制密文时在输出中报告错误且不输出操作结果。",
			giveArgs:    []string{"des", "--key", "12345678", "--data", "not-hex", "--encrypt=false"},
			wantContains: []string{
				"发生错误:",
			},
			wantNotContain: []string{
				"操作结果:",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			gotOutput, gotErr := executeKratosConfigCommand(t, tt.giveArgs...)

			if tt.wantErr {
				require.Error(t, gotErr)
				assert.Contains(t, gotErr.Error(), tt.wantErrContain)
				return
			}

			require.NoError(t, gotErr)
			for _, want := range tt.wantContains {
				assert.Contains(t, gotOutput, want)
			}
			for _, notWant := range tt.wantNotContain {
				assert.NotContains(t, gotOutput, notWant)
			}
		})
	}
}

// TestRootCommand_Behavior 验证根命令的命令元数据和参数分支。
//
// 该测试覆盖 root 命令的基本 Cobra 配置、无子命令执行示例分支，以及存在非子命令参数时的空操作兼容分支。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestRootCommand_Behavior(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveArgs    []string
		assert      func(t *testing.T, output string, err error)
	}{
		{
			name:        "success/metadata",
			description: "验证 root 命令保留 config 用法、说明和可执行入口。",
			giveArgs:    []string{"ignored"},
			assert: func(t *testing.T, output string, err error) {
				assert.Equal(t, "config", rootCmd.Use)
				assert.Equal(t, "配置工具", rootCmd.Short)
				assert.Contains(t, rootCmd.Long, "DES 加密解密")
				require.NotNil(t, rootCmd.RunE)
				require.NoError(t, err)
				assert.Empty(t, output)
			},
		},
		{
			name:        "success/non-command-argument-noop",
			description: "验证 root 命令收到非子命令参数时按兼容契约返回成功且不运行示例加载。",
			giveArgs:    []string{"plain-argument"},
			assert: func(t *testing.T, output string, err error) {
				require.NoError(t, err)
				assert.Empty(t, output)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			var gotErr error
			gotOutput := captureStdoutForCommand(t, func() {
				gotErr = rootCmd.RunE(rootCmd, tt.giveArgs)
			})

			require.NotNil(t, tt.assert)
			tt.assert(t, gotOutput, gotErr)
		})
	}
}

// TestExecute_ErrorExitsProcess 验证 Execute 在 Cobra 执行失败时退出进程。
//
// 该测试通过子进程触发未知命令错误，避免 os.Exit 终止当前测试进程，同时覆盖 Execute 的错误处理契约。
//
// 参数：
//   - t: 测试上下文，用于启动子进程和报告断言失败。
func TestExecute_ErrorExitsProcess(t *testing.T) {
	if os.Getenv("KIT_TEST_KRATOS_CONFIG_EXECUTE_ERROR") == "1" {
		resetKratosConfigCommandState(t)
		rootCmd.SetArgs([]string{"des"})
		rootCmd.SetOut(io.Discard)
		rootCmd.SetErr(io.Discard)
		Execute()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExecute_ErrorExitsProcess")
	cmd.Env = append(os.Environ(), "KIT_TEST_KRATOS_CONFIG_EXECUTE_ERROR=1")
	output, err := cmd.CombinedOutput()

	require.Error(t, err)
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 1, exitErr.ExitCode())
	assert.Contains(t, string(output), "required flag(s) \"data\" not set")
}

// TestRootCommand_LoadsExampleConfig 验证根命令能加载示例配置并输出解码后的字段。
//
// 该测试在示例配置目录和仓库根目录两种稳定 cwd 下执行根命令，使用临时 cwd 和环境变量隔离全局状态，确保 DES、base64 和 env 配置解析链路可回归。
//
// 参数：
//   - t: 测试上下文，用于运行子测试、设置临时状态、捕获输出和报告断言失败。
func TestRootCommand_LoadsExampleConfig(t *testing.T) {
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	configDir := filepath.Clean(filepath.Join(oldWd, ".."))
	repoRoot := filepath.Clean(filepath.Join(oldWd, "../../../.."))

	tests := []struct {
		name        string
		description string
		giveWd      string
		givePWD     string
	}{
		{
			name:        "success/config-directory",
			description: "验证在 example/kratos/config 目录下运行时会加载当前目录的 config.yaml。",
			giveWd:      configDir,
			givePWD:     configDir,
		},
		{
			name:        "success/repository-root-directory",
			description: "验证在仓库根目录运行时会加载 example/kratos/config/config.yaml。",
			giveWd:      repoRoot,
			givePWD:     repoRoot,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			require.NoError(t, os.Chdir(tt.giveWd))
			t.Cleanup(func() {
				require.NoError(t, os.Chdir(oldWd))
			})
			t.Setenv("PWD", tt.givePWD)
			t.Setenv("LANG", "zh_CN.UTF-8")

			gotOutput, gotErr := executeKratosConfigCommand(t)

			require.NoError(t, gotErr)
			assert.Contains(t, gotOutput, "App Name: example-kratos-config")
			assert.Contains(t, gotOutput, "App Password: 中文配置示例")
			assert.Contains(t, gotOutput, "App Addr: 我是 base64 过的 test")
			assert.Contains(t, gotOutput, "Env Lang: zh_CN.UTF-8")
		})
	}
}

// mustEncryptDESForCommand 使用指定密钥生成 DES 命令测试密文。
//
// 该辅助函数集中构造命令输入密文，保证加密失败时立即报告夹具错误。
//
// 参数：
//   - t: 测试上下文，用于报告夹具构造失败并标记辅助函数调用栈。
//   - giveKey: DES 字符串密钥。
//   - giveData: 需要加密的明文。
//
// 返回：
//   - string: 可传给 des 命令的 16 进制密文。
func mustEncryptDESForCommand(t *testing.T, giveKey, giveData string) string {
	t.Helper()

	got, err := kitcryptodes.EncryptStringCBCPkCS7PaddingStringHex(giveKey, giveData)
	require.NoError(t, err)
	return got
}

// executeKratosConfigCommand 执行配置示例根命令并捕获标准输出。
//
// 该辅助函数重置 Cobra 参数和 DES 包级标志状态，串行捕获 fmt 直接写入 stdout 的命令输出。
//
// 参数：
//   - t: 测试上下文，用于注册清理逻辑和报告执行错误。
//   - giveArgs: 传递给 rootCmd 的命令行参数。
//
// 返回：
//   - string: 命令执行期间写入标准输出的完整内容。
//   - error: Cobra Execute 返回的错误。
func executeKratosConfigCommand(t *testing.T, giveArgs ...string) (string, error) {
	t.Helper()

	resetKratosConfigCommandState(t)
	if giveArgs == nil {
		giveArgs = []string{}
	}
	rootCmd.SetArgs(giveArgs)
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})

	var gotErr error
	gotOutput := captureStdoutForCommand(t, func() {
		gotErr = rootCmd.Execute()
	})
	return gotOutput, gotErr
}

// resetKratosConfigCommandState 恢复配置示例命令的包级可变状态。
//
// 该辅助函数在每次执行命令前清理全局 flag 绑定变量和 Cobra 标志解析状态，避免子测试之间相互污染。
//
// 参数：
//   - t: 测试上下文，用于报告状态恢复错误并标记辅助函数调用栈。
func resetKratosConfigCommandState(t *testing.T) {
	t.Helper()

	key = ""
	data = ""
	encrypt = true
	resetCommandFlags(t, rootCmd)
}

// resetCommandFlags 递归重置 Cobra 命令树中的 flag changed 状态。
//
// 该辅助函数保持已注册标志和值绑定不变，仅清理上一次 Execute 留下的解析痕迹，保证后续用例可重复解析参数。
//
// 参数：
//   - t: 测试上下文，用于报告 flag 状态重置错误并标记辅助函数调用栈。
//   - command: 需要重置的 Cobra 命令节点。
func resetCommandFlags(t *testing.T, command *cobra.Command) {
	t.Helper()

	command.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Changed = false
	})
	command.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		flag.Changed = false
	})
	command.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
		flag.Changed = false
	})
	for _, child := range command.Commands() {
		resetCommandFlags(t, child)
	}
}

// captureStdoutForCommand 捕获命令执行期间写入标准输出的内容。
//
// 该辅助函数替换 os.Stdout 并在测试结束时恢复，适用于命令实现中直接使用 fmt.Print 的输出断言。
//
// 参数：
//   - t: 测试上下文，用于报告管道读写错误并标记辅助函数调用栈。
//   - giveFunc: 在 stdout 被捕获期间执行的函数。
//
// 返回：
//   - string: 捕获到的标准输出内容。
func captureStdoutForCommand(t *testing.T, giveFunc func()) string {
	t.Helper()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = writer
	t.Cleanup(func() {
		os.Stdout = originalStdout
		_ = writer.Close()
		_ = reader.Close()
	})

	var output bytes.Buffer
	readDone := make(chan error, 1)
	go func() {
		_, copyErr := io.Copy(&output, reader)
		readDone <- copyErr
	}()

	giveFunc()

	require.NoError(t, writer.Close())
	os.Stdout = originalStdout
	require.NoError(t, <-readDone)
	require.NoError(t, reader.Close())

	return strings.ReplaceAll(output.String(), "\r\n", "\n")
}
