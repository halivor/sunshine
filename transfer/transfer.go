package transfer

import (
	mw "github.com/halivor/goutility/middleware"
)

type transfer struct {
	*mw.MwTmpl
}

func init() {
	mw.Register(mw.T_TRANSFER, New)
}

func New() mw.Mwer {
	return &transfer{
		MwTmpl: mw.NewTmpl(),
	}
}
