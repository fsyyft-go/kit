// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package snowflake

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

var (
	// Epoch 表示 Snowflake 算法的起始时间戳，单位为毫秒。
	// 默认值为项目约定的毫秒时间戳 1740515125000。
	// 调用方如需自定义，应在调用 NewNode 创建节点前完成配置。
	Epoch int64 = 1740515125000

	// NodeBits 表示节点编号占用的比特数。
	// NodeBits 和 StepBits 总和不能超过 22，调用方应在调用 NewNode 前完成配置。
	NodeBits uint8 = 10

	// StepBits 表示同一毫秒内序列号占用的比特数。
	// NodeBits 和 StepBits 总和不能超过 22，调用方应在调用 NewNode 前完成配置。
	StepBits uint8 = 12

	// 以下派生变量用于兼容旧版本的组件解析方法，并由 NewNode 按当前位宽重新计算。
	mu        sync.Mutex                         // 全局互斥锁，用于保护全局派生变量。
	nodeMax   int64      = -1 ^ (-1 << NodeBits) // 节点编号的最大值。
	nodeMask             = nodeMax << StepBits   // 节点掩码，用于提取节点编号。
	stepMask  int64      = -1 ^ (-1 << StepBits) // 序列号掩码，用于提取序列号。
	timeShift            = NodeBits + StepBits   // 时间戳左移位数。
	nodeShift            = StepBits              // 节点编号左移位数。
)

const (
	// encodeBase32Map 定义了 z-base-32 字符集，用于 Base32 编码。
	encodeBase32Map = "ybndrfg8ejkmcpqxot1uwisza345h769"
	// encodeBase58Map 定义了 Base58 字符集，用于 Base58 编码。
	encodeBase58Map = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
)

var (
	// decodeBase32Map 用于 Base32 解码，将字符映射为索引。
	decodeBase32Map [256]byte
	// decodeBase58Map 用于 Base58 解码，将字符映射为索引。
	decodeBase58Map [256]byte
	// ErrInvalidBase58 表示 Base58 解析遇到未定义字符。
	// ParseBase58 返回该错误时结果值为 -1，调用方可使用 errors.Is 判断该错误。
	ErrInvalidBase58 = errors.New("invalid base58")
	// ErrInvalidBase32 表示 z-base-32 解析遇到未定义字符。
	// ParseBase32 返回该错误时结果值为 -1，调用方可使用 errors.Is 判断该错误。
	ErrInvalidBase32 = errors.New("invalid base32")
)

type (
	// JSONSyntaxError 表示 JSON 反序列化时遇到非法 ID 形态的错误类型。
	JSONSyntaxError struct{ original []byte }
)

// Error 返回 JSONSyntaxError 的错误描述。
//
// 参数：无。
//
// 返回：
//   - string: 包含原始 JSON 字节内容的错误消息。
func (j JSONSyntaxError) Error() string {
	return fmt.Sprintf("invalid snowflake ID %q", string(j.original))
}

// init 初始化 Base58 和 Base32 的解码映射表，加速解码过程。
//
// 参数：无。
func init() {
	// 初始化 Base58 解码表，所有字符默认无效。
	for i := 0; i < len(decodeBase58Map); i++ {
		decodeBase58Map[i] = 0xFF
	}
	// 填充 Base58 有效字符的索引。
	for i := 0; i < len(encodeBase58Map); i++ {
		decodeBase58Map[encodeBase58Map[i]] = byte(i)
	}
	// 初始化 Base32 解码表，所有字符默认无效。
	for i := 0; i < len(decodeBase32Map); i++ {
		decodeBase32Map[i] = 0xFF
	}
	// 填充 Base32 有效字符的索引。
	for i := 0; i < len(encodeBase32Map); i++ {
		decodeBase32Map[encodeBase32Map[i]] = byte(i)
	}
}

type (
	// Node 定义生成 Snowflake ID 的节点能力。
	Node interface {
		// Generate 生成一个新的 Snowflake ID。
		//
		// 参数：无。
		//
		// 返回：
		//   - ID: 由当前节点生成的唯一 ID。
		Generate() ID
	}
	// node 实现 Node 接口，保存生成 Snowflake ID 所需的位布局和运行状态。
	node struct {
		mu    sync.Mutex // 互斥锁，保证同一节点并发生成 ID 时的状态安全。
		epoch time.Time  // 起始时间，保留单调时钟信息。
		time  int64      // 上一次生成 ID 的时间戳，单位为毫秒。
		node  int64      // 当前节点编号。
		step  int64      // 当前毫秒内的序列号。

		nodeMax   int64 // 节点编号最大值。
		nodeMask  int64 // 节点掩码。
		stepMask  int64 // 序列号掩码。
		timeShift uint8 // 时间戳左移位数。
		nodeShift uint8 // 节点编号左移位数。
	}
	// ID 表示 Snowflake 生成的唯一标识。
	ID int64
)

