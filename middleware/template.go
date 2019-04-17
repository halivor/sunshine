package middleware

import ()

type MwTmpl struct {
	Id  QId
	QId map[string]QId
	Cs  map[QId][]Consumer
}

func NewTmpl() *MwTmpl {
	return &MwTmpl{
		Id:  10000,
		QId: make(map[string]QId),
		Cs:  make(map[QId][]Consumer),
	}
}

func (t *MwTmpl) Bind(q string, a Action, c interface{}) QId {
	id, ok := t.QId[q]
	if !ok {
		t.Id++
		id = t.Id
		t.QId[q] = id
		t.Cs[id] = make([]Consumer, 0)
	}
	if cc, ok := c.(Consumer); ok && a == A_CONSUME {
		t.Cs[id] = append(t.Cs[id], cc)
	}
	return id
}

func (t *MwTmpl) Produce(id QId, message interface{}) interface{} {
	if cs, ok := t.Cs[id]; ok {
		for _, c := range cs {
			c.Consume(message)
		}
	}
	return nil
}

func (t *MwTmpl) GetQId(q string) QId {
	if id, ok := t.QId[q]; ok {
		return id
	}
	return -1
}
