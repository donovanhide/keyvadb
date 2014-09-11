package keyvadb

// . "gopkg.in/check.v1"

var neighbourValues = KeySlice{
	{MustHash("0033F3A564EA5A6A5DA1CA4C13DE4243081771717FFFB0D81CF7ACC75652063F"), 0},
	{MustHash("0044F9B447A1F677E92BFCA281C9A1B15CDD98B5B1032C54E173E76D2DE27A71"), 0},
	{MustHash("008F2D4343ACDDB1B3A44BE0EC3EB624C1F921FBFB58A8F3A1B4EA749CDD72ED"), 0},
	{MustHash("010368116F73952808ED4C4E580C4CF9C0C231B047F616A031A404076DC817CC"), 0},
	{MustHash("011974AD2AF411A6650DB7591DFFB51C645388F78FFC1BBCD73AE2C860602559"), 0},
}

// TODO: Fix per balancer
// func (s *KeyVaSuite) TestBalancers(c *C) {
// 	for _, b := range Balancers {
// 		node := &Node{
// 			Start: MustHash("0000000000000000000000000000000000000000000000000000000000000001"),
// 			End:   MustHash("0300000000000000000000000000000000000000000000000000000000000000"),
// 		}
// 		remainder := b.Balancer.Balance(node, neighbourValues)
// 		insertions := ItemCount - len(remainder)
// 		c.Check(insertions, Equals, len(neighbourValues), Commentf("%s wrong number of insertions", b.Name))
// 		c.Check(node.SanityCheck(), Equals, true, Commentf("%s is not sane", b.Name))
// 	}
// }
