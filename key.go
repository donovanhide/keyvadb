package keyvadb

import (
	"fmt"
	"sort"
)

type Key struct {
	Hash Hash
	Id   uint64
}

func (k Key) String() string {
	return fmt.Sprintf("%s:%d", k.Hash, k.Id)
}

func (k Key) Empty() bool {
	return k.Hash.Empty()
}

func (a Key) Less(b Key) bool {
	return a.Hash.Less(b.Hash)
}

func (a Key) Equals(b Key) bool {
	return a.Hash.Equals(b.Hash) && a.Id == b.Id
}

func (k Key) Clone() *Key {
	return &Key{
		Hash: k.Hash,
		Id:   k.Id,
	}
}

type KeySlice []Key

func (s KeySlice) Len() int           { return len(s) }
func (s KeySlice) Less(i, j int) bool { return s[i].Less(s[j]) }
func (s KeySlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s KeySlice) Sort()              { sort.Sort(s) }
func (s KeySlice) IsSorted() bool     { return sort.IsSorted(s) }

// Returns all keys in s between, but not including start and end
func (s KeySlice) GetRange(start, end Hash) KeySlice {
	if len(s) == 0 {
		return nil
	}
	first := sort.Search(len(s), func(i int) bool {
		return s[i].Hash.Compare(start) >= 0
	})
	last := sort.Search(len(s)-first, func(i int) bool {
		return s[i+first].Hash.Compare(end) >= 0
	}) + first
	switch {
	case first == len(s):
		return nil
	case s[first].Hash.Equals(start) && first+1 == len(s):
		return nil
	case s[first].Hash.Equals(start):
		first++
	}
	if last < 0 || first > last {
		return nil
	}
	return s[first:last]
}

func (s KeySlice) find(hash Hash) int {
	return sort.Search(len(s), func(j int) bool {
		return !s[j].Hash.Less(hash)
	})
}

func (s *KeySlice) Remove(key Key) {
	i := s.find(key.Hash)
	if i < len(*s) && (*s)[i].Hash.Equals(key.Hash) {
		*s = append((*s)[:i], (*s)[i+1:]...)
	}
}

func (s KeySlice) Clone() KeySlice {
	c := make(KeySlice, len(s))
	copy(c, s)
	return c
}

func (s KeySlice) String() string {
	return dumpWithTitle("Keys", s.Hashes(), 0)
}

func (s KeySlice) Hashes() HashSlice {
	var hashes HashSlice
	for _, key := range s {
		hashes = append(hashes, key.Hash)
	}
	return hashes
}
