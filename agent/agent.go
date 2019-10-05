package agent

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"

	bp "github.com/halivor/goutility/bufferpool"
	ep "github.com/halivor/goutility/eventpool"
	log "github.com/halivor/goutility/logger"
	m "github.com/halivor/goutility/middleware"
	sc "github.com/halivor/sunshine/conf"
	c "github.com/halivor/sunshine/connection"
	p "github.com/halivor/sunshine/packet"
)

type Agent struct {
	addr string
	ev   ep.EP_EVENT
	buf  []byte
	pos  int

	mwTfs m.MwId
	tqid  m.QId

	c.Conn
	ep.EventPool
	m.Middleware
	log.Logger
}

func New(addr string, epr ep.EventPool, mw m.Middleware) *Agent {
	C, e := c.NewTcpConn()
	if e != nil {
		panic(e)
	}
	ad, e := net.ResolveTCPAddr("tcp", addr)
	if e != nil {
		panic(e)
	}
	saddr := &syscall.SockaddrInet4{Port: ad.Port}
	copy(saddr.Addr[:], ad.IP.To4())

	if e = syscall.Connect(C.Fd(), saddr); e != nil {
		panic(e)
	}

	a := &Agent{
		ev:         ep.EV_READ,
		addr:       addr,
		buf:        bp.Alloc(4096 - 96),
		Conn:       C,
		EventPool:  epr,
		Middleware: mw,
		Logger: log.NewLog("sunshine.log",
			fmt.Sprintf("[agent(%d)]", C.Fd()), log.LstdFlags, sc.LogLvlAgent),
	}

	if e = a.AddEvent(a); e != nil {
		panic(e)
	}

	return a
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

	// TODO: 如果是只有一个包，直接发送，不做copy
	//a.Trace("parse", a.pos, len(a.buf), cap(a.buf))
	for a.pos-beg > p.HLen && a.pos-beg >= p.HLen+h.Len() {
		//a.Trace(h.Cmd, string(a.buf[beg+p.HLen:beg+p.HLen+h.Len()]))
		pd := p.Alloc(p.HLen + h.Len())
		copy(pd.Buf, a.buf[beg:beg+p.HLen+h.Len()])
		a.Trace("parse", p.HLen+h.Len(), string(pd.Buf[p.HLen:]))
		a.Produce(a.mwTfs, a.tqid, pd)
		//pd.Release()

		beg += p.HLen + h.Len()
		h = (*p.Header)(unsafe.Pointer(&a.buf[beg]))
	}
	// 不利用buffer, 防止发送端阻塞时，数据被覆盖
	if a.pos-beg > 0 {
		copy(a.buf, a.buf[beg:a.pos])
	}
	a.pos -= beg
}

func (a *Agent) BindMsg(id m.MwId) {
	a.mwTfs = id
	a.tqid = a.Bind(id, "down", m.A_PRODUCE, a)
	a.Bind(id, "up", m.A_CONSUME, a)
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
	bp.Free(a.buf)
	a.DelEvent(a)
	a.Conn.Close()
}
