// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package snowflake

// 本测试文件覆盖 Snowflake 节点生成、编码转换、解析错误和全局配置边界行为。

import (
	"encoding/base64"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type snowflakeGlobalState struct {
	epoch     int64
	nodeBits  uint8
	stepBits  uint8
	nodeMax   int64
	nodeMask  int64
	stepMask  int64
	timeShift uint8
	nodeShift uint8
}

// TestNewNode_ConfigurationAndValidation 验证 NewNode 对节点范围和位宽配置的校验行为。
//
// 该测试通过表驱动用例覆盖最小节点、最大节点、越界节点和位宽总量超限场景，确保节点初始化契约稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNewNode_ConfigurationAndValidation(t *testing.T) {
	const testEpoch = int64(1740515125000)

	tests := []struct {
		name            string
		description     string
		giveNodeBits    uint8
		giveStepBits    uint8
		giveNodeID      int64
		wantNodeMax     int64
		wantErrContains string
	}{
		{
			name:         "success/min-node",
			description:  "验证 NewNode 接受最小节点编号并初始化派生位掩码。",
			giveNodeBits: 10,
			giveStepBits: 12,
			giveNodeID:   0,
			wantNodeMax:  1023,
		},
		{
			name:         "success/max-node",
			description:  "验证 NewNode 接受当前位宽下的最大节点编号。",
			giveNodeBits: 10,
			giveStepBits: 12,
			giveNodeID:   1023,
			wantNodeMax:  1023,
		},
		{
			name:            "error/negative-node",
			description:     "验证 NewNode 拒绝负数节点编号并返回范围错误。",
			giveNodeBits:    10,
			giveStepBits:    12,
			giveNodeID:      -1,
			wantErrContains: "Node number must be between 0 and 1023",
		},
		{
			name:            "error/node-above-max",
			description:     "验证 NewNode 拒绝超过当前节点位宽上限的节点编号。",
			giveNodeBits:    10,
			giveStepBits:    12,
			giveNodeID:      1024,
			wantErrContains: "Node number must be between 0 and 1023",
		},
		{
			name:            "error/bit-allocation-over-limit",
			description:     "验证 NewNode 在节点位和序列位总量超过二十二位时拒绝初始化。",
			giveNodeBits:    11,
			giveStepBits:    12,
			giveNodeID:      0,
			wantErrContains: "remember, you have a total 22 bits to share between Node/Step",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			configureSnowflakeGlobals(t, testEpoch, tt.giveNodeBits, tt.giveStepBits)

			got, err := NewNode(tt.giveNodeID)

			if tt.wantErrContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			actual, ok := got.(*node)
			require.True(t, ok, "NewNode 应返回包内 node 实现")
			assert.Equal(t, tt.giveNodeID, actual.node)
			assert.Equal(t, tt.wantNodeMax, actual.nodeMax)
			assert.Equal(t, int64(-1^(-1<<tt.giveStepBits)), actual.stepMask)
			assert.Equal(t, tt.giveNodeBits+tt.giveStepBits, actual.timeShift)
			assert.Equal(t, tt.giveStepBits, actual.nodeShift)
		})
	}
}

// TestNode_Generate_MonotonicUniqueIDs 验证同一节点连续生成的 ID 保持唯一且单调递增。
//
// 该测试使用默认位宽生成多组 ID，并断言节点编号、序列号范围和唯一性，覆盖常规生成路径。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNode_Generate_MonotonicUniqueIDs(t *testing.T) {
	// 配置稳定的全局位宽，避免其他测试或调用者修改包级变量后影响本用例。
	configureSnowflakeGlobals(t, 1740515125000, 10, 12)
	n := newTestNode(t, 1)

	const generateCount = 10000
	seen := make(map[ID]struct{}, generateCount)
	var last ID

	for i := 0; i < generateCount; i++ {
		got := n.Generate()
		if _, exists := seen[got]; exists {
			assert.Failf(t, "生成的 ID 必须唯一", "重复 ID: %d", got)
		}
		seen[got] = struct{}{}

		if i > 0 {
			assert.Greater(t, got.Int64(), last.Int64())
		}
		assert.Equal(t, int64(1), got.Node())
		assert.GreaterOrEqual(t, got.Step(), int64(0))
		assert.LessOrEqual(t, got.Step(), stepMask)
		last = got
	}

	assert.Len(t, seen, generateCount)
}

