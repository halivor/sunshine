package packet

import (
	"unsafe"
)

type Header struct {
	Ver  int8
	Type int8
	Len  int16
	room int32
	uid  uint64
	res  [64]byte
}

const (
	HLen = unsafe.Sizeof(Header{})
)
