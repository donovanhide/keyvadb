package keyvadb

import . "gopkg.in/check.v1"

func (s *KeyVaSuite) TestHash(c *C) {
	h0 := MustHash("0000000000000000000000000000000000000000000000000000000000000000")
	h1 := MustHash("1111111111111111111111111111111111111111111111111111111111111111")
	h2 := MustHash("2222222222222222222222222222222222222222222222222222222222222222")
	h3 := MustHash("3333333333333333333333333333333333333333333333333333333333333333")
	h4 := MustHash("4444444444444444444444444444444444444444444444444444444444444444")
	c.Assert(h2.Distance(h1).Equals(h1), Equals, true)
	c.Assert(h1.Distance(h2).Equals(h1), Equals, true)
	c.Assert(h4.Stride(h1, 3).Equals(h1), Equals, true)
	c.Assert(h4.Stride(h2, 2).Equals(h1), Equals, true)
	c.Assert(h2.Closest(h1, h4), Equals, true)
	index, distance := h3.NearestStride(h0.Big(), h2.Big(), h2.Divide(2).Big())
	c.Assert(index, Equals, 1)
	c.Assert(distance.Equals(h1), Equals, true)
	index, distance = h3.NearestStride(h0.Big(), h1.Big(), h1.Divide(2).Big())
	c.Assert(index, Equals, 3)
	c.Assert(distance.Equals(h0), Equals, true)
}
