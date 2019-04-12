package peer

type peerStat int32

const (
	PS_INIT peerStat = 1 + iota
	PS_ESTAB
	PS_NORMAL
	PS_END
)
