package keyvadb

type NodeCache map[NodeId]*Node
type KeyValueCache []*KeyValue

type MemoryKeyStore struct {
	cache NodeCache
}

type MemoryValueStore struct {
	cache KeyValueCache
}

func NewMemoryKeyStore() KeyStore {
	return &MemoryKeyStore{
		cache: make(NodeCache),
	}
}

func (m *MemoryKeyStore) New(start, end Hash, degree uint64) (*Node, error) {
	id := NodeId(len(m.cache))
	node := NewNode(start, end, id, degree)
	m.cache[node.Id] = node
	return node, nil
}

func (m *MemoryKeyStore) Set(node *Node) error {
	m.cache[node.Id] = node
	return nil
}

func (m *MemoryKeyStore) Get(id NodeId) (*Node, error) {
	if node, ok := m.cache[id]; ok {
		return node, nil
	}
	return nil, ErrNotFound
}

func NewMemoryValueStore() ValueStore {
	return &MemoryValueStore{}
}

func (m *MemoryValueStore) Append(key Hash, value []byte) (*KeyValue, error) {
	kv := &KeyValue{
		Key: Key{
			Id:   ValueId(len(m.cache)),
			Hash: key,
		},
		Value: value,
	}
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
