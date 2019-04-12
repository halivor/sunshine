package connection

import (
	"syscall"
)

func (c *C) NonBlock() {
	if e := syscall.SetsockoptInt(c.fd, syscall.SOL_TCP, syscall.SOCK_NONBLOCK, 1); e != nil {
		c.Println(e)
	}
}

func (c *C) KeepAlive(idle int) {
	if e := syscall.SetsockoptInt(c.fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); e != nil {
		c.Println(e)
	}
	if e := syscall.SetsockoptInt(c.fd, syscall.SOL_TCP, syscall.TCP_KEEPIDLE, 15); e != nil {
		c.Println(e)
	}
	if e := syscall.SetsockoptInt(c.fd, syscall.SOL_TCP, syscall.TCP_KEEPINTVL, 5); e != nil {
		c.Println(e)
	}
	if e := syscall.SetsockoptInt(c.fd, syscall.SOL_TCP, syscall.TCP_KEEPCNT, 3); e != nil {
		c.Println(e)
	}
}

func (c *C) Reuse() {
	if e := syscall.SetsockoptInt(c.fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); e != nil {
		c.Println(e)
	}
	if e := syscall.SetsockoptInt(c.fd, syscall.SOL_SOCKET, 15, 1); e != nil {
		c.Println(e)
	}
}

func (c *C) NoDelay() {
	if e := syscall.SetsockoptInt(c.fd, syscall.SOL_SOCKET, syscall.TCP_NODELAY, 1); e != nil {
		c.Println(e)
	}
}
