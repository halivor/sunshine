package connection

import (
	"fmt"
	"syscall"

	cnf "github.com/halivor/sunshine/config"
)

type Conn interface {
	Fd() int
	Send(message []byte) (int, error)
	Recv(buf []byte) (int, error)
	Close()
}

func NewConn(fd int) Conn {
	SetSndBuf(fd, DEFAULT_BUFFER_SIZE)
	SetRcvBuf(fd, DEFAULT_BUFFER_SIZE)
	return &C{
		fd:     fd,
		ss:     ESTAB,
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
		Logger: cnf.NewLogger(fmt.Sprintf("[tcp(%d)] ", fd)),
	}, nil
}
