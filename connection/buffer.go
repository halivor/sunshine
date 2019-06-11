package connection

import (
	"fmt"
	"syscall"

	log "github.com/halivor/goutility/logger"
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

	logger, _ := log.New("/data/logs/sunshine/c.log", fmt.Sprintf("[sock(%d)]", fd), log.LstdFlags, log.TRACE)
	return &c{
		fd:     fd,
		ss:     ESTAB,
		wb:     make([]buffer, 0, MAX_SENDQ_SIZE),
		Logger: logger,
	}
}

func NewTcpConnWithBuffer() (*c, error) {
	fd, e := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if e != nil {
		return nil, e
	}
	SetSndBuf(fd, DEFAULT_BUFFER_SIZE)
	SetRcvBuf(fd, DEFAULT_BUFFER_SIZE)

	logger, _ := log.New("/data/logs/sunshine/c.log", fmt.Sprintf("[tcp(%d)]", fd), log.LstdFlags, log.TRACE)
	return &c{
		fd:     fd,
		ss:     CREATE,
		wb:     make([]buffer, 0, MAX_SENDQ_SIZE),
		Logger: logger,
	}, nil
}
