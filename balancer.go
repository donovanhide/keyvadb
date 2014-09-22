package keyvadb

import (
	"fmt"
	"sort"
)

var balancerRandom = MustRand()

var Balancers = []struct {
	Name     string
	Balancer Balancer
}{
	{"Buffer", &BufferBalancer{}},
	{"Distance", &DistanceBalancer{}},
}

func newBalancer(name string) (Balancer, error) {
	for _, balancer := range Balancers {
		if balancer.Name == name {
			return balancer.Balancer, nil
		}
	}
	return nil, fmt.Errorf("unknown balancer: %s", name)
}

type BufferBalancer struct{}
type DistanceBalancer struct{}

func (b *BufferBalancer) Balance(n *Node, s KeySlice) (*Node, KeySlice) {
	occupied := n.Occupancy()
	// Check for duplicate entries in full node
	if occupied == n.MaxEntries() {
		for _, key := range n.Keys {
			s.Remove(key)
		}
		return n, s
	}
	union := n.NonEmptyKeys().Union(s)
	// Shortcut all duplicates in non full node
	if len(union) == occupied {
		return n, nil
	}
	// Copy on write
	n = n.CloneIfClean()
	// Node not full, insert entries on the right
	if len(union) <= n.MaxEntries() {
		copy(n.Keys[n.MaxEntries()-len(union):], union)
		return n, nil
	}
	// Randomly select entries from union
	picks := balancerRandom.Perm(len(union))[:n.MaxEntries()]
	sort.Ints(picks)
	for i, pick := range picks {
		n.UpdateEntry(i, union[pick])
	}
	// Remove added entries from set
	for _, key := range n.Keys {
		union.Remove(key)
	}
	return n, union
}

func (b *DistanceBalancer) Balance(n *Node, s KeySlice) (*Node, KeySlice) {
	occupied := n.Occupancy()
	// Check for duplicate entries in full node
	if occupied == n.MaxEntries() {
		for _, key := range n.Keys {
			s.Remove(key)
		}
		return n, s
	}
	union := n.NonEmptyKeys().Union(s)
	// Shortcut all duplicates in non full node
	if len(union) == occupied {
		return n, nil
	}
	// Copy on write
	n = n.CloneIfClean()
	// Node not full, insert entries on the right
	if len(union) <= n.MaxEntries() {
		copy(n.Keys[n.MaxEntries()-len(union):], union)
		return n, nil
	}
	// Merge and place in order
	n.AddSyntheticKeys()
	for index, distance := range union.FindNearestKeys(n) {
		n.UpdateEntry(index-1, distance.Key)
		union.Remove(distance.Key)
	}
	return n, union
}
