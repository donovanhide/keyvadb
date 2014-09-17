package keyvadb

import . "gopkg.in/check.v1"

func (s *KeyVaSuite) TestKeys(c *C) {
	gen := NewRandomValueGenerator(100, 500, s.R)
	kv, err := gen.Take(100)
	c.Check(err, IsNil)
	keys := kv.Keys()
	keys.Sort()
	c.Assert(keys[0].Hash.String(), Equals, "0033F3A564EA5A6A5DA1CA4C13DE4243081771717FFFB0D81CF7ACC75652063F")
	c.Assert(keys[99].Hash.String(), Equals, "FE86CD86639D2F85490FD133A482D836FE523BE36C2BF7B3A387C426B9A5171A")
	start := MustHash("4000000000000000000000000000000000000000000000000000000000000000")
	end := MustHash("7000000000000000000000000000000000000000000000000000000000000000")
	maxDistance := MustHash("3000000000000000000000000000000000000000000000000000000000000000")
	section := keys.GetRange(start, end)
	c.Assert(len(section) > 0, Equals, true)
	c.Assert(section[0].Hash.Compare(start) > 0, Equals, true)
	c.Assert(section[0].Hash.Compare(end) < 0, Equals, true)
	c.Assert(section[0].Hash.Distance(section[len(section)-1].Hash).Compare(maxDistance) <= 0, Equals, true)
}

var k1 = KeySlice{
	{Hash: MustHash("4000000000000000000000000000000000000000000000000000000000000000"), Id: 0},
	{Hash: MustHash("6000000000000000000000000000000000000000000000000000000000000000"), Id: 1},
	{Hash: MustHash("4000000000000000000000000000000000000000000000000000000000000000"), Id: 0},
	{Hash: MustHash("4000000000000000000000000000000000000000000000000000000000000000"), Id: 2},
	{Hash: MustHash("6000000000000000000000000000000000000000000000000000000000000000"), Id: 0},
}
var k2 = KeySlice{
	{Hash: MustHash("4000000000000000000000000000000000000000000000000000000000000000"), Id: 0},
	{Hash: MustHash("6000000000000000000000000000000000000000000000000000000000000000"), Id: 1},
	{Hash: MustHash("8000000000000000000000000000000000000000000000000000000000000000"), Id: 0},
}

func (s *KeyVaSuite) TestKeyUnique(c *C) {
	k := k1.Clone()
	k.Sort()
	k.Unique()
	c.Assert(len(k), Equals, 2)
}
func (s *KeyVaSuite) TestKeyUnion(c *C) {
	u1 := k1.Clone()
	u2 := k2.Clone()
	u1.Sort()
	u2.Sort()
	u1.Unique()
	u2.Unique()
	u3 := u1.Union(u2)
	c.Assert(len(u3), Equals, 3)
}
