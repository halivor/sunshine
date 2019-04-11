package peer

import (
	"syscall"

	c "github.com/halivor/frontend/connection"
	evp "github.com/halivor/frontend/eventpool"
)

type Peer struct {
	ev uint32
	rb []byte

	*c.C
	evp.EventPool

	Manage
}

func New(C *c.C, ep evp.EventPool, pm Manage) *Peer {
	if _, ok := pm.(*manager); !ok {
		return nil
	}
	return &Peer{
		ev: syscall.EPOLLIN,
		rb: make([]byte, 4096),

		C:         C,
		EventPool: ep,

		Manage: pm.(*manager),
	}
}

func (c *Peer) CallBack(ev uint32) {
	switch {
	case ev&syscall.EPOLLIN != 0:
		n, e := syscall.Read(c.Fd(), c.rb)
		if e != nil {
			c.Println(e)
			c.DelEvent(c)
			c.Release()
			return
		}
		c.Println(string(c.rb[:n]))
	case ev&syscall.EPOLLERR != 0:
	case ev&syscall.EPOLLOUT != 0:
	}
}

func (c *Peer) Event() uint32 {
	return c.ev
}
func (c *Peer) Release() {
}
