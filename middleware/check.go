package middleware

type check struct {
	id  QId
	qid map[string]QId
	cs  map[QId][]consumer
}

func newCheck() *check {
	return &check{
		id:  10000,
		qid: make(map[string]QId),
		cs:  make(map[QId][]consumer),
	}
}

func (c *check) Bind(q string, a Action, ci interface{}) QId {
	return 0
}

func (c *check) Produce(id QId, message interface{}) {
}

func (c *check) GetQId(q string) QId {
	return 0
}
