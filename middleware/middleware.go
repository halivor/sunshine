package middleware

type Middleware interface {
	Bind(category string, a Action, i interface{}) TypeID
	Produce(id TypeID, message interface{})
}

type middleware struct {
	category map[TypeID][]Consume
}

func New() *middleware {
	return &middleware{
		category: make(map[TypeID][]Consume),
	}
}

func (m *middleware) Bind(category string, a Action, i interface{}) TypeID {
	if _, ok := tId[category]; !ok {
		id++
		tId[category] = id
	}
	id := tId[category]
	switch a {
	case A_PRODUCER:
		if _, ok := m.category[id]; !ok {
			m.category[id] = make([]Consume, 0, MAX_CONSUMER)
		}
	case A_CONSUMER:
		if cs, ok := m.category[id]; ok {
			if consumer, ok := i.(Consume); ok {
				m.category[id] = append(cs, consumer)
			}
		} else {
			if consumer, ok := i.(Consume); ok {
				m.category[id] = append(make([]Consume, 0, MAX_CONSUMER), consumer)
			}
		}
	}
	return id
}

func (m *middleware) Produce(id TypeID, message interface{}) {
	if cs, ok := m.category[id]; ok {
		for _, c := range cs {
			c.Consume(message)
		}
	}
}
