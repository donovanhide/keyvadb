package keyva

import (
	"crypto/sha512"
	"encoding/binary"
	"io"
	"sync/atomic"
)

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

func (r *RandomValueGenerator) Next() (*Value, error) {
	hasher := sha512.New()
	var length uint16
	if err := binary.Read(r.r, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	length = (length % (r.max - r.min)) + r.min
	v := &Value{
		Id:    atomic.AddUint64(&r.count, 1),
		Value: make([]byte, length),
	}
	if _, err := r.r.Read(v.Value[:]); err != nil {
		return nil, err
	}
	hasher.Write(v.Value)
	copy(v.Key[:], hasher.Sum(nil))
	return v, nil
}

func (r *RandomValueGenerator) Take(n int) (ValueSlice, error) {
	values := make(ValueSlice, n)
	for i := 0; i < n; i++ {
		v, err := r.Next()
		if err != nil {
			return nil, err
		}
		values[i] = *v
	}
	return values, nil
}
