package keyvadb

import (
	"math"

	"math/big"
)

const (
	HashSize       = 32
	EmptyChild     = uint64(0)
	SyntheticChild = math.MaxUint64
)

var (
	EmptyKey  = MustHash("0000000000000000000000000000000000000000000000000000000000000000")
	FirstHash = MustHash("0000000000000000000000000000000000000000000000000000000000000001")
	LastHash  = MustHash("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	minBig    = big.NewInt(0).SetBytes(FirstHash[:])
	maxBig    = big.NewInt(0).SetBytes(LastHash[:])
)
