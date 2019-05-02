package bufferpool

var gbp *bufferpool

func init() {
}

func Alloc(length int) ([]byte, error) {
	return gbp.Alloc(length)
}
