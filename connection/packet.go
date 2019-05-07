package connection

import (
	"unsafe"

	bp "github.com/halivor/sunshine/bufferpool"
)

type buffer interface {
	Buffer() []byte
	Release()
}

type packet struct {
	buffer
	buf []byte
	pos int
}

func newPacket() *packet {
	buf := bp.Alloc(unsafe.Sizeof(packet{}))
	return nil
}

func resPacket(p *packet) {
}
