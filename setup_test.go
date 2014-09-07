package keyva

import (
	"io"
	"math/rand"
	"testing"

	"github.com/dustin/randbo"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type KeyVaSuite struct {
	R      io.Reader
	Keys   KeyStore
	Values ValueStore
}

var _ = Suite(&KeyVaSuite{})

func (s *KeyVaSuite) SetUpTest(c *C) {
	s.R = randbo.NewFrom(rand.NewSource(0))
	s.Keys = NewMemoryKeyStore()
	s.Values = NewMemoryValueStore()
}
