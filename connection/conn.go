package connection

import (
	"fmt"
	"syscall"

	log "github.com/halivor/goutility/logger"
	sc "github.com/halivor/sunshine/conf"
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

	return &c{
		fd: fd,
		ss: ESTAB,
		Logger: log.NewLog("sunshine.log", fmt.Sprintf("[sock(%d)]", fd),
			log.LstdFlags, sc.LogLvlConn),
	}
}

func NewTcpConn() (*c, error) {
	fd, e := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if e != nil {
		return nil, e
	}
	SetSndBuf(fd, DEFAULT_BUFFER_SIZE)
	SetRcvBuf(fd, DEFAULT_BUFFER_SIZE)
	return &c{
		fd: fd,
		ss: CREATE,
		Logger: log.NewLog("sunshine.log", fmt.Sprintf("[tcp(%d)]", fd),
			log.LstdFlags, sc.LogLvlConn),
	}, nil
}
