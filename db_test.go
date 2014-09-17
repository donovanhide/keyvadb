package keyvadb

import (
	"os"

	. "gopkg.in/check.v1"
)

func (s *KeyVaSuite) fillDB(db *DB, c *C) {
	gen := NewRandomValueGenerator(10, 40, s.R)
	kvs, err := gen.Take(100)
	c.Assert(err, IsNil)
	for _, kv := range kvs {
		c.Assert(db.Add(kv.Hash, kv.Value), IsNil)
	}
	for i, hash := range kvs.Keys().Hashes() {
		result, err := db.Get(hash)
		c.Assert(err, IsNil)
		c.Assert(kvs[i].Hash, Equals, result.Hash)
		c.Assert(kvs[i].Value, DeepEquals, result.Value)
	}
}

func (s *KeyVaSuite) TestMemoryDB(c *C) {
	db, err := NewMemoryDB(8, 10, "Distance")
	c.Assert(err, IsNil)
	s.fillDB(db, c)
}

func (s *KeyVaSuite) TestFileDB(c *C) {
	os.Remove("test.values")
	os.Remove("test.keys")
	db, err := NewFileDB(8, 10, "Distance", "test")
	c.Assert(err, IsNil)
	s.fillDB(db, c)
}
