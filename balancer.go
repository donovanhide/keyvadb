package keyva

import (
	"fmt"
	"math/rand"
	"sort"
)

type RandomBalancer struct{}
type BufferBalancer struct{}

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

func (b *RandomBalancer) Balance(n *Node, v KeySlice) (insertions int) {
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
				n.UpdateEntry(empty.StartIndex+i, sub[pick])
				insertions++
			}
		default:
			//Place random
			locations := r.Perm(empty.Len())[:len(sub)]
			sort.Ints(locations)
			for i, location := range locations {
				n.UpdateEntry(empty.StartIndex+location, sub[i])
				insertions++
			}
		}
	}
	return
}

func (b *BufferBalancer) Balance(n *Node, s KeySlice) (insertions int) {
	occupied := n.Occupancy()
	switch {
	case occupied+len(s) <= ItemCount:
		// No children yet
		// Add items at the start and sort node
		for i, key := range s {
			n.UpdateEntry(i, key)
		}
		sort.Sort(&nodeByKey{n})
		insertions = len(s)
	case occupied < ItemCount:
		// Merge random
		r := rand.New(rand.NewSource(int64(n.Id)))
		picks := r.Perm(len(s))[:ItemCount-occupied]
		sort.Ints(picks)
		for i, pick := range picks {
			n.UpdateEntry(i, s[pick])
			insertions++
		}
		sort.Sort(&nodeByKey{n})
	default:
		// Nothing to do
		// Node is full
	}
	return
}
