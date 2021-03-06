package keyvadb

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
)

type Hash [HashSize]byte

type HashSlice []Hash

func (s HashSlice) Len() int           { return len(s) }
func (s HashSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s HashSlice) Less(i, j int) bool { return s[i].Compare(s[j]) < 0 }
func (s HashSlice) Sort()              { sort.Sort(s) }
func (s HashSlice) IsSorted() bool     { return sort.IsSorted(s) }
func (s HashSlice) String() string     { return dumpWithTitle("Hashes", s, 0) }

func NewHash(s string) (*Hash, error) {
	b, err := hex.DecodeString(s)
	switch {
	case err != nil:
		return nil, err
	case len(b) != HashSize:
		return nil, fmt.Errorf("Hash wrong length")
	default:
		var hash Hash
		copy(hash[:], b)
		return &hash, nil
	}
}

func MustHash(s string) Hash {
	hash, err := NewHash(s)
	if err != nil {
		panic(err)
	}
	return *hash
}

// Clamps and ensures value is absolute
func newHash(n *big.Int) Hash {
	var h Hash
	n.Abs(n)
	if n.Cmp(maxBig) >= 0 {
		h = LastHash
	} else {
		b := n.Bytes()
		copy(h[HashSize-len(b):], b)
	}
	return h
}

func (h Hash) Empty() bool {
	return h == EmptyKey
}

func (a Hash) Equals(b Hash) bool {
	return a == b
}

func (a Hash) Compare(b Hash) int {
	return bytes.Compare(a[:], b[:])
}

func (a Hash) Less(b Hash) bool {
	return a.Compare(b) < 0
}

func (a Hash) Greater(b Hash) bool {
	return a.Compare(b) > 0
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

// Returns multiple of stride and distance.
// Rounds up and down if the extents are matched
func (a Hash) NearestStride(start, stride, halfStride *big.Int, entries int64) (int, Hash) {
	quot := a.Big()
	rem := big.NewInt(0)
	quot.Sub(quot, start).QuoRem(quot, stride, rem)
	i := quot.Int64()
	// fmt.Println(i, a, newHash(rem), newHash(halfStride))
	switch {
	case i == 0:
		// Shift up
		i++
		rem.Sub(a.Big(), stride).Sub(rem, start)
	case i < entries && rem.Cmp(halfStride) > 0:
		// Round up
		i++
		rem.Sub(stride, rem)
	}
	// fmt.Println(i, newHash(rem))
	return int(i), newHash(rem)
}

func (a Hash) Stride(b Hash, n int64) Hash {
	return a.Distance(b).Divide(n)
}

func (h Hash) String() string {
	return fmt.Sprintf("%X", h[:])
}
