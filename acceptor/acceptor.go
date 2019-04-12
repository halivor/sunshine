package acceptor

import (
	"fmt"
	"log"
	"net"
	"os"
	"syscall"

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

func NewTcpAcceptor(addr string, ep e.EventPool, mw m.Middleware) (a *Acceptor) {
	defer func() {
		a.Println("add event")
		a.AddEvent(a)
	}()
	c := c.NewTcp()
	ad, e := net.ResolveTCPAddr("tcp", addr)
	if e != nil {
		log.Panicln(e.Error())
	}
	saddr := &syscall.SockaddrInet4{Port: ad.Port}
	copy(saddr.Addr[:], ad.IP[0:4])
	if err := syscall.Bind(c.Fd(), saddr); err != nil {
		log.Panicln(err.Error())
	}

	if err := syscall.Listen(c.Fd(), 1024); err != nil {
		log.Panicln(err.Error())
	}
	return &Acceptor{
		ev:        syscall.EPOLLIN,
		addr:      addr,
		C:         c,
		EventPool: ep,
		Manager:   p.NewManager(mw),
		Logger:    log.New(os.Stderr, fmt.Sprintf("[lsn(%d)] ", c.Fd()), log.LstdFlags|log.Lmicroseconds),
	}
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
		a.Println(e)
		a.DelEvent(a)
		// TODO: 释放并重启acceptor
	}
}

func (a *Acceptor) Event() uint32 {
	return a.ev
}