// TestNode_Generate_Concurrency 验证同一节点在并发调用 Generate 时仍保持 ID 唯一。
//
// 该测试通过多个 goroutine 同时请求 ID，并在主测试 goroutine 中统一断言结果，确保并发路径可由 race 检查验证。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNode_Generate_Concurrency(t *testing.T) {
	// 配置稳定的全局位宽，确保并发生成结果可按相同掩码解析。
	configureSnowflakeGlobals(t, 1740515125000, 10, 12)
	n := newTestNode(t, 2)

	const generateCount = 256
	ids := make(chan ID, generateCount)
	var wg sync.WaitGroup

	for i := 0; i < generateCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ids <- n.Generate()
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(ids)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		require.FailNow(t, "并发生成 ID 超时")
	}

	seen := make(map[ID]struct{}, generateCount)
	for id := range ids {
		if _, exists := seen[id]; exists {
			assert.Failf(t, "并发生成的 ID 必须唯一", "重复 ID: %d", id)
		}
		seen[id] = struct{}{}
		assert.Equal(t, int64(2), id.Node())
	}

	assert.Len(t, seen, generateCount)
}

// TestNode_Generate_SequenceOverflow_PublicAPI 验证公开配置下序列号耗尽时 Generate 等待到下一毫秒。
//
// 该测试将 StepBits 配置为零，使同一毫秒内没有额外序列容量，并通过公开 NewNode 与 Generate API 断言连续 ID 的时间分量递增。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestNode_Generate_SequenceOverflow_PublicAPI(t *testing.T) {
	// StepBits 为零时 stepMask 为零，连续生成必须通过推进时间分量保持唯一性。
	configureSnowflakeGlobals(t, 1740515125000, 10, 0)
	n := newTestNode(t, 3)

	const generateCount = 8
	ids := make([]ID, 0, generateCount)
	for i := 0; i < generateCount; i++ {
		ids = append(ids, n.Generate())
	}

	for i, id := range ids {
		assert.Equal(t, int64(3), id.Node())
		assert.Zero(t, id.Step())
		if i > 0 {
			assert.Greater(t, id.Time(), ids[i-1].Time())
			assert.Greater(t, id.Int64(), ids[i-1].Int64())
		}
	}
}

// TestID_ConvertersAndParsers_RoundTrip 验证 ID 的主要编码形式均可解析回原始值。
//
// 该测试通过表驱动用例覆盖十进制、二进制、Base32、Base36、Base58、Base64、字节和整数大端字节表示。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestID_ConvertersAndParsers_RoundTrip(t *testing.T) {
	const giveID = ID(1234567890)

	tests := []struct {
		name        string
		description string
		roundTrip   func(ID) (ID, error)
	}{
		{
			name:        "success/int64",
			description: "验证 Int64 与 ParseInt64 保持数值一致。",
			roundTrip: func(id ID) (ID, error) {
				return ParseInt64(id.Int64()), nil
			},
		},
		{
			name:        "success/string",
			description: "验证十进制字符串编码可解析回原始 ID。",
			roundTrip: func(id ID) (ID, error) {
				return ParseString(id.String())
			},
		},
		{
			name:        "success/base2",
			description: "验证二进制字符串编码可解析回原始 ID。",
			roundTrip: func(id ID) (ID, error) {
				return ParseBase2(id.Base2())
			},
		},
		{
			name:        "success/base32",
			description: "验证 z-base-32 编码可解析回原始 ID。",
			roundTrip: func(id ID) (ID, error) {
				return ParseBase32([]byte(id.Base32()))
			},
		},
		{
			name:        "success/base36",
			description: "验证 Base36 编码可解析回原始 ID。",
			roundTrip: func(id ID) (ID, error) {
				return ParseBase36(id.Base36())
			},
		},
		{
			name:        "success/base58",
			description: "验证 Base58 编码可解析回原始 ID。",
			roundTrip: func(id ID) (ID, error) {
				return ParseBase58([]byte(id.Base58()))
			},
		},
		{
			name:        "success/base64",
			description: "验证 Base64 编码可解析回原始 ID。",
			roundTrip: func(id ID) (ID, error) {
				return ParseBase64(id.Base64())
			},
		},
		{
			name:        "success/bytes",
			description: "验证十进制字节切片可解析回原始 ID。",
			roundTrip: func(id ID) (ID, error) {
				return ParseBytes(id.Bytes())
			},
		},
		{
			name:        "success/int-bytes",
			description: "验证大端整数二进制表示可解析回原始 ID。",
			roundTrip: func(id ID) (ID, error) {
				return ParseIntBytes(id.IntBytes()), nil
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := tt.roundTrip(giveID)

			require.NoError(t, err)
			assert.Equal(t, giveID, got)
		})
	}
}

