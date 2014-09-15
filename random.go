package keyvadb

import (
	crand "crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"io"
	"math/rand"
	"sync/atomic"
)

// PRNG seeded by CRNG
func MustRand() *rand.Rand {
	var seed int64
	if err := binary.Read(crand.Reader, binary.BigEndian, &seed); err != nil {
		panic(err)
	}
	return rand.New(rand.NewSource(seed))
}

type RandomValueGenerator struct {
	r        io.Reader
	min, max uint16
	count    uint64
}

func NewRandomValueGenerator(minValue, maxValue uint16, r io.Reader) *RandomValueGenerator {
	return &RandomValueGenerator{
		r:   r,
		min: minValue,
		max: maxValue,
	}
}

func (r *RandomValueGenerator) Next() (*KeyValue, error) {
	hasher := sha512.New()
	var length uint16
	if err := binary.Read(r.r, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	length = (length % (r.max - r.min)) + r.min
	kv := &KeyValue{
		Key: Key{
			Id: atomic.AddUint64(&r.count, 1),
		},
		Value: make([]byte, length),
	}
	if _, err := r.r.Read(kv.Value[:]); err != nil {
		return nil, err
	}
	hasher.Write(kv.Value)
	copy(kv.Key.Hash[:], hasher.Sum(nil))
	return kv, nil
}

func (r *RandomValueGenerator) Take(n int) (KeyValueSlice, error) {
	kv := make(KeyValueSlice, n)
	for i := 0; i < n; i++ {
		v, err := r.Next()
		if err != nil {
			return nil, err
		}
		kv[i] = *v
	}
	return kv, nil
}
