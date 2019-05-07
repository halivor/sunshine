package peer

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"

	ep "github.com/halivor/goevent/eventpool"
	bp "github.com/halivor/sunshine/bufferpool"
	cnf "github.com/halivor/sunshine/config"
	c "github.com/halivor/sunshine/connection"
	pkt "github.com/halivor/sunshine/packet"
)

type Peer struct {
	ev     ep.EP_EVENT
	ps     peerStat
	rp     *pkt.P
	wl     []*pkt.P // write packet list
	wp     int      // previous packet write position
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
		wl: make([]*pkt.P, 1024),
		wp: 0,

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

func (p *Peer) Send(ps *pkt.P) {
	if len(p.wl) > 0 {
		switch {
		case len(p.wl) > 128:
			p.Release()
		default:
			p.wl = append(p.wl, ps)
		}
		return
	}
	switch n, e := p.Conn.Send(ps.Buf); {
	case e == nil || e == syscall.EAGAIN:
		if n < 0 {
			n = 0
		}
		if n < ps.Len {
			// 本次写入，导致socket write buffer满，剩余数据send again
			p.wl = append(p.wl, ps)
			p.wp += n

			p.ev |= ep.EV_WRITE
			p.ModEvent(p)
		}
	default:
		p.Release()
	}
}

func (p *Peer) sendAgain() {
	for len(p.wl) > 0 {
		switch n, e := p.Conn.Send(p.wl[0].Buf[p.wp:]); e {
		case nil:
			switch {
			case n+p.wp == p.wl[0].Len:
				// 本次写入完成后，socket write buffer未满时
				// 本次会写入Buf中全部数据，同时e=nil
				p.wl = p.wl[1:]
				p.wp = 0
			case n+p.wp < p.wl[0].Len:
				// 本次写入导致socket write buffer满时
				// 本次会写入Buf中全部/部分数据，同时e=nil
				// 在弱网时，此处应为大部分情况
				p.wp += n
				return
			}
		case syscall.EAGAIN:
			// 当本次写入时，本方buffer已满时，e=11，n=-1
			if n > 0 {
				p.wp += n
				if p.wp == p.wl[0].Len {
					p.wl = p.wl[1:]
				}
			}
			return
		default:
			p.Release()
			return
		}
	}
	p.ev = ep.EV_READ
	p.ModEvent(p)
}

func (p *Peer) SendBuffer(ps *pkt.P) {
	switch e := p.Conn.SendBuffer(ps); e {
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
	for _, ps := range p.wl {
		bp.ReleasePointer(unsafe.Pointer(ps))
	}
	syscall.Close(p.Fd())
	p.DelEvent(p)
	p.Del(p)
}
