package middleware

type TypeID int32
type CategoryID int32

var id TypeID = 't'<<16 | 'i'<<8 | 'd'
var tId map[string]TypeID

// 消息类型
//		1. transfer
//		2. verify
const (
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
