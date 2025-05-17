# message

## 简介

message 包提供高性能的自定义消息协议与连接封装，支持消息类型注册、心跳包、字符串消息、自动分包、并发安全等，适用于分布式服务、长连接、定制协议等多种场景。

### 主要特性

- 统一消息接口，支持自定义消息类型、封包与解包
- 消息工厂机制，支持类型注册与动态生成
- 高性能连接封装，支持并发安全、自动分包、心跳机制
- 内置心跳消息、字符串消息实现
- 支持 bufio.Scanner 自动分割消息包
- 完整单元测试覆盖

### 设计理念

message 包遵循"高效、灵活、可扩展"的设计理念，接口与实现分离，工厂注册机制灵活，连接层支持自动分包与心跳，适合高并发、定制协议、长连接等场景。

## 安装

### 前置条件

- Go 版本要求：Go 1.18+
- 依赖要求：
  - github.com/stretchr/testify（仅测试）

### 安装命令

```bash
go get -u github.com/fsyyft-go/kit/net/message
```

## 快速开始

### 基础用法

```go
package main

import (
    "context"
    "fmt"
    "net"
    kitmsg "github.com/fsyyft-go/kit/net/message"
)

func main() {
    // 启动 TCP 服务端
    ln, _ := net.Listen("tcp", ":9000")
    go func() {
        conn, _ := ln.Accept()
        msgConn := kitmsg.WrapConn(conn, 0)
        msgConn.Start(context.Background())
        for msg := range msgConn.Message() {
            fmt.Println("收到消息:", msg.MessageType())
        }
    }()

    // 客户端连接
    conn, _ := net.Dial("tcp", ":9000")
    msgConn := kitmsg.WrapConn(conn, 0)
    msgConn.Start(context.Background())
    // 发送字符串消息
    msgConn.SendMessage(kitmsg.NewSingleStringMessage("hello"))
}
```

### 注册自定义消息类型

```go
// 定义自定义消息类型常量
type MyMessage struct { ... }
const MyType uint16 = 0x10
// 实现 Message 接口并注册到工厂
kitmsg.FactoryRegister(MyType, func(typ uint16, payload []byte) (kitmsg.Message, error) { ... })
```

## 详细指南

### 核心概念

- **Message 接口**：统一封装消息类型、封包与解包方法
- **消息工厂**：支持类型注册与动态生成，便于扩展
- **Conn 接口**：自定义消息连接，支持并发安全、自动分包、心跳
- **心跳消息/字符串消息**：内置常用消息类型
- **Scanner**：支持 bufio.Scanner 自动分割消息包

### 常见用例

- 分布式服务自定义协议通信
- 长连接心跳与消息收发
- 物联网、游戏、IM 等定制消息协议
- 高并发消息网关

### 最佳实践

- 注册所有自定义消息类型，避免类型冲突
- 合理设置心跳间隔，防止连接假死
- 使用消息工厂统一管理类型与生成逻辑
- 充分利用并发安全的连接封装

### 添加自定义消息类型的详细步骤

要扩展一个新的消息类型，建议遵循如下流程：

#### 1. 在 `message_type.go` 中定义消息类型常量并注册

```go
// 步骤1：定义唯一的消息类型常量
const MyCustomMessageType uint16 = 0x20 // 需保证唯一，避免与已有类型冲突

// 步骤4：在 init() 中注册工厂函数
func init() {
    // ...已有注册...
    _ = FactoryRegister(MyCustomMessageType, GenerateMyCustomMessage)
}
```
> 说明：所有消息类型常量建议集中管理，注册必须在 `init()` 完成，确保工厂可用。

#### 2. 定义扩展接口（如有需要）和消息结构体

```go
// 步骤2：如有扩展需求，定义扩展接口
type MyCustomMessage interface {
    Message // 必须嵌入基础 Message 接口
    // 可扩展自定义方法
    CustomField() string
}

// 步骤2：实现消息结构体
type myCustomMessage struct {
    messageType uint16
    customField string
    // 其它字段...
}

// 实现扩展接口方法
func (m *myCustomMessage) CustomField() string {
    return m.customField
}
```
> 说明：扩展接口便于类型断言和业务扩展，结构体需包含类型和自定义字段。

