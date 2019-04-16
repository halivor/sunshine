package packet

type Type int8

const (
	T_JSON Type = 1 << iota
	T_STRING
	T_BINARY
	T_RPOTO
)
