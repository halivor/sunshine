package bufferpool

import (
	"container/list"
)

const (
	MIN_LEN = 4096
	MAX_LEN = 4 * 1024 * 1024
)

type BufferPool struct {
	idle  *list.List
	large map[uint32]*list.List
}

func New() *BufferPool {
	return &BufferPool{}
}

// length <= 4K
// length >= 4M
func (bp *BufferPool) New() {
}

func (bp *BufferPool) Get(length uint32) []byte {
	if length <= MIN_LEN {
		if bp.idle.Len() > 0 {
			if buffer, ok := bp.idle.Remove(bp.idle.Front()).([]byte); ok {
				return buffer
			}
		}
		return make([]byte, MIN_LEN)
	}
	if length >= 4*1024*1024 {
		return make([]byte, length+(^length&(MIN_LEN-1)+1)&(MIN_LEN-1))
	}
	if _, ok := bp.large[length+(^length&(MIN_LEN-1)+1)&(MIN_LEN-1)]; ok {
	}
	return make([]byte, length+(^length&(MIN_LEN-1)+1)&(MIN_LEN-1))
}

func (bp *BufferPool) Release(buffer []byte) {
	switch {
	case cap(buffer) == MIN_LEN:
	case cap(buffer) >= MAX_LEN:
	default:
	}
}