#### 3. 实现 Message 接口

```go
// 步骤3：实现 Message 接口
func (m *myCustomMessage) MessageType() uint16 {
    return m.messageType
}

func (m *myCustomMessage) Pack() ([]byte, error) {
    // 将 customField 等内容序列化为字节数组
    // 伪代码：return []byte(m.customField), nil
}

func (m *myCustomMessage) Unpack(payload []byte) error {
    // 从 payload 反序列化 customField
    // 伪代码：m.customField = string(payload); return nil
}
```
> 说明：Pack/Unpack 需保证与协议一致，错误处理要健壮。

#### 4. 实现 GenerateMessageFunc 工厂函数

```go
// 步骤4：实现工厂函数
func GenerateMyCustomMessage(messageType uint16, payload []byte) (Message, error) {
    if messageType != MyCustomMessageType {
        // 类型校验，防止误用
        return nil, errors.New("消息类型不匹配")
    }
    m := &myCustomMessage{messageType: MyCustomMessageType}
    // 调用 Unpack 解析 payload
    if err := m.Unpack(payload); err != nil {
        return nil, err
    }
    return m, nil
}
```
> 说明：工厂函数用于工厂动态生成消息实例，需校验类型并处理 payload。

---

**补充说明**：
- 推荐为新类型编写单元测试，覆盖 Pack/Unpack/工厂函数的正常与异常场景。
- 若有复杂字段，建议采用二进制序列化/反序列化方案，保证兼容性和健壮性。

## API 文档

### 主要类型

```go
// Message 消息接口
type Message interface {
    MessageType() uint16
    Pack() ([]byte, error)
    Unpack(payload []byte) error
}

// Conn 消息连接接口
type Conn interface {
    Close() error
    LocalAddr() net.Addr
    RemoteAddr() net.Addr
    SetDeadline(time.Time) error
    SetReadDeadline(time.Time) error
    SetWriteDeadline(time.Time) error
    Closed() bool
    Start(context.Context)
    SendMessage(Message) error
    Message() <-chan Message
}

// 消息工厂注册与生成
func FactoryRegister(messageType uint16, fn GenerateMessageFunc) error
func FactoryGenerate(messageType uint16, payload []byte) (Message, error)

// 内置消息类型
const (
    HeartbeatMessageType    uint16 = 0x80
    SingleStringMessageType uint16 = 0x09
)

// 内置消息构造
func NewHeartbeatMessage(sn uint64) *heartbeatMessage
func NewSingleStringMessage(msg string) *singleStringMessage

// 连接封装
func WrapConn(c net.Conn, heartbeatInterval time.Duration) *conn
```

### 关键函数

- `WrapConn`：将 net.Conn 封装为消息连接，支持心跳与自动分包
- `SendMessage`：发送消息（并发安全）
- `Message`：接收消息通道（只读）
- `FactoryRegister/FactoryGenerate`：注册与生成自定义消息类型
- `NewHeartbeatMessage/NewSingleStringMessage`：内置消息构造
- `NewScanner`：创建自定义分包 Scanner

## 错误处理

- 所有接口方法均返回 error，需检查
- 消息类型未注册、payload 非法等均有详细错误
- 连接关闭、超时、网络异常均有详细提示

## 性能指标

- 单连接高并发安全，消息收发吞吐量高
- 自动分包与缓冲队列，适合大规模消息场景
- 心跳与超时机制对性能影响可控

## 测试覆盖率

- 单元测试覆盖所有接口、边界、异常场景
- 使用 testify，覆盖率高

## 调试指南

- 检查消息类型注册与工厂逻辑
- 合理设置心跳与缓冲区参数
- 利用测试用例覆盖边界与异常

## 相关文档

- [Go net 官方文档](https://pkg.go.dev/net)
- [bufio.Scanner 官方文档](https://pkg.go.dev/bufio#Scanner)

## 贡献指南

欢迎提交 Issue、PR 或建议，详见 [贡献指南](../../CONTRIBUTING.md)。

## 许可证

本项目采用 MIT License 许可证。详见 [LICENSE](../../LICENSE)。 