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
	values   ValueStore
	balancer Balancer
}

func NewTree(degree uint64, keys KeyStore, values ValueStore, balancer Balancer) (*Tree, error) {
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
	err = n.Each(func(id uint64, start, end Hash) (uint64, error) {
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

// Returns number of values inserted and an error if encountered
func (t *Tree) Add(values KeySlice) (int, error) {
	if !values.IsSorted() {
		return 0, fmt.Errorf("unsorted values provided")
	}
	return t.add(t.root, values)
}

func (t *Tree) each(level int, n *Node, f NodeFunc) error {
	if err := f(level, n); err != nil {
		return err
	}
	return n.Each(func(id uint64, start, end Hash) (uint64, error) {
		if id == EmptyChild {
			return id, nil
		}
		child, err := t.keys.Get(id)
		if err != nil {
			return id, err
		}
		return id, t.each(level+1, child, f)
	})
}

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
