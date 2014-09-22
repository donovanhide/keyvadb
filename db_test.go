package keyvadb

import (
	"os"

	. "gopkg.in/check.v1"
)

func (s *KeyVaSuite) fillDB(rounds, n int, db *DB, c *C) {
	gen := NewRandomValueGenerator(10, 40, s.R)
	for i := 0; i < rounds; i++ {
		kvs, err := gen.Take(n)
		c.Assert(err, IsNil)
		for _, kv := range kvs {
			c.Assert(db.Add(kv.Hash, kv.Value), IsNil)
		}
		c.Logf("Added Round(%d/%d)", i+1, rounds)
		for i, hash := range kvs.Keys().Hashes() {
			result, err := db.Get(hash)
			c.Assert(err, IsNil, Commentf("%s", hash))
			c.Assert(kvs[i].Hash, Equals, result.Hash)
			c.Assert(kvs[i].Value, DeepEquals, result.Value)
		}
		c.Logf("Searched Round(%d/%d)", i+1, rounds)
		c.Log(db)
	}
}

func (s *KeyVaSuite) TestMemoryDB(c *C) {
	db, err := NewMemoryDB(10, 1000, "Distance")
	c.Assert(err, IsNil)
	s.fillDB(10, 10000, db, c)
}

func (s *KeyVaSuite) TestFileDB(c *C) {
	os.Remove("test.values")
	os.Remove("test.keys")
	db, err := NewFileDB(84, 3, 10000, "Distance", "test")
	c.Assert(err, IsNil)
	s.fillDB(10, 10000, db, c)
}
