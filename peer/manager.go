package peer

import (
	mw "github.com/halivor/frontend/middleware"
)

type Manager interface {
	Add(id uint64, p *Peer)
	Del(id uint64)

	Produce(message interface{})
}

type manager struct {
	peers map[uint64]*Peer
	ctid  mw.TypeID
	ptid  mw.TypeID

	mw.Middleware
}

func NewManager(mdw mw.Middleware) (pm *manager) {
	defer func() {
		pm.Bind("up", mw.A_PRODUCER, pm)
		pm.Bind("down", mw.A_CONSUMER, pm)
	}()
	return &manager{
		peers:      make(map[uint64]*Peer),
		Middleware: mdw,
		ctid:       mw.TId("down"),
		ptid:       mw.TId("up"),
	}
}

func (pm *manager) Add(id uint64, p *Peer) {
	if pp, ok := pm.peers[id]; ok {
		pp.DelEvent(pp)
		p.Release()
	}
	pm.peers[id] = p
}

func (pm *manager) Del(id uint64) {
}

func (pm *manager) Unicast(id uint64, message interface{}) {
}

func (pm *manager) Broadcast(message interface{}) {
}

func (pm *manager) Produce(message interface{}) {
	pm.Middleware.Produce(pm.ptid, message)
}

func (pm *manager) Consume(message interface{}) {
}
