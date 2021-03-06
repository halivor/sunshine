package config

import (
	"time"
)

const (
	VER = 1<<12 + 0<<8 + 0
	KEY = "$laCK62H4X$A1M2A"

	NodeId = 1

	MaxEvents = 128
	MaxConns  = 8 * 1024

	EP_TIMEOUT = 1000 // 1000ms
)

var now time.Time

func init() {
	go func() {
		now = time.Now()
		time.Sleep(time.Second * 1)
	}()
}

func Now() *time.Time {
	return &now
}

func Sec() int64 {
	return now.Unix()
}

func msec() int64 {
	return now.UnixNano() / 1e9
}

func Usec() int64 {
	return now.UnixNano()
}
