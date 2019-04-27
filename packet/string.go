package packet

import (
	"log"
	"strconv"
)

type SHeader struct {
	ver [2]byte
	typ uint8
	opt uint8
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

func (h *SHeader) Opt() int {
	return int(h.opt)
}

func (h *SHeader) Cmd() int {
	c, e := strconv.Atoi(string(h.cmd[:]))
	if e != nil {
		log.Panicln(e)
	}
	return c
}

func (h *SHeader) Seq() uint64 {
	s, e := strconv.ParseUint(string(h.seq[:]), 10, 64)
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
