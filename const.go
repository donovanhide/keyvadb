package keyvadb

import (
	"math"
	"math/big"
)

const (
	HashSize       = 32
	EmptyChild     = NodeId(0)
	SyntheticValue = ValueId(math.MaxUint64)
)

var (
	EmptyKey  = MustHash("0000000000000000000000000000000000000000000000000000000000000000")
	FirstHash = MustHash("0000000000000000000000000000000000000000000000000000000000000001")
	LastHash  = MustHash("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	minBig    = big.NewInt(0).SetBytes(FirstHash[:])
	maxBig    = big.NewInt(0).SetBytes(LastHash[:])
)