// TestID_BaseEncodingBoundaries 验证 Base32 与 Base58 在单字符和进位边界上的编码契约。
//
// 该测试覆盖小于编码基数的直接映射分支，以及等于编码基数时的多字符编码分支。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestID_BaseEncodingBoundaries(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveID      ID
		encode      func(ID) string
		parse       func(string) (ID, error)
		wantEncoded string
	}{
		{
			name:        "boundary/base32-zero",
			description: "验证 Base32 对零值使用字符表首字符编码。",
			giveID:      0,
			encode:      func(id ID) string { return id.Base32() },
			parse:       func(s string) (ID, error) { return ParseBase32([]byte(s)) },
			wantEncoded: "y",
		},
		{
			name:        "boundary/base32-single-character-max",
			description: "验证 Base32 对小于三十二的最大值仍使用单字符编码。",
			giveID:      31,
			encode:      func(id ID) string { return id.Base32() },
			parse:       func(s string) (ID, error) { return ParseBase32([]byte(s)) },
			wantEncoded: "9",
		},
		{
			name:        "boundary/base32-carry",
			description: "验证 Base32 在三十二处进位为多字符编码。",
			giveID:      32,
			encode:      func(id ID) string { return id.Base32() },
			parse:       func(s string) (ID, error) { return ParseBase32([]byte(s)) },
			wantEncoded: "by",
		},
		{
			name:        "boundary/base58-zero",
			description: "验证 Base58 对零值使用字符表首字符编码。",
			giveID:      0,
			encode:      func(id ID) string { return id.Base58() },
			parse:       func(s string) (ID, error) { return ParseBase58([]byte(s)) },
			wantEncoded: "1",
		},
		{
			name:        "boundary/base58-single-character-max",
			description: "验证 Base58 对小于五十八的最大值仍使用单字符编码。",
			giveID:      57,
			encode:      func(id ID) string { return id.Base58() },
			parse:       func(s string) (ID, error) { return ParseBase58([]byte(s)) },
			wantEncoded: "Z",
		},
		{
			name:        "boundary/base58-carry",
			description: "验证 Base58 在五十八处进位为多字符编码。",
			giveID:      58,
			encode:      func(id ID) string { return id.Base58() },
			parse:       func(s string) (ID, error) { return ParseBase58([]byte(s)) },
			wantEncoded: "21",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			encoded := tt.encode(tt.giveID)
			got, err := tt.parse(encoded)

			assert.Equal(t, tt.wantEncoded, encoded)
			require.NoError(t, err)
			assert.Equal(t, tt.giveID, got)
		})
	}
}

// TestID_JSONSerialization 验证 ID 的 JSON 序列化和反序列化错误语义。
//
// 该测试覆盖合法字符串、非法 JSON 形态和带引号非数字内容，确保错误类型与成功赋值行为稳定。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestID_JSONSerialization(t *testing.T) {
	tests := []struct {
		name               string
		description        string
		giveJSON           string
		wantID             ID
		wantErr            bool
		wantSyntaxError    bool
		wantErrIs          error
		wantErrMsgContains string
	}{
		{
			name:        "success/quoted-decimal",
			description: "验证带引号十进制 JSON 字符串可反序列化为 ID。",
			giveJSON:    `"13587"`,
			wantID:      13587,
		},
		{
			name:               "error/not-quoted",
			description:        "验证未加引号的 JSON 数字被识别为 Snowflake JSON 语法错误。",
			giveJSON:           `1`,
			wantErr:            true,
			wantSyntaxError:    true,
			wantErrMsgContains: `invalid snowflake ID "1"`,
		},
		{
			name:               "error/missing-closing-quote",
			description:        "验证缺少闭合引号的 JSON 字节串被识别为 Snowflake JSON 语法错误。",
			giveJSON:           `"invalid`,
			wantErr:            true,
			wantSyntaxError:    true,
			wantErrMsgContains: `invalid snowflake ID`,
		},
		{
			name:               "error/quoted-nonnumeric",
			description:        "验证带引号但非数字的 JSON 字符串返回 strconv 语法错误。",
			giveJSON:           `"invalid"`,
			wantErr:            true,
			wantErrIs:          strconv.ErrSyntax,
			wantErrMsgContains: `invalid syntax`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)
			var got ID

			err := got.UnmarshalJSON([]byte(tt.giveJSON))

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantSyntaxError {
					var syntaxErr JSONSyntaxError
					require.ErrorAs(t, err, &syntaxErr)
				}
				if tt.wantErrIs != nil {
					assert.ErrorIs(t, err, tt.wantErrIs)
				}
				assert.Contains(t, err.Error(), tt.wantErrMsgContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantID, got)
		})
	}

	got, err := ID(13587).MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, []byte(`"13587"`), got)
}

