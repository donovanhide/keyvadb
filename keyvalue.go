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

var lengthSize = binary.Size(uint64(0))

func (kv *KeyValue) MarshalBinary(w io.Writer) error {
	length := uint64(len(kv.Key.Hash) + len(kv.Value) + lengthSize)
	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, kv.Key.Hash); err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, kv.Value)
}

func (kv *KeyValue) UnmarshalBinary(r io.Reader) error {
	var length uint64
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return err
	}
	if err := binary.Read(r, binary.BigEndian, &kv.Hash); err != nil {
		return err
	}
	kv.Value = make([]byte, int(length)-len(kv.Hash)-lengthSize)
	return binary.Read(r, binary.BigEndian, &kv.Value)
}

func (kv *KeyValue) String() string {
	return fmt.Sprintf("%s:%X", kv.Key, kv.Value)
}
