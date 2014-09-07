package keyva

import (
	"fmt"
	"strings"
)

type nodeFunc func(int, *Node) error

type Item struct {
	Key    Hash
	Offset uint64
}

type Node struct {
	Id       uint64
	Start    Hash
	End      Hash
	Keys     [ItemCount]Hash
	Values   [ItemCount]uint64
	Children [ChildCount]uint64
}

func (n *Node) Ranges() HashSlice {
	return append(append(HashSlice{n.Start}, n.Keys[:]...), n.End)
}

func (n *Node) SanityCheck() bool {
	return n.Ranges().IsSorted()
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
		if key.Compare(EmptyItem) > 0 {
			count++
		}
	}
	return count
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
