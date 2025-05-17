// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// 本文件实现了 Snowflake 算法，用于生成分布式唯一 ID。
// Snowflake 算法通过时间戳、节点编号和序列号组合生成 64 位唯一 ID。
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
	// 默认为 Twitter Snowflake 的起始时间 2025-02-25 20:25:25 UTC。
	// 可根据实际业务需求自定义。
	Epoch int64 = 1740515125000

	// NodeBits 表示节点编号占用的比特数。
	// 节点编号和序列号总共最多占用 22 位。
	NodeBits uint8 = 10

	// StepBits 表示序列号占用的比特数。
	// 节点编号和序列号总共最多占用 22 位。
	StepBits uint8 = 12

	// 以下四个变量为兼容旧版本，未来版本将移除。
	mu        sync.Mutex                         // 全局互斥锁，用于保护全局变量。
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
	// ErrInvalidBase58 表示解析 Base58 编码时遇到非法数据。
	ErrInvalidBase58 = errors.New("invalid base58")
	// ErrInvalidBase32 表示解析 Base32 编码时遇到非法数据。
	ErrInvalidBase32 = errors.New("invalid base32")
)

type (
	// JSONSyntaxError 表示 JSON 反序列化时遇到非法 ID 的错误类型。
	JSONSyntaxError struct{ original []byte }
)

// Error 返回 JSONSyntaxError 的错误描述。
//
// 返回值：
//   - string ：错误描述字符串。
func (j JSONSyntaxError) Error() string {
	return fmt.Sprintf("invalid snowflake ID %q", string(j.original))
}

// init 初始化 Base58 和 Base32 的解码映射表，加速解码过程。
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
	// Node 接口定义了生成唯一 ID 的方法。
	Node interface {
		// Generate 生成一个唯一的 Snowflake ID。
		Generate() ID
	}
	// node 结构体实现了 Node 接口，包含生成 Snowflake ID 所需的全部信息。
	node struct {
		mu    sync.Mutex // 互斥锁，保证并发安全。
		epoch time.Time  // 起始时间。
		time  int64      // 上一次生成 ID 的时间戳（毫秒）。
		node  int64      // 当前节点编号。
		step  int64      // 当前毫秒内的序列号。

		nodeMax   int64 // 节点编号最大值。
		nodeMask  int64 // 节点掩码。
		stepMask  int64 // 序列号掩码。
		timeShift uint8 // 时间戳左移位数。
		nodeShift uint8 // 节点编号左移位数。
	}
	// ID 类型用于表示 Snowflake 生成的唯一 ID。
	ID int64
)

// NewNode 创建并返回一个新的 node 实例，用于生成唯一 ID。
//
// 入参：
//   - nodeid ：int64，当前节点编号，需保证在 0 到 nodeMax 之间。
//
// 返回值：
//   - Node ：node 实例指针。
//   - error ：错误信息。
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
// 保证唯一性需确保系统时间准确，且节点编号唯一。
//
// 返回值：
//   - ID ：生成的唯一 ID。
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
// 返回值：
//   - int64 ：ID 的 int64 表示。
func (f ID) Int64() int64 {
	return int64(f)
}

// String 返回当前 ID 的字符串表示（十进制）。
//
// 返回值：
//   - string ：ID 的十进制字符串。
func (f ID) String() string {
	return strconv.FormatInt(int64(f), 10)
}

// Base2 返回当前 ID 的二进制字符串表示。
//
// 返回值：
//   - string ：ID 的二进制字符串。
func (f ID) Base2() string {
	return strconv.FormatInt(int64(f), 2)
}

// Base32 返回当前 ID 的 z-base-32 编码字符串。
// 注意：不同实现的 base32 可能不兼容，跨系统需谨慎。
//
// 返回值：
//   - string ：ID 的 base32 编码字符串。
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
// 返回值：
//   - string ：ID 的 base36 编码字符串。
func (f ID) Base36() string {
	return strconv.FormatInt(int64(f), 36)
}

// Base58 返回当前 ID 的 base58 编码字符串。
//
// 返回值：
//   - string ：ID 的 base58 编码字符串。
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

// Base64 返回当前 ID 的 base64 编码字符串。
//
// 返回值：
//   - string ：ID 的 base64 编码字符串。
func (f ID) Base64() string {
	return base64.StdEncoding.EncodeToString(f.Bytes())
}