// TestParse_InvalidInputs 验证各解析函数对非法输入返回可诊断错误。
//
// 该测试通过表驱动用例覆盖 strconv 支持的解析器、自定义 Base32/Base58 解析器和 Base64 解码错误。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestParse_InvalidInputs(t *testing.T) {
	tests := []struct {
		name        string
		description string
		parse       func() (ID, error)
		wantID      ID
		wantErrIs   error
	}{
		{
			name:        "error/string-syntax",
			description: "验证十进制字符串解析器拒绝非数字内容。",
			parse:       func() (ID, error) { return ParseString("not-number") },
			wantID:      0,
			wantErrIs:   strconv.ErrSyntax,
		},
		{
			name:        "error/base2-syntax",
			description: "验证二进制解析器拒绝非二进制数字。",
			parse:       func() (ID, error) { return ParseBase2("2") },
			wantID:      0,
			wantErrIs:   strconv.ErrSyntax,
		},
		{
			name:        "error/base36-syntax",
			description: "验证 Base36 解析器拒绝超出字符集的输入。",
			parse:       func() (ID, error) { return ParseBase36("!") },
			wantID:      0,
			wantErrIs:   strconv.ErrSyntax,
		},
		{
			name:        "error/bytes-syntax",
			description: "验证十进制字节解析器拒绝无法表示数字的字节序列。",
			parse:       func() (ID, error) { return ParseBytes([]byte{0xff, 0xff}) },
			wantID:      0,
			wantErrIs:   strconv.ErrSyntax,
		},
		{
			name:        "error/base64-syntax",
			description: "验证 Base64 解析器在解码失败时返回错误并使用 -1 作为结果。",
			parse:       func() (ID, error) { return ParseBase64("not_base64!!") },
			wantID:      -1,
		},
		{
			name:        "error/base32-invalid-character",
			description: "验证 z-base-32 解析器遇到未定义字符时返回 ErrInvalidBase32。",
			parse:       func() (ID, error) { return ParseBase32([]byte("B8WJM1ZROYYYY")) },
			wantID:      -1,
			wantErrIs:   ErrInvalidBase32,
		},
		{
			name:        "error/base58-invalid-character",
			description: "验证 Base58 解析器遇到未定义字符时返回 ErrInvalidBase58。",
			parse:       func() (ID, error) { return ParseBase58([]byte("0jgmnx8Js8A")) },
			wantID:      -1,
			wantErrIs:   ErrInvalidBase58,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := tt.parse()

			require.Error(t, err)
			if tt.wantErrIs != nil {
				assert.ErrorIs(t, err, tt.wantErrIs)
			}
			assert.Equal(t, tt.wantID, got)
		})
	}
}

