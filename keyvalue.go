package keyvadb

import (
	"encoding/binary"
	"fmt"
	"io"
)

type KeyValue struct {
	Key
	Value []byte
}

type KeyValueSlice []KeyValue

func (s KeyValueSlice) Keys() KeySlice {
	var k KeySlice
	for _, kv := range s {
		k = append(k, kv.Key)
	}
	return k
}

func NewKeyValue(id ValueId, key Hash, value []byte) *KeyValue {
	return &KeyValue{
		Key: Key{
			Hash: key,
			Id:   id,
		},
		Value: value,
	}
}

func (kv *KeyValue) CloneKey() *Key {
	return &Key{
		Hash: kv.Hash,
		Id:   kv.Id,
	}
}

var lengthSize = binary.Size(uint64(0))

func SizeOfKeyValue(value []byte) uint64 {
	return uint64(lengthSize + SizeOfHash + len(value))
}

func (kv *KeyValue) WriteTo(w io.Writer) (int64, error) {
	length := SizeOfKeyValue(kv.Value)
	b := make([]byte, length)
	binary.BigEndian.PutUint64(b, length)
	pos := 8
	pos += copy(b[pos:], kv.Hash[:])
	pos += copy(b[pos:], kv.Value)
	n, err := w.Write(b)
	return int64(n), err
}

func (kv *KeyValue) ReadFrom(r io.Reader) (int64, error) {
	var length uint64
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return int64(lengthSize), err
	}
	n, err := r.Read(kv.Hash[:])
	if err != nil {
		return int64(lengthSize + n), err
	}
	kv.Value = make([]byte, int(length)-n-lengthSize)
	n, err = r.Read(kv.Value)
	return int64(lengthSize + HashSize + n), err
}

func (kv *KeyValue) String() string {
	return fmt.Sprintf("%s:%X", kv.Key, kv.Value)
}
