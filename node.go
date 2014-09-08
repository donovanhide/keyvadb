package keyva

import (
	"fmt"
	"math/big"
	"strings"
)

type nodeFunc func(int, *Node) error

type Node struct {
	Id       uint64
	Start    Hash
	End      Hash
	Keys     [ItemCount]Hash
	Values   [ItemCount]uint64
	Children [ChildCount]uint64
}

type nodeByKey struct {
	*Node
}

type nodeByDistance struct {
	*Node
	Stride     *big.Int
	HalfStride *big.Int
}

func (n *nodeByKey) Less(i, j int) bool { return n.Keys[i].Compare(n.Keys[j]) < 0 }

func (n *nodeByDistance) Less(i, j int) bool {
	leftIndex, leftDistance := n.Keys[i].NearestStride(n.Stride, n.HalfStride)
	rightIndex, rightDistance := n.Keys[j].NearestStride(n.Stride, n.HalfStride)
	if leftIndex == rightIndex {
		return leftDistance.Compare(rightDistance) < 0
	}
	return leftIndex < rightIndex
}

func (n *Node) Len() int { return ItemCount }
func (n *Node) Swap(i, j int) {
	if n.Children[i] != EmptyChild ||
		n.Children[j] != EmptyChild ||
		n.Children[i+1] != EmptyChild ||
		n.Children[j+1] != EmptyChild {
		panic(fmt.Sprintf("Cannot swap:\n%s", n))
	}
	n.Keys[i], n.Keys[j], n.Values[i], n.Values[j] = n.Keys[j], n.Keys[i], n.Values[j], n.Values[i]
}

func (n *Node) UpdateEntry(i int, key Hash, id uint64) {
	n.Keys[i] = key
	n.Values[i] = id
}

func (n *Node) NonEmptyKeys() HashSlice {
	var keys HashSlice
	for _, key := range n.Keys {
		if !key.Empty() {
			keys = append(keys, key)
		}
	}
	return keys
}

func (n *Node) Ranges() HashSlice {
	return append(append(HashSlice{n.Start}, n.Keys[:]...), n.End)
}

func (n *Node) NonEmptyRanges() HashSlice {
	return append(append(HashSlice{n.Start}, n.NonEmptyKeys()...), n.End)
}

func (n *Node) ChildCount() int {
	count := 0
	for _, child := range n.Children {
		if child != EmptyChild {
			count++
		}
	}
	return count
}

func (n *Node) Occupancy() int {
	count := 0
	for _, key := range n.Keys {
		if !key.Empty() {
			count++
		}
	}
	return count
}

func (n *Node) SanityCheck() bool {
	return n.NonEmptyRanges().IsSorted()
}

func (n *Node) Items() string {
	var items []string
	for i := range n.Keys {
		items = append(items, fmt.Sprintf("%08d\t%s %d %d", i, n.Keys[i], n.Values[i], n.Children[i]))
	}
	return strings.Join(items, "\n")
}

func (n *Node) String() string {
	distance := n.Start.Distance(n.End)
	stride := n.Start.Stride(n.End, ItemCount+1)
	format := "Id:\t\t%d\nWell Formed:\t%t\nOccupancy:\t%d\nChildren:\t%d\nStart:\t\t%s\nEnd:\t\t%s\nDistance:\t%s\nStride:\t\t%s\n--------\n%s\n--------"
	return fmt.Sprintf(format, n.Id, n.SanityCheck(), n.Occupancy(), n.ChildCount(), n.Start, n.End, distance, stride, n.Items())
}
