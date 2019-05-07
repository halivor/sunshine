package bufferpool

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

var gbp *bufferpool
var locker uint32

func init() {
	gbp = New()
}

func Alloc(length int) []byte {
	for !atomic.CompareAndSwapUint32(&locker, 0, 1) {
		runtime.Gosched()
	}
	b, _ := gbp.Alloc(length)
	atomic.StoreUint32(&locker, 0)
	return b
}

func AllocPointer(length int) unsafe.Pointer {
	for !atomic.CompareAndSwapUint32(&locker, 0, 1) {
		runtime.Gosched()
	}
	b, _ := gbp.AllocPointer(length)
	atomic.StoreUint32(&locker, 0)
	return unsafe.Pointer(b)
}

func Release(buf []byte) {
	for atomic.CompareAndSwapUint32(&locker, 0, 1) {
		runtime.Gosched()
	}
	gbp.ReleasePointer(uintptr(unsafe.Pointer(&buf[0])))
	atomic.StoreUint32(&locker, 0)
}

func ReleasePointer(ptr unsafe.Pointer) {
	for atomic.CompareAndSwapUint32(&locker, 0, 1) {
		runtime.Gosched()
	}
	gbp.ReleasePointer(uintptr(ptr))
	atomic.StoreUint32(&locker, 0)

}
