// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package message

import (
	"sync"

	cockroachdberrors "github.com/cockroachdb/errors"
)

var (
	// 断言 messageFactory 实现 MessageFactory 接口。
	_ MessageFactory = (*messageFactory)(nil)
)

var (
	// defaultFactory 保存包级注册和生成函数使用的默认消息工厂。
	defaultFactory MessageFactory = NewMessageFactory()
)

type (
	// MessageFactory 定义消息类型注册和消息生成契约。
	//
	// Register 负责将消息类型映射到生成函数；Generate 根据消息类型和 payload 还原消息实例。
	MessageFactory interface {
		// Register 将消息类型注册到工厂。
		//
		// 同一 messageType 只能注册一次；messageFunc 不能为空。
		//
		// 参数：
		//   - messageType: 待注册的消息类型。
		//   - messageFunc: 对应的消息生成函数。
		//
		// 返回：
		//   - error: 消息类型已存在或 messageFunc 为空时返回错误。
		Register(messageType MessageType, messageFunc GenerateMessageFunc) error
		// Generate 根据消息类型和 payload 创建消息实例。
		//
		// payload 不能为空，但可以是非 nil 的空切片。
		//
		// 参数：
		//   - messageType: 目标消息类型。
		//   - payload: 待解码的消息 payload。
		//
		// 返回：
		//   - Message: 生成的消息实例。
		//   - error: 消息类型未注册、payload 为 nil 或生成失败时返回错误。
		Generate(messageType MessageType, payload []byte) (Message, error)
	}

	// messageFactory 是 [MessageFactory] 的默认实现。
	messageFactory struct {
		funcs          map[MessageType]GenerateMessageFunc // 消息类型与生成方法映射表。
		registerLocker sync.Locker                         // 注册操作互斥锁。
	}
)

// Register 将消息类型注册到工厂。
//
// 参数：
//   - messageType: 待注册的消息类型。
//   - messageFunc: 对应的消息生成函数。
//
// 返回：
//   - error: 消息类型已存在或 messageFunc 为空时返回错误。
func (f *messageFactory) Register(messageType MessageType, messageFunc GenerateMessageFunc) error {
	var err error

	f.registerLocker.Lock()
	defer f.registerLocker.Unlock()

	if _, exists := f.funcs[messageType]; exists {
		err = cockroachdberrors.Newf("类型 %[1]d 已经存在。", messageType)
	} else if nil == messageFunc {
		err = cockroachdberrors.Newf("方法不允许为空。")
	} else {
		f.funcs[messageType] = messageFunc
	}

	return err
}

// Generate 根据消息类型和 payload 创建消息实例。
//
// 参数：
//   - messageType: 目标消息类型。
//   - payload: 待解码的消息 payload。
//
// 返回：
//   - Message: 生成的消息实例。
//   - error: 消息类型未注册、payload 为 nil 或生成失败时返回错误。
func (f *messageFactory) Generate(messageType MessageType, payload []byte) (Message, error) {
	var message Message
	var err error

	if mf, exists := f.funcs[messageType]; !exists {
		err = cockroachdberrors.Newf("类型 %[1]d 不存在。", messageType)
	} else if nil == payload {
		err = cockroachdberrors.Newf("有效负载不能为空。")
	} else {
		message, err = mf(messageType, payload)
	}

	return message, err
}

// NewMessageFactory 创建新的消息工厂实例。
//
// 返回的工厂可并发调用 Register；Generate 依赖底层 map 读取，通常应在完成注册后再供并发生成使用。
//
// 参数：无。
//
// 返回：
//   - *messageFactory: 新创建的消息工厂实例。
func NewMessageFactory() *messageFactory {
	mf := &messageFactory{
		funcs:          make(map[MessageType]GenerateMessageFunc),
		registerLocker: &sync.Mutex{},
	}

	return mf
}

// FactoryRegister 将消息类型注册到默认工厂。
//
// 参数：
//   - messageType: 待注册的消息类型。
//   - messageFunc: 对应的消息生成函数。
//
// 返回：
//   - error: 默认工厂注册失败时返回错误。
func FactoryRegister(messageType MessageType, messageFunc GenerateMessageFunc) error {
	return defaultFactory.Register(messageType, messageFunc)
}

// FactoryGenerate 使用默认工厂根据消息类型和 payload 创建消息实例。
//
// 参数：
//   - messageType: 目标消息类型。
//   - payload: 待解码的消息 payload。
//
// 返回：
//   - Message: 生成的消息实例。
//   - error: 默认工厂生成失败时返回错误。
func FactoryGenerate(messageType MessageType, payload []byte) (Message, error) {
	return defaultFactory.Generate(messageType, payload)
}
