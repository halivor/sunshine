package peer

import (
	"os"
	"syscall"
	"unsafe"

	bp "github.com/halivor/sunshine/bufferpool"
	cnf "github.com/halivor/sunshine/config"
	pkt "github.com/halivor/sunshine/packet"
)

func (p *Peer) recv() error {
	for {
		switch n, e := syscall.Read(p.Fd(), p.rp.Buf[p.rp.Len:]); {
		case e != nil:
			return e
		case n == 0:
			return os.ErrClosed
		default:
			p.rp.Len += n
			if e := p.process(); e != nil {
				return e
			}
			if n < p.rp.Cap {
				return nil
			}
		}
	}
}

func (p *Peer) process() error {
	for {
		switch p.ps {
		case PS_ESTAB:
			//p.Println("estab", p.rp.Len, string(p.rp.Buf[:p.rp.Len]))
			if e := p.auth(); e != nil {
				return e
			}
			p.Send([]byte(pkt.AUTH_SUCC))
		case PS_NORMAL:
			if e := p.parse(); e != nil {
				return e
			}
		default:
			return os.ErrInvalid
		}
	}
	return nil
}
func (p *Peer) auth() (e error) {
	defer func() {
		switch r := recover(); {
		case r != nil:
			p.Println("check =>", r)
			e = os.ErrInvalid
		case e == nil:
			p.Add(p)
		}
	}()
	rp := p.rp
	a := (*pkt.Auth)(unsafe.Pointer(&rp.Buf[0]))
	plen := pkt.ALen + a.Len()
	if rp.Len < pkt.ALen || rp.Len < plen {
		return syscall.EAGAIN
	}

	p.header.Ver = uint16(a.Ver())
	p.header.Nid = cnf.NodeId
	p.header.Uid = uint32(a.Uid())
	p.header.Cid = uint32(a.Cid())
	p.ps = PS_NORMAL
	// verify failed os.ErrInvalid

	// 转发packet消息
	tlen := pkt.HLen + plen
	tb := bp.Alloc(tlen)
	*(*pkt.Header)(unsafe.Pointer(&tb[0])) = p.header
	copy(tb[pkt.HLen:tlen], rp.Buf[:plen])
	p.Transfer(tb[:tlen])
	bp.Release(tb)

	if rp.Len > plen {
		copy(rp.Buf, rp.Buf[plen:rp.Len])
	}
	rp.Len = rp.Len - plen
	return nil
}
func (p *Peer) parse() (e error) {
	defer func() {
		if r := recover(); r != nil {
			p.Println("parse =>", r)
			e = os.ErrInvalid
		}
	}()
	rp := p.rp
	// 用户包长度校验
	uh := pkt.Parse(rp.Buf)
	plen := pkt.SHLen + uh.Len()
	if rp.Len < plen {
		return syscall.EAGAIN
	}
	if uh.Len() > 4*1024 {
		return os.ErrInvalid
	}

	p.header.Cmd = pkt.CmdID(uh.Cmd())
	p.header.SetLen(uint32(pkt.SHLen + uh.Len()))

	switch p.header.Cmd {
	case pkt.C_PING:
		p.Send([]byte(pkt.PONG))
	default:
		tlen := pkt.HLen + plen
		tb := bp.Alloc(tlen)
		*(*pkt.Header)(unsafe.Pointer(&tb[0])) = p.header
		copy(tb[pkt.HLen:tlen], rp.Buf[:plen])
		p.Transfer(tb)
		bp.Release(tb)
	}

	if rp.Len > plen {
		copy(rp.Buf, rp.Buf[plen:rp.Len])
	}
	rp.Len = rp.Len - plen
	return nil
}
