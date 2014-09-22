package keyvadb

import (
	"encoding/binary"
	"fmt"
	"io"
	"sort"
	"strings"
)

type NodeId uint64

type NodeFunc func(int, *Node) error
type ChildFunc func(int, NodeId, Hash, Hash) error

func (id NodeId) Empty() bool {
	return id == EmptyChild
}

type Node struct {
	Id       NodeId
	Start    Hash
	End      Hash
	Keys     KeySlice
	Children []NodeId
	Dirty    bool
}

func NewNode(start, end Hash, id NodeId, degree uint64) *Node {
	return &Node{
		Id:       id,
		Start:    start,
		End:      end,
		Keys:     make(KeySlice, int(degree-1)),
		Children: make([]NodeId, int(degree)),
	}
}

func (n *Node) CloneIfClean() *Node {
	if !n.Dirty {
		n = n.Clone()
		n.Dirty = true
	}
	return n
}

func (n *Node) Clone() *Node {
	c := &Node{
		Id:    n.Id,
		Start: n.Start,
		End:   n.End,
		Keys:  n.Keys.Clone(),
	}
	c.Children = make([]NodeId, len(n.Children))
	copy(c.Children, n.Children)
	return c
}

func (n *Node) Empty(i int) bool {
	return n.Keys[i].Empty()
}

func (n *Node) HasChild(i int) bool {
	return !n.Children[i].Empty() && !n.Children[i+1].Empty()
}

func (n *Node) UpdateEntry(i int, key Key) {
	if n.HasChild(i) {
		panic("cannot update entry with child")
	}
	n.Keys[i] = key
}

func (n *Node) AddSyntheticKeys() {
	stride := n.Stride()
	cursor := n.Start.Add(stride)
	for i := range n.Keys {
		n.UpdateEntry(i, Key{
			Hash: cursor,
			Id:   SyntheticValue,
		})
		cursor = cursor.Add(stride)
	}
}

func (n *Node) Synthetics() int {
	count := 0
	for _, key := range n.Keys {
		if key.Id.Synthetic() {
			count++
		}
	}
	return count
}

func (n *Node) ChildCount() int {
	count := 0
	for _, child := range n.Children {
		if !child.Empty() {
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

func (n *Node) NonEmptyKeys() KeySlice {
	var keys KeySlice
	for _, key := range n.Keys {
		if !key.Empty() {
			keys = append(keys, key)
		}
	}
	return keys
}

func (n *Node) NonEmptyRanges() HashSlice {
	return append(append(HashSlice{n.Start}, n.NonEmptyKeys().Hashes()...), n.End)
}

func (n *Node) SanityCheck() bool {
	return n.NonEmptyRanges().IsSorted()
}

func (n *Node) Stride() Hash {
	return n.Start.Stride(n.End, int64(len(n.Children)))
}

func (n *Node) Distance() Hash {
	return n.Start.Distance(n.End)
}

func (n *Node) MaxEntries() int {
	return len(n.Keys)
}

func (n *Node) MaxChildren() int {
	return len(n.Children)
}

func (n *Node) GetChildRange(i int) (Hash, Hash) {
	switch i {
	case 0:
		return n.Start, n.Keys[0].Hash
	case n.MaxEntries():
		return n.Keys[i-1].Hash, n.End
	default:
		return n.Keys[i-1].Hash, n.Keys[i].Hash
	}
}

func (n *Node) Each(f ChildFunc) error {
	for i := range n.Children {
		start, end := n.GetChildRange(i)
		if start.Empty() || end.Empty() {
			continue
		}
		if err := f(i, n.Children[i], start, end); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) String() string {
	var items []string
	for i := range n.Keys {
		items = append(items, fmt.Sprintf("%08d\t%s", i, n.Keys[i]))
	}
	format := "Id:\t\t%d\nDirty:\t\t%t\nWell Formed:\t%t\nOccupancy:\t%d\n"
	format += "Children:\t%+v\nStart:\t\t%s\nEnd:\t\t%s\nDistance:\t%s\n"
	format += "Stride:\t\t%s\n--------\n%s\n--------"
	return fmt.Sprintf(format, n.Id, n.Dirty, n.SanityCheck(), n.Occupancy(), n.Children, n.Start, n.End, n.Distance(), n.Stride(), strings.Join(items, "\n"))
}

func (n *Node) Len() int           { return len(n.Keys) }
func (n *Node) Less(i, j int) bool { return n.Keys[i].Less(n.Keys[j]) }
func (n *Node) Swap(i, j int) {
	if n.HasChild(i) || n.HasChild(j) {
		panic(fmt.Sprintf("Cannot swap:\n%s", n))
	}
	n.Keys[i], n.Keys[j] = n.Keys[j], n.Keys[i]
}
func (n *Node) Sort()          { sort.Sort(n) }
func (n *Node) IsSorted() bool { return sort.IsSorted(n) }

func (node *Node) WriteTo(w io.Writer) (int64, error) {
	b := make([]byte, NodeBlockSize)
	n := copy(b, node.Start[:])
	n += copy(b[n:], node.End[:])
	for _, key := range node.Keys {
		n += copy(b[n:], key.Hash[:])
		binary.BigEndian.PutUint64(b[n:], uint64(key.Id))
		n += 8
	}
	for _, child := range node.Children {
		binary.BigEndian.PutUint64(b[n:], uint64(child))
		n += 8
	}
	n, err := w.Write(b)
	return int64(n), err
}

func (node *Node) ReadFrom(r io.Reader) (int64, error) {
	b := make([]byte, NodeBlockSize)
	if n, err := r.Read(b); err != nil {
		return int64(n), err
	}
	n := copy(node.Start[:], b)
	n += copy(node.End[:], b[n:])
	for i := range node.Keys {
		n += copy(node.Keys[i].Hash[:], b[n:])
		node.Keys[i].Id = ValueId(binary.BigEndian.Uint64(b[n:]))
		n += 8
	}
	for i := range node.Children {
		node.Children[i] = NodeId(binary.BigEndian.Uint64(b[n:]))
		n += 8
	}
	return NodeBlockSize, nil
}
