package transfer

import (
	"log"

	mw "github.com/halivor/goevent/middleware"
	cnf "github.com/halivor/sunshine/config"
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
