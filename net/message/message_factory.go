package message

import (
	"sync"

	cockroachdbErrors "github.com/cockroachdb/errors"
)

var (
	_ MessageFactory = (*messageFactory)(nil)
)

var (
	// defaultFactory 默认消息工厂。
	defaultFactory MessageFactory = NewMessageFactory()
)

type (
	// MessageFactory 消息工厂。
	MessageFactory interface {
		// Register 注册消息类型到工厂。
		Register(messageType uint16, messageFunc GenerateMessageFunc) error
		// Generate 根据消息类型和消息负载数据创建消息。
		Generate(messageType uint16, payload []byte) (Message, error)
	}

	// messageFactory 消息工厂。
	messageFactory struct {
		funcs          map[uint16]GenerateMessageFunc
		registerLocker sync.Locker
	}
)

// Register 注册消息类型到工厂。
func (f *messageFactory) Register(messageType uint16, messageFunc GenerateMessageFunc) error {
	var err error

	f.registerLocker.Lock()
	defer f.registerLocker.Unlock()

	if _, exists := f.funcs[messageType]; exists {
		err = cockroachdbErrors.Newf("类型 %[1]d 已经存在。", messageType)
	} else if nil == messageFunc {
		err = cockroachdbErrors.Newf("方法不允许为空。")
	} else {
		f.funcs[messageType] = messageFunc
	}

	return err
}

// Generate 根据消息类型和消息负载数据创建消息。
func (f messageFactory) Generate(messageType uint16, payload []byte) (Message, error) {
	var message Message
	var err error

	if mf, exists := f.funcs[messageType]; !exists {
		err = cockroachdbErrors.Newf("类型 %[1]d 不存在。", messageType)
	} else if nil == payload {
		err = cockroachdbErrors.Newf("有效负载不能为空。")
	} else {
		message, err = mf(messageType, payload)
	}

	return message, err
}

// NewMessageFactory 创建一个新消息工厂。
func NewMessageFactory() *messageFactory {
	mf := &messageFactory{
		funcs:          make(map[uint16]GenerateMessageFunc),
		registerLocker: &sync.Mutex{},
	}

	return mf
}

// FactoryRegister 注册消息类型到工厂。
func FactoryRegister(messageType uint16, messageFunc GenerateMessageFunc) error {
	return defaultFactory.Register(messageType, messageFunc)
}

// FactoryGenerate 根据消息类型和消息负载数据创建消息。
func FactoryGenerate(messageType uint16, payload []byte) (Message, error) {
	return defaultFactory.Generate(messageType, payload)
}
