package client

import (
	"syscall"

	c "frontend/connection"
	evp "frontend/eventpool"
)

type Client struct {
	ev uint32
	rb []byte

	*c.C
	evp.EventPool
}

func New(C *c.C, ep evp.EventPool) *Client {
	return &Client{
		C:         C,
		ev:        syscall.EPOLLIN,
		rb:        make([]byte, 4096),
		EventPool: ep,
	}
}

func (c *Client) CallBack(ev uint32) {
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

func (c *Client) Event() uint32 {
	return c.ev
}
func (c *Client) Release() {
}
