package keyvadb

import (
	"sync/atomic"
)

func NewMemoryKeyStore() KeyStore {
	return &MemoryKeyStore{
		cache: make(map[NodeId]*Node),
	}
}

type MemoryKeyStore struct {
	length int64
	cache  map[NodeId]*Node
}

func (m *MemoryKeyStore) New(start, end Hash, degree uint64) (*Node, error) {
	id := NodeId(atomic.AddInt64(&m.length, 1))
	debugPrintln("Memory New Key:", id)
	node := NewNode(start, end, id, degree)
	return node, nil
}

func (m *MemoryKeyStore) Set(node *Node) error {
	debugPrintln("Memory Set Key:", node.Id)
	m.cache[node.Id] = node
	return nil
}

func (m *MemoryKeyStore) Get(id NodeId, degree uint64) (*Node, error) {
	debugPrintln("Memory Get Key:", id)
	if node, ok := m.cache[id]; ok {
		return node.Clone(), nil
	}
	return nil, ErrNotFound
}

func (m *MemoryKeyStore) Sync() error {
	return nil
}

func (m *MemoryKeyStore) Close() error {
	return nil
}

func (m *MemoryKeyStore) Length() int64 {
	return atomic.LoadInt64(&m.length)
}

func NewMemoryValueStore() ValueStore {
	return &MemoryValueStore{}
}

type MemoryValueStore struct {
	length int64
	cache  []*KeyValue
}

func (m *MemoryValueStore) Append(key Hash, value []byte) (*KeyValue, error) {
	id := ValueId(atomic.AddInt64(&m.length, 1) - 1)
	kv := NewKeyValue(id, key, value)
	m.cache = append(m.cache, kv)
	return kv, nil
}

func (m *MemoryValueStore) Get(id ValueId) (*KeyValue, error) {
	if int(id) >= len(m.cache) {
		return nil, ErrNotFound
	}
	return m.cache[id], nil
}

func (m *MemoryValueStore) Each(f func(*KeyValue)) error {
	for _, v := range m.cache {
		f(v)
	}
	return nil
}

func (m *MemoryValueStore) Close() error {
	return nil
}

func (m *MemoryValueStore) Sync() error {
	return nil
}

func (m *MemoryValueStore) Length() int64 {
	return atomic.LoadInt64(&m.length)
}
