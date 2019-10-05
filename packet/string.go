package packet

import (
	"strconv"
	"unsafe"

	cp "common/golang/packet"
)

type SHeader struct {
	Ver [2]byte
	Typ uint8
	Opt uint8
	Cmd [4]byte
	Seq [8]byte
	Len [4]byte
	Res [8]byte
}

func (h *SHeader) IVer() int {
	v, e := strconv.Atoi(string(h.Ver[:]))
	if e != nil {
		plog.Panic(e)
	}
	return v
}

func (h *SHeader) IOpt() int {
	return int(h.Opt)
}

func (h *SHeader) ICmd() cp.CmdID {
	c, e := strconv.Atoi(string(h.Cmd[:]))
	if e != nil {
		plog.Panic(e)
	}
	return cp.CmdID(c)
}

func (h *SHeader) UISeq() uint64 {
	s, e := strconv.ParseUint(string(h.Seq[:]), 10, 64)
	if e != nil {
		plog.Panic(e)
	}
	return s
}

func (h *SHeader) ILen() int {
	l, e := strconv.Atoi(string(h.Len[:]))
	if e != nil {
		plog.Panic(e)
	}
	return l
}

func (h *SHeader) String() string {
	return string((*(*[HLen]byte)(unsafe.Pointer(h)))[:HLen:HLen])
}