// NewNode 创建一个使用当前全局位布局的 Snowflake 节点。
//
// NewNode 会按当前 Epoch、NodeBits 和 StepBits 初始化节点，并为兼容旧版本组件
// 解析方法重新计算包级派生掩码。调用方应为每个节点分配唯一的 nodeid，并避免在节点
// 创建后修改全局位宽配置。
//
// 参数：
//   - nodeid: 节点编号，必须位于 0 到当前 NodeBits 可表示的最大值之间。
//
// 返回：
//   - Node: 初始化完成的节点，可用于并发生成 ID。
//   - error: NodeBits 与 StepBits 总和超过 22，或 nodeid 超出可用范围时返回错误。
func NewNode(nodeid int64) (Node, error) {

	if NodeBits+StepBits > 22 {
		return nil, errors.New("remember, you have a total 22 bits to share between Node/Step")
	}
	// 重新计算全局变量，兼容旧版本，未来将移除。
	mu.Lock()
	nodeMax = -1 ^ (-1 << NodeBits)
	nodeMask = nodeMax << StepBits
	stepMask = -1 ^ (-1 << StepBits)
	timeShift = NodeBits + StepBits
	nodeShift = StepBits
	mu.Unlock()

	n := node{}
	n.node = nodeid
	n.nodeMax = -1 ^ (-1 << NodeBits)
	n.nodeMask = n.nodeMax << StepBits
	n.stepMask = -1 ^ (-1 << StepBits)
	n.timeShift = NodeBits + StepBits
	n.nodeShift = StepBits

	if n.node < 0 || n.node > n.nodeMax {
		return nil, errors.New("Node number must be between 0 and " + strconv.FormatInt(n.nodeMax, 10))
	}

	var curTime = time.Now()
	// 通过时间差设置 epoch，保证使用单调时钟。
	n.epoch = curTime.Add(time.Unix(Epoch/1000, (Epoch%1000)*1000000).Sub(curTime))

	return &n, nil
}

// Generate 生成并返回一个唯一的 Snowflake ID。
//
// Generate 使用节点内部互斥锁保护时间戳和序列号状态；当同一毫秒内序列号耗尽时，
// 会等待到下一毫秒后继续生成。调用方需要保证不同节点使用不同 nodeid。
//
// 参数：无。
//
// 返回：
//   - ID: 当前节点生成的唯一 ID。
func (n *node) Generate() ID {

	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Since(n.epoch).Milliseconds()

	if now == n.time {
		n.step = (n.step + 1) & n.stepMask

		if n.step == 0 {
			// 当前毫秒内序列号溢出，等待到下一毫秒。
			for now <= n.time {
				now = time.Since(n.epoch).Milliseconds()
			}
		}
	} else {
		n.step = 0
	}

	n.time = now

	r := ID((now)<<n.timeShift |
		(n.node << n.nodeShift) |
		(n.step),
	)

	return r
}

// Int64 返回当前 ID 的 int64 表示。
//
// 参数：无。
//
// 返回：
//   - int64: ID 的原始整数值。
func (f ID) Int64() int64 {
	return int64(f)
}

// String 返回当前 ID 的十进制字符串表示。
//
// 参数：无。
//
// 返回：
//   - string: 使用十进制编码的 ID 字符串。
func (f ID) String() string {
	return strconv.FormatInt(int64(f), 10)
}

// Base2 返回当前 ID 的二进制字符串表示。
//
// 参数：无。
//
// 返回：
//   - string: 使用二进制编码的 ID 字符串。
func (f ID) Base2() string {
	return strconv.FormatInt(int64(f), 2)
}

// Base32 返回当前 ID 的 z-base-32 编码字符串。
//
// Base32 使用本包内置的 z-base-32 字符表，不保证与标准库 base32 编码互通。
//
// 参数：无。
//
// 返回：
//   - string: 使用 z-base-32 编码的 ID 字符串。
func (f ID) Base32() string {

	if f < 32 {
		return string(encodeBase32Map[f])
	}

	b := make([]byte, 0, 12)
	for f >= 32 {
		b = append(b, encodeBase32Map[f%32])
		f /= 32
	}
	b = append(b, encodeBase32Map[f])

	// 反转字节顺序，得到正确编码。
	for x, y := 0, len(b)-1; x < y; x, y = x+1, y-1 {
		b[x], b[y] = b[y], b[x]
	}

	return string(b)
}

// Base36 返回当前 ID 的 base36 编码字符串。
//
// 参数：无。
//
// 返回：
//   - string: 使用 base36 编码的 ID 字符串。
func (f ID) Base36() string {
	return strconv.FormatInt(int64(f), 36)
}

