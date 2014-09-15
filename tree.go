package keyvadb

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type Tree struct {
	Degree   uint64
	root     *Node
	keys     KeyStore
	values   ValueStore
	balancer Balancer
}

func NewTree(degree uint64, keys KeyStore, values ValueStore, balancer Balancer) (*Tree, error) {
	if degree < 2 {
		return nil, fmt.Errorf("degree must be 2 or above")
	}
	root, err := keys.Get(0)
	switch {
	case err == ErrNotFound:
		root = NewNode(FirstHash, LastHash, 0, degree)
		// root.AddSyntheticKeys()
	case err != nil:
		return nil, err
	}
	if err := keys.Set(root); err != nil {
		return nil, err
	}
	return &Tree{
		Degree:   degree,
		root:     root,
		keys:     keys,
		values:   values,
		balancer: balancer,
	}, nil
}

func (t *Tree) add(n *Node, v KeySlice) (insertions int, err error) {
	if len(v) == 0 {
		panic("no values to add")
	}
	maxInsertions := n.TotalEmpty()
	debugPrintln(n)
	remainder := t.balancer.Balance(n, v)
	debugPrintln(n)
	insertions = len(v) - len(remainder)
	if insertions > maxInsertions {
		panic(fmt.Sprintf("too many insertions: %d max: %d", insertions, maxInsertions))
	}
	if *debug && !n.SanityCheck() {
		panic(fmt.Sprintf("not sane:\n%s", n))
	}
	if insertions == len(v) {
		return
	}
	err = n.Each(func(id NodeId, start, end Hash) (NodeId, error) {
		candidates := remainder.GetRange(start, end)
		if len(candidates) == 0 {
			return id, nil
		}
		var child *Node
		if id == EmptyChild {
			if child, err = t.keys.New(start, end, t.Degree); err != nil {
				return id, nil
			}
			id = child.Id
		} else {
			if child, err = t.keys.Get(id); err != nil {
				return id, err
			}
		}
		childInsertions, err := t.add(child, candidates)
		insertions += childInsertions
		return id, err
	})
	if len(v) != insertions {
		panic(fmt.Sprintf("Wrong number of insertions: Expected:%d Got:%d\n", len(v), insertions))
	}
	return
}

// Returns number of keys inserted and an error if encountered
func (t *Tree) Add(keys KeySlice) (int, error) {
	if !keys.IsSorted() {
		return 0, fmt.Errorf("unsorted values provided")
	}
	return t.add(t.root, keys)
}

func (t *Tree) get(n *Node, hash Hash) (*Key, error) {
	key, cid, err := n.GetKeyOrChild(hash)
	switch {
	case err != nil:
		return nil, err
	case key != nil:
		return key, nil
	default:
		child, err := t.keys.Get(cid)
		if err != nil {
			return nil, err
		}
		return t.get(child, hash)
	}
}

func (t *Tree) Get(hash Hash) (*Key, error) {
	return t.get(t.root, hash)
}

type WalkFunc func(key *Key)

var finishedWalkErr = errors.New("finished walking")

func (t *Tree) walk(n *Node, start, end Hash, f WalkFunc) error {
	for i, cid := range n.Children {
		key := n.Keys[min(i, n.MaxEntries()-1)]
		if start.Less(key.Hash) && cid != EmptyChild {
			child, err := t.keys.Get(cid)
			if err != nil {
				return err
			}
			if err := t.walk(child, start, end, f); err != nil {
				return err
			}
		}
		if end.Compare(key.Hash) < 0 {
			return finishedWalkErr
		}
		if i < n.MaxEntries() && start.Compare(key.Hash) <= 0 && !key.Id.Synthetic() {
			f(key.Clone())
		}
	}
	return nil
}

// Walk the tree in key order from start to end inclusive
func (t *Tree) Walk(start, end Hash, f WalkFunc) error {
	if err := t.walk(t.root, start, end, f); err != finishedWalkErr {
		return err
	}
	return nil
}

func (t *Tree) each(level int, n *Node, f NodeFunc) error {
	if err := f(level, n); err != nil {
		return err
	}
	return n.Each(func(id NodeId, start, end Hash) (NodeId, error) {
		if id.Empty() {
			return id, nil
		}
		child, err := t.keys.Get(id)
		if err != nil {
			return id, err
		}
		return id, t.each(level+1, child, f)
	})
}

// Visit each node
func (t *Tree) Each(f NodeFunc) error {
	return t.each(0, t.root, f)
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
