package connection

import (
	"syscall"
)

func SetKeepAlive(fd int) error {
	if e := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); e != nil {
		return e
	}
	if e := syscall.SetsockoptInt(fd, syscall.SOL_TCP, syscall.TCP_KEEPIDLE, 15); e != nil {
		return e
	}
	if e := syscall.SetsockoptInt(fd, syscall.SOL_TCP, syscall.TCP_KEEPINTVL, 5); e != nil {
		return e
	}
	if e := syscall.SetsockoptInt(fd, syscall.SOL_TCP, syscall.TCP_KEEPCNT, 3); e != nil {
		return e
	}
	return nil
}

func Reuse(fd int, reusePort bool) error {
	// reuse addr
	if e := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); e != nil {
		return e
	}
	// reuse port
	if reusePort {
		return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, 15, 1)
	}
	return nil
}

func SetNoDelay(fd int) error {
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.TCP_NODELAY, 1)
}

// TODO: 确认Linger情况下，close对于Block和NonBlock socket的影响
func SetLinger(fd int) error {
	return syscall.SetsockoptLinger(fd, syscall.SOL_SOCKET, syscall.SO_LINGER, &syscall.Linger{Onoff: 1, Linger: 3})
}

func SetSndBuf(fd int, length int) error {
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_SNDBUF, length)
}

func SetRcvBuf(fd int, length int) error {
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_RCVBUF, length)
}
