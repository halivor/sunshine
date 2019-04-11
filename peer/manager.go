package peer

type Manage interface {
	Add(id uint64, p *Peer)
	Del(id uint64)

	Unicast(id uint64, message interface{})
	Broadcast(message interface{})
}

type manager struct {
	peers map[uint64]*Peer
}

func NewManager() *manager {
	return &manager{
		peers: make(map[uint64]*Peer),
	}
}

func (m *manager) Add(id uint64, p *Peer) {
	if pp, ok := m.peers[id]; ok {
		pp.DelEvent(pp)
		p.Release()
	}
	m.peers[id] = p
}

func (m *manager) Del(id uint64) {
}

func (m *manager) Unicast(id uint64, message interface{}) {
}

func (m *manager) Broadcast(message interface{}) {
}
