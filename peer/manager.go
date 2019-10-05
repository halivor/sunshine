package peer

import (
	"fmt"
	"sync/atomic"
	"unsafe"

	log "github.com/halivor/goutility/logger"
	mw "github.com/halivor/goutility/middleware"
	sc "github.com/halivor/sunshine/conf"
	up "github.com/halivor/sunshine/packet"
)

type Manager interface {
	AddPeer(p *Peer)
	DelPeer(p *Peer)

	GenSeq(pd *up.P)
	Transfer(message []byte)
	BindMsg(id mw.MwId)
	BindStatOn(id mw.MwId, qid mw.QId)
	BindStatOff(id mw.MwId, qid mw.QId)
}

type mwc struct {
	mwid mw.MwId
	qid  mw.QId
}

type manager struct {
	peers map[*Peer]struct{}            // 仅在全体消息时使用
	users map[uint32]map[*Peer]struct{} // 仅在指定用户ID发送时使用
	rooms map[uint32]map[*Peer]struct{} // 仅在指定房间发送时使用

	mwcTfs   mwc
	mwcStOn  []mwc
	mwcStOff []mwc

	mw.Middleware
	log.Logger
}

func NewManager(mdw mw.Middleware) (pm *manager) {
	// TODO: 把绑定中间件放到外部

	return &manager{
		peers:      make(map[*Peer]struct{}, 1024),
		users:      make(map[uint32]map[*Peer]struct{}, 1024),
		rooms:      make(map[uint32]map[*Peer]struct{}, 1024),
		mwcStOn:    make([]mwc, 0, 8),
		mwcStOff:   make([]mwc, 0, 8),
		Middleware: mdw,
		Logger: log.NewLog("sunshine.peer.log", "[mgr]",
			log.LstdFlags, sc.LogLvlManager),
	}
}

// uid = 0 匿名用户
// rid = 0 非房间用户
func (pm *manager) AddPeer(p *Peer) {
	if p.header.Uid != 0 {
		if _, ok := pm.users[p.header.Uid]; !ok {
			pm.users[p.header.Uid] = make(map[*Peer]struct{}, 16)
		}
		pm.users[p.header.Uid][p] = struct{}{}
	}
	if p.header.Cid != 0 {
		if _, ok := pm.rooms[p.header.Cid]; !ok {
			pm.rooms[p.header.Cid] = make(map[*Peer]struct{}, 128)
		}
		pm.rooms[p.header.Cid][p] = struct{}{}
	}
	pm.peers[p] = struct{}{}
	pm.ChangeStatus(p.header.Uid, p.header.Cid, "on")
}

func (pm *manager) DelPeer(p *Peer) {
	if ps, ok := pm.users[p.header.Uid]; ok {
		delete(ps, p)
	}
	if cp, ok := pm.users[p.header.Cid]; ok {
		delete(cp, p)
	}

	delete(pm.peers, p)
	pm.ChangeStatus(p.header.Uid, p.header.Cid, "off")
}

func (pm *manager) unicast(uid uint32, pd *up.P) {
	switch us, ok := pm.users[uid]; {
	case !ok:
	case ok && len(us) != 0:
		for usr, _ := range us {
			usr.Send(pd)
		}
	case ok && len(us) == 0:
		delete(pm.users, uid)
	}
}

func (pm *manager) broadcast(cid uint32, pd *up.P) {
	switch {
	case cid > 0:
		switch r, ok := pm.rooms[cid]; {
		case !ok:
		case ok:
			for p, _ := range r {
				p.Send(pd.Reference())
			}
		}
	default:
		for p, _ := range pm.peers {
			p.Send(pd.Reference())
		}
	}
}

func (pm *manager) GenSeq(pd *up.P) {
	sh := (*up.SHeader)(unsafe.Pointer(&pd.Buf[0]))
	seq := fmt.Sprintf("%08d", atomic.AddInt32(&gSeq, 2))
	copy(sh.Seq[:], []byte(seq)[:cap(sh.Seq)])
}

func (pm *manager) BindMsg(id mw.MwId) {
	pm.mwcTfs.mwid = id
	pm.Bind(id, "down", mw.A_CONSUME, pm)
	pm.mwcTfs.qid = pm.Bind(id, "up", mw.A_PRODUCE, pm)
}

func (pm *manager) BindStatOn(id mw.MwId, qid mw.QId) {
	pm.mwcStOn = append(pm.mwcStOn, mwc{id, qid})
}

func (pm *manager) BindStatOff(id mw.MwId, qid mw.QId) {
	pm.mwcStOn = append(pm.mwcStOn, mwc{id, qid})
}

func (pm *manager) Consume(message interface{}) interface{} {
	if p, ok := message.(*up.P); ok {
		h := (*up.Header)(unsafe.Pointer(&p.Buf[0]))
		p.Buf = p.Buf[up.HLen:]
		p.Len -= up.HLen
		pm.GenSeq(p)
		pm.Trace("send", string(p.Buf))
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
		pm.Trace("produce", string(message))
		pm.Produce(pm.mwcTfs.mwid, pm.mwcTfs.qid, message)
	}
}

func (pm *manager) ChangeStatus(uid, cid uint32, status string) {
	switch status {
	case "on":
		for _, ms := range pm.mwcStOn {
			pm.Produce(ms.mwid, ms.qid, "")
		}
	case "off":
		for _, ms := range pm.mwcStOff {
			pm.Produce(ms.mwid, ms.qid, "")
		}
	}
}
