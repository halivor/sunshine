package packet

import (
	"unsafe"
)

type Type int8

const (
	T_STRING Type = 1 << iota
	T_BINARY
	T_RPOTO
)

type Header struct {
	Ver uint16
	Nid uint16 // node id
	Uid uint32 // user id
	Cid uint32 // user type
	Res [64]byte
}

type UHeader interface {
	Ver() uint16
	Cmd() uint16
	Uid() uint32
	Cid() uint32
	Len() uint32
}

const (
	HLen  = int(unsafe.Sizeof(Header{}))
	BHLen = int(unsafe.Sizeof(BHeader{}))
	SHLen = int(unsafe.Sizeof(SHeader{}))
)

func Parse(data []byte) (UHeader, error) {
	pt := data[HLen]
	switch pt {
	case 'B':
		h := (*BHeader)(unsafe.Pointer(nil))
		return h.Parse()
	case 'J':
		h := (*SHeader)(unsafe.Pointer(nil))
		return h.Parse()
	}
	return nil, nil
}
