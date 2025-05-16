package message

type (
	// Message 消息包。
	Message interface {
		// MessageType 消息类型；
		MessageType() uint16

		// Pack 封包；
		//
		// 返回消息对应的 payload 的字节数组表示形式（不包含消息类型和长度）。
		Pack() ([]byte, error)

		// Unpack 拆包；
		//
		// payload 消息对应的字节数组的表示形式（不包含消息类型和长度）。
		Unpack(payload []byte) error
	}
)

var (
	_ Generator = (GenerateMessageFunc)(nil)
)

type (
	// Generator 消息生成器。
	Generator interface {
		// GenerateMessage 生成消息包结构体；
		//
		// messageType 消息类型；
		// payload 消息对应的字节数组的表示形式（不包含消息类型和长度）；
		//
		// 返回消息包。
		GenerateMessage(messageType uint16, payload []byte) (Message, error)
	}

	// GenerateMessageFunc 生成消息包结构体的方法，实现自接口 Generator ；
	//
	// messageType 消息类型；
	// payload 消息对应的字节数组的表示形式（不包含消息类型和长度）；
	//
	// 返回消息包。
	GenerateMessageFunc func(uint16, []byte) (Message, error)
)

// GenerateMessage 生成消息包结构体；
//
// messageType 消息类型；
// payload 消息对应的字节数组的表示形式（不包含消息类型和长度）；
//
// 返回消息包。
func (f GenerateMessageFunc) GenerateMessage(messageType uint16, payload []byte) (Message, error) {
	return f(messageType, payload)
}
