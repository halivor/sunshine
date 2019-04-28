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
		p.recv()
	case ev&syscall.EPOLLOUT != 0:
		p.send()
	default:
		p.Println("event error", ev)
		p.Release()
	}
}

func (p *Peer) recv() {
	for {
		n, e := syscall.Read(p.Fd(), p.rb[p.pos:])
		switch e {
		case nil:
			if n == 0 {
				p.Release()
				return
			}
			p.pos += n
			if e := p.process(); e != nil && e != syscall.EAGAIN {
				p.Release()
			}
		case syscall.EAGAIN:
			return
		default:
			p.Release()
			return
		}
	}
}

func (p *Peer) process() (e error) {
	for {
		switch p.ps {
		case PS_ESTAB:
			p.Println("estab", p.pos, string(p.rb[pkt.HLen:p.pos]))
			// 头长度不够，继续读取
			if e = p.auth(); e != nil {
				return e
			}
		case PS_NORMAL:
			if p.pos < pkt.HLen+pkt.SHLen {
				return syscall.EAGAIN
			}
			if e := p.parse(); e != nil {
				return e
			}
			h := (*pkt.Header)(unsafe.Pointer(&p.pkt[0]))
			p.Println("normal header", h)
			switch h.Cmd {
			case pkt.C_PING:
				p.Send([]byte(pkt.PONG))
			default:
				p.Transfer(p.pkt)
			}
		default:
			return os.ErrInvalid
		}
	}

	return nil
}

func (p *Peer) auth() (e error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("check =>", r)
			e = os.ErrInvalid
		} else {
			// 用户信息结构
			p.Add(p)
		}
	}()
	a := (*pkt.Auth)(unsafe.Pointer(&p.rb[pkt.HLen]))
	plen := pkt.HLen + pkt.ALen + a.Len()
	if p.pos < pkt.HLen+pkt.ALen || p.pos < plen {
		return syscall.EAGAIN
	}

	p.Ver = uint16(a.Ver())
	p.Nid = cnf.NodeId
	p.Uid = uint32(a.Uid())
	p.Cid = uint32(a.Cid())
	p.ps = PS_NORMAL
	// verify failed os.ErrInvalid

	// 转发packet消息
	p.pkt = p.rb[:plen]
	p.rb = bp.Alloc()
	p.pos = pkt.HLen
	if p.pos > plen {
		copy(p.rb[pkt.HLen:], p.rb[plen:p.pos])
		p.pos = pkt.HLen + p.pos - plen
	}
	*(*pkt.Header)(unsafe.Pointer(&p.rb[0])) = p.Header
	p.Transfer(p.pkt)
	p.Send([]byte(pkt.AUTH_SUCC))

	return nil
}

func (p *Peer) parse() (e error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("parse =>", r)
			e = os.ErrInvalid
		}
	}()
	// 用户包长度校验
	sh := pkt.Parse(p.rb[pkt.HLen:])
	plen := pkt.HLen + pkt.SHLen + sh.Len()
	if p.pos < plen {
		return syscall.EAGAIN
	}

	if sh.Len() > 4*1024 {
		return os.ErrInvalid
	}
	p.Header.Cmd = uint32(sh.Cmd())
	p.Header.SetLen(uint32(pkt.SHLen + sh.Len()))

	p.pkt = p.rb[:plen]
	p.rb = bp.Alloc()
	p.pos = pkt.HLen
	if rl := p.pos - plen; rl > 0 {
		copy(p.rb[pkt.HLen:], p.pkt[plen:p.pos])
		p.pos += rl
	}

	p.pkt = p.pkt[:plen]
	*(*pkt.Header)(unsafe.Pointer(&p.rb[0])) = p.Header
	return nil
}

func (p *Peer) send() {
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
