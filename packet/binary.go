package packet

type BHeader struct {
	ver uint16
	typ uint8
	opt uint8
	cmd uint32
	seq uint64
	len uint32
	res [8]byte
}

func (h *BHeader) Ver() int {
	return int(h.ver)
}

func (h *BHeader) Opt() int {
	return int(h.opt)
}

func (h *BHeader) Cmd() int {
	return int(h.cmd)
}

func (h *BHeader) Seq() int {
	return int(h.seq)
}

func (h *BHeader) Len() int {
	return int(h.len)
}
