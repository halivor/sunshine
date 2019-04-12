package middleware

var id TypeID = 't'<<16 | 'i'<<8 | 'd'
var tId map[string]TypeID

type TypeID uint32
type Action uint32

// 消息类型
//		1. transfer
//		2. verify
const (
	A_PRODUCER Action = 1
	A_CONSUMER Action = 2

	MAX_CONSUMER = 32
)

type Consume interface {
	Consume(message interface{})
}

func TId(t string) TypeID {
	if id, ok := tId[t]; ok {
		return id
	}
	id++
	tId[t] = id
	return id
}

func (tid TypeID) String() string {
	return ""
}
