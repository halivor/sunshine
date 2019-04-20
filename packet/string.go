package packet

type SHeader struct {
	ver [4]byte
	cmd [4]byte
	uid [8]byte
	cid [8]byte
	len [8]byte
	seq [8]byte
}

func (h *SHeader) Parse() (UHeader, error) {
	return nil, nil
}
func (h *SHeader) Ver() uint16 {
	return 0
}
func (h *SHeader) Cmd() uint16 {
	return 0
}
func (h *SHeader) Uid() uint32 {
	return 0
}
func (h *SHeader) Cid() uint32 {
	return 0
}
func (h *SHeader) Len() uint32 {
	return 0
}
