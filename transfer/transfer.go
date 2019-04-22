package transfer

import (
	"log"

	cnf "github.com/halivor/frontend/config"
	mw "github.com/halivor/goevent/middleware"
)

type transfer struct {
	*mw.MwTmpl
	*log.Logger
}

func init() {
	mw.Register(mw.T_TRANSFER, New)
}

func New() mw.Mwer {
	return &transfer{
		MwTmpl: mw.NewTmpl(),
		Logger: cnf.NewLogger("[transfer] "),
	}
}
