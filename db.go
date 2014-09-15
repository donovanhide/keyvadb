package keyvadb

type DB struct {
	tree   *Tree
	keys   KeyStore
	values ValueStore
}

func newDB(degree uint64, name string, keys KeyStore, values ValueStore) (*DB, error) {
	balancer, err := newBalancer(name)
	if err != nil {
		return nil, err
	}
	tree, err := NewTree(degree, keys, values, balancer)
	if err != nil {
		return nil, err
	}
	return &DB{
		tree: tree,
	}, nil
}

func NewMemoryDB(degree uint64, balancer string) (*DB, error) {
	return newDB(degree, balancer, NewMemoryKeyStore(), NewMemoryValueStore())
}

func NewFileDB(degree uint64, balancer, filename string) (*DB, error) {
	values, err := NewFileValueStore(filename)
	if err != nil {
		return nil, err
	}
	keys, err := NewFileKeyStore(filename)
	if err != nil {
		return nil, err
	}
	return newDB(degree, balancer, keys, values)
}

func (db *DB) Add(s KeyValueSlice) error              { return nil }
func (db *DB) Get(s HashSlice) (KeyValueSlice, error) { return nil, nil }
