package packet

import (
	"sync/atomic"
	"unsafe"

	bp "github.com/halivor/goutility/bufferpool"
)

const (
	STD_PKT_LEN = 1024
	PKT_SIZE    = int(unsafe.Sizeof(P{}))
)

type P struct {
	Len int
	Cap int
	Buf []byte
	ref int64
	ptr uintptr
}

func NewPkt() *P {
	ptr := bp.AllocPointer(STD_PKT_LEN)
	p := (*P)(unsafe.Pointer(ptr))
	p.Len = 0
	p.Cap = STD_PKT_LEN - PKT_SIZE
	p.Buf = (*(*[bp.BUF_MAX_LEN]byte)(unsafe.Pointer(uintptr(ptr) + uintptr(PKT_SIZE))))[:p.Cap:p.Cap]
	p.ref = 1
	p.ptr = uintptr(ptr)
	//pl.Trace("alloc  ", unsafe.Pointer(&buf[0]))
	return p
}

func Alloc(length int) *P {
	alen := length + PKT_SIZE
	ptr := bp.AllocPointer(alen)
	p := (*P)(unsafe.Pointer(ptr))
	p.Len = length
	p.Cap = alen - PKT_SIZE
	p.Buf = (*(*[bp.BUF_MAX_LEN]byte)(unsafe.Pointer(uintptr(ptr) + uintptr(PKT_SIZE))))[:p.Cap:p.Cap]
	p.ref = 1
	p.ptr = uintptr(ptr)
	//pl.Trace("alloc  ", unsafe.Pointer(&buf[0]))
	return p
}

func (p *P) Buffer() []byte {
	return (*(*[bp.BUF_MAX_LEN]byte)(unsafe.Pointer(&p.ptr)))[:p.Cap]
}

func (p *P) Reference() *P {
	atomic.AddInt64(&p.ref, 1)
	np := *p
	return &np
}

func (p *P) Release() {
	if atomic.AddInt64(&p.ref, -1) == 0 {
		bp.ReleasePointer(unsafe.Pointer(p.ptr))
	}
}
