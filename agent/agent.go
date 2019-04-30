package agent

import (
	"fmt"
	"log"
	"net"
	"syscall"
	"unsafe"

	bp "github.com/halivor/frontend/bufferpool"
	cnf "github.com/halivor/frontend/config"
	c "github.com/halivor/frontend/connection"
	p "github.com/halivor/frontend/packet"
	ep "github.com/halivor/goevent/eventpool"
	m "github.com/halivor/goevent/middleware"
)

type Agent struct {
	addr string
	ev   ep.EP_EVENT
	buf  []byte
	pos  int

	*c.C
	ep.EventPool
	tqid m.QId
	m.Middleware
	*log.Logger
}

func New(addr string, epr ep.EventPool, mw m.Middleware) (a *Agent, e error) {
	defer func() {
		if e == nil {
			if e = a.AddEvent(a); e != nil {
				a.Println("add event failed:", e)
				return
			}
			a.tqid = a.Bind(m.T_TRANSFER, "down", m.A_PRODUCE, a)
			a.Bind(m.T_TRANSFER, "up", m.A_CONSUME, a)
		}
	}()

	C, e := c.NewTcpConn()
	if e != nil {
		return nil, e
	}
	ad, e := net.ResolveTCPAddr("tcp", addr)
	if e != nil {
		return nil, e
	}
	saddr := &syscall.SockaddrInet4{Port: ad.Port}
	copy(saddr.Addr[:], ad.IP.To4())

	if e = syscall.Connect(C.Fd(), saddr); e != nil {
		return nil, e
	}

	return &Agent{
		ev:         ep.EV_READ,
		addr:       addr,
		buf:        bp.Alloc(),
		C:          C,
		EventPool:  epr,
		Middleware: mw,
		Logger:     cnf.NewLogger(fmt.Sprintf("[agent(%d)]", C.Fd())),
	}, nil
}

func (a *Agent) CallBack(ev ep.EP_EVENT) {
	switch {
	case ev&ep.EV_READ != 0:
		n, e := syscall.Read(a.Fd(), a.buf[a.pos:])
		if e != nil || n == 0 {
			a.Release()
			return
		}
		a.pos += n
		h := (*p.Header)(unsafe.Pointer(&a.buf[0]))
		//a.Println("recv", string(a.buf[:a.pos]))
		if a.pos < p.HLen || a.pos < p.HLen+h.Len() {
			// 消息超大，增大buffer
			if a.pos == cap(a.buf) {
				buf := bp.AllocLarge(cap(a.buf) * 2)
				copy(buf, a.buf)
				a.buf = buf
			}
			// 消息接收不完整，继续接收
			return
		}
		a.process()
	case ev&ep.EV_WRITE != 0:
	case ev&ep.EV_ERROR != 0:
	default:
	}
}

func (a *Agent) process() {
	buf := a.buf
	beg := 0
	end := a.pos
	h := (*p.Header)(unsafe.Pointer(&buf[0]))
	for {
		//a.Println(h.Cmd, string(a.buf[beg+p.HLen:beg+p.HLen+h.Len()]))
		a.Produce(m.T_TRANSFER, a.tqid, buf[beg:beg+p.HLen+h.Len()])
		beg += p.HLen + h.Len()
		h = (*p.Header)(unsafe.Pointer(&buf[beg]))
		if end-beg <= p.HLen || end-beg < p.HLen+h.Len() {
			break
		}
	}
	// 不利用buffer, 防止发送端阻塞时，数据被覆盖
	a.buf = bp.Alloc()
	a.pos = 0
	if beg < end {
		copy(a.buf, buf[beg:end])
		a.pos = end - beg
	}
}

func (a *Agent) Consume(message interface{}) interface{} {
	if msg, ok := message.([]byte); ok {
		// TODO: 通过中间件发送消息
		switch h := (*p.Header)(unsafe.Pointer(&msg[0])); h.Cmd {
		case 2000:
		case 4000:
		}
	}
	return nil
}

func (a *Agent) Event() ep.EP_EVENT {
	return a.ev
}

func (a *Agent) Release() {
	a.Release()
	a.DelEvent(a)
	a.C.Close()
}
