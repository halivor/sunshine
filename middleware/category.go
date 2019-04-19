package middleware

type MwId int32 // 中间件类型ID
type QId int32  // 队列ID
type AId int32  // 身份ID
type Action uint32

type Mwer interface {
	Bind(q string, a Action, c interface{}) QId
	Produce(id QId, message interface{}) interface{}
	GetQId(q string) QId
}

type Consumer interface {
	Consume(m interface{}) interface{}
}

// 中间件类型
const (
	T_TRANSFER MwId = 1 << iota // 透明转发
	T_CHECK                     // 消息校验
	T_EXISTS                    // Peer ID 校验
	T_CHAT                      // 私聊消息
	T_BULLET                    // 聊天消息
)

// 行为
const (
	A_PRODUCE Action = 1 + iota
	A_CONSUME
)

type newCp func() Mwer

var components map[MwId]newCp

func init() {
	components = make(map[MwId]newCp, 64)
}

func Register(id MwId, New newCp) {
	components[id] = New
}
