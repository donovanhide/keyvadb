package keyva

import (
	"fmt"
	"sort"
)

func NewNeighbour(v Value, t Target) *Neighbour {
	return &Neighbour{
		Id:       v.Id,
		Key:      v.Key,
		Distance: v.Key.Distance(t.Key),
		Index:    t.Index,
	}
}

type Neighbour struct {
	Id       uint64
	Key      Hash
	Distance Hash
	Index    int
}
type NeighbourSlice []Neighbour

func (n Neighbour) String() string {
	return fmt.Sprintf("%s:%s:%d:%d", n.Key, n.Distance, n.Index, n.Id)
}

func (n NeighbourSlice) Len() int           { return len(n) }
func (n NeighbourSlice) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n NeighbourSlice) Less(i, j int) bool { return n[i].Distance.Compare(n[j].Distance) < 0 }
func (n NeighbourSlice) Sort()              { sort.Sort(n) }

func (n NeighbourSlice) String() string {
	return dumpWithTitle("Neighbours", n, 0)
}
