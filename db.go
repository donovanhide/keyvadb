package keyvadb

import "github.com/golang/glog"

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
	tree     *Tree
	buffer   *Buffer
	incoming chan *Key
	flushed  chan bool
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
	db := &DB{
		tree:     tree,
		buffer:   NewBuffer(conf.batch),
		incoming: make(chan *Key, conf.batch*2),
		flushed:  make(chan (bool), 1),
		DBConfig: conf,
	}
	go db.run()
	return db, nil
}

func (db *DB) Close() error {
	if err := db.values.Close(); err != nil {
		return err
	}
	if err := db.keys.Close(); err != nil {
		return err
	}
	return db.journal.Close()
}

func (db *DB) Add(key Hash, value []byte) error {
	kv, err := db.values.Append(key, value)
	if err != nil {
		return err
	}
	db.incoming <- kv.CloneKey()
	return nil
}

func (db *DB) Get(hash Hash) (*KeyValue, error) {
	if key := db.buffer.Get(hash); key != nil {
		return db.values.Get(key.Id)
	}
	key, err := db.tree.Get(hash)
	if err != nil {
		return nil, err
	}
	return db.values.Get(key.Id)
}

func (db *DB) run() {
	flushed := true
	for {
		select {
		case flushed = <-db.flushed:
			//flushing updated
		case key := <-db.incoming:
			if n := db.buffer.Add(key); n >= db.batch && flushed {
				flushed = false
				go db.flush()
			}
		}
	}
}

func (db *DB) flush() {
	keys := db.buffer.Keys()
	keys.Sort()
	n, err := db.tree.Add(keys, db.journal)
	switch {
	case err != nil:
		glog.Fatalf("Tree Add Error: %s Closing with result:%+v", err.Error(), db.Close())
	case n != len(keys):
		glog.Fatalf("Too few keys added: %d expected %d: %s Closing with result: %+v", n, len(keys), db.Close())
	}
	if err := db.journal.Commit(); err != nil {
		glog.Fatalf("Commit Error: %s Closing with result: %+v", err, db.Close())
	}
	db.buffer.Remove(keys)
	db.flushed <- true
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
