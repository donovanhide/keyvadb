package keyva

import (
	"errors"
)

var ErrNotFound = errors.New("key not found")

type KeyStore interface {
	New(start, end Hash) (*Node, error)
	Set(*Node) error
	Get(id uint64) (*Node, error)
}

type ValueStore interface {
	Append(*Value) (uint64, error)
	Get(id uint64) (*Value, error)
	Each(func(int, *Value) error) error
}
