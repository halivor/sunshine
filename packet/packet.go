package packet

import (
	"unsafe"
)

type Type int8

const (
	T_AUTH Type = 1 + iota
	T_STRING
	T_BINARY
	T_RPOTO
)

type Header struct {
	Ver uint16
	Nid uint16 // node id
	Uid uint32 // user id
	Cid uint32 // user type
	Cmd uint32
	Len uint32
	Res [12]byte
}

type UHeader interface {
	Ver() int
	Opt() int
	Cmd() int
	Seq() int
	Len() int
}

const (
	HLen  = int(unsafe.Sizeof(Header{}))
	ALen  = int(unsafe.Sizeof(Auth{}))
	SHLen = int(unsafe.Sizeof(SHeader{}))
)

func Parse(pb []byte) UHeader {
	switch pb[2] {
	case 'B':
		return (*BHeader)(unsafe.Pointer(&pb[0]))
	case 'S':
		return (*SHeader)(unsafe.Pointer(&pb[0]))
	}
	panic("undefined head type")
	return nil
}
