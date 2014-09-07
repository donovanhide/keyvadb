package keyva

import . "gopkg.in/check.v1"

func (s *KeyVaSuite) TestHash(c *C) {
	h1 := MustHash("1111111111111111111111111111111111111111111111111111111111111111")
	h2 := MustHash("2222222222222222222222222222222222222222222222222222222222222222")
	h3 := MustHash("4444444444444444444444444444444444444444444444444444444444444444")
	c.Assert(h2.Distance(h1).Equals(h1), Equals, true)
	c.Assert(h1.Distance(h2).Equals(h1), Equals, true)
	c.Assert(h3.Stride(h1, 3).Equals(h1), Equals, true)
	c.Assert(h3.Stride(h2, 2).Equals(h1), Equals, true)
	c.Assert(h2.Closest(h1, h3), Equals, true)
}
