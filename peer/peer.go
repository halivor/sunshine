package peer

import (
	"fmt"
	"log"
	"syscall"

	ep "github.com/halivor/goevent/eventpool"
	cnf "github.com/halivor/sunshine/config"
	c "github.com/halivor/sunshine/connection"
	pkt "github.com/halivor/sunshine/packet"
)

type Peer struct {
	ev     ep.EP_EVENT
	ps     peerStat
	rp     *pkt.P
	header pkt.Header
	Manager
	c.Conn
	ep.EventPool
	*log.Logger
}

func New(conn c.Conn, epr ep.EventPool, pm Manager) *Peer {
	return &Peer{
		ev: ep.EV_READ,
		ps: PS_ESTAB,
		rp: pkt.NewPkt(),

		Manager:   pm,
		Conn:      conn,
		EventPool: epr,
		Logger:    cnf.NewLogger(fmt.Sprintf("[peer(%d)]", conn.Fd())),
	}
}

func (p *Peer) CallBack(ev ep.EP_EVENT) {
	switch {
	case ev&ep.EV_READ != 0:
		switch e := p.recv(); e {
		case nil, syscall.EAGAIN:
		default:
			p.Release()
		}
	case ev&ep.EV_WRITE != 0:
		p.sendAgain()
	default:
		p.Println("event error", ev)
		p.Release()
	}
}

func (p *Peer) sendAgain() {
	switch e := p.SendAgain(); e {
	case nil:
		p.ev = ep.EV_READ
		p.ModEvent(p)
	case syscall.EAGAIN:
	default: // os.ErrClosed...
		p.Release()
	}
}

func (p *Peer) Send(data []byte) {
	switch e := p.Conn.Send(data); e {
	case syscall.EAGAIN:
		p.ev |= ep.EV_WRITE
		p.ModEvent(p)
	case nil:
		return
	default:
		p.Release()
	}
}

func (p *Peer) Event() ep.EP_EVENT {
	return p.ev
}

func (p *Peer) Release() {
	syscall.Close(p.Fd())
	p.DelEvent(p)
	p.Del(p)
}
