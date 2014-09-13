package keyvadb

import (
	"fmt"
	"math/big"
	"sort"
)

var balancerRandom = MustRand()

var Balancers = []struct {
	Name     string
	Balancer Balancer
}{
	{"Random", &RandomBalancer{}},
	{"Buffer", &BufferBalancer{}},
	{"Distance", &DistanceBalancer{}},
}

type RandomBalancer struct{}
type BufferBalancer struct{}
type DistanceBalancer struct{}

type EmptyRange struct {
	Start      Hash
	End        Hash
	StartIndex int
	EndIndex   int
}

func (e EmptyRange) String() string {
	return fmt.Sprintf("%d:%d:%s:%s", e.StartIndex, e.EndIndex, e.Start, e.End)
}

func (e EmptyRange) Len() int {
	return e.EndIndex - e.StartIndex
}

func EmptyRanges(n *Node) []EmptyRange {
	var empties []EmptyRange
	r := n.Ranges()
	for i := 0; i < n.MaxEntries(); i++ {
		if n.Keys[i].Empty() {
			empty := EmptyRange{
				Start:      r[i],
				StartIndex: i,
			}
			for ; i < n.MaxEntries() && n.Keys[i].Empty(); i++ {
			}
			empty.End = r[i+1]
			empty.EndIndex = i
			empties = append(empties, empty)
		}
	}
	return empties
}

func (b *RandomBalancer) Balance(n *Node, s KeySlice) KeySlice {
	remainder := s.Clone()
	for _, empty := range EmptyRanges(n) {
		sub := s.GetRange(empty.Start, empty.End)
		switch {
		case len(sub) == 0:
			//Nothing to do
			continue
		case empty.Len() <= len(sub):
			//Pick random
			picks := balancerRandom.Perm(len(sub))[:empty.Len()]
			sort.Ints(picks)
			for i, pick := range picks {
				n.UpdateEntry(empty.StartIndex+i, sub[pick])
				remainder.Remove(sub[pick])
			}
		default:
			//Place random
			locations := balancerRandom.Perm(empty.Len())[:len(sub)]
			sort.Ints(locations)
			for i, location := range locations {
				n.UpdateEntry(empty.StartIndex+location, sub[i])
				remainder.Remove(sub[i])
			}
		}
	}
	return remainder
}

func (b *BufferBalancer) Balance(n *Node, s KeySlice) KeySlice {
	occupied := n.Occupancy()
	switch {
	case occupied+len(s) <= n.MaxEntries():
		// No children yet
		// Add items at the start and sort node
		for i, key := range s {
			n.UpdateEntry(i, key)
		}
		n.SortByKey()
		return nil
	case occupied < n.MaxEntries():
		// Merge random
		remainder := s.Clone()
		picks := balancerRandom.Perm(len(s))[:n.MaxEntries()-occupied]
		sort.Ints(picks)
		for i, pick := range picks {
			n.UpdateEntry(i, s[pick])
			remainder.Remove(s[pick])
		}
		n.SortByKey()
		return remainder
	default:
		// Nothing to do
		// Node is full
		return s
	}
}

type DistanceSorter struct {
	KeySlice
	Start, End         *big.Int
	Stride, HalfStride *big.Int
	Entries            int64
}

func NewDistanceSorter(s KeySlice, start, end Hash, entries int64) *DistanceSorter {
	stride := start.Stride(end, entries)
	return &DistanceSorter{
		KeySlice:   s,
		Start:      start.Big(),
		End:        end.Big(),
		Stride:     stride.Big(),
		HalfStride: stride.Divide(2).Big(),
		Entries:    entries,
	}
}

func (d *DistanceSorter) Less(i, j int) bool {
	_, ld := d.KeySlice[i].Key.NearestStride(d.Start, d.Stride, d.HalfStride, d.Entries)
	_, rd := d.KeySlice[j].Key.NearestStride(d.Start, d.Stride, d.HalfStride, d.Entries)
	return ld.Less(rd)
}

func (b *DistanceBalancer) Balance(n *Node, s KeySlice) KeySlice {
	occupied := n.Occupancy()
	switch {
	case occupied+len(s) <= n.MaxEntries():
		// No children yet
		// Add items at the start and sort node
		for i, key := range s {
			n.UpdateEntry(i, key)
		}
		n.SortByKey()
		return nil
	case occupied < n.MaxEntries():
		// Merge and place in order
		candidates := append(s.Clone(), n.NonEmptyKeys()...)
		n.AddSyntheticKeys()
		dist := NewDistanceSorter(candidates, n.Start, n.End, int64(n.MaxEntries()))
		sort.Sort(dist)
		used := make(map[int]Key)
		for _, candidate := range dist.KeySlice {
			// Use distance to make calculation
			// index, distance := candidate.Key.NearestStride(dist.Start, dist.Stride, dist.HalfStride)
			index, _ := candidate.Key.NearestStride(dist.Start, dist.Stride, dist.HalfStride, int64(n.MaxEntries()))
			if _, ok := used[index]; !ok {
				used[index] = candidate
				n.UpdateEntry(index-1, candidate)
			}
			if len(used) == n.MaxEntries() {
				break
			}
		}
		n.SortByKey()
		candidates.Sort()
		for _, key := range used {
			candidates.Remove(key)
		}
		return candidates
	default:
		// Nothing to do
		// Node is full
		return s
	}
}
