package bufferpool

import (
	"unsafe"
)

type P struct {
	Len uint16
	Cap uint16
	ref uint32
	bp  *bufferpool
	ptr uintptr
}

func (p *P) Buffer() []byte {
	return (*(*[BUF_MAX_LEN]byte)(unsafe.Pointer(&p.ptr)))[:p.Cap]
}

func (p *P) Reference() []byte {
	p.ref++
	return (*(*[BUF_MAX_LEN]byte)(unsafe.Pointer(&p.ptr)))[:p.Len]
}

func (p *P) Release() {
	p.ref--
}
