package peer

import (
	"os"
	"syscall"
	"unsafe"

	bp "github.com/halivor/frontend/bufferpool"
	cnf "github.com/halivor/frontend/config"
	pkt "github.com/halivor/frontend/packet"
)

func (p *Peer) recv() error {
	for {
		n, e := syscall.Read(p.Fd(), p.rb[p.pos:])
		switch {
		case e != nil:
			if e == syscall.EAGAIN {
				p.Println("read eagain", e, n)
			}
			return e
		case n == 0:
			return os.ErrClosed
		default:
			p.pos += n
			if e := p.process(); e != nil /*&& e != syscall.EAGAIN */ {
				return e
			}
		}
	}
}

func (p *Peer) process() error {
	for {
		switch p.ps {
		case PS_ESTAB:
			p.Println("estab", p.pos, string(p.rb[pkt.HLen:p.pos]))
			if e := p.auth(); e != nil {
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
			//p.Println(p.Header, string(p.pkt[pkt.HLen:pkt.HLen+h.Len()]))
			switch h.Cmd {
			case pkt.C_PING:
				p.Send([]byte(pkt.PONG))
			default:
				p.Println(h.Cmd)
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
			p.Println("check =>", r)
			e = os.ErrInvalid
		} else {
			// 用户信息结构
			p.Add(p)
		}
	}()
	a := (*pkt.Auth)(unsafe.Pointer(&p.rb[pkt.HLen]))
	plen := pkt.HLen + pkt.ALen + a.Len()
	if p.pos < pkt.HLen+pkt.ALen || p.pos < plen {
		// 头长度不够，继续读取
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
	if p.pos > plen {
		copy(p.rb[pkt.HLen:], p.pkt[plen:p.pos])
	}
	p.pos = pkt.HLen + p.pos - plen
	p.Println("auth remain", p.pos, string(p.rb[pkt.HLen:p.pos]))

	*(*pkt.Header)(unsafe.Pointer(&p.rb[0])) = p.Header
	p.Transfer(p.pkt)
	p.Send([]byte(pkt.AUTH_SUCC))

	return nil
}
func (p *Peer) parse() (e error) {
	defer func() {
		if r := recover(); r != nil {
			p.Println("parse =>", r)
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

	*(*pkt.Header)(unsafe.Pointer(&p.rb[0])) = p.Header

	p.pkt = p.rb[:plen]
	p.rb = bp.Alloc()
	if p.pos > plen {
		copy(p.rb[pkt.HLen:], p.pkt[plen:p.pos])
	}
	p.pos = pkt.HLen + p.pos - plen
	return nil
}
