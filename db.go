package keyvadb

import "fmt"

type DBConfig struct {
	name     string
	degree   uint64
	batch    uint64
	balancer string
	keys     KeyStore
	values   ValueStore
	journal  Journal
}

type DB struct {
	*DBConfig
	tree   *Tree
	buffer map[Hash]Key
}

func newDB(conf *DBConfig) (*DB, error) {
	balancer, err := newBalancer(conf.balancer)
	if err != nil {
		return nil, err
	}
	tree, err := NewTree(conf.degree, conf.keys, balancer)
	if err != nil {
		return nil, err
	}
	return &DB{
		tree:     tree,
		buffer:   make(map[Hash]Key, int(conf.batch)),
		DBConfig: conf,
	}, nil
}

func NewMemoryDB(degree, batch uint64, balancer string) (*DB, error) {
	keys, values := NewMemoryKeyStore(), NewMemoryValueStore()
	return newDB(&DBConfig{
		degree:   degree,
		batch:    batch,
		balancer: balancer,
		keys:     keys,
		values:   values,
		journal:  NewSimpleJournal("Simple Journal", keys, values),
	})
}

func NewFileDB(degree, batch uint64, balancer, filename string) (*DB, error) {
	values, err := NewFileValueStore(filename)
	if err != nil {
		return nil, err
	}
	keys, err := NewFileKeyStore(degree, filename)
	if err != nil {
		return nil, err
	}
	journal, err := NewFileJournal(filename, keys, values)
	return newDB(&DBConfig{
		degree:   degree,
		batch:    batch,
		balancer: balancer,
		keys:     keys,
		values:   values,
		journal:  journal,
	})
}

func (db *DB) Add(key Hash, value []byte) error {
	kv, err := db.values.Append(key, value)
	if err != nil {
		return err
	}
	db.buffer[key] = kv.Key
	if uint64(len(db.buffer)) >= db.batch {
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

type KeyValueFunc func(*KeyValue)

func (db *DB) All(f KeyValueFunc) error {
	return db.values.Each(f)
}

func (db *DB) Range(start, end Hash, f KeyValueFunc) error {
	return db.tree.Walk(start, end, func(key *Key) error {
		kv, err := db.values.Get(key.Id)
		if err != nil {
			return err
		}
		f(kv)
		return nil
	})
}
