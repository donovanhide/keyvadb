package keyva

import (
	"fmt"
	"strings"
)

type nodeFunc func(int, *Node) error

type Node struct {
	Id       uint64
	Start    Hash
	End      Hash
	Keys     [ItemCount]Key
	Children [ChildCount]uint64
}

func (n *Node) Empty(i int) bool {
	return n.Keys[i].Empty()
}

func (n *Node) HasChild(i int) bool {
	return n.Children[i] != EmptyChild || n.Children[i+1] != EmptyChild
}

func (n *Node) UpdateEntry(i int, key Key) {
	if n.HasChild(i) {
		panic("cannot update entry with child")
	}
	n.Keys[i] = key
}

func (n *Node) NonEmptyKeys() KeySlice {
	var keys KeySlice
	for _, key := range n.Keys {
		if !key.Empty() {
			keys = append(keys, key)
		}
	}
	return keys
}

func (n *Node) Ranges() HashSlice {
	return append(append(HashSlice{n.Start}, KeySlice(n.Keys[:]).Keys()...), n.End)
}

func (n *Node) NonEmptyRanges() HashSlice {
	return append(append(HashSlice{n.Start}, n.NonEmptyKeys().Keys()...), n.End)
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

func (n *Node) Stride() Hash {
	return n.Start.Stride(n.End, ItemCount+1)
}

func (n *Node) Distance() Hash {
	return n.Start.Distance(n.End)
}

func (n *Node) Items() string {
	var items []string
	for i := range n.Keys {
		items = append(items, fmt.Sprintf("%08d\t%s", i, n.Keys[i]))
	}
	return strings.Join(items, "\n")
}

func (n *Node) String() string {
	format := "Id:\t\t%d\nWell Formed:\t%t\nOccupancy:\t%d\nChildren:\t%d\nStart:\t\t%s\nEnd:\t\t%s\nDistance:\t%s\nStride:\t\t%s\n--------\n%s\n--------"
	return fmt.Sprintf(format, n.Id, n.SanityCheck(), n.Occupancy(), n.ChildCount(), n.Start, n.End, n.Distance(), n.Stride(), n.Items())
}

// sorting helpers

func (n *Node) Len() int { return ItemCount }
func (n *Node) Swap(i, j int) {
	if n.Children[i] != EmptyChild ||
		n.Children[j] != EmptyChild ||
		n.Children[i+1] != EmptyChild ||
		n.Children[j+1] != EmptyChild {
		panic(fmt.Sprintf("Cannot swap:\n%s", n))
	}
	n.Keys[i], n.Keys[j] = n.Keys[j], n.Keys[i]
}

type nodeByKey struct {
	*Node
}

func (n nodeByKey) Less(i, j int) bool { return n.Keys[i].Less(n.Keys[j]) }
