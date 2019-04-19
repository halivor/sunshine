package acceptor

import (
	"fmt"
	"log"
	"syscall"

	"github.com/halivor/frontend/config"
	c "github.com/halivor/frontend/connection"
	e "github.com/halivor/frontend/eventpool"
	m "github.com/halivor/frontend/middleware"
	p "github.com/halivor/frontend/peer"
)

type Acceptor struct {
	ev   uint32
	addr string

	*c.C
	e.EventPool // event: add, del, mod
	p.Manager

	*log.Logger
}

func NewTcpAcceptor(addr string, ep e.EventPool, mw m.Middleware) (a *Acceptor, e error) {
	defer func() {
		a.AddEvent(a)
	}()

	C, e := c.NewTcp()
	if e != nil {
		return nil, e
	}
	saddr, e := c.ParseAddr4("tcp", addr)
	if e != nil {
		return nil, e
	}
	if e = syscall.Bind(C.Fd(), saddr); e != nil {
		return nil, e
	}

	if e = syscall.Listen(C.Fd(), 1024); e != nil {
		return nil, e
	}

	a = &Acceptor{
		ev:        syscall.EPOLLIN,
		addr:      addr,
		C:         C,
		EventPool: ep,
		Manager:   p.NewManager(mw),
		Logger:    config.NewLogger(fmt.Sprintf("[accept(%d)] ", C.Fd())),
	}

	a.Println("non block", c.NonBlock(C.Fd()))
	a.Println("reuse addr port", c.Reuse(C.Fd(), true))
	return a, nil
}

// TODO: 细化异常处理流程
func (a *Acceptor) CallBack(ev uint32) {
	if ev&syscall.EPOLLERR != 0 {
		a.Println("epoll error", ev)
		a.DelEvent(a)
		return
	}
	switch fd, _, e := syscall.Accept(a.Fd()); e {
	case syscall.EAGAIN, syscall.EINTR:
	case nil:
		a.Println("accept connection", fd)
		a.AddEvent(p.New(c.NewSock(fd), a.EventPool, a.Manager))
	default:
		a.DelEvent(a)
		// TODO: 释放并重启acceptor
	}
}

func (a *Acceptor) Event() uint32 {
	return a.ev
}

func (a *Acceptor) Release() {
}
