package keyvadb

import "fmt"

type DB struct {
	tree   *Tree
	batch  uint64
	buffer map[Hash]Key
	keys   KeyStore
	values ValueStore
}

func newDB(degree, batch uint64, name string, keys KeyStore, values ValueStore) (*DB, error) {
	balancer, err := newBalancer(name)
	if err != nil {
		return nil, err
	}
	tree, err := NewTree(degree, keys, balancer)
	if err != nil {
		return nil, err
	}
	return &DB{
		tree:   tree,
		batch:  batch,
		buffer: make(map[Hash]Key, int(batch)),
		keys:   keys,
		values: values,
	}, nil
}

func NewMemoryDB(degree, batch uint64, balancer string) (*DB, error) {
	return newDB(degree, batch, balancer, NewMemoryKeyStore(), NewMemoryValueStore())
}

func NewFileDB(degree, batch uint64, balancer, filename string) (*DB, error) {
	values, err := NewFileValueStore(filename)
	if err != nil {
		return nil, err
	}
	keys, err := NewFileKeyStore(filename)
	if err != nil {
		return nil, err
	}
	return newDB(degree, batch, balancer, keys, values)
}

func (db *DB) Add(key Hash, value []byte) error {
	kv, err := db.values.Append(key, value)
	if err != nil {
		return err
	}
	db.buffer[key] = kv.Key
	if uint64(len(db.buffer)) > db.batch {
		var keys KeySlice
		for _, key := range db.buffer {
			keys = append(keys, key)
		}
		keys.Sort()
		n, err := db.tree.Add(keys)
		switch {
		case err != nil:
			return err
		case n != len(keys):
			return fmt.Errorf("Too few keys added: %d expected %d", n, len(keys))
		default:
			db.buffer = make(map[Hash]Key, int(db.batch))
		}
	}
	return nil
}

func (db *DB) Get(hash Hash) (*KeyValue, error) {
	if key, ok := db.buffer[hash]; ok {
		return db.values.Get(key.Id)
	}
	key, err := db.tree.Get(hash)
	if err != nil {
		return nil, err
	}
	return db.values.Get(key.Id)

}
