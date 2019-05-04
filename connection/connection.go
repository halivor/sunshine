package connection

import (
	"fmt"
	"log"
	"os"
	"syscall"

	_ "github.com/halivor/sunshine/bufferpool"
	cnf "github.com/halivor/sunshine/config"
)

const (
	MAX_SENDQ_SIZE      = 32 // 超过队列，写入报错
	DEFAULT_BUFFER_SIZE = 32 * 1024
)

type Conn interface {
	Fd() int
	SendAgain() error
	Send(message []byte) error
	Recv(buf []byte) (int, error)
	Close()
}

type packet struct {
	buf []byte
	pos int
}

type C struct {
	fd int
	ss SockStat
	wl []*packet

	*log.Logger
}

func NewConn(fd int) *C {
	SetSndBuf(fd, DEFAULT_BUFFER_SIZE)
	SetRcvBuf(fd, DEFAULT_BUFFER_SIZE)
	return &C{
		fd:     fd,
		ss:     ESTAB,
		wl:     make([]*packet, 0, MAX_SENDQ_SIZE),
		Logger: cnf.NewLogger(fmt.Sprintf("[sock(%d)] ", fd)),
	}
}

func NewTcpConn() (*C, error) {
	fd, e := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if e != nil {
		return nil, e
	}
	SetSndBuf(fd, DEFAULT_BUFFER_SIZE)
	SetRcvBuf(fd, DEFAULT_BUFFER_SIZE)
	return &C{
		fd:     fd,
		ss:     CREATE,
		wl:     make([]*packet, 0, MAX_SENDQ_SIZE),
		Logger: cnf.NewLogger(fmt.Sprintf("[tcp(%d)] ", fd)),
	}, nil
}

func (c *C) Send(message []byte) error {
	if c.Closed() || len(c.wl) >= MAX_SENDQ_SIZE {
		return os.ErrClosed
	}
	if len(c.wl) > 0 {
		//c.Println("append", len(message), "bytes")
		c.wl = append(c.wl, &packet{
			buf: message,
			pos: 0,
		})
		return nil
	}
	n, e := syscall.Write(c.fd, message)
	//c.Println("write", n, ", error", e)
	if e == syscall.EAGAIN {
		if n < 0 {
			n = 0
		}
		c.wl = append(c.wl, &packet{
			buf: message,
			pos: n,
		})
	}
	if e == nil {
		if n < len(message) {
			c.wl = append(c.wl, &packet{
				buf: message,
				pos: n,
			})
			return syscall.EAGAIN
		}
	}
	return e
}

func (c *C) SendAgain() error {
	if c.Closed() {
		return os.ErrClosed
	}
	for len(c.wl) > 0 {
		switch n, e := syscall.Write(c.fd, c.wl[0].buf[c.wl[0].pos:]); e {
		case syscall.EAGAIN:
			// c.Println("write", n, "bytes", uintptr(syscall.EAGAIN), e)
			// 当双方buffer全满时，e=11，n=-1
			if n > 0 {
				c.wl[0].pos += n
			}
			return e
		case nil:
			//c.Println("pos", c.wl[0].pos, "len", len(c.wl[0].buf), "write", n)
			// 当对方buffer满，本方本次会写入部分数据，同时e=nil
			if n == len(c.wl[0].buf[c.wl[0].pos:]) {
				c.wl = c.wl[1:]
			} else {
				// 本次数据未完全写入时，表示Send-Q已满
				c.wl[0].pos += n
				return syscall.EAGAIN
			}
		default:
			return e
		}
	}
	return nil
}

func (c *C) Recv(buf []byte) (int, error) {
	if c.Closed() {
		return 0, os.ErrClosed
	}
	switch c.ss {
	case ESTAB:
		return syscall.Read(c.fd, buf)
	case CLOSED:
		return 0, os.ErrClosed
	}
	return 0, os.ErrInvalid
}

func (c *C) Closed() bool {
	if c.ss == CLOSED {
		return true
	}
	return false
}

func (c *C) Close() {
	syscall.Close(c.fd)
	c.ss = CLOSED
	c.wl = nil
}

func (c *C) Fd() int {
	return c.fd
}
