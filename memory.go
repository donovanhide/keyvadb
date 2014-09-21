package keyvadb

import (
	"sync/atomic"
)

type MemoryKeyStore struct {
	length uint64
	cache  map[NodeId]*Node
}

type MemoryValueStore struct {
	cache []*KeyValue
}

func NewMemoryKeyStore() KeyStore {
	return &MemoryKeyStore{
		cache: make(map[NodeId]*Node),
	}
}

func (m *MemoryKeyStore) New(start, end Hash, degree uint64) (*Node, error) {
	id := NodeId(atomic.AddUint64(&m.length, 1))
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

func NewMemoryValueStore() ValueStore {
	return &MemoryValueStore{}
}

func (m *MemoryValueStore) Append(key Hash, value []byte) (*KeyValue, error) {
	kv := NewKeyValue(ValueId(len(m.cache)), key, value)
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
