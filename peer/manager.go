package peer

import (
	"log"
	"unsafe"

	cnf "github.com/halivor/frontend/config"
	pkt "github.com/halivor/frontend/packet"
	mw "github.com/halivor/goevent/middleware"
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
		pm.Bind(mw.T_TRANSFER, "dchat", mw.A_CONSUME, pm)
		pm.Bind(mw.T_TRANSFER, "dbullet", mw.A_CONSUME, pm)
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
	pm.Println("add", p.Uid)
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
	if u, ok := pm.peers[uid]; ok {
		u.Send(message[pkt.HLen:])
	}
}

func (pm *manager) broadcast(message []byte) {
	for _, p := range pm.peers {
		p.Send(message[pkt.HLen:])
	}
}

func (pm *manager) Consume(message interface{}) interface{} {
	if data, ok := message.([]byte); ok {
		u := (*pkt.SHeader)(unsafe.Pointer(&data[pkt.HLen]))
		h := (*pkt.Header)(unsafe.Pointer(&data[0]))
		switch u.Cmd() {
		case pkt.C_BULLET:
			//pm.Println("consume bullet", string(data[pkt.HLen:]))
			pm.broadcast(data)
		case pkt.C_CHAT:
			pm.Println("consume chat", string(data[pkt.HLen:]))
			pm.unicast(h.Uid, data)
		default:
			pm.Println("consume default", string(data[pkt.HLen:]))
			pm.Println(string(data[pkt.HLen:]))
		}

	}
	return nil
}

func (pm *manager) Transfer(message []byte) {
	if message != nil {
		pm.Println("produce", string(message))
		pm.Produce(mw.T_TRANSFER, pm.uqid, message)
	}
}
