package peer

import (
	"log"
	"unsafe"

	mw "github.com/halivor/goevent/middleware"
	cnf "github.com/halivor/sunshine/config"
	pkt "github.com/halivor/sunshine/packet"
)

type Manager interface {
	Add(p *Peer)
	Del(p *Peer)

	Transfer(message []byte)
}

type manager struct {
	peers map[*Peer]struct{}            // 仅在全体消息广播时使用
	users map[uint32]map[*Peer]struct{} // 仅在指定用户ID发送时使用
	rooms map[uint32]map[*Peer]struct{} // 仅在指定房间发送时使用
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
		peers:      make(map[*Peer]struct{}, 1024),
		users:      make(map[uint32]map[*Peer]struct{}, 1024),
		rooms:      make(map[uint32]map[*Peer]struct{}, 1024),
		Middleware: mdw,
		Logger:     cnf.NewLogger("[pm] "),
	}
}

// uid = 0 匿名用户
// rid = 0 非房间用户
func (pm *manager) Add(p *Peer) {
	// 超时重连
	pm.peers[p] = struct{}{}
	pm.Println("add", p.header.Uid)
	if _, ok := pm.users[p.header.Uid]; !ok {
		pm.users[p.header.Uid] = make(map[*Peer]struct{}, 16)
	}
	pm.users[p.header.Uid][p] = struct{}{}

	if _, ok := pm.rooms[p.header.Cid]; !ok {
		pm.rooms[p.header.Cid] = make(map[*Peer]struct{}, 128)
	}
	pm.rooms[p.header.Cid][p] = struct{}{}
}

func (pm *manager) Del(p *Peer) {
	delete(pm.peers, p)
	if ps, ok := pm.users[p.header.Uid]; ok {
		delete(ps, p)
	}
	if cp, ok := pm.users[p.header.Cid]; ok {
		delete(cp, p)
	}
}

func (pm *manager) unicast(uid uint32, pd *pkt.P) {
	if us, ok := pm.users[uid]; ok {
		for usr, _ := range us {
			usr.Send(pd)
		}
	}
}

func (pm *manager) broadcast(cid uint32, pd *pkt.P) {
	switch {
	case cid > 0:
		if r, ok := pm.rooms[cid]; ok {
			for p, _ := range r {
				p.Send(pd)
			}
		}
	default:
		for p, _ := range pm.peers {
			p.Send(pd)
		}
	}
}

func (pm *manager) Consume(message interface{}) interface{} {
	if p, ok := message.(*pkt.P); ok {
		h := (*pkt.Header)(unsafe.Pointer(&p.Buf[0]))
		switch {
		case h.Uid > 0:
			pm.unicast(h.Uid, p)
		default:
			pm.broadcast(h.Cid, p)
		}
	}
	return nil
}

func (pm *manager) Transfer(message []byte) {
	if message != nil {
		//pm.Println("produce", string(message))
		pm.Produce(mw.T_TRANSFER, pm.uqid, message)
	}
}
