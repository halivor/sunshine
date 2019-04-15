package transfer

import (
	mw "github.com/halivor/frontend/middleware"
)

type transfer struct {
	id  mw.QId
	qid map[string]mw.QId
	cs  map[mw.QId][]mw.Consumer
}

func init() {
	mw.Register(mw.T_TRANSFER, New)
}

func New() *transfer {
	return &transfer{
		id:  10000,
		qid: make(map[string]mw.QId),
		cs:  make(map[mw.QId][]mw.Consumer),
	}
}

func (t *transfer) Bind(q string, a mw.Action, c interface{}) mw.QId {
	id, ok := t.qid[q]
	if !ok {
		t.id++
		id = t.id
		t.qid[q] = id
		t.cs[id] = make([]mw.Consumer, 0)
	}
	if cc, ok := c.(mw.Consumer); ok && a == mw.A_CONSUME {
		t.cs[id] = append(t.cs[id], cc)
	}
	return id
}

func (t *transfer) Produce(id mw.QId, message interface{}) {
}

func (t *transfer) GetQId(q string) mw.QId {
	if id, ok := t.qid[q]; ok {
		return id
	}
	return -1
}
