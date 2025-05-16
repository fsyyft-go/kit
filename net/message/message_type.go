package message

const (
	HeartbeatMessageType    uint16 = 0x80 // 心跳消息。
	SingleStringMessageType uint16 = 0x09 // 简单的字符串消息。
)

func init() {
	_ = FactoryRegister(HeartbeatMessageType, GenerateHeartbeatMessage)
	_ = FactoryRegister(SingleStringMessageType, GenerateSingleStringMessage)
}
