package keyva

import (
	"fmt"
	"strings"
)

type nodeFunc func(int, *Node) error

type EmptyRange struct {
	Start      Hash
	End        Hash
	StartIndex int
	EndIndex   int
}

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

func (n *Node) NonEmptyKeys() HashSlice {
	var keys HashSlice
	for _, key := range n.Keys {
		if !key.Equals(EmptyItem) {
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

func (n *Node) EmptyRanges() []EmptyRange {
	var empties []EmptyRange
	r := n.Ranges()
	for i := 0; i < ItemCount; i++ {
		if n.Keys[i].Empty() {
			empty := EmptyRange{
				Start:      r[i],
				StartIndex: i,
			}
			for ; i < ItemCount && n.Keys[i].Empty(); i++ {
			}
			empty.End = r[i+1]
			empty.EndIndex = i
			empties = append(empties, empty)
		}
	}
	return empties
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
	return len(n.NonEmptyKeys())
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
