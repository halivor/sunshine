package peer

import (
	"fmt"
	"syscall"
	"unsafe"

	bp "github.com/halivor/goutility/bufferpool"
	ep "github.com/halivor/goutility/eventpool"
	log "github.com/halivor/goutility/logger"
	c "github.com/halivor/sunshine/connection"
	pkt "github.com/halivor/sunshine/packet"
)

type Peer struct {
	ev     ep.EP_EVENT
	ps     peerStat
	rp     *pkt.P   // read  buffer
	wl     []*pkt.P // write buffer list
	wp     int      // previous packet write position
	header pkt.Header
	Manager
	c.Conn
	ep.EventPool
	log.Logger
}

func New(conn c.Conn, epr ep.EventPool, pm Manager) *Peer {
	logger, _ := log.New("/data/logs/sunshine/peer.log", fmt.Sprintf("[peer(%d)]", conn.Fd()), log.LstdFlags, log.TRACE)
	return &Peer{
		ev: ep.EV_READ,
		ps: PS_ESTAB,
		rp: pkt.NewPkt(),
		wl: make([]*pkt.P, 0, 1024),
		wp: 0,

		Manager:   pm,
		Conn:      conn,
		EventPool: epr,
		Logger:    logger,
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

func (p *Peer) Send(pd *pkt.P) {
	if len(p.wl) > 0 {
		switch {
		case len(p.wl) > 128:
			p.Release()
		default:
			p.wl = append(p.wl, pd)
		}
		return
	}
	switch n, e := p.Conn.Send(pd.Buf[:pd.Len]); {
	case e == nil || e == syscall.EAGAIN:
		if n < 0 {
			n = 0
		}
		if n < pd.Len {
			// 本次写入，导致socket write buffer满，剩余数据send again
			// 网络正常时，不会出现此情况
			p.wl = append(p.wl, pd)
			p.wp += n

			p.ev |= ep.EV_WRITE
			p.ModEvent(p)
		} else {
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
				p.wp += n
				if p.wp == p.wl[0].Len {
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
	for _, ps := range p.wl {
		bp.ReleasePointer(unsafe.Pointer(ps))
	}
	syscall.Close(p.Fd())
	p.DelEvent(p)
	p.Del(p)
}
