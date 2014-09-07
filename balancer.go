package keyva

import (
	"fmt"
	"math/rand"
	"sort"
)

type RandomBalancer struct{}
type NaiveBalancer struct{}
type MatchingBalancer struct{}

func (b *RandomBalancer) Balance(n *Node, v ValueSlice) NeighbourSlice {
	var neighbours NeighbourSlice
	r := rand.New(rand.NewSource(int64(n.Id)))
	for _, empty := range n.EmptyRanges() {
		sub := v.GetRange(empty.Start, empty.End)
		length := empty.EndIndex - empty.StartIndex
		if length > len(sub) {
			length = len(sub)
		}
		selection := r.Perm(len(sub))[:length]
		sort.Ints(selection)
		// fmt.Println(empty)
		// fmt.Println(selection)
		j := 0
		for i := empty.StartIndex; j < length && i < empty.EndIndex; i++ {
			// fmt.Println(sub[selection[j]].Key)
			neighbours = append(neighbours, Neighbour{
				Id:    sub[selection[j]].Id,
				Key:   sub[selection[j]].Key,
				Index: i,
			})
			j++
		}
	}
	neighbours.SortByIndex()
	return neighbours
}

func (b *NaiveBalancer) Balance(n *Node, v ValueSlice) NeighbourSlice {
	median := v[len(v)/2]
	occupancy := n.Occupancy()
	length := ItemCount - n.Occupancy()
	swing := median.Key.Distance(n.Start).Compare(median.Key.Distance(n.End))
	fmt.Println(median.Key, length, swing)
	neighbours := make(NeighbourSlice, length)
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
	return neighbours
}

func (b *MatchingBalancer) Balance(n *Node, v ValueSlice) NeighbourSlice {
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
	return filtered
}
