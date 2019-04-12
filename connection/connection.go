package connection

import (
	"fmt"
	"log"
	"os"
	"syscall"

	bp "github.com/halivor/frontend/bufferpool"
)

const (
	MAX_SENDQ_SIZE = 32
)

type Conn interface {
	Fd() int
	SendAgain() error
	Send(message []byte) error
	Recv()
	Close()
}

type packet struct {
	buf []byte
	pos int
}

type C struct {
	stat bool
	fd   int
	rb   []byte
	wl   []*packet

	*log.Logger
}

func NewSock(fd int) *C {
	return &C{
		fd:     fd,
		stat:   true,
		rb:     bp.Alloc(),
		wl:     make([]*packet, MAX_SENDQ_SIZE),
		Logger: log.New(os.Stderr, fmt.Sprintf("[tcp(%d)] ", fd), log.LstdFlags|log.Lmicroseconds),
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
		rb:     make([]byte, 4096),
		wl:     make([]*packet, 0),
	}
}

func (c *C) SendAgain() error {
	for {
		if len(c.wl) > 0 {
			switch n, e := syscall.Write(c.fd, c.wl[0].buf[c.wl[0].pos:]); e {
			case syscall.EAGAIN:
				c.wl[0].pos += n
			case nil:
				if n == len(c.wl[0].buf[c.wl[0].pos:]) {
					bp.Release(c.wl[0].buf)
					c.wl = c.wl[1:]
				} else {
					c.wl[0].pos += n
				}
			}
		}
	}
	return nil
}

func (c *C) Send(message []byte) error {
	if len(c.wl) > 0 {
		c.wl = append(c.wl, &packet{
			buf: message,
			pos: 0,
		})
		return syscall.EAGAIN
	}
	n, e := syscall.Write(c.fd, message)
	if n == len(message) {
		bp.Release(message)
	} else {
		c.wl = append(c.wl, &packet{
			buf: message,
			pos: n,
		})
	}
	return e
}

func (c *C) Recv() {
}

func (c *C) Close() {
}

func (c *C) Fd() int {
	return c.fd
}
