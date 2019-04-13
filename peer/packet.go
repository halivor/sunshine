package peer

// message type :
// b - binary
// j - json
// p - protobuf
// s - string
type packet struct {
	version [1]byte
	length  [4]byte
	uid     [12]byte
	room    [8]byte
	mac     [32]byte
	seq     [12]byte
	body    []byte
}
