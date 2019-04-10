package connection

import (
	"container/list"
	"fmt"
	"log"
	"os"
	"syscall"
)

type C struct {
	fd int
	*log.Logger
	rl *list.List
	wl *list.List
}

type Sock interface {
	Fd() int
}

func NewSock(fd int) *C {
	return &C{
		fd:     fd,
		Logger: log.New(os.Stderr, fmt.Sprintf("[tcp(%d)] ", fd), log.LstdFlags|log.Lmicroseconds),
		rl:     list.New(),
		wl:     list.New(),
	}
}

func NewTcp() *C {
	fd, e := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if e != nil {
		log.Panicln(e.Error())
	}
	return &C{
		fd:     fd,
		Logger: log.New(os.Stderr, fmt.Sprintf("[tcp(%d)] ", fd), log.LstdFlags|log.Lmicroseconds),
		rl:     list.New(),
		wl:     list.New(),
	}
}

func (c *C) Fd() int {
	return c.fd
}

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
