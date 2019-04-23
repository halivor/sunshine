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
	pkt "github.com/halivor/frontend/packet"
	evp "github.com/halivor/goevent/eventpool"
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
			n, e := syscall.Read(p.Fd(), p.rb[p.pos:])
			switch e {
			case nil:
				if n == 0 {
					p.Release()
					return
				}
				p.pos += n
				switch e := p.Process(); e {
				case nil, syscall.EAGAIN:
				default:
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
			return
		case ev&syscall.EPOLLOUT != 0:
			switch e := p.SendAgain(); e {
			case nil:
				p.ev = syscall.EPOLLIN
				p.ModEvent(p)
			case os.ErrClosed:
				p.Release()
			}
			return
		default:
			p.Release()
			return
		}
	}
}

func (p *Peer) Process() (e error) {
	switch p.ps {
	case PS_ESTAB:
		p.Println("estab", p.pos, string(p.rb[pkt.HLen:p.pos]))
		// 头长度不够，继续读取
		if e = p.Auth(); e != nil {
			return e
		}
	case PS_NORMAL:
		for {
			packet, e := p.Parse()
			if e != nil {
				return e
			}
			p.Transfer(packet)
		}
	}

	return nil
}

func (p *Peer) Auth() (e error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("check =>", r)
			e = os.ErrInvalid
		}
		// 用户信息结构
		p.Add(p)
	}()
	a := (*pkt.Auth)(unsafe.Pointer(&p.rb[pkt.HLen]))
	if p.pos < pkt.HLen+pkt.ALen || p.pos < pkt.HLen+pkt.ALen+a.Len() {
		p.Println("not enough")
		return syscall.EAGAIN
	}

	h := (*pkt.Header)(unsafe.Pointer(&p.rb[0]))
	h.Ver = uint16(a.Ver())
	h.Nid = cnf.NodeId
	h.Uid = uint32(a.Uid())
	h.Cid = uint32(a.Cid())
	p.ps = PS_NORMAL
	p.Header = *(*pkt.Header)(unsafe.Pointer(&p.rb[0]))
	// verify failed os.ErrInvalid

	// 转发packet消息
	packet := p.rb[:pkt.HLen+pkt.ALen+a.Len()]
	if p.pos > pkt.HLen+pkt.ALen+a.Len() {
		nb := bp.Alloc()
		copy(nb[pkt.HLen:], p.rb[len(packet):p.pos])
		p.rb = nb
		p.pos = pkt.HLen + p.pos - len(packet)
	} else {
		p.rb = bp.Alloc()
		p.pos = pkt.HLen
	}
	*(*pkt.Header)(unsafe.Pointer(&p.rb[0])) = p.Header
	p.Transfer(packet)

	return nil
}

func (p *Peer) Parse() (packet []byte, e error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("parse =>", r)
			e = os.ErrInvalid
		}
	}()
	sh := (*pkt.SHeader)(unsafe.Pointer(&p.rb[pkt.HLen]))
	if p.pos < pkt.HLen+pkt.SHLen || p.pos < pkt.HLen+pkt.SHLen+sh.Len() {
		return nil, syscall.EAGAIN
	}

	h := (*pkt.Header)(unsafe.Pointer(&p.rb[0]))
	if sh.Len() > 4*1024 {
		return nil, os.ErrInvalid
	}
	h.Cmd = uint32(sh.Cmd())
	h.Len = uint32(pkt.SHLen + sh.Len())
	plen := pkt.HLen + pkt.SHLen + sh.Len()
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
	return packet, nil
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
