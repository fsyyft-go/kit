// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.
//
// 本测试文件针对 Snowflake 算法实现的所有核心方法进行单元测试。
// 设计思路：
//  1. 采用表格驱动法，便于批量、系统性测试各类输入输出。
//  2. 断言全部采用 stretchr/testify 包，提升可读性和一致性。
//  3. 保留原有测试用例的核心逻辑，补充边界和异常场景，提升覆盖率。
//  4. 注释详细，便于理解每个测试的目的和断言依据。
//  5. 所有测试均可直接 go test 运行，无需额外依赖。
//
// 使用方法：
//   - 直接运行 go test -v ./algorithms/snowflake
//   - 推荐配合 -cover 检查覆盖率
package snowflake

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewNode_ValidAndInvalid 测试 NewNode 的正常与异常分支。
func TestNewNode_ValidAndInvalid(t *testing.T) {
	cases := []struct {
		name   string
		nodeid int64
		wantOk bool
	}{
		{"valid nodeid 0", 0, true},
		{"valid nodeid max", int64(-1 ^ (-1 << NodeBits)), true},
		{"invalid nodeid < 0", -1, false},
		{"invalid nodeid > max", int64(-1^(-1<<NodeBits)) + 1, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			n, err := NewNode(c.nodeid)
			if c.wantOk {
				assert.NoError(t, err)
				assert.NotNil(t, n)
			} else {
				assert.Error(t, err)
				assert.Nil(t, n)
			}
		})
	}
}

// TestNode_Generate_Unique 测试同一节点生成的 ID 唯一性。
func TestNode_Generate_Unique(t *testing.T) {
	// 创建节点
	n, err := NewNode(1)
	assert.NoError(t, err)
	assert.NotNil(t, n)

	idSet := make(map[ID]struct{})
	var last ID
	for i := 0; i < 10000; i++ {
		id := n.Generate()
		_, exists := idSet[id]
		assert.False(t, exists, "ID 重复: %v", id)
		idSet[id] = struct{}{}
		if i > 0 {
			assert.NotEqual(t, last, id)
		}
		last = id
	}
}

// TestNode_Generate_Concurrency 并发生成 ID，验证无重复。
func TestNode_Generate_Concurrency(t *testing.T) {
	n, err := NewNode(2)
	assert.NoError(t, err)
	assert.NotNil(t, n)

	idCh := make(chan ID, 1000)
	count := 1000

	// 并发生成
	for i := 0; i < count; i++ {
		go func() {
			idCh <- n.Generate()
		}()
	}
	idSet := make(map[ID]struct{})
	for i := 0; i < count; i++ {
		id := <-idCh
		_, exists := idSet[id]
		assert.False(t, exists, "并发 ID 重复: %v", id)
		idSet[id] = struct{}{}
	}
}

// TestID_Converters_And_Parsers 测试所有编码/解码方法的互逆性。
func TestID_Converters_And_Parsers(t *testing.T) {
	n, _ := NewNode(0)
	id := n.Generate()

	// 十进制字符串
	s := id.String()
	id2, err := ParseString(s)
	assert.NoError(t, err)
	assert.Equal(t, id, id2)

	// int64
	i := id.Int64()
	id3 := ParseInt64(i)
	assert.Equal(t, id, id3)

	// base2
	b2 := id.Base2()
	id4, err := ParseBase2(b2)
	assert.NoError(t, err)
	assert.Equal(t, id, id4)

	// base32
	b32 := id.Base32()
	id5, err := ParseBase32([]byte(b32))
	assert.NoError(t, err)
	assert.Equal(t, id, id5)

	// base36
	b36 := id.Base36()
	id6, err := ParseBase36(b36)
	assert.NoError(t, err)
	assert.Equal(t, id, id6)

	// base58
	b58 := id.Base58()
	id7, err := ParseBase58([]byte(b58))
	assert.NoError(t, err)
	assert.Equal(t, id, id7)

	// base64
	b64 := id.Base64()
	id8, err := ParseBase64(b64)
	assert.NoError(t, err)
	assert.Equal(t, id, id8)

	// Bytes
	by := id.Bytes()
	id9, err := ParseBytes(by)
	assert.NoError(t, err)
	assert.Equal(t, id, id9)

	// IntBytes
	iby := id.IntBytes()
	id10 := ParseIntBytes(iby)
	assert.Equal(t, id, id10)
}

