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
	pkt "github.com/halivor/frontend/packet"
	evp "github.com/halivor/goevent/eventpool"
	m "github.com/halivor/goevent/middleware"
)

type Agent struct {
	addr string
	ev   uint32
	buf  []byte
	pos  int

	*c.C
	evp.EventPool

	tqid m.QId
	cqid m.QId
	bqid m.QId
	m.Middleware

	*log.Logger
}

func New(addr string, ep evp.EventPool, mw m.Middleware) (a *Agent, e error) {
	defer func() {
		if e == nil {
			if e = a.AddEvent(a); e != nil {
				a.Println("add event failed:", e)
			}
			a.tqid = a.Bind(m.T_TRANSFER, "down", m.A_PRODUCE, a)
			a.cqid = a.Bind(m.T_TRANSFER, "dchat", m.A_PRODUCE, a)
			a.bqid = a.Bind(m.T_TRANSFER, "dbullet", m.A_PRODUCE, a)
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
	// TODO: 远程地址会解析失败，确认问题
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
		n, e := syscall.Read(a.Fd(), a.buf[a.pos:])
		if e != nil || n == 0 {
			a.Release()
			return
		}
		a.pos += n
		h := (*pkt.Header)(unsafe.Pointer(&a.buf[0]))
		a.Println("recv", string(a.buf[:a.pos]))
		if a.pos < pkt.HLen || a.pos < pkt.HLen+h.Len() {
			// 消息超大，增大buffer
			if a.pos == cap(a.buf) {
				buf := bp.AllocLarge(cap(a.buf) * 2)
				copy(buf, a.buf)
				a.buf = buf
			}
			// 消息接收不完整，继续接收
			return
		}
		a.Process()
	case ev&syscall.EPOLLOUT != 0:
	case ev&syscall.EPOLLERR != 0:
	default:
	}
}

func (a *Agent) Process() {
	buf := a.buf
	beg := 0
	end := a.pos
	h := (*pkt.Header)(unsafe.Pointer(&buf[0]))
	for {
		switch h.Cmd {
		case pkt.C_BULLET:
			//a.Println("bullet", string(a.buf[beg+pkt.HLen:beg+pkt.HLen+h.Len()]))
			a.Produce(m.T_TRANSFER, a.cqid, buf[beg:beg+pkt.HLen+h.Len()])
		case pkt.C_CHAT:
			//a.Println("chat", string(a.buf[beg+pkt.HLen:beg+pkt.HLen+h.Len()]))
			a.Produce(m.T_TRANSFER, a.bqid, buf[beg:beg+pkt.HLen+h.Len()])
		default:
			a.Println("default", string(buf[beg+pkt.HLen:beg+pkt.HLen+h.Len()]))
		}
		beg += pkt.HLen + h.Len()
		h = (*pkt.Header)(unsafe.Pointer(&buf[beg]))
		if end-beg <= pkt.HLen || end-beg < pkt.HLen+h.Len() {
			break
		}
	}
	a.pos = 0
	if beg != end {
		copy(a.buf, buf[beg:end])
		a.pos = end - beg
	}
}

func (a *Agent) Consume(message interface{}) interface{} {
	if msg, ok := message.([]byte); ok {
		switch h := (*pkt.Header)(unsafe.Pointer(&msg[0])); h.Cmd {
		case 2000:
			//a.Produce(m.T_CHAT, a.cqid, msg)
		case 4000:
			//a.Produce(m.T_BULLET, a.bqid, msg)
		}
	}
	return nil
}

func (a *Agent) Event() uint32 {
	return a.ev
}

func (a *Agent) Release() {
	a.DelEvent(a)
	a.C.Close()
}
