package keyvadb

import . "gopkg.in/check.v1"

func (s *KeyVaSuite) TestTree(c *C) {
	for _, b := range Balancers {
		ks := NewMemoryKeyStore()
		vs := NewMemoryValueStore()
		msg := Commentf(b.Name)
		var allKeys KeySlice
		tree, err := NewTree(8, ks, vs, b.Balancer)
		c.Assert(err, IsNil, msg)
		batch := 1000
		rounds := 10
		gen := NewRandomValueGenerator(10, 50, s.R)
		sum := 0
		for i := 0; i < rounds; i++ {
			kv, err := gen.Take(batch)
			c.Assert(err, IsNil, msg)
			keys := kv.Keys()
			keys.Sort()
			allKeys = append(allKeys, keys...)
			n, err := tree.Add(keys)
			c.Assert(err, IsNil, msg)
			c.Assert(n, Equals, len(keys), msg)
			sum += n
			summary, err := NewSummary(tree)
			c.Assert(err, IsNil, msg)
			c.Logf("%08d: %12s: %s", i, b.Name, summary.Overall())
			// c.Assert(tree.Dump(os.Stdout), IsNil)
			c.Assert(summary.Total.NonSyntheticEntries(), Equals, uint64(sum), msg)
			// Add them again
			c.Log(b.Name)
			n, err = tree.Add(keys)
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
		err = tree.Walk(FirstHash, LastHash, func(key *Key) {
			// c.Log(key, allKeys[i])
			c.Assert(key.Equals(allKeys[i]), Equals, true)
			i++
		})
		c.Assert(err, IsNil)
		c.Assert(i, Equals, len(allKeys))
		// Check subset of keys are walked in order
		j := 100
		start, end := allKeys[j].Hash, allKeys[len(allKeys)-100].Hash
		err = tree.Walk(start, end, func(key *Key) {
			// c.Log(key, allKeys[j])
			c.Assert(key.Equals(allKeys[j]), Equals, true)
			j++
		})
		c.Assert(err, IsNil)
		c.Assert(j, Equals, len(allKeys)-99)
	}
}
