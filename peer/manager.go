package peer

import (
	mw "github.com/halivor/frontend/middleware"
)

type Manager interface {
	Add(id uint64, p *Peer)
	Del(id uint64)
}

type manager struct {
	peers map[uint64]*Peer
	cqid  mw.QId
	pqid  mw.QId

	mw.Middleware
}

func NewManager(mdw mw.Middleware) (pm *manager) {
	defer func() {
		pm.Bind(mw.T_TRANSFER, "up", mw.A_PRODUCE, pm)
		pm.Bind(mw.T_TRANSFER, "down", mw.A_CONSUME, pm)
	}()
	return &manager{
		peers:      make(map[uint64]*Peer),
		Middleware: mdw,
		cqid:       mdw.GetQId(mw.T_TRANSFER, "up"),
		pqid:       mdw.GetQId(mw.T_TRANSFER, "down"),
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

func (pm *manager) Consume(message interface{}) {
}

func (pm *manager) Produce(message interface{}) {
	pm.Middleware.Produce(mw.T_TRANSFER, pm.pqid, message)
}
