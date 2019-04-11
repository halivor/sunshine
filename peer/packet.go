package peer

// message type :
// b - binary
// j - json
// p - protobuf
// s - string
type packet struct {
	mt   byte
	uid  uint64
	data string
}
