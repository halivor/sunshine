package listener

import (
	"fmt"
	"log"
	"net"
	"os"
	"syscall"

	clnt "frontend/client"
	c "frontend/connection"
	evp "frontend/eventpool"
)

type Listener struct {
	ev   uint32
	addr string

	*c.C
	evp.EventPool

	*log.Logger
}

func NewTcpListeners(addrs []string, ep evp.EventPool) []*Listener {
	l := make([]*Listener, 0)
	for _, addr := range addrs {
		l = append(l, NewTcpListener(addr, ep))
	}
	return l
}

func NewTcpListener(addr string, ep evp.EventPool) (l *Listener) {
	defer func() {
		l.Println("add event")
		l.AddEvent(l)
	}()
	c := c.NewTcp()
	a, e := net.ResolveTCPAddr("tcp", addr)
	if e != nil {
		log.Panicln(e.Error())
	}
	saddr := &syscall.SockaddrInet4{Port: a.Port}
	copy(saddr.Addr[:], a.IP[0:4])
	if err := syscall.Bind(c.Fd(), saddr); err != nil {
		log.Panicln(err.Error())
	}

	if err := syscall.Listen(c.Fd(), 1024); err != nil {
		log.Panicln(err.Error())
	}
	return &Listener{
		C:         c,
		ev:        syscall.EPOLLIN,
		addr:      addr,
		Logger:    log.New(os.Stderr, fmt.Sprintf("[lsn(%d)] ", c.Fd()), log.LstdFlags|log.Lmicroseconds),
		EventPool: ep,
	}
}

func (l *Listener) CallBack(ev uint32) {
	if ev&syscall.EPOLLERR != 0 {
		l.Println("epoll error", ev)
		l.DelEvent(l)
		return
	}
	fd, _, e := syscall.Accept(l.Fd())
	l.Println("accept connection", fd)
	if e == nil {
		l.AddEvent(clnt.New(c.NewSock(fd), l.EventPool))
		return
	}
	l.Println(e)
}

func (l *Listener) Event() uint32 {
	return l.ev
}
