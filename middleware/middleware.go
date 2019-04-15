package middleware

type Middleware interface {
	Bind(id MwId, q string, a Action, c interface{}) QId
	Produce(mwid MwId, qid QId, msg interface{})
	GetQId(id MwId, q string) QId
}

type middleware struct {
	category map[TypeID][]Consume
	mwers    map[MwId]mwer
}

func New() *middleware {
	cs := make(map[MwId]mwer, 32)
	for id, cm := range components {
		if f, ok := cm.(func() mwer); ok {
			cs[id] = f()
		}
	}
	return &middleware{
		category: make(map[TypeID][]Consume),
		mwers:    cs,
	}
}

func (m *middleware) GetQId(id MwId, q string) QId {
	if mw, ok := m.mwers[id]; ok {
		return mw.GetQId(q)
	}
	return -1
}

func (m *middleware) Bind(mwid MwId, q string, a Action, c interface{}) QId {
	if mw, ok := m.mwers[mwid]; ok {
		return mw.Bind(q, a, c)
	}
	return -1
}

func (m *middleware) Produce(id MwId, qid QId, msg interface{}) {
	if mw, ok := m.mwers[id]; ok {
		mw.Produce(qid, msg)
	}
}
