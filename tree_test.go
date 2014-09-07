package keyva

import . "gopkg.in/check.v1"

func (s *KeyVaSuite) TestTree(c *C) {
	tree, err := NewTree(s.Keys, s.Values, balancer)
	n := 100000
	rounds := 1
	values, err := newRandomValues(n, 100, 500, s.R)
	c.Assert(err, IsNil)
	for i := 0; i < rounds; i++ {
		start := (n / rounds) * i
		end := start + (n / rounds)
		section := values[start:end]
		section.Sort()
		err = tree.Add(values)
		c.Assert(err, IsNil)
		levels, err := tree.Levels()
		c.Assert(err, IsNil)
		c.Log(levels)
		// c.Assert(tree.Dump(os.Stdout), IsNil)
	}
}
