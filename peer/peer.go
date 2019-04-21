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
			p.Println("callback------", p.pos)
			n, e := syscall.Read(p.Fd(), p.rb[p.pos:])
			switch e {
			case nil:
				if n == 0 {
					p.Release()
					return
				}
				p.pos += n
				if e := p.Process(); e != nil {
					p.Release()
				}
			case syscall.EAGAIN:
				return
			default:
				p.Release()
				return
			}
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

func (p *Peer) Process() (e error) {
	switch p.ps {
	case PS_ESTAB:
		p.Println("estab", p.pos, string(p.rb[pkt.HLen:p.pos]))
		// 头长度不够，继续读取
		if p.pos < pkt.HLen+pkt.ALen {
			p.Println("not enough")
			return
		}
		ul, e := p.check()
		if e != nil {
			return e
		}
		// 用户信息结构
		p.Manager.Add(p)
		// 转发packet消息
		packet := p.rb[:pkt.HLen+pkt.ALen+ul]

		nb := bp.Alloc()
		rl := p.pos - len(packet)
		if rl > 0 {
			copy(nb[pkt.HLen:], p.rb[len(packet):p.pos])
		}

		p.rb = nb
		*(*pkt.Header)(unsafe.Pointer(&p.rb[0])) = p.Header
		p.pos = pkt.HLen + rl
		p.Transfer(packet)
	case PS_NORMAL:
		for {
			p.Println("normal", p.pos, string(p.rb[pkt.HLen:p.pos]))
			if p.pos < pkt.HLen+pkt.SHLen {
				break
			}
			packet, e := p.Parse()
			if e != nil {
				return e
			}
			p.Transfer(packet)
		}
	}

	p.Println("done", p.pos)
	return nil
}

func (p *Peer) check() (len int, e error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("check =>", r)
			e = os.ErrInvalid
		}
	}()
	h := (*pkt.Header)(unsafe.Pointer(&p.rb[0]))
	a := (*pkt.Auth)(unsafe.Pointer(&p.rb[pkt.HLen]))
	h.Ver = uint16(a.Ver())
	h.Nid = cnf.NodeId
	h.Uid = uint32(a.Uid())
	h.Cid = uint32(a.Cid())
	p.ps = PS_NORMAL
	p.Header = *(*pkt.Header)(unsafe.Pointer(&p.rb[0]))
	// verify failed os.ErrInvalid
	p.Println("checked", h, a)
	return a.Len(), nil
}

func (p *Peer) Parse() (packet []byte, e error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("parse =>", r)
			e = os.ErrInvalid
		}
	}()
	h := (*pkt.SHeader)(unsafe.Pointer(&p.rb[pkt.HLen]))
	if h.Len() > 4*1024 {
		return nil, os.ErrInvalid
	}
	p.Println("parse", h.Len())
	plen := pkt.HLen + pkt.SHLen + h.Len()
	if plen > p.pos {
		return nil, nil
	}
	packet = p.rb[:plen]
	nb := bp.Alloc()

	rl := p.pos - plen
	if rl > 0 {
		copy(nb[pkt.HLen:], p.rb[plen:p.pos])
	}

	p.rb = nb
	*(*pkt.Header)(unsafe.Pointer(&p.rb[0])) = p.Header
	p.pos = pkt.HLen + rl
	p.Println("parse end", p.pos)
	return packet, nil
}

func (p *Peer) Send(data []byte) {
	switch e := p.Conn.Send(data); e {
	case syscall.EAGAIN:
		p.ev |= syscall.EPOLLOUT
		p.ModEvent(p)
	}
}

func (p *Peer) Event() uint32 {
	return p.ev
}

func (p *Peer) Release() {
	syscall.Close(p.Fd())
	p.DelEvent(p)
}
