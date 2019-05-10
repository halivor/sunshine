package bufferpool

import (
	"unsafe"
)

var gbp *bufferpool
var locker uint32

func init() {
	gbp = New()
}

func Alloc(length int) []byte {
	b, _ := gbp.Alloc(length)
	return b
}

func AllocPointer(length int) unsafe.Pointer {
	b, _ := gbp.AllocPointer(length)
	return b
}

func Realloc(src []byte, length int) []byte {
	dst := Alloc(length)
	copy(dst, src)
	Release(src)
	return dst
}

func Release(buf []byte) {
	ReleasePointer(unsafe.Pointer(&buf[0]))
}

func ReleasePointer(ptr unsafe.Pointer) {
	gbp.ReleasePointer(ptr)
}
