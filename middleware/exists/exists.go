package exists

import (
	"log"

	cnf "github.com/halivor/frontend/config"
	mw "github.com/halivor/frontend/middleware"
)

type exists struct {
	*mw.MwTmpl
	*log.Logger
}

func init() {
	mw.Register(mw.T_TRANSFER, New)
}

func New() mw.Mwer {
	return &exists{
		MwTmpl: mw.NewTmpl(),
		Logger: cnf.NewLogger("[exists] "),
	}
}

func (t *exists) Produce(id mw.QId, message interface{}) interface{} {
	if cs, ok := t.Cs[id]; ok {
		for _, c := range cs {
			if i := c.Consume(message); i != nil {
				return i
			}
		}
	}
	return nil
}
