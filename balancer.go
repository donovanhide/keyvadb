package keyvadb

import "sort"

var balancerRandom = MustRand()

var Balancers = []struct {
	Name     string
	Balancer Balancer
}{
	{"Buffer", &BufferBalancer{}},
	{"Distance", &DistanceBalancer{}},
}

type BufferBalancer struct{}
type DistanceBalancer struct{}

func (b *BufferBalancer) Balance(n *Node, s KeySlice) KeySlice {
	occupied := n.Occupancy()
	switch {
	case occupied+len(s) <= n.MaxEntries():
		// No children yet
		// Add items at the start and sort node
		for i, key := range s {
			n.UpdateEntry(i, key)
		}
		n.Sort()
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
		n.Sort()
		return remainder
	default:
		// Nothing to do
		// Node is full
		return s
	}
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
		n.Sort()
		return nil
	case occupied < n.MaxEntries():
		// Merge and place in order
		candidates := append(n.NonEmptyKeys(), s...)
		candidates.Sort()
		n.AddSyntheticKeys()
		for index, distance := range candidates.FindNearestKeys(n) {
			n.UpdateEntry(index-1, distance.Key)
			candidates.Remove(distance.Key)
		}
		return candidates
	default:
		// Nothing to do
		// Node is full
		return s
	}
}
