package peer

import (
	"log"

	cnf "github.com/halivor/frontend/config"
	mw "github.com/halivor/frontend/middleware"
)

type Manager interface {
	Add(p *Peer)
	Del(p *Peer)

	Transfer(message []byte)
}

type manager struct {
	peers map[uint64]*Peer
	rooms map[uint32]map[*Peer]struct{}
	cqid  mw.QId
	pqid  mw.QId

	mw.Middleware
	*log.Logger
}

func NewManager(mdw mw.Middleware) (pm *manager) {
	defer func() {
		pm.cqid = pm.Bind(mw.T_TRANSFER, "up", mw.A_PRODUCE, pm)
		pm.pqid = pm.Bind(mw.T_TRANSFER, "down", mw.A_CONSUME, pm)
	}()
	return &manager{
		peers:      make(map[uint64]*Peer),
		rooms:      make(map[uint32]map[*Peer]struct{}),
		Middleware: mdw,
		Logger:     cnf.NewLogger("[pm] "),
	}
}

func (pm *manager) Add(p *Peer) {
	// 超时重连
	if pp, ok := pm.peers[p.id]; ok {
		pp.Release()
	}
	pm.peers[p.id] = p

	if _, ok := pm.rooms[p.room]; !ok {
		pm.rooms[p.room] = make(map[*Peer]struct{}, 1024)
	}
	pm.rooms[p.room][p] = struct{}{}
}

func (pm *manager) Del(p *Peer) {
	delete(pm.peers, p.id)
}

func (pm *manager) unicast(message interface{}) {
}

func (pm *manager) broadcast(message interface{}) {
	if msg, ok := message.([]byte); ok {
		for _, p := range pm.peers {
			p.Send(msg)
		}
	}
}

func (pm *manager) Consume(message interface{}) {
	if buf, ok := message.([]byte); ok {
		pm.Println("consume", string(buf))
	}
}

func (pm *manager) Transfer(message []byte) {
	pm.Produce(mw.T_TRANSFER, pm.pqid, message)
}
