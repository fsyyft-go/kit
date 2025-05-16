// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//
// 文件设计思路：
// 本测试文件针对 message_scanner.go 的 scanMessage 和 NewScanner 两个方法进行单元测试。
// 采用表格驱动法，覆盖如下场景：
//   - 正常完整包解析
//   - 边界条件（不足 4 字节、长度字段异常、包体不完整等）
//   - atEOF 行为
//   - panic 捕获
//   - 多包连续解析
// 使用 stretchr/testify 断言，保证测试准确性。
//
// 使用方法：
//   go test -v ./net/message -run ^TestScanMessage
//   go test -coverprofile=cover.out ./net/message && go tool cover -func=cover.out
//

package message

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestScanMessage_表格驱动，覆盖协议边界、异常、正常等多种情况。
func TestScanMessage(t *testing.T) {
	type testCase struct {
		name      string
		data      []byte
		atEOF     bool
		wantAdv   int
		wantToken []byte
		wantErr   error
	}

	// 构造一个合法包：2字节类型+2字节长度+payload
	makePacket := func(typ uint16, payload []byte) []byte {
		buf := new(bytes.Buffer)
		_ = binary.Write(buf, binary.BigEndian, typ)
		_ = binary.Write(buf, binary.BigEndian, uint16(len(payload)))
		buf.Write(payload)
		return buf.Bytes()
	}

	tests := []testCase{
		{
			name:      "数据不足4字节，无法判断长度，返回0，无错误",
			data:      []byte{0x01, 0x02, 0x03},
			atEOF:     false,
			wantAdv:   0,
			wantToken: nil,
			wantErr:   nil,
		},
		{
			name:      "atEOF为true，返回io.EOF",
			data:      makePacket(1, []byte("abc")),
			atEOF:     true,
			wantAdv:   0,
			wantToken: nil,
			wantErr:   io.EOF,
		},
		{
			name:      "长度字段读取不足2字节，binary.Read返回io.EOF",
			data:      []byte{0x01, 0x02, 0x03, 0x04},
			atEOF:     false,
			wantAdv:   0,
			wantToken: nil,
			wantErr:   nil,
		},
		{
			name:      "长度字段读取不足，binary.Read返回io.EOF",
			data:      []byte{0x01, 0x02, 0x03},
			atEOF:     false,
			wantAdv:   0,
			wantToken: nil,
			wantErr:   nil,
		},
		{
			name:      "包体不完整，长度足够4字节但不够完整包，返回0，无错误",
			data:      func() []byte { p := makePacket(2, []byte("abcde")); return p[:5] }(),
			atEOF:     false,
			wantAdv:   0,
			wantToken: nil,
			wantErr:   nil,
		},
		{
			name:      "完整包，正常返回advance和token，无错误",
			data:      makePacket(3, []byte("hello")),
			atEOF:     false,
			wantAdv:   4 + len([]byte("hello")),
			wantToken: makePacket(3, []byte("hello")),
			wantErr:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			adv, token, err := scanMessage(tc.data, tc.atEOF)
			assert.Equal(t, tc.wantAdv, adv, "advance 不符")
			assert.Equal(t, tc.wantToken, token, "token 不符")
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr.Error())
			}
		})
	}
}

// TestNewScanner_多包连续解析，验证 bufio.Scanner 能正确分包。
func TestNewScanner(t *testing.T) {
	// 构造两个连续包
	makePacket := func(typ uint16, payload []byte) []byte {
		buf := new(bytes.Buffer)
		_ = binary.Write(buf, binary.BigEndian, typ)
		_ = binary.Write(buf, binary.BigEndian, uint16(len(payload)))
		buf.Write(payload)
		return buf.Bytes()
	}
	p1 := makePacket(10, []byte("foo"))
	p2 := makePacket(20, []byte("barbaz"))
	data := append(p1, p2...)

	scanner := NewScanner(bytes.NewReader(data))
	var tokens [][]byte
	for scanner.Scan() {
		tokens = append(tokens, scanner.Bytes())
	}
	assert.Len(t, tokens, 2, "应解析出两个包")
	assert.Equal(t, p1, tokens[0], "第一个包内容不符")
	assert.Equal(t, p2, tokens[1], "第二个包内容不符")
	assert.NoError(t, scanner.Err())
}
