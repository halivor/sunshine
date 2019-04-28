package peer

import (
	"fmt"
	"log"
	"syscall"

	bp "github.com/halivor/frontend/bufferpool"
	cnf "github.com/halivor/frontend/config"
	c "github.com/halivor/frontend/connection"
	pkt "github.com/halivor/frontend/packet"
	evp "github.com/halivor/goevent/eventpool"
)

type Peer struct {
	rb  []byte
	pos int
	pkt []byte
	ev  uint32
	ps  peerStat

	pkt.Header
	Manager
	c.Conn
	evp.EventPool
	*log.Logger
}

func New(conn c.Conn, ep evp.EventPool, pm Manager) (p *Peer) {
	p = &Peer{
		ev:  syscall.EPOLLIN,
		rb:  bp.Alloc(),
		pos: pkt.HLen,
		ps:  PS_ESTAB,

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
		switch e := p.recv(); e {
		case nil, syscall.EAGAIN:
		default:
			p.Release()
		}
	case ev&syscall.EPOLLOUT != 0:
		p.sendAgain()
	default:
		p.Println("event error", ev)
		p.Release()
	}
}

func (p *Peer) sendAgain() {
	switch e := p.SendAgain(); e {
	case nil:
		p.ev = syscall.EPOLLIN
		p.ModEvent(p)
	case syscall.EAGAIN:
	default: // os.ErrClosed...
		p.Release()
	}
}

func (p *Peer) Send(data []byte) {
	switch e := p.Conn.Send(data); e {
	case syscall.EAGAIN:
		p.ev |= syscall.EPOLLOUT
		p.ModEvent(p)
	case nil:
		return
	default:
		p.Release()
	}
}

func (p *Peer) Event() uint32 {
	return p.ev
}

func (p *Peer) Release() {
	syscall.Close(p.Fd())
	p.DelEvent(p)
	p.Del(p)
}
