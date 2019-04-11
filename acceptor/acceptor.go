package acceptor

import (
	"fmt"
	"log"
	"net"
	"os"
	"syscall"

	c "github.com/halivor/frontend/connection"
	evp "github.com/halivor/frontend/eventpool"
	m "github.com/halivor/frontend/middleware"
	p "github.com/halivor/frontend/peer"
)

type Acceptor struct {
	ev   uint32
	addr string

	*c.C
	evp.EventPool // event: add, del, mod
	p.Manage      // event: unicast, broadcast
	m.Middleware

	*log.Logger
}

func NewTcpAcceptors(addrs []string, ep evp.EventPool, mw m.Middleware) []*Acceptor {
	a := make([]*Acceptor, 0)
	for _, addr := range addrs {
		a = append(a, NewTcpAcceptor(addr, ep, mw))
	}
	return a
}

func NewTcpAcceptor(addr string, ep evp.EventPool, mw m.Middleware) (a *Acceptor) {
	defer func() {
		a.Println("add event")
		a.AddEvent(a)
		a.Register(m.C_PEER, a)
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
		ev:   syscall.EPOLLIN,
		addr: addr,

		C:         c,
		EventPool: ep,
		Manage:    p.NewManager(),

		Logger: log.New(os.Stderr, fmt.Sprintf("[lsn(%d)] ", c.Fd()), log.LstdFlags|log.Lmicroseconds),
	}
}

func (a *Acceptor) CallBack(ev uint32) {
	if ev&syscall.EPOLLERR != 0 {
		a.Println("epoll error", ev)
		a.DelEvent(a)
		return
	}
	fd, _, e := syscall.Accept(a.Fd())
	a.Println("accept connection", fd)
	if e == nil {
		a.AddEvent(p.New(c.NewSock(fd), a.EventPool, a.Manage))
		return
	}
	a.Println(e)
}

func (a *Acceptor) Event() uint32 {
	return a.ev
}
