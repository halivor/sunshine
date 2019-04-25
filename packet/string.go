package packet

import (
	"log"
	"strconv"
)

type SHeader struct {
	ver [4]byte
	typ uint16
	opt uint16
	cmd [4]byte
	seq [8]byte
	len [4]byte
	res [8]byte
}

func (h *SHeader) Ver() int {
	v, e := strconv.Atoi(string(h.ver[:]))
	if e != nil {
		log.Panicln(e)
	}
	return v
}

func (h *SHeader) Cmd() int {
	c, e := strconv.Atoi(string(h.cmd[:]))
	if e != nil {
		log.Panicln(e)
	}
	return c
}

func (h *SHeader) Seq() int {
	s, e := strconv.Atoi(string(h.seq[:]))
	if e != nil {
		log.Panicln(e)
	}
	return s
}

func (h *SHeader) Len() int {
	l, e := strconv.Atoi(string(h.len[:]))
	if e != nil {
		log.Panicln(e)
	}
	return l
}
