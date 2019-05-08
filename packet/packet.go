package packet

import (
	_ "log"
	"sync/atomic"
	"unsafe"

	bp "github.com/halivor/sunshine/bufferpool"
)

const (
	STD_PKT_LEN = 1024
)

type P struct {
	Len int
	Cap int
	Buf []byte
	ref int64
	ptr uintptr
}

func NewPkt() *P {
	buf := bp.Alloc(STD_PKT_LEN)
	p := (*P)(unsafe.Pointer(&buf[0]))
	p.Len = 0
	p.Cap = cap(buf) - int(unsafe.Offsetof(p.ptr))
	p.Buf = (*(*[bp.BUF_MAX_LEN]byte)(unsafe.Pointer(&p.ptr)))[:p.Cap]
	p.ref = 1
	//log.Println("alloc  ", unsafe.Pointer(&buf[0]))
	return p
}

func Alloc(length int) *P {
	alen := length + int(unsafe.Sizeof(P{}))
	buf := bp.Alloc(alen)
	p := (*P)(unsafe.Pointer(&buf[0]))
	p.Len = length
	p.Cap = cap(buf) - int(unsafe.Offsetof(p.ptr))
	p.Buf = (*(*[bp.BUF_MAX_LEN]byte)(unsafe.Pointer(&p.ptr)))[:p.Cap]
	p.ref = 1
	//log.Println("alloc  ", unsafe.Pointer(&buf[0]))
	return p
}

func (p *P) Buffer() []byte {
	return (*(*[bp.BUF_MAX_LEN]byte)(unsafe.Pointer(&p.ptr)))[:p.Cap]
}

func (p *P) Refefence() *P {
	atomic.AddInt64(&p.ref, 1)
	return p
}

func (p *P) Release() {
	if atomic.AddInt64(&p.ref, -1) == 0 {
		//log.Println("release", unsafe.Pointer(p))
		bp.Release((*(*[1]byte)(unsafe.Pointer(p)))[:])
	}
}