// Base58 返回当前 ID 的 Base58 编码字符串。
//
// Base58 使用本包内置字符表，排除了 0、O、I 和 l 等易混淆字符。
//
// 参数：无。
//
// 返回：
//   - string: 使用 Base58 编码的 ID 字符串。
func (f ID) Base58() string {

	if f < 58 {
		return string(encodeBase58Map[f])
	}

	b := make([]byte, 0, 11)
	for f >= 58 {
		b = append(b, encodeBase58Map[f%58])
		f /= 58
	}
	b = append(b, encodeBase58Map[f])

	// 反转字节顺序，得到正确编码。
	for x, y := 0, len(b)-1; x < y; x, y = x+1, y-1 {
		b[x], b[y] = b[y], b[x]
	}

	return string(b)
}

// Base64 返回当前 ID 十进制字节表示的 base64 编码字符串。
//
// 参数：无。
//
// 返回：
//   - string: 对 Bytes 返回值执行 base64.StdEncoding 后得到的字符串。
func (f ID) Base64() string {
	return base64.StdEncoding.EncodeToString(f.Bytes())
}

// Bytes 返回当前 ID 十进制字符串的字节表示。
//
// 参数：无。
//
// 返回：
//   - []byte: ID 的十进制字符串字节切片。
func (f ID) Bytes() []byte {
	return []byte(f.String())
}

// IntBytes 返回当前 ID 的八字节大端整数表示。
//
// 参数：无。
//
// 返回：
//   - [8]byte: 按 binary.BigEndian 写入的 ID 整数字节数组。
func (f ID) IntBytes() [8]byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(f))
	return b
}

// Time 返回当前 ID 对应的绝对时间戳，单位为毫秒。
//
// Time 使用当前包级 Epoch 和 timeShift 解析 ID；如果解析时的全局配置与生成时不一致，
// 返回值可能不代表原始生成时间。
//
// 参数：无。
//
// 返回：
//   - int64: ID 对应的毫秒时间戳。
//
// Deprecated: Time 依赖当前包级位布局解析旧 ID，保留用于兼容旧版本。
func (f ID) Time() int64 {
	return (int64(f) >> timeShift) + Epoch
}

// Node 返回当前 ID 编码的节点编号。
//
// Node 使用当前包级 nodeMask 和 nodeShift 解析 ID；如果解析时的全局配置与生成时不一致，
// 返回值可能不代表原始节点编号。
//
// 参数：无。
//
// 返回：
//   - int64: ID 中编码的节点编号。
//
// Deprecated: Node 依赖当前包级位布局解析旧 ID，保留用于兼容旧版本。
func (f ID) Node() int64 {
	return int64(f) & nodeMask >> nodeShift
}

// Step 返回当前 ID 编码的序列号。
//
// Step 使用当前包级 stepMask 解析 ID；如果解析时的全局配置与生成时不一致，
// 返回值可能不代表原始序列号。
//
// 参数：无。
//
// 返回：
//   - int64: ID 中编码的序列号。
//
// Deprecated: Step 依赖当前包级位布局解析旧 ID，保留用于兼容旧版本。
func (f ID) Step() int64 {
	return int64(f) & stepMask
}

// MarshalJSON 实现 JSON 序列化，将 ID 编码为带引号的十进制字符串。
//
// 参数：无。
//
// 返回：
//   - []byte: JSON 字符串字面量形式的 ID。
//   - error: 当前实现始终返回 nil。
func (f ID) MarshalJSON() ([]byte, error) {
	buff := make([]byte, 0, 22)
	buff = append(buff, '"')
	buff = strconv.AppendInt(buff, int64(f), 10)
	buff = append(buff, '"')
	return buff, nil
}

// UnmarshalJSON 实现 JSON 反序列化，将带引号的十进制字符串解析为 ID。
//
// 调用方应使用非 nil 的 ID 指针接收结果；解析成功后才会覆盖接收者。
//
// 参数：
//   - b: JSON 字节切片，必须是带引号的十进制字符串。
//
// 返回：
//   - error: b 不是带引号字符串时返回 JSONSyntaxError；十进制内容无法解析或超出 int64 范围时返回 strconv.ParseInt 的错误。
func (f *ID) UnmarshalJSON(b []byte) error {
	if len(b) < 3 || b[0] != '"' || b[len(b)-1] != '"' {
		return JSONSyntaxError{b}
	}

	i, err := strconv.ParseInt(string(b[1:len(b)-1]), 10, 64)
	if err != nil {
		return err
	}

	*f = ID(i)
	return nil
}

