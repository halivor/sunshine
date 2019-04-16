package check

import (
	mw "github.com/halivor/frontend/middleware"
)

type check struct {
	id  mw.QId
	qid map[string]mw.QId
	cs  map[mw.QId][]mw.Consumer
}

func init() {
	mw.Register(mw.T_CHECK, New)
}

func New() Mwer {
	return &check{
		id:  10000,
		qid: make(map[string]mw.QId),
		cs:  make(map[mw.QId][]mw.Consumer),
	}
}

func (c *check) Bind(q string, a mw.Action, ci interface{}) mw.QId {
	return 0
}

func (c *check) Produce(id mw.QId, message interface{}) {
}

func (c *check) GetQId(q string) mw.QId {
	return 0
}
