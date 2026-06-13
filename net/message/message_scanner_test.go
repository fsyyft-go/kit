// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import (
	"bytes"
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScanMessage 验证协议分割函数在完整帧、不完整帧和 EOF 场景下的行为。
//
// 该测试通过表驱动用例覆盖头部不足、完整空 payload、普通 payload、最大 payload 和 EOF，确保 Scanner 分割契约稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestScanMessage(t *testing.T) {
	maxPayload := bytes.Repeat([]byte{'x'}, math.MaxUint16)
	tests := []struct {
		name        string
		description string
		giveData    []byte
		giveAtEOF   bool
		wantAdvance int
		wantToken   []byte
	}{
		{
			name:        "boundary/short-header",
			description: "验证不足 4 字节头部时分割函数等待更多数据且不消费缓冲区。",
			giveData:    []byte{0x01, 0x02, 0x03},
			wantAdvance: 0,
			wantToken:   nil,
		},
		{
			name:        "boundary/header-only-empty-payload",
			description: "验证 payload 长度为 0 的头部本身就是完整合法消息包。",
			giveData:    buildTestPacket(t, SingleStringMessageType, []byte{}),
			wantAdvance: messageHeaderLength,
			wantToken:   buildTestPacket(t, SingleStringMessageType, []byte{}),
		},
		{
			name:        "boundary/incomplete-payload",
			description: "验证头部声明的 payload 尚未完整到达时分割函数等待更多数据。",
			giveData:    buildTestPacket(t, SingleStringMessageType, []byte("abcde"))[:5],
			wantAdvance: 0,
			wantToken:   nil,
		},
		{
			name:        "success/complete-packet",
			description: "验证完整消息包会被一次性作为 token 返回并消费对应字节数。",
			giveData:    buildTestPacket(t, SingleStringMessageType, []byte("hello")),
			wantAdvance: messageHeaderLength + len("hello"),
			wantToken:   buildTestPacket(t, SingleStringMessageType, []byte("hello")),
		},
		{
			name:        "success/complete-packet-with-trailing-data",
			description: "验证缓冲区包含后续数据时只返回第一条完整消息包。",
			giveData:    append(buildTestPacket(t, SingleStringMessageType, []byte("hello")), []byte{0x01, 0x02}...),
			wantAdvance: messageHeaderLength + len("hello"),
			wantToken:   buildTestPacket(t, SingleStringMessageType, []byte("hello")),
		},
		{
			name:        "boundary/max-payload-length",
			description: "验证 uint16 最大 payload 长度的完整包可以被分割函数识别。",
			giveData:    buildTestPacket(t, SingleStringMessageType, maxPayload),
			wantAdvance: messageHeaderLength + len(maxPayload),
			wantToken:   buildTestPacket(t, SingleStringMessageType, maxPayload),
		},
		{
			name:        "boundary/at-eof-with-complete-packet",
			description: "验证 EOF 状态下如果缓冲区已有完整消息包，分割函数仍优先返回该 token。",
			giveData:    buildTestPacket(t, SingleStringMessageType, []byte("abc")),
			giveAtEOF:   true,
			wantAdvance: messageHeaderLength + len("abc"),
			wantToken:   buildTestPacket(t, SingleStringMessageType, []byte("abc")),
		},
		{
			name:        "boundary/at-eof-without-complete-packet",
			description: "验证 EOF 状态下如果没有完整消息包，分割函数不返回 token 并结束扫描。",
			giveData:    []byte{0x01, 0x02, 0x03},
			giveAtEOF:   true,
			wantAdvance: 0,
			wantToken:   nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			advance, token, err := scanMessage(tt.giveData, tt.giveAtEOF)

			require.NoError(t, err)
			assert.Equal(t, tt.wantAdvance, advance)
			assert.Equal(t, tt.wantToken, token)
		})
	}
}

