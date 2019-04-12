package config

import (
	"log"
	"os"
)

const (
	MaxEvents = 128
	MaxConns  = 1024 * 1024

	BUF_MIN_LEN = 4096
	BUF_MAX_LEN = 4 * 1024 * 1024
)

func NewLogger(prefix string) *log.Logger {
	return log.New(os.Stderr, prefix, log.LstdFlags)
}
