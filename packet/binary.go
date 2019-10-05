package packet

import (
	"fmt"

	cp "common/golang/packet"
)

type BHeader struct {
	ver uint16
	typ uint8
	opt uint8
	cmd uint32
	seq uint64
	len uint32
	res uint64
}

func (h *BHeader) IVer() int {
	return int(h.ver)
}

func (h *BHeader) IOpt() int {
	return int(h.opt)
}

func (h *BHeader) ICmd() cp.CmdID {
	return cp.CmdID(h.cmd)
}

func (h *BHeader) UISeq() uint64 {
	return h.seq
}

func (h *BHeader) ILen() int {
	return int(h.len)
}

func (h *BHeader) String() string {
	return fmt.Sprintf("")
}