// TestNewScanner 验证消息 Scanner 可以从连续字节流中解析完整协议包。
//
// 该测试通过表驱动用例覆盖连续多包、空 payload 包和最大 payload 包，确保 Scanner 配置与协议长度边界一致。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewScanner(t *testing.T) {
	maxPayload := bytes.Repeat([]byte{'m'}, math.MaxUint16)
	tests := []struct {
		name        string
		description string
		givePackets [][]byte
		wantTokens  [][]byte
	}{
		{
			name:        "success/multiple-packets",
			description: "验证 Scanner 可以从连续字节流中依次解析多个完整消息包。",
			givePackets: [][]byte{
				buildTestPacket(t, HeartbeatMessageType, buildHeartbeatPayload(1)),
				buildTestPacket(t, SingleStringMessageType, []byte("barbaz")),
			},
		},
		{
			name:        "boundary/empty-payload-packet",
			description: "验证 Scanner 可以解析 payload 长度为 0 的合法消息包。",
			givePackets: [][]byte{
				buildTestPacket(t, SingleStringMessageType, []byte{}),
			},
		},
		{
			name:        "boundary/max-payload-packet",
			description: "验证 Scanner 的缓冲区上限允许协议定义的最大 payload 消息包。",
			givePackets: [][]byte{
				buildTestPacket(t, SingleStringMessageType, maxPayload),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			data := bytes.Join(tt.givePackets, nil)
			scanner := NewScanner(bytes.NewReader(data))
			var tokens [][]byte

			for scanner.Scan() {
				tokens = append(tokens, append([]byte(nil), scanner.Bytes()...))
			}

			wantTokens := tt.wantTokens
			if nil == wantTokens {
				wantTokens = tt.givePackets
			}
			require.NoError(t, scanner.Err())
			assert.Equal(t, wantTokens, tokens)
		})
	}
}

// TestNewScanner_IncompletePacket 验证 Scanner 对不完整协议包的结束行为。
//
// 该测试覆盖输入流结束时仍未收到完整 payload 的场景，确保不会返回半包 token 或错误。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestNewScanner_IncompletePacket(t *testing.T) {
	// 验证半包不会被 Scanner 作为有效 token 返回，避免上层生成错误消息。
	packet := buildTestPacket(t, SingleStringMessageType, []byte("abcdef"))
	scanner := NewScanner(bytes.NewReader(packet[:len(packet)-2]))

	assert.False(t, scanner.Scan())
	assert.NoError(t, scanner.Err())
}

// TestNewScanner_ReadCompletePacketWithEOF 验证 Reader 同时返回完整包和 io.EOF 时 Scanner 仍产出 token。
//
// 该测试覆盖 bufio.Scanner 在读取到最后一批数据并收到 EOF 后调用 split 的路径，确保最后一个完整消息包不会被丢弃。
//
// 参数：
//   - t: 测试上下文，用于报告断言失败。
func TestNewScanner_ReadCompletePacketWithEOF(t *testing.T) {
	// 验证底层 Reader 在同一次 Read 中返回 n>0 和 io.EOF 时，Scanner 仍能产出完整 token。
	packet := buildTestPacket(t, SingleStringMessageType, []byte{})
	scanner := NewScanner(&eofWithDataReader{data: packet})

	require.True(t, scanner.Scan())
	assert.Equal(t, packet, scanner.Bytes())
	assert.False(t, scanner.Scan())
	assert.NoError(t, scanner.Err())
}

type eofWithDataReader struct {
	data []byte
	done bool
}

// Read 在首次读取时同时返回全部数据和 io.EOF。
//
// 该辅助方法模拟标准库允许的 Reader 行为，用于验证 Scanner split 函数不会在 atEOF=true 时丢弃完整 token。
//
// 参数：
//   - p: 调用方提供的读取缓冲区。
//
// 返回：
//   - int: 首次读取复制的数据长度。
//   - error: 首次读取返回 io.EOF，后续读取也返回 io.EOF。
func (r *eofWithDataReader) Read(p []byte) (int, error) {
	if r.done {
		return 0, io.EOF
	}

	copy(p, r.data)
	r.done = true
	return len(r.data), io.EOF
}
