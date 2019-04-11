package middleware

type Category uint32

const (
	C_PEER Category = 1 << iota
	C_AGENT
)

type Broadcast interface {
	Broadcast(message interface{})
}

type Unicast interface {
	Unicast(id uint64, message interface{})
}

func (c Category) check(i interface{}) bool {
	switch i.(type) {
	case Broadcast, Unicast:
		return true
	default:
		return false
	}
}
