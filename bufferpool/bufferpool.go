package bufferpool

import (
	"container/list"

	"github.com/halivor/frontend/config"
)

type BufferPool struct {
	idle  *list.List
	large map[uint32]*list.List
}

var gbp *BufferPool

func init() {
	gbp = &BufferPool{
		idle:  list.New(),
		large: make(map[uint32]*list.List),
	}
}

func New() *BufferPool {
	return &BufferPool{}
}

func Alloc() []byte {
	return gbp.Alloc()
}

func AllocLarge(length uint32) []byte {
	return gbp.AllocLarge(length)
}

func Release(buffer []byte) {
	gbp.Release(buffer)
}

// length <= 4K
// length >= 4M
func (bp *BufferPool) Alloc() []byte {
	return make([]byte, config.BUF_MIN_LEN)
	if bp.idle.Len() > 0 {
		if buffer, ok := bp.idle.Remove(bp.idle.Front()).([]byte); ok {
			return buffer
		}
	}
	return make([]byte, config.BUF_MIN_LEN)
}

func (bp *BufferPool) AllocLarge(length uint32) []byte {
	if length <= config.BUF_MIN_LEN {
		return bp.Alloc()
	}
	if length >= 4*1024*1024 {
		return make([]byte, length+(^length&(config.BUF_MIN_LEN-1)+1)&(config.BUF_MIN_LEN-1))
	}
	if _, ok := bp.large[length+(^length&(config.BUF_MIN_LEN-1)+1)&(config.BUF_MIN_LEN-1)]; ok {
	}
	return make([]byte, length+(^length&(config.BUF_MIN_LEN-1)+1)&(config.BUF_MIN_LEN-1))
}

func (bp *BufferPool) Release(buffer []byte) {
	switch {
	case cap(buffer) == config.BUF_MIN_LEN:
	case cap(buffer) >= config.BUF_MAX_LEN:
	default:
	}
}
