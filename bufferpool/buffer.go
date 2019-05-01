package bufferpool

type P struct {
	Buf []byte
	bp  *bufferpool
	ref uint16
	Pos uint16
	ptr uintptr
}

func (p *P) Reference() *P {
	p.ref++
	return p
}

func (p *P) Release() {
}
