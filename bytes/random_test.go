// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package bytes 单元测试覆盖 GenerateNonce 的长度校验、随机源读取和错误传播行为。
package bytes

import (
	stdbytes "bytes"
	"crypto/rand"
	"errors"
	"io"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateNonce_LengthValidation 验证 GenerateNonce 对长度参数的边界处理行为。
//
// 该测试通过表驱动用例覆盖零长度和负数长度语义，确保函数对边界输入返回稳定结果。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGenerateNonce_LengthValidation(t *testing.T) {
	tests := []struct {
		name            string
		description     string
		giveLength      int
		wantLen         int
		wantNilNonce    bool
		wantErr         bool
		wantErrContains string
	}{
		{
			name:        "boundary/zero-length",
			description: "验证 GenerateNonce 在长度为 0 时返回可用的空切片且不产生错误。",
			giveLength:  0,
			wantLen:     0,
		},
		{
			name:            "error/negative-length",
			description:     "验证 GenerateNonce 拒绝负数长度并返回说明非法参数的错误。",
			giveLength:      -1,
			wantNilNonce:    true,
			wantErr:         true,
			wantErrContains: "长度不能为负数：-1",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := GenerateNonce(tt.giveLength)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrContains)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
			if tt.wantNilNonce {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
			}
		})
	}
}

// TestGenerateNonce_ReadsRequestedBytes 验证 GenerateNonce 按请求长度从随机源填充字节。
//
// 该测试使用确定性随机源覆盖完整读取和截取读取场景，避免依赖真实随机值的概率性断言。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGenerateNonce_ReadsRequestedBytes(t *testing.T) {
	tests := []struct {
		name           string
		description    string
		giveLength     int
		giveRandomData []byte
		wantNonce      []byte
	}{
		{
			name:           "success/exact-length",
			description:    "验证 GenerateNonce 在随机源恰好提供请求字节数时返回完整字节序列。",
			giveLength:     4,
			giveRandomData: []byte{0xde, 0xad, 0xbe, 0xef},
			wantNonce:      []byte{0xde, 0xad, 0xbe, 0xef},
		},
		{
			name:           "success/truncate-extra-source-bytes",
			description:    "验证 GenerateNonce 只读取请求长度的数据，不把随机源中的额外字节写入结果。",
			giveLength:     3,
			giveRandomData: []byte{0x01, 0x02, 0x03, 0x04},
			wantNonce:      []byte{0x01, 0x02, 0x03},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			withRandReader(t, stdbytes.NewReader(tt.giveRandomData))

			got, err := GenerateNonce(tt.giveLength)

			require.NoError(t, err)
			assert.Equal(t, tt.wantNonce, got)
		})
	}
}

// TestGenerateNonce_RandomReaderErrors 验证 GenerateNonce 对随机源读取失败的错误传播行为。
//
// 该测试通过表驱动用例覆盖随机源立即失败和短读失败场景，确保调用方能够感知底层读取错误。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestGenerateNonce_RandomReaderErrors(t *testing.T) {
	giveReadErr := errors.New("random reader failed")
	tests := []struct {
		name        string
		description string
		giveLength  int
		setup       func(t *testing.T) io.Reader
		wantNonce   []byte
		wantErrIs   error
	}{
		{
			name:        "error/reader-failure",
			description: "验证 GenerateNonce 在随机源立即返回错误时保留目标长度并透传底层错误。",
			giveLength:  4,
			setup: func(t *testing.T) io.Reader {
				t.Helper()
				return iotest.ErrReader(giveReadErr)
			},
			wantNonce: []byte{0x00, 0x00, 0x00, 0x00},
			wantErrIs: giveReadErr,
		},
		{
			name:        "error/short-reader",
			description: "验证 GenerateNonce 在随机源短读时返回 io.ErrUnexpectedEOF 并保留已读取字节。",
			giveLength:  4,
			setup: func(t *testing.T) io.Reader {
				t.Helper()
				return stdbytes.NewReader([]byte{0xca, 0xfe})
			},
			wantNonce: []byte{0xca, 0xfe, 0x00, 0x00},
			wantErrIs: io.ErrUnexpectedEOF,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			withRandReader(t, tt.setup(t))

			got, err := GenerateNonce(tt.giveLength)

			require.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErrIs)
			assert.Equal(t, tt.wantNonce, got)
		})
	}
}

// BenchmarkGenerateNonce 度量 GenerateNonce 在常见 nonce 长度下的生成开销。
//
// 该基准覆盖 16、32 和 64 字节长度，帮助观察真实随机源在不同输出规模下的性能表现。
//
// 参数：
//   - b: 基准测试上下文，用于运行子基准并报告性能结果。
func BenchmarkGenerateNonce(b *testing.B) {
	benchmarks := []struct {
		name        string
		description string
		giveLength  int
	}{
		{
			name:        "16-bytes",
			description: "度量 GenerateNonce 生成 16 字节 nonce 的基准性能。",
			giveLength:  16,
		},
		{
			name:        "32-bytes",
			description: "度量 GenerateNonce 生成 32 字节 nonce 的基准性能。",
			giveLength:  32,
		},
		{
			name:        "64-bytes",
			description: "度量 GenerateNonce 生成 64 字节 nonce 的基准性能。",
			giveLength:  64,
		},
	}

	for _, bm := range benchmarks {
		bm := bm
		b.Run(bm.name, func(b *testing.B) {
			b.Log(bm.description)

			for i := 0; i < b.N; i++ {
				_, err := GenerateNonce(bm.giveLength)
				if err != nil {
					b.Fatalf("GenerateNonce(%d) returned error: %v", bm.giveLength, err)
				}
			}
		})
	}
}

// withRandReader 临时替换 crypto/rand.Reader 以便测试 GenerateNonce 的确定性读取行为。
//
// 该辅助函数在当前测试结束时恢复原始随机源，避免全局状态泄漏到其他用例。
//
// 参数：
//   - t: 测试上下文，用于注册清理函数并标记辅助函数调用栈。
//   - giveReader: 当前用例希望 GenerateNonce 使用的随机源。
func withRandReader(t *testing.T, giveReader io.Reader) {
	t.Helper()

	originalReader := rand.Reader
	rand.Reader = giveReader
	t.Cleanup(func() {
		rand.Reader = originalReader
	})
}
