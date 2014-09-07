package keyva

type NodeCache map[uint64]*Node
type ValueCache []*Value

type MemoryKeyStore struct {
	cache NodeCache
}

type MemoryValueStore struct {
	cache ValueCache
}

func NewMemoryKeyStore() KeyStore {
	return &MemoryKeyStore{
		cache: make(NodeCache),
	}
}

func (m *MemoryKeyStore) New(start, end Hash) (*Node, error) {
	node := &Node{
		Id:    uint64(len(m.cache)),
		Start: start,
		End:   end,
	}
	m.cache[node.Id] = node
	return node, nil
}

func (m *MemoryKeyStore) Set(node *Node) error {
	m.cache[node.Id] = node
	return nil
}

func (m *MemoryKeyStore) Get(id uint64) (*Node, error) {
	if node, ok := m.cache[id]; ok {
		return node, nil
	}
	return nil, ErrNotFound
}

func NewMemoryValueStore() ValueStore {
	return &MemoryValueStore{}
}

func (m *MemoryValueStore) Append(v *Value) (uint64, error) {
	id := uint64(len(m.cache))
	m.cache = append(m.cache, v)
	return id, nil
}

func (m *MemoryValueStore) Get(id uint64) (*Value, error) {
	if int(id) >= len(m.cache) {
		return nil, ErrNotFound
	}
	return m.cache[id], nil
}

func (m *MemoryValueStore) Each(f func(int, *Value) error) error {
	for i, v := range m.cache {
		if err := f(i, v); err != nil {
			return err
		}
	}
	return nil
}
