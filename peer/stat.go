package peer

type peerStat int32

const (
	PS_INIT peerStat = 1 + iota
	PS_ESTAB
	PS_NORMAL
	PS_END
)

const (
	MAX_QUEUE_SIZE = 256
)

var (
	gSeq int32 = 10 * 1000 * 1000
)
