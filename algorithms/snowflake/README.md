# snowflake

## 简介

`snowflake` 包实现了经典的 Snowflake 分布式唯一 ID 生成算法，适用于高并发、分布式系统中的唯一标识需求。该包支持多节点部署，生成的 64 位整数 ID 具备全局唯一性、趋势递增、可多种编码表示，广泛应用于数据库主键、分布式事务、消息队列等场景。

### 主要特性

- 分布式唯一 ID 生成，支持多节点高并发
- 生成 64 位整型 ID，趋势递增，适合排序
- 支持多种编码格式（十进制、二进制、Base32、Base36、Base58、Base64）
- 提供多种解析与转换方法，兼容多系统
- 并发安全，适合高并发场景
- 支持 JSON 序列化与反序列化
- 完善的错误处理与边界校验
- 单元测试覆盖率高

### 设计理念

`snowflake` 包遵循"高性能、易用性、兼容性"原则，采用互斥锁保证并发安全，节点号与序列号灵活配置，API 设计简洁直观。所有编码/解析方法均支持互逆，便于跨系统数据交换。

#### ID 结构说明

该库的 ID 由 64 位的整数组成，其中：

```
+---------------------------------------------------------------+
|  1 bit  |    41 bits    |  10 bits  |   12 bits   |
+---------------------------------------------------------------+
|  符号位 |  时间戳 (ms)  |  机器 ID  |  序列号    |
+---------------------------------------------------------------+
```

字段解释：

- **1 bit**：符号位（始终为 0）。
- **41 bits**：毫秒级时间戳（支持约 69 年）。
- **10 bits**：节点 ID（最多支持 1024 台机器）。
- **12 bits**：每毫秒最多生成 4096 个 ID。

## 安装

### 前置条件

- Go 版本要求：Go 1.16+
- 依赖要求：无外部依赖，仅用标准库

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/algorithms/snowflake
```

## 快速开始

### 基础用法

```go
package main

import (
    "fmt"
    "github.com/fsyyft-go/kit/algorithms/snowflake"
)

func main() {
    // 创建节点（节点编号需唯一，范围 0~1023）
    node, err := snowflake.NewNode(1)
    if err != nil {
        panic(err)
    }
    // 生成唯一 ID
    id := node.Generate()
    fmt.Println("ID:", id.Int64())
    fmt.Println("Base58:", id.Base58())
}
```

## 详细指南

### 核心概念

- **ID 结构**：64 位整数，由时间戳、节点编号、序列号组成，趋势递增。
- **节点编号**：每个节点需分配唯一编号，最大支持 1024 个节点（可配置）。
- **序列号**：同一毫秒内自增，避免并发冲突。
- **多种编码**：支持十进制、二进制、Base32、Base36、Base58、Base64 等多种编码与解析。

### 常见用例

#### 1. 生成唯一 ID 并多种编码

```go
id := node.Generate()
fmt.Println("十进制:", id.String())
fmt.Println("Base32:", id.Base32())
fmt.Println("Base58:", id.Base58())
fmt.Println("Base64:", id.Base64())
```

#### 2. ID 解析与互逆

```go
id := node.Generate()
s := id.Base58()
parsed, err := snowflake.ParseBase58([]byte(s))
if err != nil {
    // 处理解析错误
}
fmt.Println(parsed == id) // true
```

#### 3. JSON 序列化与反序列化

```go
import "encoding/json"

id := node.Generate()
b, _ := json.Marshal(id)
var id2 snowflake.ID
_ = json.Unmarshal(b, &id2)
fmt.Println(id == id2) // true
```

### 最佳实践

- 每个节点分配唯一编号，避免冲突
- 系统时间需准确，避免回拨
- 并发场景下可安全调用 Generate
- 合理选择编码格式，跨系统建议用十进制或 Base58
- 检查解析函数返回的错误，防止非法输入

## API 文档

### 主要类型

```go
// Node 接口定义唯一 ID 生成方法
 type Node interface {
     Generate() ID
 }

// node 结构体实现 Node 接口
// ID 类型为 int64
```

### 关键函数

#### NewNode

创建新节点。

```go
func NewNode(nodeid int64) (*node, error)
```
- nodeid：节点编号，需唯一且在合法范围内
- 返回 node 实例和错误

#### Generate

生成唯一 ID。

```go
func (n *node) Generate() ID
```

#### ID 编码与解析

- `String()`：十进制字符串
- `Base2()`：二进制字符串
- `Base32()`：z-base-32 编码
- `Base36()`：base36 编码
- `Base58()`：base58 编码
- `Base64()`：base64 编码
- `Bytes()`：十进制字节切片
- `IntBytes()`：大端字节数组

#### 解析函数

- `ParseString(s string) (ID, error)`
- `ParseBase2(s string) (ID, error)`
- `ParseBase32(b []byte) (ID, error)`
- `ParseBase36(s string) (ID, error)`
- `ParseBase58(b []byte) (ID, error)`
- `ParseBase64(s string) (ID, error)`
- `ParseBytes(b []byte) (ID, error)`
- `ParseIntBytes([8]byte) ID`

#### JSON 支持

- `MarshalJSON() ([]byte, error)`
- `UnmarshalJSON([]byte) error`

### 错误处理

- 节点编号非法、编码解析失败等均返回详细错误
- JSON 反序列化非法时返回自定义错误类型
- 建议始终检查解析和创建节点的错误返回值

## 性能指标

| 操作 | 性能 | 说明 |
|------|------|------|
| ID 生成 | O(1) | 支持高并发 |
| 编码/解析 | O(n) | n 为字符串长度 |

## 测试覆盖率

| 包 | 覆盖率 |
|------|--------|
| snowflake | >95% |

所有核心功能、边界和异常场景均有单元测试覆盖。

## 调试指南

### 常见问题排查

#### ID 重复
- 检查节点编号是否唯一
- 检查系统时间是否回拨

#### 解析失败
- 检查输入编码格式是否正确
- 检查是否有非法字符

## 相关文档

- [Twitter Snowflake 算法原理](https://github.com/twitter-archive/snowflake)
- [Go strconv 包文档](https://pkg.go.dev/strconv)
- [golang每日一库之snowflake](https://mp.weixin.qq.com/s/dewzsCAmE_xIEwDSHjc7FQ)
- [A simple to use Go (golang) package to generate or parse Twitter snowflake IDs ](https://github.com/bwmarrin/snowflake)

## 贡献指南

欢迎任何形式的贡献，包括但不限于：
- 报告问题
- 提交功能建议
- 提交代码改进
- 完善文档

请参考[贡献指南](../../CONTRIBUTING.md)了解详细信息。

## 许可证

本项目采用 MIT License 许可证。查看 [LICENSE](../../LICENSE) 文件了解更多信息。 