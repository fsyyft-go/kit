// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

// HeartbeatMessageType 心跳消息类型常量。
const (
	HeartbeatMessageType    uint16 = 0x80 // 心跳消息。
	SingleStringMessageType uint16 = 0x09 // 简单的字符串消息。
)

// init 注册心跳消息和简单字符串消息的生成方法到工厂。
func init() {
	_ = FactoryRegister(HeartbeatMessageType, GenerateHeartbeatMessage)
	_ = FactoryRegister(SingleStringMessageType, GenerateSingleStringMessage)
}
