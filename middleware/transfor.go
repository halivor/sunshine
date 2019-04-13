package middleware

type transfer struct {
	id  QId
	qid map[string]QId
	cs  map[QId][]consumer
}

func newTransfor() *transfer {
	return &transfer{
		id:  10000,
		qid: make(map[string]QId),
		cs:  make(map[QId][]consumer),
	}
}

func (t *transfer) Bind(q string, a Action, c interface{}) QId {
	id, ok := t.qid[q]
	if !ok {
		t.id++
		id = t.id
		t.qid[q] = id
		t.cs[id] = make([]consumer, 0)
	}
	if cc, ok := c.(consumer); ok && a == A_CONSUME {
		t.cs[id] = append(t.cs[id], cc)
	}
	return id
}

func (t *transfer) Produce(id QId, message interface{}) {
}

func (t *transfer) GetQId(q string) QId {
	if id, ok := t.qid[q]; ok {
		return id
	}
	return -1
}
