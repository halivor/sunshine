package middleware

type Middleware interface {
	Register(ctg Category, i interface{}) bool
}

type middleware struct {
	category map[Category]interface{}
}

func New() *middleware {
	return &middleware{
		category: make(map[Category]interface{}),
	}
}

func (m *middleware) Register(ctg Category, i interface{}) bool {
	if !ctg.check(i) {
		return false
	}
	m.category[ctg] = i
	return true
}

func (m *middleware) Unicast(ctg Category, id uint64, message interface{}) {
	if c, ok := m.category[ctg]; ok {
		if i, ok := c.(Unicast); ok {
			i.Unicast(id, message)
		}
	}
}

func (m *middleware) Broadcast(ctg Category, message interface{}) {
	if c, ok := m.category[ctg]; ok {
		if i, ok := c.(Broadcast); ok {
			i.Broadcast(message)
		}
	}
}
