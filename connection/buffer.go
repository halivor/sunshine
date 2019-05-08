package connection

import (
	"fmt"
	"syscall"

	cnf "github.com/halivor/sunshine/config"
)

type buffer interface {
	Len() int
	Buffer() []byte
	Release()
}

type BConn interface {
	Fd() int
	SendBuffer(buf buffer) error
	SendBufferAgain() error
	Recv(buf []byte) (int, error)
	Close()
}

func NewConnWithBuffer(fd int) BConn {
	SetSndBuf(fd, DEFAULT_BUFFER_SIZE)
	SetRcvBuf(fd, DEFAULT_BUFFER_SIZE)
	return &C{
		fd:     fd,
		ss:     ESTAB,
		wb:     make([]buffer, 0, MAX_SENDQ_SIZE),
		Logger: cnf.NewLogger(fmt.Sprintf("[sock(%d)] ", fd)),
	}
}

func NewTcpConnWithBuffer() (*C, error) {
	fd, e := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if e != nil {
		return nil, e
	}
	SetSndBuf(fd, DEFAULT_BUFFER_SIZE)
	SetRcvBuf(fd, DEFAULT_BUFFER_SIZE)
	return &C{
		fd:     fd,
		ss:     CREATE,
		wb:     make([]buffer, 0, MAX_SENDQ_SIZE),
		Logger: cnf.NewLogger(fmt.Sprintf("[tcp(%d)] ", fd)),
	}, nil
}
