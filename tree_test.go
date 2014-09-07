package keyva

import . "gopkg.in/check.v1"

func (s *KeyVaSuite) TestTree(c *C) {
	tree, err := NewTree(s.Keys, s.Values, &MatchingBalancer{})
	c.Assert(err, IsNil)
	n := 1000
	rounds := 1
	gen := NewRandomValueGenerator(100, 500, s.R)
	for i := 0; i < rounds; i++ {
		values, err := gen.Take(n)
		c.Assert(err, IsNil)
		values.Sort()
		c.Assert(tree.Add(values), IsNil)
		levels, err := tree.Levels()
		c.Assert(err, IsNil)
		c.Log(levels)
		// c.Assert(tree.Dump(os.Stdout), IsNil)
	}
}
