package keyvadb

import . "gopkg.in/check.v1"

func (s *KeyVaSuite) TestTree(c *C) {
	for _, b := range Balancers {
		keys := NewMemoryKeyStore()
		values := NewMemoryValueStore()
		msg := Commentf(b.Name)
		var allKeys KeySlice
		tree, err := NewTree(8, keys, b.Balancer)
		c.Assert(err, IsNil, msg)
		batch := 1000
		rounds := 10
		gen := NewRandomValueGenerator(10, 50, s.R)
		sum := 0
		journal := NewSimpleJournal("test", keys, values)
		for i := 0; i < rounds; i++ {
			kv, err := gen.Take(batch)
			c.Assert(err, IsNil, msg)
			keys := kv.Keys()
			keys.Sort()
			allKeys = append(allKeys, keys...)
			n, err := tree.Add(keys, journal)
			c.Assert(err, IsNil, msg)
			c.Assert(n, Equals, len(keys), msg)
			c.Assert(journal.Len(), Not(Equals), 0)
			// c.Log(journal)
			c.Assert(journal.Commit(), IsNil)
			c.Assert(journal.Len(), Equals, 0)
			sum += n
			summary, err := NewSummary(tree)
			c.Assert(err, IsNil, msg)
			c.Logf("%08d: %12s: %s", i, b.Name, summary.Overall())
			// c.Assert(tree.Dump(os.Stdout), IsNil)
			c.Assert(summary.Total.NonSyntheticEntries(), Equals, uint64(sum), msg)
			// Add them again
			n, err = tree.Add(keys, journal)
			c.Assert(err, IsNil, msg)
			c.Assert(n, Equals, batch, msg)
		}
		// Check all keys were added
		for _, key := range allKeys {
			found, err := tree.Get(key.Hash)
			c.Assert(err, IsNil)
			c.Assert(found.Equals(key), Equals, true)
		}
		// Check all keys can be walked in order
		i := 0
		allKeys.Sort()
		err = tree.Walk(FirstHash, LastHash, func(key *Key) error {
			// c.Log(key, allKeys[i])
			c.Assert(key.Equals(allKeys[i]), Equals, true)
			i++
			return nil
		})
		c.Assert(err, IsNil)
		c.Assert(i, Equals, len(allKeys))
		// Check subset of keys are walked in order
		j := 100
		start, end := allKeys[j].Hash, allKeys[len(allKeys)-100].Hash
		err = tree.Walk(start, end, func(key *Key) error {
			// c.Log(key, allKeys[j])
			c.Assert(key.Equals(allKeys[j]), Equals, true)
			j++
			return nil
		})
		c.Assert(err, IsNil)
		c.Assert(j, Equals, len(allKeys)-99)
	}
}
