package types

import "encoding/binary"

const (
	// ModuleName defines the module name
	ModuleName = "icaauth"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_icaauth"

	// Version defines the current version the IBC module supports
	Version = "icaauth-1"
)

// prefix bytes for the icaauth persistent store
const (
	paramsKey = iota + 1
	prefixPacketResult
)

// KVStore key prefixes
var (
	// ParamsKey is the key for params.
	ParamsKey = []byte{paramsKey}

	KeyPrefixPacketResult = []byte{prefixPacketResult}
)

// SequenceToPacketResultKey defines the store key for sequence to packet result
func SequenceToPacketResultKey(sequence uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, sequence)
	return append(KeyPrefixPacketResult, b...)
}
