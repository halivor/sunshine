package packet

//系统消息
const (
	C_PING = 100 + iota
	C_PONG
)

// 应用消息
const (
	C_BULLET = 2000 + iota
	C_CHAT
)
