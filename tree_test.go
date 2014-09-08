package keyva

import . "gopkg.in/check.v1"

func (s *KeyVaSuite) TestTree(c *C) {
	tree, err := NewTree(s.Keys, s.Values, &RandomBalancer{})
	c.Assert(err, IsNil)
	n := 100000
	rounds := 10
	gen := NewRandomValueGenerator(10, 50, s.R)
	sum := 0
	for i := 0; i < rounds; i++ {
		values, err := gen.Take(n)
		c.Assert(err, IsNil)
		values.Sort()
		n, err := tree.Add(values)
		c.Assert(err, IsNil)
		c.Assert(n, Equals, len(values))
		levels, err := tree.Levels()
		c.Assert(err, IsNil)
		c.Log(levels)
		// c.Assert(tree.Dump(os.Stdout), IsNil)
		sum += n
		c.Assert(levels.Total().Entries, Equals, sum)
	}
}
