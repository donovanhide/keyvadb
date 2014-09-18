package keyvadb

import (
	"fmt"
	"io"
	"strings"
)

type Tree struct {
	Degree   uint64
	root     *Node
	keys     KeyStore
	balancer Balancer
}

func NewTree(degree uint64, keys KeyStore, balancer Balancer) (*Tree, error) {
	if degree < 2 {
		return nil, fmt.Errorf("degree must be 2 or above")
	}
	root := NewNode(FirstHash, LastHash, rootNodeId, degree)
	root.AddSyntheticKeys()
	if err := keys.Set(root); err != nil {
		return nil, err
	}
	return &Tree{
		Degree:   degree,
		root:     root,
		keys:     keys,
		balancer: balancer,
	}, nil
}

func (t *Tree) add(n *Node, v KeySlice) (insertions int, err error) {
	if len(v) == 0 {
		panic("no values to add")
	}
	debugPrintln(n)
	remainder, dirty := t.balancer.Balance(n, v)
	debugPrintln(n)
	insertions = len(v) - len(remainder)
	if *debug && !n.SanityCheck() {
		panic(fmt.Sprintf("not sane:\n%s", n))
	}
	err = n.Each(func(id NodeId, start, end Hash) (NodeId, error) {
		candidates := remainder.GetRange(start, end)
		if len(candidates) == 0 {
			return id, nil
		}
		var child *Node
		if id.Empty() {
			if child, err = t.keys.New(start, end, t.Degree); err != nil {
				return id, nil
			}
			id = child.Id
			dirty = true
		} else {
			if child, err = t.keys.Get(id, t.Degree); err != nil {
				return id, err
			}
		}
		childInsertions, err := t.add(child, candidates)
		insertions += childInsertions
		return id, err
	})
	if err != nil {
		return
	}
	if dirty {
		err = t.keys.Set(n)
	}
	return
}

// Returns number of keys inserted and an error if encountered
func (t *Tree) Add(keys KeySlice) (int, error) {
	if !keys.IsSorted() {
		return 0, fmt.Errorf("unsorted values provided")
	}
	unique := keys.Clone()
	unique.Unique()
	if len(unique) < len(keys) {
		return 0, fmt.Errorf("values provided are not unique")
	}
	return t.add(t.root, unique)
}

type WalkFunc func(key *Key) error

func (t *Tree) walk(id NodeId, start, end Hash, f WalkFunc) error {
	n, err := t.keys.Get(id, t.Degree)
	if err != nil {
		return err
	}
	for i, cid := range n.Children {
		if !cid.Empty() {
			if s, e := n.GetChildRange(i); !end.Less(s) && !start.Greater(e) {
				if err := t.walk(cid, start, end, f); err != nil {
					return err
				}
			}
		}
		if i < n.MaxEntries() {
			key := n.Keys[i]
			if start.Compare(key.Hash) <= 0 && end.Compare(key.Hash) >= 0 && !key.Id.Synthetic() {
				if err := f(key.Clone()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Walk the tree in key order from start to end inclusive
func (t *Tree) Walk(start, end Hash, f WalkFunc) error {
	return t.walk(t.root.Id, start, end, f)
}

func (t *Tree) Get(hash Hash) (*Key, error) {
	var result *Key
	return result, t.walk(t.root.Id, hash, hash, func(key *Key) error {
		result = key
		return nil
	})
}

func (t *Tree) each(id NodeId, level int, f NodeFunc) error {
	n, err := t.keys.Get(id, t.Degree)
	if err != nil {
		return err
	}
	if err := f(level, n); err != nil {
		return err
	}
	return n.Each(func(id NodeId, start, end Hash) (NodeId, error) {
		if id.Empty() {
			return id, nil
		}
		return id, t.each(id, level+1, f)
	})
}

// Visit each node
func (t *Tree) Each(f NodeFunc) error {
	return t.each(t.root.Id, 0, f)
}

func (t *Tree) Dump(w io.Writer) error {
	return t.Each(func(level int, n *Node) error {
		indent := strings.Repeat("\t", level)
		for _, line := range strings.Split(n.String(), "\n") {
			if _, err := fmt.Fprintf(w, "%s%s\n", indent, line); err != nil {
				return err
			}
		}
		return nil
	})
}
