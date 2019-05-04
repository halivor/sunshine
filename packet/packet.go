package packet

import (
	"unsafe"

	bp "github.com/halivor/sunshine/bufferpool"
)

type P struct {
	Len int
	Cap int
	Buf []byte
	ptr uintptr
}

func NewPkt() *P {
	buf := bp.Alloc(2048)
	p := (*P)(unsafe.Pointer(&buf[0]))
	p.Len = 0
	p.Cap = 2048 - int(unsafe.Offsetof(p.ptr))
	p.Buf = (*(*[bp.BUF_MAX_LEN]byte)(unsafe.Pointer(&p.ptr)))[:p.Cap]
	return p
}

func Alloc(length int) *P {
	alen := length + int(unsafe.Sizeof(P{}))
	buf := bp.Alloc(alen)
	p := (*P)(unsafe.Pointer(&buf[0]))
	p.Len = 0
	p.Cap = alen - int(unsafe.Offsetof(p.ptr))
	p.Buf = (*(*[bp.BUF_MAX_LEN]byte)(unsafe.Pointer(&p.ptr)))[:p.Cap]
	return p
}

func (p *P) Buffer() []byte {
	return (*(*[bp.BUF_MAX_LEN]byte)(unsafe.Pointer(&p.ptr)))[:p.Cap]
}

func (p *P) Release() {
	bp.Release((*(*[1]byte)(unsafe.Pointer(p)))[:])
}
