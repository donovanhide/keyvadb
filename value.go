package keyva

import (
	"bytes"
	"fmt"
	"sort"
)

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

func (v ValueSlice) Sort() {
	sort.Sort(v)
}

// Returns all values in v between, but not including start and end
func (v ValueSlice) GetRange(start, end Hash) ValueSlice {
	if len(v) == 0 {
		return nil
	}
	first := sort.Search(len(v), func(i int) bool {
		return v[i].Key.Compare(start) >= 0
	})
	switch {
	case first == len(v):
		return nil
	case v[first].Key.Equals(start) && first+1 == len(v):
		return nil
	case v[first].Key.Equals(start):
		first++
	}
	last := sort.Search(len(v)-first, func(i int) bool {
		return v[i+first].Key.Compare(end) >= 0
	}) + first
	switch {
	case last == len(v):
		last--
	case v[last].Key.Equals(end):
		last--
	}
	if last < 0 || first > last {
		return nil
	}
	return v[first:last]
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
