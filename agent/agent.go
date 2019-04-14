package agent

import (
	"fmt"
	"log"
	"net"
	"os"
	"syscall"

	c "github.com/halivor/frontend/connection"
	evp "github.com/halivor/frontend/eventpool"
	m "github.com/halivor/frontend/middleware"
)

type Agent struct {
	addr string
	ev   uint32
	*c.C
	evp.EventPool

	pmid m.QId
	cmid m.QId
	m.Middleware

	*log.Logger
}

func New(addr string, ep evp.EventPool, mw m.Middleware) (a *Agent, e error) {
	defer func() {
		a.Println("add event")
		a.AddEvent(a)
		a.pmid = a.Bind(m.T_TRANSFER, "down", m.A_PRODUCE, a)
		a.cmid = a.Bind(m.T_TRANSFER, "up", m.A_CONSUME, a)
	}()

	C := c.NewTcp()
	ad, e := net.ResolveTCPAddr("tcp", addr)
	if e != nil {
		return nil, e
	}
	saddr := &syscall.SockaddrInet4{Port: ad.Port}
	copy(saddr.Addr[:], ad.IP[0:4])
	/*if e := syscall.Bind(C.Fd(), saddr); e != nil {*/
	//return nil, e
	/*}*/

	/*if e := syscall.Listen(C.Fd(), 1024); e != nil {*/
	//return nil, e
	/*}*/

	return &Agent{
		ev:   syscall.EPOLLIN,
		addr: addr,

		C:          C,
		EventPool:  ep,
		Middleware: mw,

		Logger: log.New(os.Stderr, fmt.Sprint("[agent(%d)]", C.Fd()), log.LstdFlags|log.Lmicroseconds),
	}, nil
}

func (a *Agent) CallBack(ev uint32) {
}

func (a *Agent) Event() uint32 {
	return 0
}
