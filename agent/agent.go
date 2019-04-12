package agent

import (
	"fmt"
	"log"
	"os"
	"syscall"

	c "github.com/halivor/frontend/connection"
	evp "github.com/halivor/frontend/eventpool"
	m "github.com/halivor/frontend/middleware"
)

type Agent struct {
	ev   uint32
	addr string
	mpid m.TypeID
	mcid m.TypeID

	*c.C
	evp.EventPool
	m.Middleware

	*log.Logger
}

func New(addr string, ep evp.EventPool, mw m.Middleware) (a *Agent) {
	defer func() {
		a.Println("add event")
		a.AddEvent(a)
		a.mpid = a.Bind("down", m.A_PRODUCER, a)
		a.mcid = a.Bind("up", m.A_CONSUMER, a)
	}()

	C := c.NewTcp()

	return &Agent{
		ev:   syscall.EPOLLIN,
		addr: addr,

		C:          C,
		EventPool:  ep,
		Middleware: mw,

		Logger: log.New(os.Stderr, fmt.Sprint("[agent(%d)]", C.Fd()), log.LstdFlags|log.Lmicroseconds),
	}
}

func (a *Agent) CallBack(ev uint32) {
}

func (a *Agent) Event() uint32 {
	return 0
}
