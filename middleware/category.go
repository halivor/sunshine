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
)

// 行为
const (
	A_PRODUCE Action = 1 + iota
	A_CONSUME
)

var components map[MwId]interface{}

func init() {
	components = make(map[MwId]interface{}, 64)
}

func Register(id MwId, New interface{}) {
	components[id] = New
}
