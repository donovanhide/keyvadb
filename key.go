package keyva

import (
	"errors"
	"fmt"
	"sort"
)

type Key struct {
	Key Hash
	Id  uint64
}

type KeySlice []Key

func (k Key) String() string {
	return fmt.Sprintf("%s:%d", k.Key, k.Id)
}

func (k Key) Empty() bool {
	return k.Key.Empty()
}

func (a Key) Less(b Key) bool {
	return a.Key.Less(b.Key)
}

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
		return s[i].Key.Compare(start) >= 0
	})
	last := sort.Search(len(s)-first, func(i int) bool {
		return s[i+first].Key.Compare(end) >= 0
	}) + first
	switch {
	case first == len(s):
		return nil
	case s[first].Key.Equals(start) && first+1 == len(s):
		return nil
	case s[first].Key.Equals(start):
		first++
	}
	if last < 0 || first > last {
		return nil
	}
	return s[first:last]
}

var (
	entryHasChild  = errors.New("entry has child")
	valueNotFound  = errors.New("value not found")
	alreadyPresent = errors.New("already present")
)

// Exchange Key and Id at position i for value
func (s KeySlice) TryExchange(n *Node, i int, key Key) error {
	if n.HasChild(i) {
		return entryHasChild
	}
	j := sort.Search(len(s), func(k int) bool {
		return !s[k].Key.Less(key.Key)
	})
	switch {
	case j == len(s) || s[j].Id != key.Id:
		return valueNotFound
	case s[j].Id == key.Id:
		return alreadyPresent
	default:
		s[j], n.Keys[i] = n.Keys[i], s[j]
		s.Sort()
		return nil
	}
}

func (s KeySlice) String() string {
	return dumpWithTitle("Values", s.Keys(), 0)
}

func (s KeySlice) Keys() HashSlice {
	var keys HashSlice
	for _, key := range s {
		keys = append(keys, key.Key)
	}
	return keys
}
