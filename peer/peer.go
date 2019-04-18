package peer

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"

	bp "github.com/halivor/frontend/bufferpool"
	cnf "github.com/halivor/frontend/config"
	c "github.com/halivor/frontend/connection"
	evp "github.com/halivor/frontend/eventpool"
	pkt "github.com/halivor/frontend/packet"
)

type uinfo struct {
	id   uint64
	room uint32
	ttl  uint32
}

type Peer struct {
	rb []byte
	ev uint32
	ps peerStat

	*pkt.Header
	Manager
	c.Conn
	evp.EventPool
	*log.Logger
}

func New(conn c.Conn, ep evp.EventPool, pm Manager) (p *Peer) {
	p = &Peer{
		ev: syscall.EPOLLIN,
		rb: bp.Alloc(),
		ps: PS_ESTAB,

		Manager:   pm,
		Conn:      conn,
		EventPool: ep,
		Logger:    cnf.NewLogger(fmt.Sprintf("[peer(%d)]", conn.Fd())),
	}
	return
}

func (p *Peer) CallBack(ev uint32) {
	switch {
	case ev&syscall.EPOLLIN != 0:
		n, e := syscall.Read(p.Fd(), p.rb[pkt.HLen:])
		if e != nil || n == 0 {
			switch e {
			case syscall.EAGAIN:
				return
			default:
				p.Release()
				return
			}
		}
		switch p.ps {
		case PS_ESTAB:
			if p.check(p.rb[:pkt.HLen+n]) {
				// 用户信息结构
				p.Manager.Add(p)
				p.Manager.Transfer(p.rb[0:n])
			} else {
				p.Release()
			}
		case PS_NORMAL:
			p.Manager.Transfer(p.rb[0:n])
		case PS_END:
		default:
		}
	case ev&syscall.EPOLLERR != 0:
		p.Release()
	case ev&syscall.EPOLLOUT != 0:
		if e := p.SendAgain(); e == nil {
			p.ev = syscall.EPOLLIN
			p.ModEvent(p)
		}
	}
}

func (p *Peer) check(message []byte) bool {
	p.Println(string(message))
	// TODO: parse message
	h := (*pkt.Header)(unsafe.Pointer(&message[0]))
	uh := (*pkt.UHeader)(unsafe.Pointer(&message[pkt.HLen]))
	h.Ver = cnf.VER
	h.Nid = cnf.NodeId
	h.Uid = uh.Uid
	h.Cid = uh.Cid
	p.ps = PS_NORMAL
	return true
}

func (p *Peer) Event() uint32 {
	return p.ev
}

func (p *Peer) Release() {
	syscall.Close(p.Fd())
	p.DelEvent(p)
}

func (p *Peer) Send(data []byte) {
	switch e := p.Conn.Send(data); e {
	case syscall.EAGAIN:
		p.ev |= syscall.EPOLLOUT
		p.ModEvent(p)
	}
}
