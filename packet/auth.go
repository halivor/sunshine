package packet

import (
	"fmt"
	"log"
	"strconv"
)

type Auth struct {
	ver  [4]byte
	uid  [8]byte
	cid  [8]byte
	sign [8]byte
	len  [4]byte
}

func (a *Auth) Ver() int {
	v, e := strconv.Atoi(string(a.ver[:]))
	if e != nil {
		log.Panicln(e)
	}
	return v
}

func (a *Auth) Uid() int {
	u, e := strconv.Atoi(string(a.uid[:]))
	if e != nil {
		log.Panicln(e)
	}
	return u
}

func (a *Auth) Cid() int {
	c, e := strconv.Atoi(string(a.cid[:]))
	if e != nil {
		log.Panicln(e)
	}
	return c
}

func (a *Auth) Sign() int {
	s, e := strconv.Atoi(string(a.sign[:]))
	if e != nil {
		log.Panicln(e)
	}
	return s
}

func (a *Auth) Len() int {
	l, e := strconv.Atoi(string(a.len[:]))
	if e != nil {
		log.Panicln(e)
	}
	return l
}

func (a *Auth) String() string {
	return fmt.Sprintf("%d|%d|%d|%d|", a.Ver(), a.Uid(), a.Cid(), a.Len())
}
