package packet

import (
	"strconv"
)

type CmdID uint32

//系统消息
const (
	C_PING CmdID = 100 + iota
	C_PONG
)

// 应用消息
const (
	C_BULLET CmdID = 2000 + iota
	C_CHAT
)

const (
	AUTH_SUCC = "10S0000000000000000700000000success"
	PING      = "10S10100123456780004RESERVEDPING"
	PONG      = "10S10100123456780004RESERVEDPONG"
)

func (id CmdID) ToString() string {
	return strconv.Itoa(int(id))
}
