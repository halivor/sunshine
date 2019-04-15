package eventpool

import (
	"syscall"
)

type TEVENT int32

const (
	EVENT_READ    TEVENT = syscall.EPOLLIN
	EVENT_WRITE   TEVENT = syscall.EPOLLOUT
	EVENT_ERROR   TEVENT = syscall.EPOLLERR
	EVENT_EPOLLET TEVENT = syscall.EPOLLET
)

var event map[TEVENT]string = map[TEVENT]string{
	EVENT_READ:    "ev in",
	EVENT_WRITE:   "ev out",
	EVENT_ERROR:   "ev err",
	EVENT_EPOLLET: "ev et",
}

var es map[uint32]TEVENT = map[uint32]TEVENT{
	syscall.EPOLLIN:  EVENT_READ,
	syscall.EPOLLOUT: EVENT_WRITE,
	syscall.EPOLLERR: EVENT_ERROR,
}

type Event interface {
	Fd() int
	CallBack(ev uint32)
	Event() uint32
}

func (t TEVENT) String() string {
	if s, ok := event[t]; ok {
		return s
	}
	return "no such event"
}
