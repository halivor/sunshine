package peer

import (
	"log"
	"unsafe"

	cnf "github.com/halivor/frontend/config"
	mw "github.com/halivor/frontend/middleware"
	pkt "github.com/halivor/frontend/packet"
)

type Manager interface {
	Add(p *Peer)
	Del(p *Peer)

	Transfer(message []byte)
}

type manager struct {
	peers map[uint32]*Peer
	rooms map[uint32]map[*Peer]struct{}
	uqid  mw.QId

	mw.Middleware
	*log.Logger
}

func NewManager(mdw mw.Middleware) (pm *manager) {
	defer func() {
		pm.uqid = pm.Bind(mw.T_TRANSFER, "up", mw.A_PRODUCE, pm)
		pm.Bind(mw.T_TRANSFER, "down", mw.A_CONSUME, pm)
	}()
	return &manager{
		peers:      make(map[uint32]*Peer),
		rooms:      make(map[uint32]map[*Peer]struct{}),
		Middleware: mdw,
		Logger:     cnf.NewLogger("[pm] "),
	}
}

func (pm *manager) Add(p *Peer) {
	// 超时重连
	if pp, ok := pm.peers[p.Uid]; ok {
		pp.Release()
	}
	pm.peers[p.Uid] = p

	if _, ok := pm.rooms[p.Cid]; !ok {
		pm.rooms[p.Cid] = make(map[*Peer]struct{}, 1024)
	}
	pm.rooms[p.Cid][p] = struct{}{}
}

func (pm *manager) Del(p *Peer) {
	delete(pm.peers, p.Uid)
}

func (pm *manager) unicast(uid uint32, message []byte) {
	if up, ok := pm.peers[uid]; ok {
		up.Send(message)
	}
}

func (pm *manager) broadcast(message []byte) {
	for _, p := range pm.peers {
		p.Send(message)
	}
}

func (pm *manager) Consume(message interface{}) interface{} {
	if buf, ok := message.([]byte); ok {
		uh := (*pkt.UHeader)(unsafe.Pointer(&buf[pkt.HLen]))
		switch uh.Ver {
		case 1:
			pm.broadcast(buf)
		case 3:
			pm.unicast(uh.Uid, buf)
		}

	}
	return nil
}

func (pm *manager) Transfer(message []byte) {
	pm.Println("produce", string(message))
	pm.Produce(mw.T_TRANSFER, pm.uqid, message)
}
