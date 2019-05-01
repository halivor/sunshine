package bufferpool

const (
	BUF_MIN_LEN = 1024
	BUF_MAX_LEN = 4 * 1024 * 1024
)

type BufferPool interface {
}

// bufferpool =
//   1024 * 8000 * 4 = 32M
//   2048 * 8000 * 2 = 32M
//   4096 * 4000     = 16M
//   8192 * 2000     = 16M
type bufferpool struct {
	memList map[uint16][][]byte
}

var memSize map[uint16]uint16 = map[uint16]uint16{
	1024: 8000 * 4,
	2048: 8000 * 2,
	4096: 4000,
	8192: 2000,
}

func New() BufferPool {
	bp := &bufferpool{
		memList: make(map[uint16][][]byte, 32),
	}
	for size, num := range memSize {
		slice := make([][]byte, num)
		pool := make([]byte, size*num)
		for pre, cur := uint16(0), uint16(1); cur < num; pre, cur = cur*size, cur+1 {
			slice[cur-1] = pool[pre : cur*size]
		}
		bp.memList[size] = slice
	}
	return bp
}

func (bp *bufferpool) newP(buf []byte) *P {
	return &P{}
}

// length <= 2K
// length >= 4M
func (bp *bufferpool) Alloc() []byte {
	/*return make([]byte, BUF_MIN_LEN)*/
	//if bp.stdList.Len() > 0 {
	//if buffer, ok := bp.stdList.Remove(bp.stdList.Front()).([]byte); ok {
	//return buffer
	//}
	/*}*/
	return nil
}

func (bp *bufferpool) AllocLarge(length int) []byte {
	/*if length <= BUF_MIN_LEN {*/
	//return bp.Alloc()
	//}
	//if length >= 4*1024*1024 {
	//return make([]byte, length+(^length&(BUF_MIN_LEN-1)+1)&(BUF_MIN_LEN-1))
	//}
	//if _, ok := bp.largeList[length+(^length&(BUF_MIN_LEN-1)+1)&(BUF_MIN_LEN-1)]; ok {
	/*}*/
	return nil
}

func (bp *bufferpool) Release(buffer []byte) {
}
