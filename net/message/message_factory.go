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
	// defaultFactory 默认消息工厂实例。
	defaultFactory MessageFactory = NewMessageFactory()
)

type (
	// MessageFactory 消息工厂接口，定义消息类型注册与消息生成方法。
	MessageFactory interface {
		// Register 注册消息类型到工厂。
		//
		// 参数：
		//   - messageType: 消息类型。
		//   - messageFunc: 消息生成方法。
		//
		// 返回值：
		//   - error: 错误信息。
		Register(messageType uint16, messageFunc GenerateMessageFunc) error
		// Generate 根据消息类型和消息负载数据创建消息。
		//
		// 参数：
		//   - messageType: 消息类型。
		//   - payload: 消息负载数据。
		//
		// 返回值：
		//   - Message: 生成的消息实例。
		//   - error: 错误信息。
		Generate(messageType uint16, payload []byte) (Message, error)
	}

	// messageFactory 消息工厂实现。
	messageFactory struct {
		funcs          map[uint16]GenerateMessageFunc // 消息类型与生成方法映射表。
		registerLocker sync.Locker                    // 注册操作互斥锁。
	}
)

// Register 注册消息类型到工厂。
//
// 参数：
//   - messageType: 消息类型。
//   - messageFunc: 消息生成方法。
//
// 返回值：
//   - error: 错误信息。
func (f *messageFactory) Register(messageType uint16, messageFunc GenerateMessageFunc) error {
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

// Generate 根据消息类型和消息负载数据创建消息。
//
// 参数：
//   - messageType: 消息类型。
//   - payload: 消息负载数据。
//
// 返回值：
//   - Message: 生成的消息实例。
//   - error: 错误信息。
func (f messageFactory) Generate(messageType uint16, payload []byte) (Message, error) {
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

// NewMessageFactory 创建一个新消息工厂实例。
//
// 无参数。
//
// 返回值：
//   - *messageFactory: 新建的消息工厂实例。
func NewMessageFactory() *messageFactory {
	mf := &messageFactory{
		funcs:          make(map[uint16]GenerateMessageFunc),
		registerLocker: &sync.Mutex{},
	}

	return mf
}

// FactoryRegister 注册消息类型到默认工厂。
//
// 参数：
//   - messageType: 消息类型。
//   - messageFunc: 消息生成方法。
//
// 返回值：
//   - error: 错误信息。
func FactoryRegister(messageType uint16, messageFunc GenerateMessageFunc) error {
	return defaultFactory.Register(messageType, messageFunc)
}

// FactoryGenerate 根据消息类型和消息负载数据创建消息。
//
// 参数：
//   - messageType: 消息类型。
//   - payload: 消息负载数据。
//
// 返回值：
//   - Message: 生成的消息实例。
//   - error: 错误信息。
func FactoryGenerate(messageType uint16, payload []byte) (Message, error) {
	return defaultFactory.Generate(messageType, payload)
}
