package agent

import (
	"fmt"
	"log"
	"net"
	"syscall"

	bp "github.com/halivor/frontend/bufferpool"
	cnf "github.com/halivor/frontend/config"
	c "github.com/halivor/frontend/connection"
	evp "github.com/halivor/frontend/eventpool"
	m "github.com/halivor/frontend/middleware"
)

type Agent struct {
	addr string
	ev   uint32
	buf  []byte

	*c.C
	evp.EventPool

	pqid m.QId
	cqid m.QId
	m.Middleware

	*log.Logger
}

func New(addr string, ep evp.EventPool, mw m.Middleware) (a *Agent, e error) {
	defer func() {
		if e == nil {
			if e := a.AddEvent(a); e != nil {
				a.Println("add event failed:", e)
			}
			a.pqid = a.Bind(m.T_TRANSFER, "down", m.A_PRODUCE, a)
			a.cqid = a.Bind(m.T_TRANSFER, "up", m.A_CONSUME, a)
		}
	}()

	C, e := c.NewTcp()
	if e != nil {
		return nil, e
	}
	ad, e := net.ResolveTCPAddr("tcp", addr)
	if e != nil {
		return nil, e
	}
	saddr := &syscall.SockaddrInet4{Port: ad.Port}
	copy(saddr.Addr[:], ad.IP[0:4])

	if e = syscall.Connect(C.Fd(), saddr); e != nil {
		return nil, e
	}

	return &Agent{
		ev:         syscall.EPOLLIN,
		addr:       addr,
		buf:        bp.Alloc(),
		C:          C,
		EventPool:  ep,
		Middleware: mw,
		Logger:     cnf.NewLogger(fmt.Sprintf("[agent(%d)]", C.Fd())),
	}, nil
}

func (a *Agent) CallBack(ev uint32) {
	switch {
	case ev&syscall.EPOLLIN != 0:
		n, e := syscall.Read(a.Fd(), a.buf)
		if e != nil {
			a.Release()
		}
		a.Println("produce", string(a.buf[:n]))
		a.Produce(m.T_TRANSFER, a.pqid, a.buf[:n])
	case ev&syscall.EPOLLOUT != 0:
	case ev&syscall.EPOLLERR != 0:
	default:
	}
}

func (a *Agent) Event() uint32 {
	return a.ev
}

func (a *Agent) Release() {
}
