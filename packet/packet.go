package packet

import (
	"unsafe"
)

type Header struct {
	Ver uint16
	Nid uint16 // node id
	Uid uint32 // user id
	Cid uint32 // user type
	Res [32]byte
}

type UHeader struct {
	Ver uint16
	Cmd uint16
	Uid uint32
	Cid uint32
	Ttl uint32
	Len uint32
	seq uint64
}

const (
	HLen  = int(unsafe.Sizeof(Header{}))
	UHLen = int(unsafe.Sizeof(UHeader{}))
)
