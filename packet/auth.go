package packet

import (
	"fmt"
	"strconv"
)

type Auth struct {
	ver  [2]byte
	typ  uint8
	opt  uint8
	uid  [8]byte
	cid  [8]byte
	sign [8]byte
	len  [4]byte
}

const (
	AUTH_SUCC = "10S0100100000000000700000000success"
	PING      = "10S1101012345678000400000000PING"
	PONG      = "10S1101112345678000400000000PONG"
)

var Pong, AuthSucc *P

func init() {
	AuthSucc = Alloc(len(AUTH_SUCC))
	copy(AuthSucc.Buf, AUTH_SUCC)
	Pong = Alloc(len(PONG))
	copy(Pong.Buf, PONG)
}

func (a *Auth) Ver() int {
	v, e := strconv.Atoi(string(a.ver[:]))
	if e != nil {
		plog.Panic(e)
	}
	return v
}

func (a *Auth) Uid() int {
	u, e := strconv.Atoi(string(a.uid[:]))
	if e != nil {
		plog.Panic(e)
	}
	return u
}

func (a *Auth) Cid() int {
	c, e := strconv.Atoi(string(a.cid[:]))
	if e != nil {
		plog.Panic(e)
	}
	return c
}

func (a *Auth) Sign() int {
	s, e := strconv.Atoi(string(a.sign[:]))
	if e != nil {
		plog.Panic(e)
	}
	return s
}

func (a *Auth) Len() int {
	l, e := strconv.Atoi(string(a.len[:]))
	if e != nil {
		plog.Panic(e)
	}
	return l
}

func (a *Auth) String() string {
	return fmt.Sprintf("%d|%d|%d|%d|", a.Ver(), a.Uid(), a.Cid(), a.Len())
}
