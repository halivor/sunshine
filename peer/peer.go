package peer

import (
	"fmt"
	"syscall"

	ep "github.com/halivor/goutility/eventpool"
	log "github.com/halivor/goutility/logger"
	sc "github.com/halivor/sunshine/conf"
	c "github.com/halivor/sunshine/connection"
	up "github.com/halivor/sunshine/packet"
)

type Peer struct {
	ev     ep.EP_EVENT
	ps     peerStat
	rp     *up.P   // read  buffer
	wl     []*up.P // write buffer list
	wp     int     // previous packet write position
	header up.Header
	Manager
	c.Conn
	ep.EventPool
	log.Logger
}

func New(conn c.Conn, epr ep.EventPool, pm Manager) *Peer {
	return &Peer{
		ev: ep.EV_READ,
		ps: PS_ESTAB,
		rp: up.NewPkt(),
		wl: make([]*up.P, 0, 1024),
		wp: 0,

		Manager:   pm,
		Conn:      conn,
		EventPool: epr,
		Logger: log.NewLog("sunshine.peer.log",
			fmt.Sprintf("[p(%d)]", conn.Fd()), log.LstdFlags, sc.LogLvlPeer),
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
		p.Warn("event error", ev)
		p.Release()
	}
}

func (p *Peer) Send(pd *up.P) {
	if len(p.wl) > 0 {
		switch {
		case len(p.wl) > MAX_QUEUE_SIZE:
			p.Release()
		default:
			p.wl = append(p.wl, pd)
		}
		return
	}
	switch n, e := p.Conn.Send(pd.Buf); e {
	case nil, syscall.EAGAIN:
		if n < 0 {
			n = 0
		}
		switch {
		case n < pd.Len:
			// 本次写入，导致socket write buffer满，剩余数据send again
			// 网络正常时，不会出现此情况
			p.wl = append(p.wl, pd)
			p.wp = n

			p.ev |= ep.EV_WRITE
			p.ModEvent(p)
		default:
			p.Trace("send", string(pd.Buf))
			pd.Release()
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
				p.wl[0].Release()
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
				// 没遇到过这种情况，防止意外
				if p.wp += n; p.wp == p.wl[0].Len {
					p.wl[0].Release()
					p.wl = p.wl[1:]
					p.wp = 0
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

func (p *Peer) Event() ep.EP_EVENT {
	return p.ev
}

func (p *Peer) Release() {
	p.Debug("release fd", p.Fd(), ", uid", p.header.Uid)
	for _, ps := range p.wl {
		ps.Release()
	}
	syscall.Close(p.Fd())
	p.DelEvent(p)
	p.DelPeer(p)
	log.Release(p.Logger)
}
