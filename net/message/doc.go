// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

// Package message 提供基于自定义二进制协议的消息类型、工厂注册和连接封装。
//
// 本包使用固定 4 字节头部编码消息：前 2 字节为 MessageType，后 2 字节为 payload 长度，
// 因此单条消息的 payload 最大为 uint16 上限。调用方可以通过 Message、MessageFactory
// 和 FactoryRegister 扩展自定义消息类型；内置实现提供心跳消息、单字符串消息以及与该协议配套的 Scanner。
//
// WrapConn 会把 net.Conn 包装为按上述协议收发消息的连接，并可按给定间隔发送心跳包。
// 连接上的并发、生命周期和共享 channel 约束以 Conn 及其方法文档为准。
package message
