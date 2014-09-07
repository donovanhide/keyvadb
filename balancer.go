package keyva

import "sort"

type NaiveBalancer struct{}
type MatchingBalancer struct{}

func (b *NaiveBalancer) Balance(n *Node, v ValueSlice) NeighbourSlice {
	var neighbours NeighbourSlice
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
		neighbours.Sort()
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
	debugPrintln(filtered)
	return filtered
}
