package peer

import (
	"fmt"
	"log"
	"syscall"

	"github.com/halivor/frontend/config"
	c "github.com/halivor/frontend/connection"
	evp "github.com/halivor/frontend/eventpool"
)

type Peer struct {
	ev uint32
	rb []byte
	ps peerStat

	c.Conn
	evp.EventPool
	Manager

	*log.Logger
}

func New(conn c.Conn, ep evp.EventPool, pm Manager) *Peer {
	return &Peer{
		ev: syscall.EPOLLIN,
		rb: make([]byte, 4096),
		ps: PS_ESTAB,

		Conn:      conn,
		EventPool: ep,
		Manager:   pm.(*manager),

		Logger: config.NewLogger(fmt.Sprint("[peer(%d)]", conn.Fd())),
	}
}

func (p *Peer) CallBack(ev uint32) {
	switch {
	case ev&syscall.EPOLLIN != 0:
		n, e := syscall.Read(p.Fd(), p.rb)
		if e != nil {
			p.Println(e)
			p.DelEvent(p)
			p.Release()
			return
		}
		switch p.ps {
		case PS_ESTAB:
			p.check(p.rb[:n])
		case PS_NORMAL:
		case PS_END:
		default:
			p.Produce(p.rb[0:n])
		}
	case ev&syscall.EPOLLERR != 0:
		p.DelEvent(p)
		p.Release()
	case ev&syscall.EPOLLOUT != 0:
		if e := p.SendAgain(); e == nil {
		}
	}
}

func (p *Peer) check(message []byte) {
}

func (p *Peer) Event() uint32 {
	return p.ev
}

func (p *Peer) Release() {
}

func (p *Peer) Send(data []byte) {
	switch e := p.Conn.Send(data); e {
	case syscall.EAGAIN:
		p.ev |= syscall.EPOLLOUT
		p.ModEvent(p)
	}
}
