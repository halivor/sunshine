package connection

import (
	"log"
	"os"
	"syscall"
)

const (
	MAX_SENDQ_SIZE      = 32 // 超过队列，写入报错
	DEFAULT_BUFFER_SIZE = 32 * 1024
)

// TODO: connection nocache/connection buffercache
type C struct {
	fd int
	ss SockStat
	cc bool     // cached
	wb []buffer // cached buffer
	wp int      // first buffer start position

	*log.Logger
}

func (c *C) Send(data []byte) (int, error) {
	if c.Closed() {
		return 0, os.ErrClosed
	}
	return syscall.Write(c.fd, data)
}

func (c *C) SendBuffer(pb buffer) (e error) {
	if c.Closed() || len(c.wb) >= MAX_SENDQ_SIZE {
		return os.ErrClosed
	}
	if len(c.wb) > 0 {
		c.wb = append(c.wb, pb)
		return syscall.EAGAIN
	}
	if c.wp, e = syscall.Write(c.fd, pb.Buffer()); c.wp != pb.Len() {
		if c.wp < 0 {
			c.wp = 0
		}
		switch {
		case e == syscall.EAGAIN:
			c.wb = append(c.wb, pb)
		case e == nil:
			c.wb = append(c.wb, pb)
			return syscall.EAGAIN
		}
	}
	return e
}

func (c *C) SendBufferAgain() error {
	if c.Closed() {
		return os.ErrClosed
	}
	for len(c.wb) > 0 {
		switch n, e := syscall.Write(c.fd, c.wb[0].Buffer()[c.wp:]); e {
		case syscall.EAGAIN:
			// 当本方buffer全满时，e=11，n=-1
			if n > 0 {
				if c.wp += n; c.wp == c.wb[0].Len() {
					c.wb = c.wb[1:]
				}
			}
			return e
		case nil:
			// 当本方buffer满，本方本次会写入部分数据，同时e=nil
			if c.wp += n; c.wp == c.wb[0].Len() {
				c.wb = c.wb[1:]
				c.wp = 0
			} else {
				// 本次数据未完全写入时，表示Send-Q已满
				c.wp += n
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
	c.wb = nil
}

func (c *C) Fd() int {
	return c.fd
}
