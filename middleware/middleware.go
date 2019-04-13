package middleware

type Middleware interface {
	Bind(id MwId, q string, a Action, c interface{}) TypeID
	Produce(mwid MwId, qid QId, msg interface{})
	GetQId(id MwId, q string) QId
}

type middleware struct {
	category map[TypeID][]Consume
}

func New() *middleware {
	return &middleware{
		category: make(map[TypeID][]Consume),
	}
}

func (m *middleware) GetQId(id MwId, q string) QId {
	if mw, ok := mws[id]; ok {
		return mw.GetQId(q)
	}
	return -1
}

func (m *middleware) Bind(mwid MwId, q string, a Action, c interface{}) QId {
	if mw, ok := mws[mwid]; ok {
		return mw.Bind(q, a, c)
	}
	return -1
}

func (m *middleware) Produce(id MwId, qid QId, msg interface{}) {
	if mw, ok := mws[id]; ok {
		mw.Produce(qid, msg)
	}
}