// Bytes 返回当前 ID 的字节切片（十进制字符串形式）。
//
// 返回值：
//   - []byte ：ID 的字节切片。
func (f ID) Bytes() []byte {
	return []byte(f.String())
}

// IntBytes 返回当前 ID 的大端字节数组。
//
// 返回值：
//   - [8]byte ：ID 的大端字节数组。
func (f ID) IntBytes() [8]byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(f))
	return b
}

// Time 返回当前 ID 对应的时间戳（毫秒）。
// 该方法已废弃，未来版本将移除。
//
// 返回值：
//   - int64 ：ID 对应的时间戳（毫秒）。
func (f ID) Time() int64 {
	return (int64(f) >> timeShift) + Epoch
}

// Node 返回当前 ID 的节点编号。
// 该方法已废弃，未来版本将移除。
//
// 返回值：
//   - int64 ：ID 的节点编号。
func (f ID) Node() int64 {
	return int64(f) & nodeMask >> nodeShift
}

// Step 返回当前 ID 的序列号。
// 该方法已废弃，未来版本将移除。
//
// 返回值：
//   - int64 ：ID 的序列号。
func (f ID) Step() int64 {
	return int64(f) & stepMask
}

// MarshalJSON 实现 JSON 序列化，将 ID 转为字符串。
//
// 返回值：
//   - []byte ：JSON 字节数组。
//   - error ：错误信息。
func (f ID) MarshalJSON() ([]byte, error) {
	buff := make([]byte, 0, 22)
	buff = append(buff, '"')
	buff = strconv.AppendInt(buff, int64(f), 10)
	buff = append(buff, '"')
	return buff, nil
}

// UnmarshalJSON 实现 JSON 反序列化，将字符串转为 ID。
//
// 入参：
//   - b ：[]byte，JSON 字节数组。
//
// 返回值：
//   - error ：错误信息。
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

// ParseInt64 将 int64 转为 ID 类型。
//
// 入参：
//   - id ：int64。
//
// 返回值：
//   - ID ：转换后的 ID。
func ParseInt64(id int64) ID {
	return ID(id)
}

// ParseString 将字符串转为 ID 类型。
//
// 入参：
//   - id ：string。
//
// 返回值：
//   - ID ：转换后的 ID。
//   - error ：错误信息。
func ParseString(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 10, 64)
	return ID(i), err
}

// ParseBase2 将二进制字符串转为 ID 类型。
//
// 入参：
//   - id ：string。
//
// 返回值：
//   - ID ：转换后的 ID。
//   - error ：错误信息。
func ParseBase2(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 2, 64)
	return ID(i), err
}

// ParseBase32 将 base32 字节切片转为 ID 类型。
// 注意：不同实现的 base32 可能不兼容，跨系统需谨慎。
//
// 入参：
//   - b ：[]byte。
//
// 返回值：
//   - ID ：转换后的 ID。
//   - error ：错误信息。
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

// ParseBase36 将 base36 字符串转为 ID 类型。
//
// 入参：
//   - id ：string。
//
// 返回值：
//   - ID ：转换后的 ID。
//   - error ：错误信息。
func ParseBase36(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 36, 64)
	return ID(i), err
}

// ParseBase58 将 base58 字节切片转为 ID 类型。
//
// 入参：
//   - b ：[]byte。
//
// 返回值：
//   - ID ：转换后的 ID。
//   - error ：错误信息。
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

// ParseBase64 将 base64 字符串转为 ID 类型。
//
// 入参：
//   - id ：string。
//
// 返回值：
//   - ID ：转换后的 ID。
//   - error ：错误信息。
func ParseBase64(id string) (ID, error) {
	b, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return -1, err
	}
	return ParseBytes(b)
}

// ParseBytes 将字节切片（十进制字符串）转为 ID 类型。
//
// 入参：
//   - id ：[]byte。
//
// 返回值：
//   - ID ：转换后的 ID。
//   - error ：错误信息。
func ParseBytes(id []byte) (ID, error) {
	i, err := strconv.ParseInt(string(id), 10, 64)
	return ID(i), err
}

// ParseIntBytes 将大端字节数组转为 ID 类型。
//
// 入参：
//   - id ：[8]byte。
//
// 返回值：
//   - ID ：转换后的 ID。
func ParseIntBytes(id [8]byte) ID {
	return ID(int64(binary.BigEndian.Uint64(id[:])))
}
