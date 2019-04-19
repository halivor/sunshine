package eventpool

import (
	"syscall"
)

type EP_EVENT int32

const (
	EVENT_READ    EP_EVENT = syscall.EPOLLIN
	EVENT_WRITE   EP_EVENT = syscall.EPOLLOUT
	EVENT_ERROR   EP_EVENT = syscall.EPOLLERR
	EVENT_EPOLLET EP_EVENT = syscall.EPOLLET
	EVENT_END     EP_EVENT = -1
)

var event map[EP_EVENT]string = map[EP_EVENT]string{
	EVENT_READ:    "ev in",
	EVENT_WRITE:   "ev out",
	EVENT_ERROR:   "ev err",
	EVENT_EPOLLET: "ev et",
}

var es map[int32]EP_EVENT = map[int32]EP_EVENT{
	syscall.EPOLLIN:  EVENT_READ,
	syscall.EPOLLOUT: EVENT_WRITE,
	syscall.EPOLLERR: EVENT_ERROR,
	syscall.EPOLLET:  EVENT_EPOLLET,
	-1:               EVENT_END,
}

type Event interface {
	Fd() int
	CallBack(ev uint32)
	Event() uint32
	Release()
}

func (t EP_EVENT) String() string {
	if s, ok := event[t]; ok {
		return s
	}
	return "no such event"
}
