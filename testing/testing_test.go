// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package testing

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	stdtesting "testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// wantLogHeader 定义测试期望的公开日志前缀，避免复用被测实现内部常量。
	wantLogHeader = "=-=       "
)

var (
	// stdoutCaptureMu 串行化标准输出捕获，避免多个用例同时替换 os.Stdout。
	stdoutCaptureMu sync.Mutex
)

// TestPrintln_Output 验证 Println 在不同参数组合下输出统一前缀并保持 fmt.Println 兼容语义。
//
// 该测试通过表驱动用例覆盖无参数、空字符串、nil 和多参数输出，确保日志前缀与换行契约稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestPrintln_Output(t *stdtesting.T) {
	tests := []struct {
		name        string
		description string
		giveArgs    []interface{}
		wantOutput  string
	}{
		{
			name:        "boundary/no-arguments",
			description: "验证 Println 在无参数时仍输出统一前缀并追加换行。",
			wantOutput:  wantLogHeader + "\n",
		},
		{
			name:        "boundary/empty-string",
			description: "验证 Println 在空字符串参数下保留统一前缀和 fmt.Println 的换行语义。",
			giveArgs:    []interface{}{""},
			wantOutput:  wantLogHeader + "\n",
		},
		{
			name:        "boundary/nil-value",
			description: "验证 Println 在 nil 参数下使用 fmt.Println 的 nil 文本表示。",
			giveArgs:    []interface{}{nil},
			wantOutput:  wantLogHeader + "<nil>\n",
		},
		{
			name:        "success/multiple-values",
			description: "验证 Println 在多参数下使用空格分隔内容并追加换行。",
			giveArgs:    []interface{}{"value:", 42, true},
			wantOutput:  wantLogHeader + "value: 42 true\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *stdtesting.T) {
			t.Log(tt.description)

			gotOutput := captureStdout(t, func() {
				Println(tt.giveArgs...)
			})

			assert.Equal(t, tt.wantOutput, gotOutput)
		})
	}
}

// TestPrintf_Output 验证 Printf 在不同格式化输入下输出统一前缀并保持 fmt.Printf 兼容语义。
//
// 该测试通过表驱动用例覆盖空格式串、普通文本、格式化参数、nil 和缺失参数场景，确保格式化输出契约稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestPrintf_Output(t *stdtesting.T) {
	tests := []struct {
		name        string
		description string
		giveFormat  string
		giveArgs    []interface{}
		wantOutput  string
	}{
		{
			name:        "boundary/empty-format",
			description: "验证 Printf 在空格式串下仍输出统一前缀且不追加换行。",
			wantOutput:  wantLogHeader,
		},
		{
			name:        "success/plain-text",
			description: "验证 Printf 在普通文本格式串下输出统一前缀和原始文本。",
			giveFormat:  "plain message",
			wantOutput:  wantLogHeader + "plain message",
		},
		{
			name:        "success/formatted-values",
			description: "验证 Printf 在格式化参数下遵循 fmt.Printf 的格式化语义。",
			giveFormat:  "progress=%d%% status=%s",
			giveArgs:    []interface{}{75, "ok"},
			wantOutput:  wantLogHeader + "progress=75% status=ok",
		},
		{
			name:        "boundary/nil-value",
			description: "验证 Printf 在 nil 参数下使用 fmt.Printf 的 nil 文本表示。",
			giveFormat:  "value=%v",
			giveArgs:    []interface{}{nil},
			wantOutput:  wantLogHeader + "value=<nil>",
		},
		{
			name:        "error/missing-format-argument",
			description: "验证 Printf 在格式化参数缺失时保留 fmt.Printf 的诊断输出。",
			giveFormat:  "value=%d",
			wantOutput:  wantLogHeader + "value=%!d(MISSING)",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *stdtesting.T) {
			t.Log(tt.description)

			gotOutput := captureStdout(t, func() {
				Printf(tt.giveFormat, tt.giveArgs...)
			})

			assert.Equal(t, tt.wantOutput, gotOutput)
		})
	}
}

// TestPrintFunctions_ConcurrentWriteFragments 验证输出辅助函数在并发调用下的片段级输出契约。
//
// 该测试只断言每次调用产生的前缀次数和唯一消息片段可被捕获；由于实现分别写入前缀和正文，不断言单条记录不可交错。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestPrintFunctions_ConcurrentWriteFragments(t *stdtesting.T) {
	tests := []struct {
		name        string
		description string
		giveWorkers int
		giveCall    func(index int)
		wantMessage func(index int) string
	}{
		{
			name:        "concurrency/println-fragments",
			description: "验证 Println 在并发调用下可观察到每次调用的前缀和唯一消息片段，但不承诺记录原子性。",
			giveWorkers: 16,
			giveCall: func(index int) {
				Println(fmt.Sprintf("println-worker-%02d", index))
			},
			wantMessage: func(index int) string {
				return fmt.Sprintf("println-worker-%02d", index)
			},
		},
		{
			name:        "concurrency/printf-fragments",
			description: "验证 Printf 在并发调用下可观察到每次调用的前缀和唯一格式化消息片段，但不承诺记录原子性。",
			giveWorkers: 16,
			giveCall: func(index int) {
				Printf("printf-worker-%02d;", index)
			},
			wantMessage: func(index int) string {
				return fmt.Sprintf("printf-worker-%02d;", index)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *stdtesting.T) {
			t.Log(tt.description)

			gotOutput := captureStdout(t, func() {
				var wg sync.WaitGroup
				for index := 0; index < tt.giveWorkers; index++ {
					index := index
					wg.Add(1)
					go func() {
						defer wg.Done()
						tt.giveCall(index)
					}()
				}
				wg.Wait()
			})

			assert.Equal(t, tt.giveWorkers, strings.Count(gotOutput, wantLogHeader))
			for index := 0; index < tt.giveWorkers; index++ {
				assert.Contains(t, gotOutput, tt.wantMessage(index))
			}
		})
	}
}

// captureStdout 捕获被测函数写入标准输出的内容。
//
// 该辅助函数串行化 os.Stdout 替换，并通过清理函数恢复全局状态和关闭管道资源，供验证直接写标准输出的函数使用。
//
// 参数：
//   - t: 测试上下文，用于报告捕获过程中的错误并标记辅助函数调用栈。
//   - giveFunc: 需要在标准输出被捕获期间执行的函数。
//
// 返回：
//   - string: giveFunc 执行期间写入标准输出的完整内容。
func captureStdout(t *stdtesting.T, giveFunc func()) string {
	t.Helper()

	stdoutCaptureMu.Lock()
	defer stdoutCaptureMu.Unlock()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	var (
		readerClosed bool
		writerClosed bool
	)
	closeReader := func() error {
		if readerClosed {
			return nil
		}
		readerClosed = true
		return reader.Close()
	}
	closeWriter := func() error {
		if writerClosed {
			return nil
		}
		writerClosed = true
		return writer.Close()
	}

	os.Stdout = writer
	restoreStdout := func() {
		os.Stdout = originalStdout
	}
	defer restoreStdout()
	t.Cleanup(func() {
		restoreStdout()
		_ = closeWriter()
		_ = closeReader()
	})

	var output bytes.Buffer
	readDone := make(chan error, 1)
	go func() {
		_, copyErr := io.Copy(&output, reader)
		readDone <- copyErr
	}()

	giveFunc()

	require.NoError(t, closeWriter())
	require.NoError(t, <-readDone)
	require.NoError(t, closeReader())

	return output.String()
}
