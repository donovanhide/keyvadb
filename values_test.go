package keyva

import . "gopkg.in/check.v1"

func (s *KeyVaSuite) TestValues(c *C) {
	gen := NewRandomValueGenerator(100, 500, s.R)
	values, err := gen.Take(100)
	c.Check(err, IsNil)
	values.Sort()
	c.Assert(values[0].Key.String(), Equals, "0033F3A564EA5A6A5DA1CA4C13DE4243081771717FFFB0D81CF7ACC75652063F")
	c.Assert(values[99].Key.String(), Equals, "FE86CD86639D2F85490FD133A482D836FE523BE36C2BF7B3A387C426B9A5171A")
	start := MustHash("4000000000000000000000000000000000000000000000000000000000000000")
	end := MustHash("7000000000000000000000000000000000000000000000000000000000000000")
	maxDistance := MustHash("3000000000000000000000000000000000000000000000000000000000000000")
	section := values.GetRange(start, end)
	c.Assert(len(section) > 0, Equals, true)
	c.Assert(section[0].Key.Compare(start) > 0, Equals, true)
	c.Assert(section[0].Key.Compare(end) < 0, Equals, true)
	c.Assert(section[0].Key.Distance(section[len(section)-1].Key).Compare(maxDistance) <= 0, Equals, true)
}
