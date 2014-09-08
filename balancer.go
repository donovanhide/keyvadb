package keyva

import (
	"fmt"
	"math/rand"
	"sort"
)

type RandomBalancer struct{}
type BufferBalancer struct{}
type BufferWithFitting struct{}

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

func (b *RandomBalancer) Balance(n *Node, v ValueSlice) (insertions int) {
	r := rand.New(rand.NewSource(int64(n.Id)))
	for _, empty := range EmptyRanges(n) {
		sub := v.GetRange(empty.Start, empty.End)
		switch {
		case len(sub) == 0:
			//Nothing to do
			continue
		case empty.Len() <= len(sub):
			//Pick random
			picks := r.Perm(len(sub))[:empty.Len()]
			sort.Ints(picks)
			for i, pick := range picks {
				n.UpdateEntry(empty.StartIndex+i, sub[pick].Key, sub[pick].Id)
				insertions++
			}
		default:
			//Place random
			locations := r.Perm(empty.Len())[:len(sub)]
			sort.Ints(locations)
			for i, location := range locations {
				n.UpdateEntry(empty.StartIndex+location, sub[i].Key, sub[i].Id)
				insertions++
			}
		}
	}
	return
}

func (b *BufferBalancer) Balance(n *Node, v ValueSlice) (insertions int) {
	occupied := n.Occupancy()
	switch {
	case occupied+len(v) <= ItemCount:
		// No children yet
		// Add items at the start and sort node
		for i, v := range v {
			n.UpdateEntry(i, v.Key, v.Id)
		}
		sort.Sort(&nodeByKey{n})
		insertions = len(v)
	case occupied < ItemCount:
		// Merge random
		r := rand.New(rand.NewSource(int64(n.Id)))
		picks := r.Perm(len(v))[:ItemCount-occupied]
		sort.Ints(picks)
		for i, pick := range picks {
			n.UpdateEntry(i, v[pick].Key, v[pick].Id)
			insertions++
		}
		sort.Sort(&nodeByKey{n})
	default:
		// Nothing to do
		// Node is full
	}
	return
}

func (b *BufferWithFitting) Balance(n *Node, v ValueSlice) (insertions int) {
	occupied := n.Occupancy()
	switch {
	case occupied+len(v) <= ItemCount:
		// No children yet
		// Add items at the start and sort node
		for i, v := range v {
			n.UpdateEntry(i, v.Key, v.Id)
		}
		sort.Sort(&nodeByKey{n})
		insertions = len(v)
	case occupied < ItemCount:
		// Choose random from v and n
		r := rand.New(rand.NewSource(int64(n.Id)))
		picks := r.Perm(len(v))[:ItemCount-occupied]
		sort.Ints(picks)
		for i, pick := range picks {
			n.UpdateEntry(i, v[pick].Key, v[pick].Id)
			insertions++
		}
		sort.Sort(&nodeByKey{n})
	default:
		// Nothing to do
		// Node is full
	}
	return
}
