package keyvadb

import (
	"errors"
)

var ErrNotFound = errors.New("key not found")

type KeyStore interface {
	New(start, end Hash, degree uint64) (*Node, error)
	Set(*Node) error
	Get(id NodeId, degree uint64) (*Node, error)
}

type ValueStore interface {
	Append(Hash, []byte) (*KeyValue, error)
	Get(id ValueId) (*KeyValue, error)
	Each(func(*KeyValue)) error
}

type Journal interface {
	Swap(current, previous *Node)
	Commit() error
	String() string
}

type Balancer interface {
	Balance(*Node, KeySlice) (KeySlice, bool)
}
