package agent

import (
	"fmt"
	"log"
	"net"
	"syscall"
	"unsafe"

	ep "github.com/halivor/goevent/eventpool"
	m "github.com/halivor/goevent/middleware"
	bp "github.com/halivor/goutility/bufferpool"
	cnf "github.com/halivor/sunshine/config"
	c "github.com/halivor/sunshine/connection"
	p "github.com/halivor/sunshine/packet"
)

type Agent struct {
	addr string
	ev   ep.EP_EVENT
	buf  []byte
	pos  int

	c.Conn
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

	buf := bp.Alloc(2048)
	return &Agent{
		ev:         ep.EV_READ,
		addr:       addr,
		buf:        buf,
		Conn:       C,
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
		if a.pos < p.HLen || a.pos < p.HLen+h.Len() {
			// 消息超大，增大buffer
			if a.pos == len(a.buf) {
				a.buf = bp.Realloc(a.buf, len(a.buf)*2)
			}
			// 消息接收不完整，继续接收
			return
		}
		a.parse()
	case ev&ep.EV_WRITE != 0:
	case ev&ep.EV_ERROR != 0:
	default:
	}
}

func (a *Agent) parse() {
	beg := 0
	h := (*p.Header)(unsafe.Pointer(&a.buf[beg]))

	//a.Println("parse", a.pos, len(a.buf), cap(a.buf))
	for a.pos-beg > p.HLen && a.pos-beg >= p.HLen+h.Len() {
		//a.Println(h.Cmd, string(a.buf[beg+p.HLen:beg+p.HLen+h.Len()]))
		pd := p.Alloc(p.HLen + h.Len())
		//a.Println("packet", p.HLen+h.Len(), beg, beg+p.HLen+h.Len())
		copy(pd.Buf, a.buf[beg:beg+p.HLen+h.Len()])
		a.Produce(m.T_TRANSFER, a.tqid, pd)

		beg += p.HLen + h.Len()
		h = (*p.Header)(unsafe.Pointer(&a.buf[beg]))
	}
	// 不利用buffer, 防止发送端阻塞时，数据被覆盖
	if a.pos-beg > 0 {
		copy(a.buf, a.buf[beg:a.pos])
	}
	a.pos -= beg
}

func (a *Agent) Consume(message interface{}) interface{} {
	if msg, ok := message.([]byte); ok {
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
	bp.Release(a.buf)
	a.DelEvent(a)
	a.Conn.Close()
}
