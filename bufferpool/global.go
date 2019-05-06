package bufferpool

import (
	"runtime"
	"sync/atomic"
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

func Release(buf []byte) {
	for atomic.CompareAndSwapUint32(&locker, 0, 1) {
		runtime.Gosched()
	}
	gbp.Release(buf)
	atomic.StoreUint32(&locker, 0)
}