// TestParse_OverflowInputs 验证基于 strconv 的解析入口对 int64 溢出输入返回范围错误。
//
// 该测试覆盖十进制字符串、二进制、Base36、字节、Base64 包装字节以及 JSON 反序列化入口，确保溢出不会被静默当作成功解析。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestParse_OverflowInputs(t *testing.T) {
	const (
		overflowDecimal = "9223372036854775808"
		maxInt64ID      = ID(1<<63 - 1)
	)

	tests := []struct {
		name        string
		description string
		parse       func() (ID, error)
		wantID      ID
	}{
		{
			name:        "error/string-overflow",
			description: "验证十进制字符串超过 int64 上限时返回 strconv 范围错误。",
			parse:       func() (ID, error) { return ParseString(overflowDecimal) },
			wantID:      maxInt64ID,
		},
		{
			name:        "error/base2-overflow",
			description: "验证二进制字符串超过 int64 上限时返回 strconv 范围错误。",
			parse:       func() (ID, error) { return ParseBase2("1" + strings.Repeat("0", 63)) },
			wantID:      maxInt64ID,
		},
		{
			name:        "error/base36-overflow",
			description: "验证 Base36 字符串超过 int64 上限时返回 strconv 范围错误。",
			parse:       func() (ID, error) { return ParseBase36("1y2p0ij32e8e8") },
			wantID:      maxInt64ID,
		},
		{
			name:        "error/bytes-overflow",
			description: "验证十进制字节切片超过 int64 上限时返回 strconv 范围错误。",
			parse:       func() (ID, error) { return ParseBytes([]byte(overflowDecimal)) },
			wantID:      maxInt64ID,
		},
		{
			name:        "error/base64-decimal-overflow",
			description: "验证 Base64 解码后的十进制字节超过 int64 上限时透传范围错误。",
			parse: func() (ID, error) {
				return ParseBase64(base64.StdEncoding.EncodeToString([]byte(overflowDecimal)))
			},
			wantID: maxInt64ID,
		},
		{
			name:        "error/json-overflow-preserves-existing-id",
			description: "验证 JSON 反序列化遇到 int64 溢出时返回范围错误且不覆盖既有 ID。",
			parse: func() (ID, error) {
				id := ID(99)
				err := id.UnmarshalJSON([]byte(`"` + overflowDecimal + `"`))
				return id, err
			},
			wantID: 99,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got, err := tt.parse()

			require.Error(t, err)
			assert.ErrorIs(t, err, strconv.ErrRange)
			var numErr *strconv.NumError
			assert.ErrorAs(t, err, &numErr)
			assert.Equal(t, tt.wantID, got)
		})
	}
}

// TestID_DeprecatedAccessors_DecodeComponents 验证 Time、Node 和 Step 按当前全局掩码解码 ID 组件。
//
// 该测试构造确定性的 ID 位布局，覆盖兼容旧版本的组件访问方法并恢复包级全局状态。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestID_DeprecatedAccessors_DecodeComponents(t *testing.T) {
	const (
		testEpoch   = int64(1000)
		giveTime    = int64(123)
		giveNodeID  = int64(17)
		giveStep    = int64(5)
		wantAbsTime = testEpoch + giveTime
	)

	// 配置固定全局掩码，使手工构造的 ID 与访问器解析规则一致。
	configureSnowflakeGlobals(t, testEpoch, 10, 12)
	id := ID((giveTime << timeShift) | (giveNodeID << nodeShift) | giveStep)

	assert.Equal(t, wantAbsTime, id.Time())
	assert.Equal(t, giveNodeID, id.Node())
	assert.Equal(t, giveStep, id.Step())
}

// TestID_IntBytes_BigEndian 验证 IntBytes 使用八字节大端序表示 ID。
//
// 该测试通过表驱动用例断言具体字节内容，并验证 ParseIntBytes 能够还原原始 ID。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestID_IntBytes_BigEndian(t *testing.T) {
	tests := []struct {
		name        string
		description string
		giveID      ID
		wantBytes   [8]byte
	}{
		{
			name:        "success/small-positive-id",
			description: "验证较小正整数 ID 被编码为高位补零的大端字节数组。",
			giveID:      13587,
			wantBytes:   [8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x35, 0x13},
		},
		{
			name:        "success/larger-positive-id",
			description: "验证较大正整数 ID 的大端字节数组可被 ParseIntBytes 还原。",
			giveID:      1234567890,
			wantBytes:   [8]byte{0x00, 0x00, 0x00, 0x00, 0x49, 0x96, 0x02, 0xd2},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.description)

			got := tt.giveID.IntBytes()

			assert.Equal(t, tt.wantBytes, got)
			assert.Equal(t, tt.giveID, ParseIntBytes(got))
		})
	}
}

