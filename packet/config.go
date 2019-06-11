package packet

import (
	log "github.com/halivor/goutility/logger"
)

var (
	pl log.Logger
)

func init() {
	pl, _ = log.New("/data/logs/sunshine/packet.log", "", log.LstdFlags, log.TRACE)
}
