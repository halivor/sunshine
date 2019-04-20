package packet

import ()

type BHeader struct {
	ver uint16
	cmd uint16
	uid uint32
	cid uint32
	len uint32
	seq uint64
}

func (h *BHeader) Parse() (UHeader, error) {
	return nil, nil
}
func (h *BHeader) Ver() uint16 {
	return h.ver
}
func (h *BHeader) Cmd() uint16 {
	return h.cmd
}
func (h *BHeader) Uid() uint32 {
	return h.uid
}
func (h *BHeader) Cid() uint32 {
	return h.cid
}
func (h *BHeader) Len() uint32 {
	return h.len
}
