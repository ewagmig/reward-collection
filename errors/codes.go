package errors

// Code is 32-bit error code
type Code uint32

// Common: 00000000--000FFFFF
const (
	EthCallError            Code = 0x00000000
	CongressGetValsError Code = 0x00000001
	UnknownError			Code = 0x00000002
	DatabaseError           Code = 0x00000003
	InvalidJSONBody			Code = 0x00000004
	IDNotFound				Code = 0x00000005
	Forbidden				Code = 0x00000006
)
