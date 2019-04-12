package eventpool

import (
	"syscall"
)

type TEVENT uint32

const (
	EVENT_READ  TEVENT = syscall.EPOLLIN
	EVENT_WRITE TEVENT = syscall.EPOLLOUT
	EVENT_ERROR TEVENT = syscall.EPOLLERR
)

var event map[TEVENT]string = map[TEVENT]string{
	EVENT_READ:  "event in",
	EVENT_WRITE: "event out",
	EVENT_ERROR: "event err",
}

var es map[uint32]TEVENT = map[uint32]TEVENT{
	syscall.EPOLLIN:  EVENT_READ,
	syscall.EPOLLOUT: EVENT_WRITE,
	syscall.EPOLLERR: EVENT_ERROR,
}

type Eventer interface {
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
