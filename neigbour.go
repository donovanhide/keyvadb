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

func NewNeighbourAt(v *Value, i int) *Neighbour {
	return &Neighbour{
		Id:    v.Id,
		Key:   v.Key,
		Index: i,
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

type byDistance struct{ NeighbourSlice }
type byIndex struct{ NeighbourSlice }
type byKey struct{ NeighbourSlice }

func (n byDistance) Less(i, j int) bool {
	return n.NeighbourSlice[i].Distance.Compare(n.NeighbourSlice[j].Distance) < 0
}

func (n byIndex) Less(i, j int) bool {
	return n.NeighbourSlice[i].Index < n.NeighbourSlice[j].Index
}

func (n byKey) Less(i, j int) bool {
	return n.NeighbourSlice[i].Key.Compare(n.NeighbourSlice[j].Key) < 0
}

func (n NeighbourSlice) Len() int                 { return len(n) }
func (n NeighbourSlice) Swap(i, j int)            { n[i], n[j] = n[j], n[i] }
func (n NeighbourSlice) SortByDistance()          { sort.Sort(byDistance{n}) }
func (n NeighbourSlice) SortByIndex()             { sort.Sort(byIndex{n}) }
func (n NeighbourSlice) SortByKey()               { sort.Sort(byKey{n}) }
func (n NeighbourSlice) IsSortedByDistance() bool { return sort.IsSorted(byDistance{n}) }
func (n NeighbourSlice) IsSortedByIndex() bool    { return sort.IsSorted(byIndex{n}) }
func (n NeighbourSlice) IsSortedByKey() bool      { return sort.IsSorted(byKey{n}) }
func (n NeighbourSlice) SanityCheck() bool        { return n.IsSortedByIndex() && n.IsSortedByKey() }

func (n NeighbourSlice) String() string {
	return dumpWithTitle("Neighbours", n, 0)
}
