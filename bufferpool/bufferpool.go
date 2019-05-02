package bufferpool

import (
	"os"
	"unsafe"
)

const (
	BUF_MIN_LEN = 1024
	BUF_MAX_LEN = 4 * 1024 * 1024
)

type BufferPool interface {
}

var arrSize [4]uint16 = [4]uint16{1024, 2048, 4096, 8192}
var memNum map[uint16]uint16 = map[uint16]uint16{
	1024: 8000 * 4,
	2048: 8000 * 2,
	4096: 4000,
	8192: 2000,
}

// bufferpool =
//   1024 * 8000 * 4 = 32M
//   2048 * 8000 * 2 = 32M
//   4096 * 4000     = 16M
//   8192 * 2000     = 16M
type bufferpool struct {
	memCache map[uint16][]uintptr
	memRef   map[uintptr]uint16
}

var gpool [][]byte

func init() {
	gpool = make([][]byte, 128)
}

func New() BufferPool {
	bp := &bufferpool{
		memCache: make(map[uint16][]uintptr, 32),
	}
	for size, num := range memNum {
		bp.memCache[size] = bp.allocMemory(size, num)
	}
	return bp
}

func (bp *bufferpool) allocMemory(size, num uint16) []uintptr {
	list := make([]uintptr, num, num*10)
	pool := make([]byte, size*num)
	gpool = append(gpool, pool)
	for pre, cur := uint16(0), uint16(1); cur-1 < num; pre, cur = cur*size, cur+1 {
		list[cur-1] = uintptr(unsafe.Pointer(&pool[pre]))
	}
	return list
}

func (bp *bufferpool) Alloc(length int) (buf []byte, e error) {
	for i := 0; i < len(arrSize); i++ {
		size := arrSize[i]
		if length < int(size) {
			if mc, ok := bp.memCache[size]; ok {
				switch {
				case len(mc) == 0:
					return nil, os.ErrInvalid
				case len(mc) == 1:
					buf = (*((*[BUF_MAX_LEN]byte)(unsafe.Pointer(mc[0]))))[:size]
					num := memNum[size]
					bp.memCache[size] = bp.allocMemory(num, size)
				default:
					buf = (*((*[BUF_MAX_LEN]byte)(unsafe.Pointer(mc[0]))))[:size]
					bp.memCache[size] = mc[1:]
				}
				bp.memRef[uintptr(unsafe.Pointer(&buf[0]))] = size
				return buf, nil
			}
		}
	}
	return make([]byte, length), nil
}

func (bp *bufferpool) Release(size uint16, buffer []byte) {
	if size, ok := bp.memRef[uintptr(unsafe.Pointer(&buffer[0]))]; ok {
		bp.memCache[size] = append(bp.memCache[size], uintptr(unsafe.Pointer(&buffer[0])))
		/*if cap(bp.memCache[size])-len(bp.memCache[size]) < 64 {*/
		//list := make([]uintptr, 0, memNum[size]*10)
		//copy(list, bp.memCache[size])
		//bp.memCache[size] = list
		/*}*/
		delete(bp.memRef, uintptr(unsafe.Pointer(&buffer[0])))
	}
}
