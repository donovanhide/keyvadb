package keyvadb

import (
	"errors"
)

var ErrNotFound = errors.New("key not found")

type KeyStore interface {
	New(start, end Hash, degree uint64) (*Node, error)
	Set(*Node) error
	Get(id uint64) (*Node, error)
}

type ValueStore interface {
	Append(*KeyValue) (uint64, error)
	Get(id uint64) (*KeyValue, error)
	Each(func(int, *KeyValue) error) error
}

type Balancer interface {
	Balance(*Node, KeySlice) KeySlice
}
