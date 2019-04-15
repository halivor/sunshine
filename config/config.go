package config

import (
	"log"
	"os"
	"time"
)

const (
	MaxEvents = 128
	MaxConns  = 8 * 1024

	BUF_MIN_LEN = 4096
	BUF_MAX_LEN = 4 * 1024 * 1024
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
func NewLogger(prefix string) *log.Logger {
	return log.New(os.Stderr, prefix, log.LstdFlags)
}
