package bufferpool

var gbp *bufferpool

func init() {
}

func Alloc() []byte {
	return gbp.Alloc()
}

func AllocLarge(length int) []byte {
	return gbp.AllocLarge(length)
}

func Release(buffer []byte) {
	gbp.Release(buffer)
}
