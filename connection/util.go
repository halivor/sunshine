package connection

import (
	"net"
	"syscall"
)

func ParseAddr4(network, addr string) (*syscall.SockaddrInet4, error) {
	ad, e := net.ResolveTCPAddr(network, addr)
	if e != nil {
		return nil, e
	}
	saddr := &syscall.SockaddrInet4{Port: ad.Port}
	copy(saddr.Addr[:], ad.IP[0:4])
	return saddr, nil
}