// ParseInt64 将 int64 值转换为 ID。
//
// 参数：
//   - id: 待转换的整数值。
//
// 返回：
//   - ID: 与 id 数值相同的 Snowflake ID。
func ParseInt64(id int64) ID {
	return ID(id)
}

// ParseString 将十进制字符串解析为 ID。
//
// 参数：
//   - id: 待解析的十进制字符串。
//
// 返回：
//   - ID: 解析成功时得到的 ID；解析失败时为 strconv.ParseInt 的中间结果，不应继续使用。
//   - error: id 不是合法十进制整数或超出 int64 范围时返回 strconv.ParseInt 的错误。
func ParseString(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 10, 64)
	return ID(i), err
}

// ParseBase2 将二进制字符串解析为 ID。
//
// 参数：
//   - id: 待解析的二进制字符串。
//
// 返回：
//   - ID: 解析成功时得到的 ID；解析失败时为 strconv.ParseInt 的中间结果，不应继续使用。
//   - error: id 不是合法二进制整数或超出 int64 范围时返回 strconv.ParseInt 的错误。
func ParseBase2(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 2, 64)
	return ID(i), err
}

// ParseBase32 将 z-base-32 字节切片解析为 ID。
//
// ParseBase32 使用本包内置的 z-base-32 字符表，不执行 int64 溢出检测。
//
// 参数：
//   - b: 待解析的 z-base-32 编码字节切片。
//
// 返回：
//   - ID: 解析成功时得到的 ID；遇到非法字符时返回 -1。
//   - error: b 包含未定义字符时返回 ErrInvalidBase32，调用方可使用 errors.Is 判断。
func ParseBase32(b []byte) (ID, error) {

	var id int64

	for i := range b {
		if decodeBase32Map[b[i]] == 0xFF {
			return -1, ErrInvalidBase32
		}
		id = id*32 + int64(decodeBase32Map[b[i]])
	}

	return ID(id), nil
}

// ParseBase36 将 base36 字符串解析为 ID。
//
// 参数：
//   - id: 待解析的 base36 字符串。
//
// 返回：
//   - ID: 解析成功时得到的 ID；解析失败时为 strconv.ParseInt 的中间结果，不应继续使用。
//   - error: id 不是合法 base36 整数或超出 int64 范围时返回 strconv.ParseInt 的错误。
func ParseBase36(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 36, 64)
	return ID(i), err
}

// ParseBase58 将 Base58 字节切片解析为 ID。
//
// ParseBase58 使用本包内置字符表，不执行 int64 溢出检测。
//
// 参数：
//   - b: 待解析的 Base58 编码字节切片。
//
// 返回：
//   - ID: 解析成功时得到的 ID；遇到非法字符时返回 -1。
//   - error: b 包含未定义字符时返回 ErrInvalidBase58，调用方可使用 errors.Is 判断。
func ParseBase58(b []byte) (ID, error) {

	var id int64

	for i := range b {
		if decodeBase58Map[b[i]] == 0xFF {
			return -1, ErrInvalidBase58
		}
		id = id*58 + int64(decodeBase58Map[b[i]])
	}

	return ID(id), nil
}

// ParseBase64 将 base64 字符串解析为 ID。
//
// ParseBase64 先使用 base64.StdEncoding 解码，再按 Bytes 的十进制字节表示解析。
//
// 参数：
//   - id: 待解析的 base64 字符串。
//
// 返回：
//   - ID: 解析成功时得到的 ID；base64 解码失败时返回 -1，十进制解析失败时返回 ParseBytes 的结果。
//   - error: id 不是合法 base64 字符串时返回解码错误；解码后内容不是合法十进制整数或超出 int64 范围时返回 strconv.ParseInt 的错误。
func ParseBase64(id string) (ID, error) {
	b, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return -1, err
	}
	return ParseBytes(b)
}

// ParseBytes 将十进制字符串字节切片解析为 ID。
//
// 参数：
//   - id: 待解析的十进制字符串字节切片。
//
// 返回：
//   - ID: 解析成功时得到的 ID；解析失败时为 strconv.ParseInt 的中间结果，不应继续使用。
//   - error: id 不是合法十进制整数或超出 int64 范围时返回 strconv.ParseInt 的错误。
func ParseBytes(id []byte) (ID, error) {
	i, err := strconv.ParseInt(string(id), 10, 64)
	return ID(i), err
}

// ParseIntBytes 将八字节大端整数表示转换为 ID。
//
// 参数：
//   - id: 按 binary.BigEndian 编码的八字节整数数组。
//
// 返回：
//   - ID: 从 id 还原得到的 Snowflake ID。
func ParseIntBytes(id [8]byte) ID {
	return ID(int64(binary.BigEndian.Uint64(id[:])))
}
