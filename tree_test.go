package keyva

import . "gopkg.in/check.v1"

func (s *KeyVaSuite) TestTree(c *C) {
	for name, balancer := range balancers {
		ks := NewMemoryKeyStore()
		vs := NewMemoryValueStore()
		msg := Commentf(name)
		tree, err := NewTree(ks, vs, balancer)
		c.Assert(err, IsNil, msg)
		n := 10000
		rounds := 10
		gen := NewRandomValueGenerator(10, 50, s.R)
		sum := 0
		for i := 0; i < rounds; i++ {
			kv, err := gen.Take(n)
			c.Assert(err, IsNil, msg)
			keys := kv.Keys()
			keys.Sort()
			n, err := tree.Add(keys)
			c.Assert(err, IsNil, msg)
			c.Assert(n, Equals, len(keys), msg)
			levels, err := tree.Levels()
			c.Assert(err, IsNil, msg)
			c.Log(levels)
			// c.Assert(tree.Dump(os.Stdout), IsNil)
			sum += n
			total := levels.Total()
			c.Assert(total.Entries-total.Synthetics, Equals, sum, msg)
		}
	}
}
