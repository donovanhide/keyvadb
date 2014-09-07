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
	balancer BalancerFunc
}

func NewTree(keys KeyStore, values ValueStore, f BalancerFunc) (*Tree, error) {
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
		balancer: f,
	}, nil
}

type BalancerFunc func(n *Node, v ValueSlice) NeighbourSlice

func (t *Tree) add(n *Node, v ValueSlice) error {
	if len(v) == 0 {
		return nil
	}
	debugPrintln(n)
	neighbours := t.balancer(n, v)
	for _, neighbour := range neighbours {
		n.Keys[neighbour.Index] = neighbour.Key
		n.Values[neighbour.Index] = neighbour.Id
	}
	childrenRanges := n.Ranges()
	for i := 1; i <= ChildCount; i++ {
		childStart, childEnd := childrenRanges[i-1], childrenRanges[i]
		if childEnd == EmptyItem {
			break
		}
		candidates := v.GetRange(childStart, childEnd)
		if len(candidates) == 0 {
			continue
		}
		var child *Node
		var err error
		if id := n.Children[i-1]; id == EmptyChild {
			child, err = t.keys.New(childStart, childEnd)
			n.Children[i-1] = child.Id
		} else {
			child, err = t.keys.Get(id)
		}
		if err != nil {
			return err
		}
		if err := t.add(child, candidates); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tree) Add(values ValueSlice) error {
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
	if err := t.Each(func(level int, n *Node) error {
		levels.Add(n, level)
		return nil
	}); err != nil {
		return nil, err
	}
	return levels, nil
}
