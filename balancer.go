package keyva

import (
	"fmt"
	"math/big"
	"math/rand"
	"sort"
)

type RandomBalancer struct{}
type BufferBalancer struct{}
type DistanceWithBufferBalancer struct{}

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
	for i := 0; i < ItemCount; i++ {
		if n.Keys[i].Empty() {
			empty := EmptyRange{
				Start:      r[i],
				StartIndex: i,
			}
			for ; i < ItemCount && n.Keys[i].Empty(); i++ {
			}
			empty.End = r[i+1]
			empty.EndIndex = i
			empties = append(empties, empty)
		}
	}
	return empties
}

func (b *RandomBalancer) Balance(n *Node, s KeySlice) KeySlice {
	r := rand.New(rand.NewSource(int64(n.Id)))
	remainder := s.Clone()
	for _, empty := range EmptyRanges(n) {
		sub := s.GetRange(empty.Start, empty.End)
		switch {
		case len(sub) == 0:
			//Nothing to do
			continue
		case empty.Len() <= len(sub):
			//Pick random
			picks := r.Perm(len(sub))[:empty.Len()]
			sort.Ints(picks)
			for i, pick := range picks {
				n.UpdateEntry(empty.StartIndex+i, sub[pick])
				remainder.Remove(sub[pick])
			}
		default:
			//Place random
			locations := r.Perm(empty.Len())[:len(sub)]
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
	case occupied+len(s) <= ItemCount:
		// No children yet
		// Add items at the start and sort node
		for i, key := range s {
			n.UpdateEntry(i, key)
		}
		n.SortByKey()
		return nil
	case occupied < ItemCount:
		// Merge random
		remainder := s.Clone()
		r := rand.New(rand.NewSource(int64(n.Id)))
		picks := r.Perm(len(s))[:ItemCount-occupied]
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
}

func NewDistanceSorter(s KeySlice, start, end Hash) *DistanceSorter {
	stride := start.Stride(end, ItemCount)
	return &DistanceSorter{
		KeySlice:   s,
		Start:      start.Big(),
		End:        end.Big(),
		Stride:     stride.Big(),
		HalfStride: stride.Divide(2).Big(),
	}
}

func (d *DistanceSorter) Less(i, j int) bool {
	_, ld := d.KeySlice[i].Key.NearestStride(d.Start, d.Stride, d.HalfStride)
	_, rd := d.KeySlice[j].Key.NearestStride(d.Start, d.Stride, d.HalfStride)
	return ld.Less(rd)
	// li, ld := d.KeySlice[i].Key.NearestStride(d.Start, d.Stride, d.HalfStride)
	// ri, rd := d.KeySlice[j].Key.NearestStride(d.Start, d.Stride, d.HalfStride)
	// if li == ri {
	// 	return ld.Less(rd)
	// }
	// return li < ri
}

func (b *DistanceWithBufferBalancer) Balance(n *Node, s KeySlice) KeySlice {
	occupied := n.Occupancy()
	switch {
	case occupied+len(s) <= ItemCount:
		// No children yet
		// Add items at the start and sort node
		for i, key := range s {
			n.UpdateEntry(i, key)
		}
		n.SortByKey()
		return nil
	case occupied < ItemCount:
		// Merge and place in order
		candidates := append(s.Clone(), n.NonEmptyKeys()...)
		n.AddSyntheticKeys()
		dist := NewDistanceSorter(candidates, n.Start, n.End)
		sort.Sort(dist)
		used := make(map[int]Key)
		for _, candidate := range dist.KeySlice {
			// Use distance to make calculation
			// index, distance := candidate.Key.NearestStride(dist.Start, dist.Stride, dist.HalfStride)
			index, _ := candidate.Key.NearestStride(dist.Start, dist.Stride, dist.HalfStride)
			if _, ok := used[index]; !ok {
				used[index] = candidate
				n.UpdateEntry(index-1, candidate)
			}
			if len(used) == ItemCount {
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
