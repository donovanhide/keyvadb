package keyva

import (
	"fmt"
	"math/rand"
	"sort"
)

type RandomBalancer struct{}
type NaiveBalancer struct{}
type MatchingBalancer struct{}

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

func (b *NaiveBalancer) Balance(n *Node, v ValueSlice) (insertions int) {
	median := v[len(v)/2]
	occupancy := n.Occupancy()
	length := ItemCount - n.Occupancy()
	swing := median.Key.Distance(n.Start).Compare(median.Key.Distance(n.End))
	fmt.Println(median.Key, length, swing)
	if occupancy == 0 {
		// 	if swing <= 0 {
		// 		for i := 0; i < length; i++ {
		// 			neighbours = append(neighbours, Neighbour{Id: v[i].Id, Key: v[i].Key, Index: i})
		// 		}
		// 	} else {
		// 		for i := length - 1; i >= 0; i-- {
		// 			neighbours = append(neighbours, Neighbour{Id: v[i].Id, Key: v[i].Key, Index: i})
		// 		}
		// 	}
	}
	return
}

func (b *MatchingBalancer) Balance(n *Node, v ValueSlice) (insertions int) {
	targets := NewTargetSlice(n.Start, n.End, n.Keys[:])
	debugPrintln(targets)
	var neighbours NeighbourSlice
	var filtered NeighbourSlice
	available := make(ValueSlice, len(v))
	copy(available, v)
	availableSlots := make(map[int]Target)
	for _, target := range targets {
		availableSlots[target.Index] = target
	}
	usedKeys := make(map[Hash]struct{})
	for len(availableSlots) > 0 && len(available) > 0 {
		debugPrintln("Matched Targets")
		for _, target := range availableSlots {
			i := sort.Search(len(available), func(j int) bool {
				return available[j].Key.Compare(target.Key) >= 0
			})
			debugPrintf("%06d:%s\n", i, target)
			if i == len(available) {
				// if len(available) == 1 {
				neighbours = append(neighbours, *NewNeighbour(available[i-1], target))
				// }
				continue
			}
			neighbours = append(neighbours, *NewNeighbour(available[i], target))
			if i > 0 {
				neighbours = append(neighbours, *NewNeighbour(available[i-1], target))
			}
			if i < len(available)-1 {
				neighbours = append(neighbours, *NewNeighbour(available[i+1], target))
			}
		}
		debugPrintln("--------")
		debugPrintln(available)
		neighbours.SortByDistance()
		for _, neighbour := range neighbours {
			_, keyUsed := usedKeys[neighbour.Key]
			_, slotAvailable := availableSlots[neighbour.Index]
			if !keyUsed && slotAvailable {
				filtered = append(filtered, neighbour)
				usedKeys[neighbour.Key] = struct{}{}
				delete(availableSlots, neighbour.Index)
				i := sort.Search(len(available), func(j int) bool {
					return available[j].Key.Compare(neighbour.Key) >= 0
				})
				available = append(available[:i], available[i+1:]...)
			}
		}
	}
	filtered.SortByIndex()
	debugPrintln(filtered)
	return
}
