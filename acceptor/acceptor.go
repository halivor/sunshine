package acceptor

import (
	"fmt"
	"log"
	"syscall"

	"github.com/halivor/frontend/config"
	c "github.com/halivor/frontend/connection"
	p "github.com/halivor/frontend/peer"
	ep "github.com/halivor/goevent/eventpool"
	m "github.com/halivor/goevent/middleware"
)

type Acceptor struct {
	ev   ep.EP_EVENT
	addr string

	*c.C
	ep.EventPool // event: add, del, mod
	p.Manager

	*log.Logger
}

func NewTcpAcceptor(addr string, epr ep.EventPool, mw m.Middleware) (a *Acceptor, e error) {
	defer func() {
		a.AddEvent(a)
	}()

	C, e := c.NewTcpConn()
	if e != nil {
		return nil, e
	}
	saddr, e := c.ParseAddr4("tcp", addr)
	if e != nil {
		return nil, e
	}
	log.Println("reuse addr port", c.Reuse(C.Fd(), true))
	if e = syscall.Bind(C.Fd(), saddr); e != nil {
		return nil, e
	}
	if e = syscall.Listen(C.Fd(), 1024); e != nil {
		return nil, e
	}

	a = &Acceptor{
		ev:        ep.EV_READ,
		addr:      addr,
		C:         C,
		EventPool: epr,
		Manager:   p.NewManager(mw),
		Logger:    config.NewLogger(fmt.Sprintf("[accept(%d)] ", C.Fd())),
	}

	return a, nil
}

// TODO: 细化异常处理流程
func (a *Acceptor) CallBack(ev ep.EP_EVENT) {
	if ev&ep.EV_ERROR != 0 {
		a.Println("epoll error", ev)
		a.DelEvent(a)
		return
	}
	switch fd, _, e := syscall.Accept4(a.Fd(), syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC); e {
	case syscall.EAGAIN, syscall.EINTR:
	case nil:
		a.Println("accept connection", fd)
		a.AddEvent(p.New(c.NewConn(fd), a.EventPool, a.Manager))
	default:
		a.DelEvent(a)
		// TODO: 释放并重启acceptor
	}
}
func (a *Acceptor) Event() ep.EP_EVENT {
	return a.ev
}
func (a *Acceptor) Release() {
	a.C.Close()
}
