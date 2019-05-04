package bufferpool

var gbp *bufferpool

func init() {
	gbp = New()
}

func Alloc(length int) []byte {
	b, _ := gbp.Alloc(length)
	return b
}

func Release(buf []byte) {
	gbp.Release(buf)
}