// newTestNode 创建用于测试的 Snowflake 节点。
//
// 该辅助函数集中处理 NewNode 的前置断言，确保调用方只在节点成功初始化后继续测试。
//
// 参数：
//   - t: 测试上下文，用于报告节点初始化失败并标记辅助函数调用栈。
//   - giveNodeID: 需要创建的节点编号。
//
// 返回：
//   - Node: 已成功初始化的 Snowflake 节点。
func newTestNode(t *testing.T, giveNodeID int64) Node {
	t.Helper()

	n, err := NewNode(giveNodeID)
	require.NoError(t, err)
	require.NotNil(t, n)
	return n
}

// configureSnowflakeGlobals 配置测试所需的 Snowflake 包级全局状态。
//
// 该辅助函数会先注册清理逻辑，再按给定位宽重建派生掩码，避免包级状态污染后续用例。
//
// 参数：
//   - t: 测试上下文，用于注册清理逻辑并标记辅助函数调用栈。
//   - giveEpoch: Snowflake 起始时间戳，单位为毫秒。
//   - giveNodeBits: 节点编号占用的比特数。
//   - giveStepBits: 序列号占用的比特数。
func configureSnowflakeGlobals(t *testing.T, giveEpoch int64, giveNodeBits, giveStepBits uint8) {
	t.Helper()

	preserveSnowflakeGlobals(t)
	applySnowflakeGlobalState(deriveSnowflakeGlobalState(giveEpoch, giveNodeBits, giveStepBits))
}

// preserveSnowflakeGlobals 保存当前 Snowflake 包级全局状态并注册恢复逻辑。
//
// 该辅助函数用于隔离会修改 Epoch、NodeBits、StepBits 或派生掩码的测试用例。
//
// 参数：
//   - t: 测试上下文，用于注册清理逻辑并标记辅助函数调用栈。
func preserveSnowflakeGlobals(t *testing.T) {
	t.Helper()

	mu.Lock()
	snapshot := snowflakeGlobalState{
		epoch:     Epoch,
		nodeBits:  NodeBits,
		stepBits:  StepBits,
		nodeMax:   nodeMax,
		nodeMask:  nodeMask,
		stepMask:  stepMask,
		timeShift: timeShift,
		nodeShift: nodeShift,
	}
	mu.Unlock()

	t.Cleanup(func() {
		applySnowflakeGlobalState(snapshot)
	})
}

// deriveSnowflakeGlobalState 根据基础配置计算 Snowflake 派生全局状态。
//
// 该辅助函数复用生产代码中的位掩码公式，使测试中的显式配置与 NewNode 初始化规则保持一致。
//
// 参数：
//   - giveEpoch: Snowflake 起始时间戳，单位为毫秒。
//   - giveNodeBits: 节点编号占用的比特数。
//   - giveStepBits: 序列号占用的比特数。
//
// 返回：
//   - snowflakeGlobalState: 包含基础配置和派生掩码的全局状态快照。
func deriveSnowflakeGlobalState(giveEpoch int64, giveNodeBits, giveStepBits uint8) snowflakeGlobalState {
	nodeMaximum := int64(-1 ^ (-1 << giveNodeBits))
	stepMaximum := int64(-1 ^ (-1 << giveStepBits))

	return snowflakeGlobalState{
		epoch:     giveEpoch,
		nodeBits:  giveNodeBits,
		stepBits:  giveStepBits,
		nodeMax:   nodeMaximum,
		nodeMask:  nodeMaximum << giveStepBits,
		stepMask:  stepMaximum,
		timeShift: giveNodeBits + giveStepBits,
		nodeShift: giveStepBits,
	}
}

// applySnowflakeGlobalState 恢复或设置 Snowflake 包级全局状态。
//
// 该辅助函数在持有全局互斥锁时更新基础配置和派生掩码，避免测试内状态切换出现部分写入。
//
// 参数：
//   - state: 需要应用到包级变量的全局状态快照。
func applySnowflakeGlobalState(state snowflakeGlobalState) {
	mu.Lock()
	defer mu.Unlock()

	Epoch = state.epoch
	NodeBits = state.nodeBits
	StepBits = state.stepBits
	nodeMax = state.nodeMax
	nodeMask = state.nodeMask
	stepMask = state.stepMask
	timeShift = state.timeShift
	nodeShift = state.nodeShift
}
