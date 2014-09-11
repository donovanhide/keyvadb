package keyva

import (
	"fmt"
	"math/big"
	"sort"
)

type Distance struct {
	Id       uint64
	Key      Hash
	Distance Hash
}

type DistanceSlice []Distance

func (s DistanceSlice) Len() int           { return len(s) }
func (s DistanceSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s DistanceSlice) Less(i, j int) bool { return s[i].Distance.Less(s[j].Distance) }
func (s DistanceSlice) Sort()              { sort.Sort(s) }
func (s DistanceSlice) IsSorted() bool     { return sort.IsSorted(s) }

type DistanceMap struct {
	m                   map[int]DistanceSlice
	start, stride, half *big.Int
}

func NewDistanceMap(n *Node) *DistanceMap {
	return &DistanceMap{
		m:      make(map[int]DistanceSlice),
		start:  n.Start.Big(),
		stride: n.Stride().Big(),
		half:   n.Stride().Divide(2).Big(),
	}
}

func (m DistanceMap) AddNode(n *Node) {
	for _, slot := range n.Movable() {
		m.add(n.Values[slot], n.Keys[slot])
	}
}

func (m DistanceMap) AddValues(v KeySlice) {
	for _, value := range v {
		m.add(value.Id, value.Key)
	}
}

func (m DistanceMap) add(id uint64, key Hash) {
	index, distance := key.NearestStride(m.start, m.stride, m.half)
	s := m.m[index]
	s = append(s, Distance{
		Id:       id,
		Key:      key,
		Distance: distance,
	})
	s.Sort()
	m.m[index] = s
}

func (m DistanceMap) Get(i int) DistanceSlice {
	return m.m[i]
}

func (m DistanceMap) GetBest(i int) *Distance {
	s := m.m[i]
	if len(s) == 0 {
		return nil
	}
	return &s[0]
}

func (m DistanceMap) GetWorst(i int) *Distance {
	s := m.m[i]
	if len(s) == 0 {
		return nil
	}
	return &s[len(s)-1]
}

func (m DistanceMap) TakeBest(i int) *Distance {
	s := m.m[i]
	if len(s) == 0 {
		return nil
	}
	best := s[0]
	m.m[i] = s[1:]
	if len(m.m[i]) == 0 {
		delete(m.m, i)
	}
	return &best
}

func (m DistanceMap) TakeWorst(i int) *Distance {
	s := m.m[i]
	if len(s) == 0 {
		return nil
	}
	worst := s[len(s)-1]
	m.m[i] = s[:len(s)-1]
	if len(m.m[i]) == 0 {
		delete(m.m, i)
	}
	return &worst
}

func (m DistanceMap) Target(i int) Hash {
	return newHash(m.start).Add(newHash(m.stride).Multiply(int64(i)))
}

func (m DistanceMap) Len() int {
	return len(m.m)
}

func (m DistanceMap) Count() int {
	var count int
	for _, s := range m.m {
		count += len(s)
	}
	return count
}

func (m DistanceMap) Matches() []int {
	var matches []int
	for i := range m.m {
		matches = append(matches, i)
	}
	sort.Ints(matches)
	return matches
}

func (m DistanceMap) SanityCheck() bool {
	for _, i := range m.Matches() {
		s := m.Get(i)
		if !s.IsSorted() {
			return false
		}
		target := m.Target(i)
		for _, d := range m.Get(i) {
			distance := d.Key.Sub(target)
			if distance.Compare(d.Distance) != 0 || i == 0 || i > ItemCount {
				return false
			}
		}
	}
	return true
}

func (m DistanceMap) String() string {
	var s []string
	for _, i := range m.Matches() {
		for _, d := range m.Get(i) {
			expected := d.Key.Sub(m.Target(i))
			s = append(s, fmt.Sprintf("%03d:%s:%s:%s:%s", i, d.Key, m.Target(i), d.Distance, expected))
		}
	}
	return dumpWithTitle("Distances", s, 0)
}
