package peer

import (
	"fmt"
	"log"
	"os"
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
	rb  []byte
	pos int
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
	for {
		switch {
		case ev&syscall.EPOLLIN != 0:
			if n, e := syscall.Read(p.Fd(), p.rb[p.pos:]); e != nil || n == 0 {
				p.pos += n
				// 头长度不够，继续读取
				if p.pos < pkt.HLen+pkt.UHLen {
					return
				}
				switch e {
				case syscall.EAGAIN:
					return
				default:
					p.Release()
					return
				}
			}
			p.Do()
		case ev&syscall.EPOLLERR != 0:
			p.Release()
		case ev&syscall.EPOLLOUT != 0:
			if e := p.SendAgain(); e == nil {
				p.ev = syscall.EPOLLIN
				p.ModEvent(p)
			} else if e == os.ErrClosed {
				p.Release()
			}
		}
	}
}

func (p *Peer) Do() error {
	switch p.ps {
	case PS_ESTAB:
		if e := p.check(); e != nil {
			return e
		}
		// 用户信息结构
		p.Manager.Add(p)
		for {
			packet, e := p.Parse()
			if e != nil {
				return e
			}
			p.Manager.Transfer(packet)
		}
	case PS_NORMAL:
		for {
			packet, e := p.Parse()
			if e != nil {
				return e
			}
			p.Manager.Transfer(packet)
		}
	}
	return nil
}

func (p *Peer) check() error {
	h := (*pkt.Header)(unsafe.Pointer(&p.rb[0]))
	uh := (*pkt.UHeader)(unsafe.Pointer(&p.rb[pkt.HLen]))
	h.Ver = cnf.VER
	h.Nid = cnf.NodeId
	h.Uid = uh.Uid
	h.Cid = uh.Cid
	p.ps = PS_NORMAL
	p.Header = *(*pkt.Header)(unsafe.Pointer(&p.rb[0]))
	// verify failed os.ErrInvalid
	return nil
}

func (p *Peer) Parse() ([]byte, error) {
	uh := (*pkt.UHeader)(unsafe.Pointer(&p.rb[pkt.HLen]))
	if uh.Len > 4*1024 {
		return nil, os.ErrInvalid
	}
	if pkt.HLen+pkt.UHLen+int(uh.Len) < p.pos {
		return nil, syscall.EAGAIN
	}
	packet := p.rb[:pkt.HLen+pkt.UHLen+int(uh.Len)]
	nb := bp.Alloc()

	if p.pos-pkt.HLen+pkt.UHLen+int(uh.Len) > 0 {
		copy(nb[pkt.HLen:], p.rb[pkt.HLen+pkt.UHLen+int(uh.Len):p.pos])
	}

	p.rb = nb
	*(*pkt.Header)(unsafe.Pointer(&p.rb[0])) = p.Header
	p.pos = pkt.HLen
	return packet, nil
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
