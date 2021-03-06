package keyvadb

import (
	"flag"
	"io"
	"math/rand"
	"testing"

	"github.com/dustin/randbo"
	. "gopkg.in/check.v1"
)

func init() {
	flag.Parse()
}

func Test(t *testing.T) { TestingT(t) }

type KeyVaSuite struct {
	R io.Reader
}

var _ = Suite(&KeyVaSuite{})

func (s *KeyVaSuite) SetUpTest(c *C) {
	s.R = randbo.NewFrom(rand.NewSource(0))
}
