package keyva

import (
	"fmt"
	"sort"
)

type Distance struct {
	Id       uint64
	Key      Hash
	Distance Hash
}

type DistanceMap struct {
	m             map[int]*Distance
	start, stride Hash
}

func NewDistanceMap(n *Node) *DistanceMap {
	return &DistanceMap{
		m:      make(map[int]*Distance),
		start:  n.Start,
		stride: n.Stride(),
	}
}

func (m DistanceMap) Add(v ValueSlice) {
	stride := m.stride.Big()
	halfStride := m.stride.Divide(2).Big()
	for _, value := range v {
		index, distance := value.Key.NearestStride(m.start, stride, halfStride)
		// Discard matches closet to Start and End
		if index == 0 || index > ItemCount {
			continue
		}
		// See if it is the best result
		d, ok := m.m[index]
		if !ok || d.Distance.Compare(distance) > 0 {
			m.m[index] = &Distance{
				Id:       value.Id,
				Key:      value.Key,
				Distance: distance,
			}
		}
	}
}

func (m DistanceMap) Get(i int) *Distance {
	return m.m[i]
}

func (m DistanceMap) Target(i int) Hash {
	return m.start.Add(m.stride.Multiply(int64(i)))
}

func (m DistanceMap) Len() int {
	return len(m.m)
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
		d := m.Get(i)
		target := m.Target(i)
		distance := d.Key.Sub(target)
		if distance.Compare(d.Distance) != 0 {
			debugPrintln("Insane distance")
			debugPrintln("Start:\t", m.start)
			debugPrintln("Stride:\t", m.stride)
			debugPrintln("Key:\t", d.Key)
			debugPrintln("Target:\t", target)
			debugPrintln("Wanted:\t", distance)
			debugPrintln("Got:\t", d.Distance)
			return false
		}
	}
	return true
}

func (m DistanceMap) String() string {
	var s []string
	for _, i := range m.Matches() {
		d := m.Get(i)
		s = append(s, fmt.Sprintf("%03d:%s:%s:%s", i, d.Key, m.Target(i), d.Distance))
	}
	return dumpWithTitle("Distances", s, 0)
}
