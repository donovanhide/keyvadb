package keyvadb

import . "gopkg.in/check.v1"

func (s *KeyVaSuite) TestMemoryDB(c *C) {
	db, err := NewMemoryDB(8, 10, "Distance")
	c.Assert(err, IsNil)
	gen := NewRandomValueGenerator(10, 40, s.R)
	kvs, err := gen.Take(1000)
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
