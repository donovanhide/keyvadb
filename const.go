package keyva

import "math/big"

const (
	HashSize   = 32
	ItemCount  = 8
	ChildCount = ItemCount + 1
)

var (
	EmptyChild = uint64(0)
	EmptyItem  = MustHash("0000000000000000000000000000000000000000000000000000000000000000")
	FirstHash  = MustHash("0000000000000000000000000000000000000000000000000000000000000001")
	LastHash   = MustHash("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	minBig     = big.NewInt(0).SetBytes(FirstHash[:])
	maxBig     = big.NewInt(0).SetBytes(LastHash[:])
)
