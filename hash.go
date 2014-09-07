package keyva

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"sort"

	"fmt"
)

type Hash [HashSize]byte

type HashSlice []Hash

func (s HashSlice) Len() int           { return len(s) }
func (s HashSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s HashSlice) Less(i, j int) bool { return s[i].Compare(s[j]) < 0 }
func (s HashSlice) Sort()              { sort.Sort(s) }
func (s HashSlice) IsSorted() bool     { return sort.IsSorted(s) }
func (s HashSlice) String() string     { return dumpWithTitle("Hashes", s, 0) }

func MustHash(s string) Hash {
	b, err := hex.DecodeString(s)
	switch {
	case err != nil:
		panic(err)
	case len(b) != 32:
		panic("Hash wrong length")
	default:
		var hash Hash
		copy(hash[:], b)
		return hash
	}
}

func newHash(n *big.Int) Hash {
	n.Abs(n)
	switch {
	case n.Cmp(maxBig) >= 0:
		return LastHash
	case n.Cmp(minBig) <= 0:
		return FirstHash
	default:
		b := n.Bytes()
		var h Hash
		copy(h[HashSize-len(b):], b)
		return h
	}
}

func (a Hash) Compare(b Hash) int {
	return bytes.Compare(a[:], b[:])
}

func (a Hash) Equals(b Hash) bool {
	return a.Compare(b) == 0
}

func (h Hash) Big() *big.Int {
	return big.NewInt(0).SetBytes(h[:])
}

func (a Hash) Distance(b Hash) Hash {
	diff := a.Big()
	return newHash(diff.Sub(diff, b.Big()))
}

// Returns true if a is closer to b than c
func (a Hash) Closest(b, c Hash) bool {
	return a.Distance(b).Compare(a.Distance(c)) <= 0
}

func (a Hash) Add(b Hash) Hash {
	sum := a.Big()
	return newHash(sum.Add(sum, b.Big()))
}

func (a Hash) Divide(n int64) Hash {
	quot := a.Big()
	return newHash(quot.Div(quot, big.NewInt(n)))
}

func (a Hash) Stride(b Hash, n int64) Hash {
	return a.Distance(b).Divide(n)
}

func (h Hash) String() string {
	return fmt.Sprintf("%X", h[:])
}