// TestID_MarshalUnmarshalJSON 测试 JSON 序列化和反序列化。
func TestID_MarshalUnmarshalJSON(t *testing.T) {
	cases := []struct {
		name      string
		jsonInput string
		wantID    ID
		wantErr   bool
	}{
		{"valid", `"13587"`, 13587, false},
		{"invalid: not quoted", `1`, 0, true},
		{"invalid: not closed", `"invalid`, 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var id ID
			err := id.UnmarshalJSON([]byte(c.jsonInput))
			if c.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.wantID, id)
			}
		})
	}

	// MarshalJSON
	id := ID(13587)
	b, err := id.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, []byte(`"13587"`), b)
}

// TestParseBase32_And_ParseBase58_ValidAndInvalid 测试 base32/base58 的合法与非法输入。
func TestParseBase32_And_ParseBase58_ValidAndInvalid(t *testing.T) {
	b32Cases := []struct {
		name    string
		input   string
		want    ID
		wantErr bool
	}{
		{"ok", "b8wjm1zroyyyy", 1427970479175499776, false},
		{"capital case invalid", "B8WJM1ZROYYYY", -1, true},
		{"l not allowed", "b8wjm1zroyyyl", -1, true},
		{"v not allowed", "b8wjm1zroyyyv", -1, true},
		{"2 not allowed", "b8wjm1zroyyy2", -1, true},
	}
	for _, c := range b32Cases {
		t.Run("base32-"+c.name, func(t *testing.T) {
			got, err := ParseBase32([]byte(c.input))
			if c.wantErr {
				assert.Error(t, err)
				assert.Equal(t, c.want, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want, got)
			}
		})
	}

	b58Cases := []struct {
		name    string
		input   string
		want    ID
		wantErr bool
	}{
		{"ok", "4jgmnx8Js8A", 1428076403798048768, false},
		{"0 not allowed", "0jgmnx8Js8A", -1, true},
		{"I not allowed", "Ijgmnx8Js8A", -1, true},
		{"O not allowed", "Ojgmnx8Js8A", -1, true},
		{"l not allowed", "ljgmnx8Js8A", -1, true},
	}
	for _, c := range b58Cases {
		t.Run("base58-"+c.name, func(t *testing.T) {
			got, err := ParseBase58([]byte(c.input))
			if c.wantErr {
				assert.Error(t, err)
				assert.Equal(t, c.want, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want, got)
			}
		})
	}
}

// TestID_IntBytes_And_ParseIntBytes 测试 IntBytes 与 ParseIntBytes 的互逆。
func TestID_IntBytes_And_ParseIntBytes(t *testing.T) {
	id := ID(13587)
	iby := id.IntBytes()
	expected := [8]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x35, 0x13}
	assert.Equal(t, expected, iby)
	id2 := ParseIntBytes(iby)
	assert.Equal(t, id, id2)
}

// TestID_Bytes_And_ParseBytes 测试 Bytes 与 ParseBytes 的互逆。
func TestID_Bytes_And_ParseBytes(t *testing.T) {
	id := ID(1234567890)
	by := id.Bytes()
	id2, err := ParseBytes(by)
	assert.NoError(t, err)
	assert.Equal(t, id, id2)

	// 非法输入
	_, err = ParseBytes([]byte{0xFF, 0xFF, 0xFF})
	assert.Error(t, err)
}

// TestID_Base64_And_ParseBase64 测试 Base64 与 ParseBase64 的互逆。
func TestID_Base64_And_ParseBase64(t *testing.T) {
	id := ID(1234567890)
	b64 := id.Base64()
	id2, err := ParseBase64(b64)
	assert.NoError(t, err)
	assert.Equal(t, id, id2)

	// 非法输入
	_, err = ParseBase64("not_base64!!")
	assert.Error(t, err)
}

// TestID_MarshalIntBytes 测试 IntBytes 的二进制内容。
func TestID_MarshalIntBytes(t *testing.T) {
	id := ID(13587)
	expected := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x35, 0x13}
	iby := id.IntBytes()
	assert.True(t, bytes.Equal(iby[:], expected))
}
