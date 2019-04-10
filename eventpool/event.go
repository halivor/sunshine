package eventpool

import (
	"syscall"
)

var event map[uint32]string = map[uint32]string{
	syscall.EPOLLIN:  "event in",
	syscall.EPOLLOUT: "event out",
	syscall.EPOLLERR: "event err",
}

type Event interface {
	Fd() int
	CallBack(ev uint32)
	Event() uint32
}
