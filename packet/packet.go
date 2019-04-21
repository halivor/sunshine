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
	Res [20]byte
}

type UHeader interface {
	Ver() int
	Cmd() int
	Seq() int
	Len() int
}

const (
	HLen  = int(unsafe.Sizeof(Header{}))
	ALen  = int(unsafe.Sizeof(Auth{}))
	SHLen = int(unsafe.Sizeof(SHeader{}))
)
