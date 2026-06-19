// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

type (
	// MessageType 标识协议中的消息类型。
	//
	// 内置消息类型包括：
	//   - HeartbeatMessageType: 心跳消息类型。
	//   - SingleStringMessageType: 仅携带单个字符串 payload 的消息类型。
	//
	// 调用方可通过 FactoryRegister 注册其它 uint16 值作为自定义消息类型。
	MessageType uint16
)

const (
	// HeartbeatMessageType 表示心跳消息类型。
	HeartbeatMessageType MessageType = 0x80
	// SingleStringMessageType 表示仅携带单个字符串 payload 的消息类型。
	SingleStringMessageType MessageType = 0x09
)

// init 注册心跳消息和简单字符串消息的生成方法到默认工厂。
//
// 参数：无。
func init() {
	if err := FactoryRegister(HeartbeatMessageType, GenerateHeartbeatMessage); nil != err {
		panic(err)
	}
	if err := FactoryRegister(SingleStringMessageType, GenerateSingleStringMessage); nil != err {
		panic(err)
	}
}
