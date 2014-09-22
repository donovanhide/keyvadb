package keyvadb

import (
	"errors"
)

var ErrNotFound = errors.New("key not found")

type KeyStore interface {
	New(start, end Hash, degree uint64) (*Node, error)
	Set(*Node) error
	Get(id NodeId, degree uint64) (*Node, error)
	Close() error
	Sync() error
}

type ValueStore interface {
	Append(Hash, []byte) (*KeyValue, error)
	Get(id ValueId) (*KeyValue, error)
	Each(func(*KeyValue)) error
	Close() error
	Sync() error
}

type Journal interface {
	Swap(current, previous *Node)
	Commit() error
	Len() int
	String() string
	Close() error
}

type Balancer interface {
	Balance(*Node, KeySlice) (*Node, KeySlice)
}
