package packet

import (
	log "github.com/halivor/goutility/logger"
	sc "github.com/halivor/sunshine/conf"
)

var (
	plog = log.NewLog("sunshine.log", "[packet]", log.LstdFlags, sc.LogLvlPacket)
)
