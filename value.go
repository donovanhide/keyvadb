package keyva

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
)

type Key struct {
	Key Hash
	Id  uint64
}

type KeyValue struct {
	Key
	Value []byte
}

type Value struct {
	Id    uint64
	Key   Hash
	Value []byte
}

type ValueSlice []Value

func (v Value) String() string {
	return fmt.Sprintf("%s:%X", v.Key, v.Value)
}

func (v ValueSlice) Len() int           { return len(v) }
func (v ValueSlice) Less(i, j int) bool { return bytes.Compare(v[i].Key[:], v[j].Key[:]) < 0 }
func (v ValueSlice) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v ValueSlice) Sort()              { sort.Sort(v) }
func (v ValueSlice) IsSorted() bool     { return sort.IsSorted(v) }

// Returns all values in v between, but not including start and end
func (v ValueSlice) GetRange(start, end Hash) ValueSlice {
	if len(v) == 0 {
		return nil
	}
	first := sort.Search(len(v), func(i int) bool {
		return v[i].Key.Compare(start) >= 0
	})
	last := sort.Search(len(v)-first, func(i int) bool {
		return v[i+first].Key.Compare(end) >= 0
	}) + first
	switch {
	case first == len(v):
		return nil
	case v[first].Key.Equals(start) && first+1 == len(v):
		return nil
	case v[first].Key.Equals(start):
		first++
	}
	if last < 0 || first > last {
		return nil
	}
	return v[first:last]
}

var (
	entryHasChild  = errors.New("entry has child")
	valueNotFound  = errors.New("value not found")
	alreadyPresent = errors.New("already present")
)

// Exchange Key and Id at position i for value
func (v ValueSlice) TryExchange(n *Node, i int, value Value) error {
	if n.HasChild(i) {
		return entryHasChild
	}
	j := sort.Search(len(v), func(k int) bool {
		return !v[k].Key.Less(value.Key)
	})
	switch {
	case j == len(v) || v[j].Id != value.Id:
		return valueNotFound
	case v[j].Id == value.Id:
		return alreadyPresent
	default:
		v[j].Id, v[j].Key, n.Values[i], n.Keys[i] = n.Values[i], n.Keys[i], v[j].Id, v[j].Key
		v.Sort()
		return nil
	}
}

func (v ValueSlice) String() string {
	return dumpWithTitle("Values", v.Keys(), 0)
}

func (v ValueSlice) Keys() HashSlice {
	var keys HashSlice
	for i := range v {
		keys = append(keys, v[i].Key)
	}
	return keys
}
