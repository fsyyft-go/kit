// Copyright 2024 fsyyft-go Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package log 提供了日志功能的测试用例。
package log

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoggers 测试所有支持的日志实现。
// 测试内容包括：
// - 控制台日志输出
// - 标准库文件日志
// - Logrus 文件日志
// - 结构化字段支持
func TestLoggers(t *testing.T) {
	// 创建临时测试目录用于存放日志文件。
	// 使用系统临时目录确保在不同环境下都能正常工作。
	tmpDir := filepath.Join(os.TempDir(), "apisix-metric-test")
	err := os.MkdirAll(tmpDir, 0755)
	assert.NoError(t, err)
	// 测试完成后清理临时目录。
	defer os.RemoveAll(tmpDir)

	// 定义测试用例，包含不同类型的日志器测试。
	testCases := []struct {
		name     string                            // 测试用例名称
		logType  LogType                           // 日志类型
		logPath  string                            // 日志文件路径
		testFunc func(t *testing.T, logger Logger) // 测试函数
	}{
		{
			name:    "Console Logger",
			logType: LogTypeConsole,
			logPath: "",
			testFunc: func(t *testing.T, logger Logger) {
				logger.Info("测试控制台日志。")
				logger.WithField("test", "field").Info("测试带字段的控制台日志。")
			},
		},
		{
			name:    "Std Logger File",
			logType: LogTypeStd,
			logPath: filepath.Join(tmpDir, "std.log"),
			testFunc: func(t *testing.T, logger Logger) {
				logger.Info("测试标准库日志文件。")
				logger.WithFields(map[string]interface{}{
					"test1": "value1",
					"test2": "value2",
				}).Info("测试带多个字段的标准库日志。")
			},
		},
		{
			name:    "Logrus Logger File",
			logType: LogTypeLogrus,
			logPath: filepath.Join(tmpDir, "logrus.log"),
			testFunc: func(t *testing.T, logger Logger) {
				logger.Debug("测试 Logrus 调试日志。")
				logger.Info("测试 Logrus 信息日志。")
				logger.Warn("测试 Logrus 警告日志。")
				logger.WithField("component", "test").Error("测试 Logrus 错误日志。")
			},
		},
	}

	// 执行所有测试用例。
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 初始化日志实例。
			err := InitLogger(tc.logType, tc.logPath)
			assert.NoError(t, err)

			// 执行测试函数。
			tc.testFunc(t, GetLogger())

			// 如果配置了日志文件，验证文件是否正确创建和写入。
			if tc.logPath != "" {
				_, err := os.Stat(tc.logPath)
				assert.NoError(t, err)

				// 读取并验证日志文件内容。
				content, err := os.ReadFile(tc.logPath)
				assert.NoError(t, err)
				assert.NotEmpty(t, content)
			}
		})
	}
}

// TestLogLevels 测试日志的各个级别。
// 测试内容包括：
// - Debug 级别日志
// - Info 级别日志
// - Warn 级别日志
// - Error 级别日志
// - 格式化和非格式化日志
func TestLogLevels(t *testing.T) {
	// 创建临时测试目录。
	tmpDir := filepath.Join(os.TempDir(), "apisix-metric-test-levels")
	err := os.MkdirAll(tmpDir, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// 设置日志文件路径。
	logPath := filepath.Join(tmpDir, "all-levels.log")

	// 初始化 Logrus 日志器。
	err = InitLogger(LogTypeLogrus, logPath)
	assert.NoError(t, err)

	logger := GetLogger()

	// 测试所有日志级别的记录功能。
	logger.Debug("调试信息。")
	logger.Debugf("带格式的调试信息：%s。", "测试")

	logger.Info("普通信息。")
	logger.Infof("带格式的普通信息：%s。", "测试")

	logger.Warn("警告信息。")
	logger.Warnf("带格式的警告信息：%s。", "测试")

	logger.Error("错误信息。")
	logger.Errorf("带格式的错误信息：%s。", "测试")

	// 验证日志文件内容。
	content, err := os.ReadFile(logPath)
	assert.NoError(t, err)
	assert.NotEmpty(t, content)
}

// TestWithFieldsAndFormat 测试结构化字段和格式化功能。
// 测试内容包括：
// - 单个字段添加
// - 多个字段添加
// - 链式字段添加
// - 不同类型的字段值
func TestWithFieldsAndFormat(t *testing.T) {
	// 创建临时测试目录。
	tmpDir := filepath.Join(os.TempDir(), "apisix-metric-test-fields")
	err := os.MkdirAll(tmpDir, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// 设置日志文件路径。
	logPath := filepath.Join(tmpDir, "fields.log")

	// 初始化 Logrus 日志器。
	err = InitLogger(LogTypeLogrus, logPath)
	assert.NoError(t, err)

	logger := GetLogger()

	// 测试不同的字段添加方式。
	logger.WithField("single", "field").Info("单字段测试。")

	logger.WithFields(map[string]interface{}{
		"field1": "value1",
		"field2": 123,
		"field3": true,
	}).Info("多字段测试。")

	logger.WithField("request_id", "123").
		WithField("user_id", "456").
		Info("链式字段测试。")

	// 验证日志文件内容。
	content, err := os.ReadFile(logPath)
	assert.NoError(t, err)
	assert.NotEmpty(t, content)
}
