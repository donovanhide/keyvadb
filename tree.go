package keyva

import (
	"fmt"
	"io"
	"strings"
)

type Tree struct {
	root     *Node
	keys     KeyStore
	values   ValueStore
	balancer Balancer
}

func NewTree(keys KeyStore, values ValueStore, balancer Balancer) (*Tree, error) {
	root, err := keys.Get(0)
	switch {
	case err == ErrNotFound:
		root = &Node{
			Start: FirstHash,
			End:   LastHash,
		}
	case err != nil:
		return nil, err
	}
	if err := keys.Set(root); err != nil {
		return nil, err
	}
	return &Tree{
		root:     root,
		keys:     keys,
		values:   values,
		balancer: balancer,
	}, nil
}

func (t *Tree) add(n *Node, v ValueSlice) (insertions int, err error) {
	if len(v) == 0 {
		panic("no values to add")
	}
	debugPrintln(n)
	insertions = t.balancer.Balance(n, v)
	if !n.SanityCheck() || insertions > len(v) {
		panic(fmt.Sprintf("not sane:\n%s", n))
	}
	if insertions == len(v) {
		return
	}
	childrenRanges := n.Ranges()
	for i := 0; i < ChildCount; i++ {
		childStart, childEnd := childrenRanges[i], childrenRanges[i+1]
		if childStart.Empty() || childEnd.Empty() {
			continue
		}
		candidates := v.GetRange(childStart, childEnd)
		if len(candidates) == 0 {
			continue
		}
		var child *Node
		if id := n.Children[i]; id == EmptyChild {
			if child, err = t.keys.New(childStart, childEnd); err != nil {
				return
			}
			n.Children[i] = child.Id
		} else {
			if child, err = t.keys.Get(id); err != nil {
				return
			}
		}
		var childInsertions int
		childInsertions, err = t.add(child, candidates)
		insertions += childInsertions
		if err != nil {
			return
		}
	}
	return
}

// Returns number of values inserted and an error if encountered
func (t *Tree) Add(values ValueSlice) (int, error) {
	if !values.IsSorted() {
		return 0, fmt.Errorf("unsorted values provided")
	}
	return t.add(t.root, values)
}

func (t *Tree) each(level int, n *Node, f nodeFunc) error {
	if err := f(level, n); err != nil {
		return err
	}
	for _, cid := range n.Children {
		if cid == EmptyChild {
			continue
		}
		child, err := t.keys.Get(cid)
		if err != nil {
			return err
		}
		if err := t.each(level+1, child, f); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tree) Each(f nodeFunc) error {
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

func (t *Tree) Levels() (LevelSlice, error) {
	var levels LevelSlice
	err := t.Each(func(level int, n *Node) error {
		levels.Add(n, level)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return levels, nil
}
