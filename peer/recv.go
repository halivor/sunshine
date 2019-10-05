package peer

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	cp "common/golang/packet"
	bp "github.com/halivor/goutility/bufferpool"
	cnf "github.com/halivor/sunshine/conf"
	up "github.com/halivor/sunshine/packet"
)

func (p *Peer) recv() error {
	for {
		switch n, e := syscall.Read(p.Fd(), p.rp.Buf[p.rp.Len:]); {
		case e != nil:
			p.Warn("read", e)
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
		switch {
		case p.ps == PS_NORMAL && p.rp.Len < up.SHLen:
			return syscall.EAGAIN
		case p.ps == PS_NORMAL:
			if e := p.parse(); e != nil {
				return e
			}
		case p.ps == PS_ESTAB && p.rp.Len < up.ALen:
			return syscall.EAGAIN
		case p.ps == PS_ESTAB:
			p.Trace("estab", p.rp.Len, string(p.rp.Buf[:p.rp.Len]))
			if e := p.auth(); e != nil {
				p.Warn("auth", e)
				return e
			}
			p.SetPrefix(fmt.Sprintf("[p(%d)u(%d)]", p.Fd(), p.header.Uid))
			p.Trace("auth success")
			p.Send(up.AuthSucc)
		default:
			p.Warn("unknown status", p.ps)
			return os.ErrInvalid
		}
	}
	return nil
}
func (p *Peer) auth() (e error) {
	defer func() {
		if r := recover(); r != nil {
			p.Warn("check =>", r)
			e = os.ErrInvalid
		}
	}()
	rp := p.rp
	a := (*up.Auth)(unsafe.Pointer(&rp.Buf[0]))
	plen := up.ALen + a.Len()
	if rp.Len < plen {
		return syscall.EAGAIN
	}

	p.header.Ver = uint16(a.Ver())
	p.header.Nid = cnf.NodeId
	p.header.Uid = uint32(a.Uid())
	p.header.Cid = uint32(a.Cid())
	p.ps = PS_NORMAL
	p.AddPeer(p)

	// 转发packet消息
	tlen := up.HLen + plen
	tb := bp.Alloc(tlen)
	*(*up.Header)(unsafe.Pointer(&tb[0])) = p.header
	copy(tb[up.HLen:tlen], rp.Buf[:plen])
	p.Transfer(tb[:tlen])
	bp.Free(tb)

	if rp.Len > plen {
		copy(rp.Buf, rp.Buf[plen:rp.Len])
	}
	rp.Len -= plen
	return nil
}

// TODO: 只有一个包, 直接返回EAGAIN
func (p *Peer) parse() (e error) {
	defer func() {
		if r := recover(); r != nil {
			p.Warn("parse =>", r)
			e = os.ErrInvalid
		}
	}()
	rp := p.rp
	// 用户包长度校验
	uh := up.Parse(rp.Buf)
	if uh.ILen() > 4*1024 {
		return os.ErrInvalid
	}

	plen := up.SHLen + uh.ILen()
	p.header.Cmd = uh.ICmd()
	p.header.SetLen(uint32(plen))

	switch p.header.Cmd {
	case cp.C_PING:
		p.Trace("ping/pong...")
		p.GenSeq(up.Pong)
		p.Send(up.Pong)
	default:
		tlen := up.HLen + plen
		tb := bp.Alloc(tlen)
		*(*up.Header)(unsafe.Pointer(&tb[0])) = p.header
		copy(tb[up.HLen:tlen], rp.Buf[:plen])
		p.Trace("recv", string(tb[up.HLen:tlen]))
		p.Transfer(tb)
		bp.Free(tb)
	}

	if rp.Len > plen {
		copy(rp.Buf, rp.Buf[plen:rp.Len])
	}
	rp.Len -= plen
	return nil
}
